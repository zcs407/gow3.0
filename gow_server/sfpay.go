package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	//amountList []int      = []int{20, 30, 50, 80, 100, 200, 300, 500, 800, 1000, 1500, 2000, 3000, 4000, 5000, 8000, 10000, 15000, 20000, 30000}
)

type QrCodes struct {
	Account  AccountInfo
	CodeList map[int]string
}

func sfPay(input map[string]interface{}) map[string]interface{} {
	fields := []string{"version", "MerchaantNo", "type", "payer", "amount", "sign"}
	if !verifyFields(input, fields) {
		return GetMissingFieldsError()
	}
	payer := input["payer"].(string)
	amount, _ := strconv.ParseFloat(input["amount"].(string), 64)
	amountStr := fmt.Sprintf("%.2f", amount)
	amount, _ = strconv.ParseFloat(amountStr, 64)
	merchanntNo := input["MerchaantNo"].(string)
	midpayType := input["type"].(string)
	version := input["version"].(string)
	userSign := input["sign"].(string)
	customSign := ""
	if v, ok := input["customSign"]; ok {
		customSign = v.(string)
	}

	data := make(map[string]interface{})
	var err error
	//檢查商戶號
	agent, err := dbManager.GetAgentByName(merchanntNo)
	if err != nil || agent == nil {
		data["code"] = 400
		data["msg"] = "Merchaant Can't find."
		log.Println("GetAgent Error:", err)
		return data
	}
	personSetting := GetPersonSetting(agent.Name)

	//檢查充值金額
	if amount < personSetting.Pay.Min || amount > personSetting.Pay.Max {
		data["code"] = 400
		data["msg"] = fmt.Sprintf("Amount need between %.f to %.f.", personSetting.Pay.Min, personSetting.Pay.Max)
		return data
	}
	sign := fmt.Sprintf("%s%s%s%s%s", version, merchanntNo, midpayType, payer, agent.Sign)
	if CONFIGS.Debug {
		log.Println("personSetting:", personSetting)
		log.Println("Sign:", sign)
	}
	serviceSign := getMD5(sign)
	if serviceSign != userSign {
		data["code"] = 400
		data["msg"] = "Sign Verify Error."
		return data
	}
	data["code"] = 200
	data["msg"] = "success"
	var person *Person

	var url string
	accounts, err := dbManager.CheckPersonAccountRequest(personSetting, payer, amount)
	if err != nil {
		data["code"] = 500
		data["msg"] = err.Error()
		return data
	}
	if len(accounts) > 1 {
		data["code"] = 500
		data["msg"] = fmt.Sprintf("Find duplicate payers.")
		return data
	}

	insertTime := time.Now()
	apiMode := "err"
	realUrl := ""
	emptyErr := false
	if len(accounts) == 1 {
		person, err = dbManager.GetPayAccount(accounts[0].Account)
		if err != nil {
			err = fmt.Errorf("Can't find any account use.")
			emptyErr = true
		}
		apiMode = accounts[0].Mode
		realUrl = accounts[0].QrUrl
		log.Println("Find Payer Info. Use Old Info.", person, accounts, apiMode)
	} else {
		useWild := true
		userType := "1"
		switch midpayType {
		case "wxapi":
			userType = "0"
		case "aliapi":
			userType = "1"
		case "qqapi":
			userType = "2"
		}
		running := true
		/*bizAccount := GetBizAccount(merchanntNo)
		if bizAccount != "" {
			payerSign := getMD5(fmt.Sprintf("%s%s%d%.2f%s", bizAccount, payer, insertTime.Unix(), amount, agent.Sign))
			realUrl = getPayAppUrl(bizAccount, userType, payerSign, amount)
			if realUrl != "" {
				log.Printf("[BizAccount]:%s Url:%s\n", bizAccount, realUrl)

				person, err = dbManager.GetPayAccount(bizAccount)
				if person == nil || err != nil {
					log.Printf("[BizAccount]:%s Person is nil. Error:%v\n", bizAccount, err)
				} else {
					running = false
					apiMode = "bizpay"
				}

			}
		}*/
		if running && CONFIGS.Deposit.Count == 0 {
			useWild = true
			log.Printf("[SFPay]:Hit the deposit code limit. use Wild.")
			person, err = dbManager.GetAvailableLockedAccount(agent, payer, userType, amount)
			if err != nil {
				log.Println("GetAvailableUnLockedAccount Failed.", err)
			}
			if person != nil {
				realUrl = person.QrUrl
				dbManager.UpdateMidpayTime(payer, person.Account)
				apiMode = "midpay"
				log.Println("[SFPay][UseWildCode]:", person.Account, realUrl)
			} else {
				log.Println("Can't find any account use.", input)
				emptyErr = true
				err = fmt.Errorf("Can't find any account use.")
			}
			running = false
		}

		if running {
			for {
				person, err = dbManager.GetAvailableSFAccount(agent, payer, userType, amount)
				if err != nil {
					log.Println("GetAvailableSFAccount Failed.", err)
					break
				}
				if person != nil {
					realUrl = person.QrUrl
					apiMode = "sfpay"
					useWild = false
					dbManager.UpdateSFPayTime(payer, person.Account, amount, person.Note)
					log.Println("SFPay Find Account:", person, amount, realUrl, apiMode)
					err = nil
				} else {
					//先檢查沒被鎖定的帳號
					person, err = dbManager.GetAvailableAccount(agent, payer, userType, amount)
					if err != nil {
						log.Println("[SFPay]GetAvailableAccount Failed.", err)
					}
					//找不到的沒被鎖定的帳號,找已經鎖了24小時的帳號
					if person == nil {
						person, err = dbManager.GetAvailableUnLockedAccount(agent, payer, userType, amount)
						if err != nil {
							log.Println("[SFPay]GetAvailableUnLockedAccount Failed.", err)
						}
						log.Println("[SFPay]GetAvailableUnLockedAccount Failed.", err)
					}
					//找到帳號就產生二維碼
					if person != nil {
						accountInfo, err := dbManager.CheckAccountExists(person.Account)
						if err != nil {
							log.Println("[SFPay]Can't find account info.", person.Account)
						}
						apiMode = "sfpay"
						//檢查這個帳號的固定碼有沒有被產生過,如果有的話就換成萬用碼
						noteMiddle := String(2)
						noteSign := fmt.Sprintf("%s-%s-%.2f", accountInfo.IP[2:], noteMiddle, amount)
						exist, _ := dbManager.CheckSFPayNoteExists(person.Account, amount)
						if !exist {
							realUrl = getPayAppUrl(person.Account, person.Type, noteSign, amount)
							if realUrl != "" {
								state, _ := dbManager.InsertSFPay(accountInfo, float64(amount), noteSign, realUrl)
								if state {
									log.Printf("[GenQRCode] Insert QrCode:Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, realUrl)
									dbManager.UpdateSFPayTime(payer, person.Account, amount, noteSign)
									dbManager.UpdateUnLockedAccount(person.Account)
								} else {
									log.Printf("[GenQRCode] Insert QrCode Error: Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, realUrl)
								}
								useWild = false
							}
						} else {
							log.Printf("[SFPay][GenQRCode] Gen QRCode Failed. The code is exists. Account:%s amount:%.2f", person.Account, amount)
						}

						if realUrl == "" {
							dbManager.UpdateLockedAccount(person.Account)
							log.Println("[SFPay]Can't gen qrcode. Lock account:", person.Account)
						}
					}
				}
				if person != nil {
					depositCount, _ := dbManager.GetDepositCount(person.Account)
					log.Println("GetDepositCount Count:", person.Account, depositCount)
					if depositCount >= CONFIGS.Deposit.Count {
						useWild = true
						log.Printf("[SFPay]:Hit the deposit code limit. use Wild. Account:%s Count:%d\n", person.Account, depositCount)
					}
				} else {
					useWild = true
					log.Printf("[SFPay]:Can't find person. use Wild. Payer:%s\n", payer)
				}

				if useWild {
					//使用萬用碼
					person, err = dbManager.GetAvailableLockedAccount(agent, payer, userType, amount)
					if err != nil {
						log.Println("GetAvailableUnLockedAccount Failed.", err)
						break
					}
					if person != nil {
						realUrl = person.QrUrl
						dbManager.UpdateMidpayTime(payer, person.Account)
						apiMode = "midpay"
						log.Println("[SFPay][UseWildCode]:", person.Account, realUrl)
					} else {
						log.Println("Can't find any account use.", input)
						err = fmt.Errorf("Can't find any account use.")
						emptyErr = true
						break
					}
				}

				if realUrl == "" {
					log.Println("Can't find QrCode Url.", input)
					err = fmt.Errorf("Can't find any account use. resource can't find.")
					emptyErr = true
					break
				}
				//中斷loop
				break
			}
		}
	}
	if emptyErr {
		dbManager.InsertMidpayRecord("resource-error", amount, payer, insertTime, customSign, fmt.Sprintf("%d", agent.Id), apiMode, realUrl, input["device"].(string), STATE_DEPOSIT_QRCODE_FAILED)
	}

	if err != nil {
		data["code"] = 501
		data["msg"] = err.Error()
		data["data"] = nil
		return data
	}
	if apiMode == "err" {
		data["code"] = 501
		data["msg"] = "System busy. Please try agian."
		data["data"] = nil
		return data
	}
	if amount == 0 {
		data["code"] = 500
		data["msg"] = "Amount cant't be zero."
		data["data"] = nil
		return data
	}
	//代表這個payer值沒被用過,可以用
	if len(accounts) == 0 {
		accounts, err := dbManager.CheckPersonAccountRequest(personSetting, payer, amount)
		if len(accounts) == 0 {
			status, err := dbManager.InsertMidpayRecord(person.Account, amount, payer, insertTime, customSign, fmt.Sprintf("%d", agent.Id), apiMode, realUrl, input["device"].(string), STATE_DEPOSIT_CHECKING)
			if !status || err != nil {
				log.Println("Insert to midpay record failed.", err)
			}
			payerSign := getMD5(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
			status, err = dbManager.InsertPayerRecord(person.Account, amount, payer, insertTime, payerSign, customSign)
			if !status || err != nil {
				log.Println("Insert to payer record failed.", err)
			}
		} else if len(accounts) == 1 {
			log.Printf("發現重複申請，使用舊資訊")
			person, err = dbManager.GetPayAccount(accounts[0].Account)
			if err != nil {
				err = fmt.Errorf("Can't find any account use.")
			}
			apiMode = accounts[0].Mode
			insertTime = accounts[0].RequestTime
		}
	} else if len(accounts) == 1 {
		insertTime = accounts[0].RequestTime
	}
	//Set the output data in data object
	var midPayObject MidPayObject
	midPayObject.Mode = apiMode
	midPayObject.Account = person.Account
	midPayObject.Amount = amount
	midPayObject.NickName = person.NickName
	midPayObject.OverTime = personSetting.ClientLockTime
	midPayObject.Payer = payer
	if realUrl != "" {
		midPayObject.QRUrl = realUrl
	} else {
		midPayObject.QRUrl = person.QrUrl
	}
	midPayObject.Timestamp = insertTime.Unix()
	midPayObject.RealName = person.RealName
	payerSign := getMD5(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
	url = fmt.Sprintf("%s/sfqrcode?account=%s&ts=%d&payer=%s&platfrom=%d&amount=%.2f&sign=%s&mode=%s", CONFIGS.Person.OutputImageServer, person.Account, insertTime.Unix(), payer, agent.Id, amount, payerSign, apiMode)
	midPayObject.Url = url
	midPayObject.RefreshTime = personSetting.AccountLockTime
	data["data"] = midPayObject

	return data
}

func IsAppOnline(account string) bool {
	for k, v := range lum.Users {
		if sfAccount, ok := v.Accounts["alipay"]; ok {
			log.Printf("IsAppOnline[%v]:%v", k, v.Accounts)
			if account == sfAccount {
				return true
			}
		}
	}
	return false
}

func GenQRCodeFromApp(item QrCodes) (string, int) {
	urlCount := 0
	lastQrCode := ""
	noteMiddle := String(2)
	for amount, _ := range item.CodeList {
		item.Account.IP = strings.Replace(item.Account.IP, " ", "", -1)
		item.Account.IP = strings.Replace(item.Account.IP, "\t", "", -1)
		item.Account.IP = strings.Replace(item.Account.IP, "\n", "", -1)
		noteSign := fmt.Sprintf("%s-%s-%d", item.Account.IP[2:], noteMiddle, amount)
		item.CodeList[amount] = getPayAppUrl(item.Account.Account, item.Account.Type, noteSign, float64(amount))
		if item.CodeList[amount] != "" {
			if lastQrCode != item.CodeList[amount] {
				lastQrCode = item.CodeList[amount]
			} else {
				log.Printf("[GenQRCode] The QRCode had duplicate. Account:%s amount:%d qrcode:%s\n", item.Account.Account, amount, item.CodeList[amount])
				break
			}
			urlCount++
			exist, _ := dbManager.CheckSFPayNoteExists(item.Account.Account, float64(amount))
			if exist {
				err := dbManager.UpdateSFPay(item.Account.Account, item.Account.Platfrom, float64(amount), noteSign, item.CodeList[amount])
				if err == nil {
					log.Printf("[GenQRCode] Update QrCode: Account:%s amount:%d qrcode:%s\n", item.Account.Account, amount, item.CodeList[amount])
				} else {
					log.Printf("[GenQRCode] Update QrCode Error: Account:%s amount:%d qrcode:%s\n", item.Account.Account, amount, item.CodeList[amount])
				}
			} else {
				state, _ := dbManager.InsertSFPay(item.Account, float64(amount), noteSign, item.CodeList[amount])
				if state {
					log.Printf("[GenQRCode] Insert QrCode:Account:%s amount:%d qrcode:%s\n", item.Account.Account, amount, item.CodeList[amount])
				} else {
					log.Printf("[GenQRCode] Insert QrCode Error: Account:%s amount:%d qrcode:%s\n", item.Account.Account, amount, item.CodeList[amount])
				}
			}
		} else {
			log.Printf("[GenQRCode] Error:Account:%s amount:%d can't gen qrcode.\n", item.Account.Account, amount)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return item.Account.Account, urlCount
}

func GenQRCode() map[string]interface{} {
	log.Println("GenQRCode Starting")
	log.Println("GenQRCode Amount List:", CONFIGS.QrCode.AmountList)
	data := make(map[string]interface{})
	GenQrCodeList := make(map[string]QrCodes)
	accountList, err := dbManager.GetAllAccountsStatus()
	if err != nil {
		log.Println("GetAllAccountsStatus Error:", err)
		data["code"] = 400
		data["msg"] = "GetAllAccountsStatus Error"
		return data
	}
	for _, account := range accountList {
		if account.Platfrom == CONFIGS.QrCode.GenPlatform {
			if IsAppOnline(account.Account) {
				var qrcodes QrCodes
				qrcodes.Account = account
				qrcodes.CodeList = make(map[int]string)
				for _, amount := range CONFIGS.QrCode.AmountList {
					qrcodes.CodeList[amount] = ""
				}
				GenQrCodeList[account.Account] = qrcodes
			}
		}
	}
	qrCode := make(map[string]int)
	for _, item := range GenQrCodeList {
		account, count := GenQRCodeFromApp(item)
		log.Printf("GenQRCodeFromApp:%s Url Count:%d\n", account, count)
		qrCode[account] = count
	}
	log.Println("GenQRCode Done.")
	data["code"] = 200
	data["msg"] = "success"
	data["info"] = qrCode
	return data
}

func GenSingleQrCode(account string) {
	if IsAppOnline(account) {
		var qrcodes QrCodes
		accountInfo, err := dbManager.CheckAccountExists(account)
		if err != nil {
			log.Println("GenSingleQrCode Error:", err, account)
			return
		}
		qrcodes.Account = accountInfo
		qrcodes.CodeList = make(map[int]string)
		for _, amount := range CONFIGS.QrCode.AmountList {
			qrcodes.CodeList[amount] = ""
		}
		GenQRCodeFromApp(qrcodes)
	} else {
		log.Printf("GenSingleQrCode:[%s] App is off line.\n", account)
	}
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func IsBizAgent(agent string) bool {
	for _, v := range CONFIGS.Biz.Agents {
		if v == agent {
			return true
		}
	}
	return false
}

func GetBizAccount(agent string) string {
	account := ""
	if IsBizAgent(agent) {
		if len(CONFIGS.Biz.Acccounts) > 0 {
			account = CONFIGS.Biz.Acccounts[seededRand.Intn(len(CONFIGS.Biz.Acccounts))]
		}
	}
	return account
}
