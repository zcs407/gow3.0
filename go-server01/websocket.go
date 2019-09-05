package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"log"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	Device     string
	Accounts   map[string]string
	PayAppList []*PayApp
	Mutex      *sync.Mutex
	Enable     bool
	Conn       *websocket.Conn
}

type LoginUserManager struct {
	Mutex *sync.Mutex
	Users map[string]*WSClient
}

var lum LoginUserManager

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetAppPay(payApp *PayApp) error {
	send := false
	data := make(map[string]interface{})
	data["type"] = payApp.AccountType
	data["payer"] = payApp.Payer
	data["amount"] = payApp.Amount
	data["info"] = payApp
	log.Println("Sent Command Orign:", data, payApp.Account)
	for _, client := range lum.Users {
		if client.Conn != nil && client.Enable {
			log.Println("client:", client.Device, client.Accounts)
			if account, ok := client.Accounts[payApp.AccountType]; ok {
				if account == payApp.Account {
					client.Mutex.Lock()
					client.PayAppList = append(client.PayAppList, payApp)
					client.Mutex.Unlock()
					log.Println("Sent Command:", data)
					go client.SendCmd("getpay", data)
					send = true
					break
				}
			}
		}
	}
	if !send {
		log.Println("GetAppPay Error: App didn't online.")
		return fmt.Errorf("GetAppPay Error: App didn't online.")
	}
	return nil
}

func (wsc *WSClient) Process() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WebSocket][%s][ProcessHandler] Recovered in send to user data:%v\n", wsc.Device, r)
			wsc.Close()
		}
	}()
	for wsc.Enable {
		_, message, err := wsc.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			wsc.Enable = false
			break
		}
		log.Printf("recv message: %s\n", message)
		var input map[string]interface{}
		err = json.Unmarshal([]byte(message), &input)
		if err != nil {
			log.Println("json unmarshal error:", err)
			wsc.Enable = false
			break
		}
		log.Printf("input message: %v\n", input)
		if _, ok := input["cmd"]; ok {
			cmd := input["cmd"].(string)
			log.Println("RecvCmd:", cmd)

			switch cmd {
			case "login":
				if _, ok := input["device"]; ok {
					wsc.Device = input["device"].(string)
				}
				if len(wsc.Device) > 0 {
					lum.Mutex.Lock()
					if v, ok := lum.Users[wsc.Device]; ok {
						if v.Enable {
							KickOff(v, "", true)
						}
						delete(lum.Users, wsc.Device)
						log.Println("Remove from LUM", wsc.Device)
					}
					log.Println("Set to LUM", wsc.Device)
					lum.Users[wsc.Device] = wsc
					lum.Mutex.Unlock()

					if _, ok := input["alipay"]; ok {
						wsc.Accounts["alipay"] = input["alipay"].(string)
					}
					if _, ok := input["wechat"]; ok {
						wsc.Accounts["wechat"] = input["wechat"].(string)
					}
					if _, ok := input["qq"]; ok {
						wsc.Accounts["qq"] = input["qq"].(string)
					}
					log.Printf("Device login.%s Accounts:%v\n", wsc.Device, wsc.Accounts)
					cookie, _ := dbManager.GetAccountCookie(wsc.Accounts["alipay"])
					setCookie := ""
					if cookie != "" && cookie != "www.baidu.com" {
						setCookie = cookie
					}
					wsc.Send(cmd, 200, setCookie)
				} else {
					log.Println("User account is empty.")
					wsc.Enable = false
					break
				}
			case "deposit":
				data := SFDeposit(input)
				code := data["code"].(int)
				msg := data["msg"].(string)
				if code == 200 {
					msg = "WS API入帳完成"
				}
				wsc.Send("receive", code, msg)
			default:
				findKey := -1
				for k, v := range wsc.PayAppList {
					if v.Command == cmd {
						switch cmd {
						case "getpay":
							if int(input["code"].(float64)) == 200 {
								if v.Payer == input["mark"].(string) && v.AccountType == input["type"].(string) {
									var result PayAppResult
									if _, ok := input["account"]; ok {
										result.Account = input["account"].(string)
									}
									if _, ok := input["code"]; ok {
										result.Code = int(input["code"].(float64))
									}
									if _, ok := input["msg"]; ok {
										result.Message = input["msg"].(string)
									}
									if _, ok := input["payurl"]; ok {
										result.Url = input["payurl"].(string)
									}

									result.Data = input
									v.SetPay(result)
									findKey = k
									break
								}
							}
						}
					}
				}
				if findKey != -1 {
					wsc.Mutex.Lock()
					if findKey == 0 {
						if len(wsc.PayAppList) > 1 {
							wsc.PayAppList = wsc.PayAppList[1:]
						} else {
							wsc.PayAppList = wsc.PayAppList[:0]
						}
					} else if findKey >= 0 {
						if findKey+1 > len(wsc.PayAppList) {
							wsc.PayAppList = wsc.PayAppList[:findKey]
						} else {
							if len(wsc.PayAppList) > 0 {
								if findKey < len(wsc.PayAppList) {
									//log.Printf("socket[%s] removeIdx:%d split:%d\n", name, idx, len(socketPool[name]))
									tmp := wsc.PayAppList[:findKey]
									wsc.PayAppList = append(tmp, wsc.PayAppList[findKey+1:]...)
								} else {
									wsc.PayAppList = wsc.PayAppList[:findKey]
								}
							}
						}
					}
					wsc.Mutex.Unlock()
				}
			}
		}
	}
	if !wsc.Enable {
		wsc.Close()
	}
}

func SFDeposit(record map[string]interface{}) map[string]interface{} {
	var depositRecord DepositRecord
	data := make(map[string]interface{})
	if record["amount"] == nil {
		log.Println("Amount not exists.")
		data["code"] = 400
		data["msg"] = "Amount not exists."
		return data
	}
	if record["account"] == nil {
		log.Println("account not exists.")
		data["code"] = 400
		data["msg"] = "account not exists."
		return data
	}
	if record["deviceAccount"] == nil {
		log.Println("deviceAccount not exists.")
		data["code"] = 400
		data["msg"] = "deviceAccount not exists."
		return data
	}
	if record["depositNumber"] == nil {
		log.Println("DepositNumber not exists.")
		data["code"] = 400
		data["msg"] = "DepositNumber not exists."
		return data
	}
	if record["note"] == nil {
		log.Println("Note not exists.")
		data["code"] = 400
		data["msg"] = "Note not exists."
		return data
	}
	if record["transferTime"] == nil {
		log.Println("transferTime not exists.")
		data["code"] = 400
		data["msg"] = "transferTime not exists."
		return data
	}
	if record["device"] == nil {
		log.Println("device not exists.")
		data["code"] = 400
		data["msg"] = "device not exists."
		return data
	}
	accountInfo, err := dbManager.CheckAccountExists(record["deviceAccount"].(string))
	if err != nil || accountInfo.Account == "" {
		log.Println("Device Account not exists.")
		data["code"] = 400
		data["msg"] = "Device account not exists in our system."
		return data
	}
	payerRecord, err := dbManager.GetPayerInfo(record["deviceAccount"].(string), record["amount"].(string), record["note"].(string))
	if err != nil {
		log.Println("SFApp Deposit PayerRecord Error:", err)
	}
	if payerRecord.Payer != "" {
		record["sign"] = record["note"]
		record["note"] = payerRecord.Payer
	} else {
		record["sign"] = ""
		log.Println("SFApp Deposit PayerRecord Not Find.")
	}
	amount, _ := strconv.ParseFloat(record["amount"].(string), 64)
	depositRecord.Amount = amount
	depositRecord.CreateUser = "auto"
	depositRecord.DepositNumber = record["depositNumber"].(string)
	depositRecord.Note = record["note"].(string)
	if record["name"].(string) != "" {
		depositRecord.PayAccount = record["name"].(string)
	} else {
		depositRecord.PayAccount = record["device"].(string)
	}
	depositRecord.Platfrom = accountInfo.Platfrom
	// Data init
	depositRecord.CreateTime = time.Now()
	depositRecord.TranTime = depositRecord.CreateTime
	depositRecord.ExcuteTime = depositRecord.CreateTime
	if record["transferTime"] == nil {
		record["transferTime"] = depositRecord.CreateTime.Format("2006-01-02 15:04:05")
	} else {
		i, err := strconv.ParseInt(record["transferTime"].(string), 10, 64)
		if err != nil {
			log.Println("transferTime Error:", err)
		}
		i = int64(i / 1000)
		tm := time.Unix(i, 0)
		record["transferTime"] = tm.Format("2006-01-02 15:04:05")
	}
	depositRecord.TransferTime = record["transferTime"].(string)
	depositRecord.State = STATE_PENDING
	depositRecord.Times = 0
	depositRecord.BillNo = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	depositRecord.WechatName = record["deviceAccount"].(string)
	log.Println("Origin DepositRecord:", depositRecord)
	data = saveDeposit(&accountInfo, &depositRecord, record)
	log.Println("saveDeposit Result:", data)
	return data
}

func (wsc *WSClient) SendCmd(cmd string, data interface{}) {
	if wsc.Conn != nil && wsc.Enable {
		wsc.Conn.SetWriteDeadline(time.Now().Add(time.Duration(10) * time.Second))
		message := make(map[string]interface{})
		message["cmd"] = cmd
		message["data"] = data
		err := websocket.WriteJSON(wsc.Conn, message)
		if err != nil {
			log.Println("write error:", err, wsc.Device)
			KickOff(wsc, fmt.Sprintf("Send Error:%s", err.Error()), false)
		}
	}
}

func (wsc *WSClient) Send(cmd string, code int, msg interface{}) {
	if wsc.Conn != nil && wsc.Enable {
		wsc.Conn.SetWriteDeadline(time.Now().Add(time.Duration(10) * time.Second))
		message := make(map[string]interface{})
		message["cmd"] = cmd
		message["code"] = code
		message["message"] = msg
		err := websocket.WriteJSON(wsc.Conn, message)
		if err != nil {
			log.Println("write error:", err, wsc.Device)
			KickOff(wsc, fmt.Sprintf("Send Error:%s", err.Error()), false)
		}
	}
}

func (wsc *WSClient) Close() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WebSocket][%s][Close] Recovered in close data:%v\n", wsc.Device, r)
		}
	}()
	if wsc.Conn != nil {
		wsc.Enable = false
		wsc.Conn.Close()
		wsc.Conn = nil
	}
	log.Printf("[WebSocket][%s] Connection force closed.\n", wsc.Device)
}

func Boradcast(cmd string, code int, data string) {
	if len(data) > 0 {
		lum.Mutex.Lock()
		for _, client := range lum.Users {
			if client.Conn != nil && client.Enable {
				go client.Send(cmd, code, data)
			}
		}
		lum.Mutex.Unlock()
	}
}

func KickOff(client *WSClient, kickMsg string, manual bool) {
	errmsg := fmt.Sprintf("[Websocket][Kick]:[%s] duplicate login.", client.Device)
	if kickMsg != "" {
		errmsg = kickMsg
	}
	if manual {
		client.SendCmd("kick", errmsg)
	}

	if client.Conn != nil {
		client.Close()
	}
	//remove old websocket client
	client.Enable = false
	log.Printf("[Websocket][Kick][%s]:%s", client.Device, errmsg)
	client = nil
}

//Main websocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	var wsClient WSClient
	wsClient.Conn = c
	wsClient.Enable = true
	wsClient.Mutex = &sync.Mutex{}
	wsClient.Accounts = make(map[string]string)
	wsClient.Process()
}

func WebSocketInit() {
	lum.Mutex = &sync.Mutex{}
	lum.Users = make(map[string]*WSClient)
}
