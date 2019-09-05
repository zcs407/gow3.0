package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	log.Println("HomeHandler:", r.URL)
	fmt.Fprint(w, "403 Forbidden")
}

func GowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	cmd := vars["cmd"]
	data := make(map[string]interface{})
	switch cmd {
	case "callback":
		input, err := getInputJSON(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data = input
	case "sfnotify":
		input, err := getInputForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("sfnotify:", input)
		SFDeposit(input)
		fmt.Fprint(w, "success")
		return
	case "sfresult":
		input, err := getInputForm(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("sfresult:", input)
		fmt.Fprint(w, "success")
		return
	case "midpay":
		var input map[string]interface{}
		var err error
		if r.Method == "POST" {
			input, err = getInputForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if r.Method == "GET" {
			input = getInputQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			input["mobile"] = "1"
		}
		input["device"] = GetDevice(r.UserAgent())
		data = sfPay(input)
		if v, ok := input["mobile"]; ok {
			if data["code"].(int) == 200 {
				log.Println("User Info:", r.UserAgent(), "UserData:", input, "GetAccount", data["data"])
				if v.(string) == "1" {
					midpayObj := data["data"].(MidPayObject)
					merchanntNo := input["MerchaantNo"].(string)
					//檢查商戶號
					agent, err := dbManager.GetAgentByName(merchanntNo)
					if err != nil || agent == nil {
						log.Println("Redirect GetAgent Error:", err)
					}
					accountInfo, err := dbManager.CheckAccountExists(midpayObj.Account)
					if err != nil {
						log.Println("Redirect CheckAccountExists Error:", err)
					}
					if agent != nil && accountInfo.Type != "" {
						realUrl := ""
						log.Println("QrCorde Verify Code:", midpayObj)
						if midpayObj.QRUrl != "" {
							realUrl = fmt.Sprintf("%s?t=%d", midpayObj.QRUrl, (time.Now().UnixNano() / 1000000))
							log.Println("QrCorde Verify Code Url:", midpayObj.Account, realUrl)
						}
						if realUrl != "" {
							dbManager.UpdateMidpayStatusRecord(midpayObj.Account, fmt.Sprintf("%d", midpayObj.Timestamp), midpayObj.Payer, fmt.Sprintf("%.2f", midpayObj.Amount), STATE_DEPOSIT_VERIFY)
							http.Redirect(w, r, realUrl, http.StatusSeeOther)
							return
						} else {
							dbManager.UpdateMidpayStatusRecord(midpayObj.Account, fmt.Sprintf("%d", midpayObj.Timestamp), midpayObj.Payer, fmt.Sprintf("%.2f", midpayObj.Amount), STATE_DEPOSIT_QRCODE_FAILED)
							tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
							data["code"] = ERROR_ACCOUNT_NOT_EXISTS
							data["msg"] = "抱歉，目前系统繁忙中，请稍后再试\n Sorry, the system is busy now, please try again later."
							tmpl.Execute(w, data)
							return
						}
					}
				}
			} else {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_ACCOUNT_NOT_EXISTS
				data["msg"] = "抱歉，目前系统繁忙中，请稍后再试\n Sorry, the system is busy now, please try again later."
				tmpl.Execute(w, data)
				return
			}
		}
		/*var input map[string]interface{}
		var err error
		if r.Method == "POST" {
			input, err = getInputForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if r.Method == "GET" {
			input = getInputQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			input["mobile"] = "1"
		}
		data = midPay(input)
		if data["code"].(int) == 200 {
			log.Println("User Info:", r.UserAgent(), "UserData:", input, "GetAccount", data["data"])
			if v, ok := input["mobile"]; ok {
				if v.(string) == "1" {
					midpayObj := data["data"].(MidPayObject)
					merchanntNo := input["MerchaantNo"].(string)
					//檢查商戶號
					agent, err := dbManager.GetAgentByName(merchanntNo)
					if err != nil || agent == nil {
						log.Println("Redirect GetAgent Error:", err)
					}
					accountInfo, err := dbManager.CheckAccountExists(midpayObj.Account)
					if err != nil {
						log.Println("Redirect CheckAccountExists Error:", err)
					}
					if agent != nil && accountInfo.Type != "" {
						realUrl := ""
						userSign := getMD5(fmt.Sprintf("%s%s%d%.2f%s", midpayObj.Account, midpayObj.Payer, midpayObj.Timestamp, midpayObj.Amount, agent.Sign))
						midPayAccount, _ := dbManager.GetMidpayAccount(midpayObj.Account)
						if !CONFIGS.HB {
							if midPayAccount != nil && midPayAccount.Lock == "0" {
								realUrl = getPayAppUrl(midpayObj.Account, accountInfo.Type, userSign, midpayObj.Amount)
							} else {
								log.Println("QrCorde Verify Code:", midpayObj)
								if midpayObj.QRUrl != "" {
									realUrl = fmt.Sprintf("%s?t=%d", midpayObj.QRUrl, (time.Now().UnixNano() / 1000000))
									log.Println("QrCorde Verify Code Url:", midpayObj.Account, realUrl)
								}
							}
						} else {
							log.Println("QrCorde Verify Code:", midpayObj)
							if midpayObj.QRUrl != "" {
								realUrl = fmt.Sprintf("%s?t=%d", midpayObj.QRUrl, (time.Now().UnixNano() / 1000000))
								log.Println("QrCorde Verify Code Url:", midpayObj.Account, realUrl)
							}
						}
						if realUrl != "" {
							http.Redirect(w, r, realUrl, http.StatusSeeOther)
							return
						} else {
							data["code"] = 401
							data["msg"] = "Can't find account used. The qrcode url is empty. Please retry again."
						}
					}
				}
			}
		}*/
	case "Depositlist":
		input, err := getInputJSON(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data = depositList(input)
	case "deposit":
		if r.Method == "POST" {
			input, err := getInputForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data = manualSaveDeposit(input)
		} else {
			sendToOldAPI(w, r)
			return
		}
	case "sysdeposit":
		if r.Method == "GET" {
			query := r.URL.Query()
			data["code"] = 400
			data["msg"] = "Sync deposit failed"
			data["data"] = nil
			if query["depositNumber"][0] == "" || query["platfrom"][0] == "" {
				data["code"] = 500
				data["msg"] = "Missing some fields"
			} else {
				data = SysDeposit(query["depositNumber"][0], query["platfrom"][0])
			}
		} else {
			sendToOldAPI(w, r)
			return
		}
	case "midpaypage":
		if r.Method == "GET" {
			input := getInputQuery(r)
			log.Println("midpaypage input:", input)
			fields := []string{"version", "MerchaantNo", "type", "payer", "amount", "sign"}
			if !verifyFields(input, fields) {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_MISSING_FIELDS
				data["msg"] = "Missing some fields.Please check agian."
				tmpl.Execute(w, data)
				return
			}
			tmpl := template.Must(template.ParseFiles("tpl/midpay.html"))
			input["device"] = GetDevice(r.UserAgent())
			data = sfPay(input)
			if data["data"] == nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				tmpl.Execute(w, data)
				return
			}
			tmpl.Execute(w, data["data"].(MidPayObject))
			return
		} else {
			data["code"] = 500
			data["msg"] = "not support this method."
		}
	case "depositstatus":
		input, err := getInputJSON(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data = getDepositStatus(input)
	/*case "qrcode":
	if r.Method == "GET" {
		input := getInputQuery(r)
		log.Println("qrcode input:", input)
		fields := []string{"account", "payer", "ts", "platfrom", "amount", "sign"}
		if !verifyFields(input, fields) {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_MISSING_FIELDS
			data["msg"] = "Missing some fields.Please check agian."
			tmpl.Execute(w, data)
			return
		}
		verifyUrl := fmt.Sprintf("%s/VerifyCode?account=%s&ts=%s&payer=%s&platfrom=%s&amount=%s&sign=%s", CONFIGS.Person.OutputImageServer, input["account"].(string), input["ts"].(string), input["payer"].(string), input["platfrom"].(string), input["amount"].(string), input["sign"].(string))
		image, err := createQrCode(verifyUrl)
		if err != nil {
			log.Println("qrcode gengerate error:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_QRCODE)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(image)))
		if _, err := w.Write(image); err != nil {
			log.Println("unable to write image.", input["account"])
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}*/

	/*case "VerifyCode":
	if r.Method == "GET" {
		input := getInputQuery(r)
		log.Println("VerifyCode input:", input)
		fields := []string{"account", "platfrom", "payer", "ts", "amount", "sign"}
		if !verifyFields(input, fields) {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_MISSING_FIELDS
			data["msg"] = "Missing some fields.Please check agian."
			tmpl.Execute(w, data)
			return
		}
		ts, err := strconv.ParseInt(input["ts"].(string), 10, 64)
		if err != nil {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_TIMESTAMP_INCORRECT
			data["msg"] = "timestamp incorrect"
			tmpl.Execute(w, data)
			return
		}
		//檢查商戶號
		agentId, err := strconv.Atoi(input["platfrom"].(string))
		if err != nil {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_PLATFROM_INCORRECT
			data["msg"] = "platfrom incorrect."
			tmpl.Execute(w, data)
			return
		}
		agent, err := dbManager.GetAgentById(agentId)
		if err != nil || agent == nil {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_PLATFROM_INCORRECT
			data["msg"] = "platfrom Can't find."
			tmpl.Execute(w, data)
			return
		}
		userSign := getMD5(fmt.Sprintf("%s%s%s%s%s", input["account"].(string), input["payer"].(string), input["ts"].(string), input["amount"].(string), agent.Sign))
		if userSign != input["sign"].(string) {
			log.Println("sign verify error:", input, userSign)
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_SIGN_INCORRECT
			data["msg"] = "sign verify error"
			tmpl.Execute(w, data)
			return
		}
		personSetting := GetPersonSetting(agent.Name)
		log.Println("Now:", time.Now().Unix(), "QRCODE TIME:", (ts + int64(personSetting.AccountLockTime)))
		//if !personSetting.SFPay {
		if time.Now().Unix() > ts+int64(personSetting.AccountLockTime) {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_TIMEOUT
			data["msg"] = "Operation timeout"
			tmpl.Execute(w, data)
			return
		}
		amount, _ := strconv.ParseFloat(input["amount"].(string), 64)
		accountInfo, err := dbManager.CheckAccountExists(input["account"].(string))
		if accountInfo.Account == "" || err != nil {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_ACCOUNT_NOT_EXISTS
			data["msg"] = "Account not exists."
			tmpl.Execute(w, data)
			return
		}
		_, err = dbManager.GetPayerInfo(input["account"].(string), input["amount"].(string), userSign)
		if err != nil {
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_PAYER_NOT_EXISTS
			data["msg"] = "payer information not exists."
			tmpl.Execute(w, data)
			return
		}
		realUrl := ""
		midPayAccount, err := dbManager.GetMidpayAccount(input["account"].(string))
		if err != nil {
			log.Println("[VerifyCode] QrCorde GetMidpayAccount Error:", input["account"], err)
		}
		if midPayAccount != nil {
			if !CONFIGS.HB {
				if midPayAccount.Lock == "0" {
					realUrl = getPayAppUrl(input["account"].(string), accountInfo.Type, userSign, amount)
				} else {
					person, err := dbManager.GetPayAccount(input["account"].(string))
					if err != nil {
						log.Println("[VerifyCode] QrCorde Verify Code Error:", input["account"], err)
					}
					log.Println("[VerifyCode] QrCorde Verify Code:", input["account"], person)
					if person != nil && person.QrUrl != "" {
						realUrl = fmt.Sprintf("%s?t=%d", person.QrUrl, (time.Now().UnixNano() / 1000000))
						log.Println("QrCorde Verify Code Url:", person.Account, realUrl)
					}
				}
			} else {
				person, err := dbManager.GetPayAccount(input["account"].(string))
				if err != nil {
					log.Println("[VerifyCode] QrCorde Verify Code Error:", input["account"], err)
				}
				log.Println("[VerifyCode] QrCorde Verify Code:", input["account"], person)
				if person != nil && person.QrUrl != "" {
					realUrl = fmt.Sprintf("%s?t=%d", person.QrUrl, (time.Now().UnixNano() / 1000000))
					log.Println("QrCorde Verify Code Url:", person.Account, realUrl)
				} else {
					log.Printf("[VerifyCode] QrCorde Verify Failed. account:%v person:%v\n ", input["account"], person)
				}
			}
		} else {
			log.Printf("[VerifyCode] Can't find midpay account. account:%v\n", input["account"])
		}
		if realUrl != "" {
			dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_VERIFY)
			http.Redirect(w, r, realUrl, http.StatusSeeOther)
			return
		} else {
			dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_QRCODE_FAILED)
			log.Printf("[VerifyCode] qrcode error: %v can't find any qrcode url.\n", input["account"])
			tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
			data["code"] = ERROR_IMAGE_INCORRECT
			data["msg"] = fmt.Sprintf("qrcode error: %v can't find any qrcode url.\n", input["account"])
			tmpl.Execute(w, data)
			return
		}
	}*/
	case "genamount":
		if r.Method == "GET" {
			input := getInputQuery(r)
			data = genAccountAmount(input)
		} else {
			data["code"] = 401
			data["msg"] = "METHOD accept GET only."
		}
	case "sfpay":
		var input map[string]interface{}
		var err error
		if r.Method == "POST" {
			input, err = getInputForm(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else if r.Method == "GET" {
			input = getInputQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			input["mobile"] = "1"
		}
		input["device"] = GetDevice(r.UserAgent())
		data = sfPay(input)
		if v, ok := input["mobile"]; ok {
			if data["code"].(int) == 200 {
				log.Println("User Info:", r.UserAgent(), "UserData:", input, "GetAccount", data["data"])
				if v.(string) == "1" {
					midpayObj := data["data"].(MidPayObject)
					merchanntNo := input["MerchaantNo"].(string)
					//檢查商戶號
					agent, err := dbManager.GetAgentByName(merchanntNo)
					if err != nil || agent == nil {
						log.Println("Redirect GetAgent Error:", err)
					}
					accountInfo, err := dbManager.CheckAccountExists(midpayObj.Account)
					if err != nil {
						log.Println("Redirect CheckAccountExists Error:", err)
					}
					if agent != nil && accountInfo.Type != "" {
						realUrl := ""
						log.Println("QrCorde Verify Code:", midpayObj)
						if midpayObj.QRUrl != "" {
							realUrl = fmt.Sprintf("%s?t=%d", midpayObj.QRUrl, (time.Now().UnixNano() / 1000000))
							log.Println("QrCorde Verify Code Url:", midpayObj.Account, realUrl)
						}
						if realUrl != "" {
							dbManager.UpdateMidpayStatusRecord(midpayObj.Account, fmt.Sprintf("%d", midpayObj.Timestamp), midpayObj.Payer, fmt.Sprintf("%.2f", midpayObj.Amount), STATE_DEPOSIT_VERIFY)
							http.Redirect(w, r, realUrl, http.StatusSeeOther)
							return
						} else {
							dbManager.UpdateMidpayStatusRecord(midpayObj.Account, fmt.Sprintf("%d", midpayObj.Timestamp), midpayObj.Payer, fmt.Sprintf("%.2f", midpayObj.Amount), STATE_DEPOSIT_QRCODE_FAILED)
							tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
							data["code"] = ERROR_ACCOUNT_NOT_EXISTS
							data["msg"] = "抱歉，目前系统繁忙中，请稍后再试\n Sorry, the system is busy now, please try again later."
							tmpl.Execute(w, data)
							return
						}
					}
				}
			} else {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_ACCOUNT_NOT_EXISTS
				data["msg"] = "抱歉，目前系统繁忙中，请稍后再试\n Sorry, the system is busy now, please try again later."
				tmpl.Execute(w, data)
				return
			}
		}
	case "sfqrcode":
		if r.Method == "GET" {
			input := getInputQuery(r)
			log.Println("sf qrcode input:", input)
			fields := []string{"account", "payer", "ts", "platfrom", "amount", "sign", "mode"}
			if !verifyFields(input, fields) {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_MISSING_FIELDS
				data["msg"] = "Missing some fields.Please check agian."
				tmpl.Execute(w, data)
				return
			}
			verifyUrl := fmt.Sprintf("%s/verifysfcode?account=%s&ts=%s&payer=%s&platfrom=%s&amount=%s&sign=%s&mode=%s", CONFIGS.Person.OutputImageServer, input["account"].(string), input["ts"].(string), input["payer"].(string), input["platfrom"].(string), input["amount"].(string), input["sign"].(string), input["mode"].(string))
			dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_QRCODE)
			image, err := createQrCode(verifyUrl)
			if err != nil {
				log.Println("qrcode gengerate error:", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Content-Length", strconv.Itoa(len(image)))
			if _, err := w.Write(image); err != nil {
				log.Println("unable to write image.", input["account"])
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
	case "verifysfcode":
		if r.Method == "GET" {
			input := getInputQuery(r)
			log.Println("VerifyCode input:", input)
			fields := []string{"account", "platfrom", "payer", "ts", "amount", "sign", "mode"}
			if !verifyFields(input, fields) {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_MISSING_FIELDS
				data["msg"] = "栏位缺省.\nMissing some fields.Please check agian."
				tmpl.Execute(w, data)
				return
			}
			ts, err := strconv.ParseInt(input["ts"].(string), 10, 64)
			if err != nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_TIMESTAMP_INCORRECT
				data["msg"] = "时间格式错误.\ntimestamp incorrect"
				tmpl.Execute(w, data)
				return
			}
			//檢查商戶號
			agentId, err := strconv.Atoi(input["platfrom"].(string))
			if err != nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_PLATFROM_INCORRECT
				data["msg"] = "支付平台错误.\nplatfrom incorrect."
				tmpl.Execute(w, data)
				return
			}
			agent, err := dbManager.GetAgentById(agentId)
			if err != nil || agent == nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_PLATFROM_INCORRECT
				data["msg"] = "商戶平台错误.\nplatfrom Can't find."
				tmpl.Execute(w, data)
				return
			}
			userSign := getMD5(fmt.Sprintf("%s%s%s%s%s", input["account"].(string), input["payer"].(string), input["ts"].(string), input["amount"].(string), agent.Sign))
			if userSign != input["sign"].(string) {
				log.Println("sf sign verify error:", input, userSign)
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_SIGN_INCORRECT
				data["msg"] = "认证失败.\nsign verify error"
				tmpl.Execute(w, data)
				return
			}
			personSetting := GetPersonSetting(agent.Name)
			if time.Now().Unix() > ts+int64(personSetting.AccountLockTime) {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_TIMEOUT
				data["msg"] = "充值超时，请重新充值.\nOperation timeout.Please deposit again."
				tmpl.Execute(w, data)
				return
			}
			//amount, _ := strconv.ParseFloat(input["amount"].(string), 64)
			accountInfo, err := dbManager.CheckAccountExists(input["account"].(string))
			if accountInfo.Account == "" || err != nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_ACCOUNT_NOT_EXISTS
				data["msg"] = "帐号不存在.\nAccount not exists."
				tmpl.Execute(w, data)
				return
			}
			payerInfo, err := dbManager.GetPayerInfo(input["account"].(string), input["amount"].(string), userSign)
			if payerInfo.Payer == "" || err != nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_PAYER_NOT_EXISTS
				data["msg"] = "订单资讯不存在.\norder information not exists."
				tmpl.Execute(w, data)
				return
			}
			log.Printf("[SFPay Verfiy Code]:Payer %v\n", payerInfo)
			realUrl := ""
			midpayRecord, err := dbManager.GetMidPayRecord(payerInfo.Account, payerInfo.Amount, payerInfo.Payer, payerInfo.RequestTime)
			if err != nil {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_PAYER_NOT_EXISTS
				data["msg"] = "订单不存在.\norder information not exists."
				tmpl.Execute(w, data)
				return
			}
			if len(midpayRecord) == 1 {
				realUrl = midpayRecord[0].QrUrl
			} else {
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_PAYER_NOT_EXISTS
				data["msg"] = "订单資訊重複.\norder no duplicate."
				tmpl.Execute(w, data)
				return
			}

			/*
				mode := input["mode"].(string)
				if mode == "sfpay" {
					sfPayAccount, err := dbManager.GetSFPayAccount(input["account"].(string), amount)
					if err != nil {
						log.Println("[VerifyCode] QrCorde GetMidpayAccount Error:", input["account"], err)
					}
					if sfPayAccount != nil {
						log.Println("[VerifySFCode] SF QrCorde Verify Code:", input["account"], sfPayAccount)
						if sfPayAccount != nil && sfPayAccount.QrUrl != "" {
							realUrl = fmt.Sprintf("%s?t=%d", sfPayAccount.QrUrl, (time.Now().UnixNano() / 1000000))
							log.Println("sfPayAccount QrCorde Verify Code Url:", sfPayAccount.Account, realUrl)
						} else {
							log.Printf("[VerifySFCode] QrCorde Verify Failed. account:%v person:%v\n ", input["account"], sfPayAccount)
						}
					} else {
						log.Printf("[VerifySFCode] Can't find sfpay account. account:%v\n", input["account"])
					}
				} else if mode == "midpay" {
					//檢查是否有使用萬用碼
					midPayAccount, err := dbManager.GetMidpayAccount(input["account"].(string))
					if err != nil {
						log.Println("[VerifySFCode] QrCorde GetMidpayAccount Error:", input["account"], err)
					}
					if midPayAccount != nil {
						person, err := dbManager.GetPayAccount(input["account"].(string))
						if err != nil {
							log.Println("[VerifySFCode] QrCorde Verify Code Error:", input["account"], err)
						}
						log.Println("[VerifySFCode] QrCorde Verify Code:", input["account"], person)
						if person != nil && person.QrUrl != "" {
							realUrl = fmt.Sprintf("%s?t=%d", person.QrUrl, (time.Now().UnixNano() / 1000000))
							log.Println("[VerifySFCode]QrCorde Verify Code Url:", person.Account, realUrl)
						} else {
							log.Printf("[VerifySFCode] QrCorde Verify Failed. account:%v person:%v\n ", input["account"], person)
						}
					} else {
						log.Printf("[VerifySFCode] Can't find midpay account. account:%v\n", input["account"])
					}
				}*/
			if realUrl != "" {
				dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_VERIFY)
				http.Redirect(w, r, realUrl, http.StatusSeeOther)
				return
			} else {
				dbManager.UpdateMidpayStatusRecord(input["account"].(string), input["ts"].(string), input["payer"].(string), input["amount"].(string), STATE_DEPOSIT_QRCODE_FAILED)
				log.Printf("[VerifySFCode] qrcode error: %v can't find any qrcode url.\n", input["account"])
				tmpl := template.Must(template.ParseFiles("tpl/midpay_error.html"))
				data["code"] = ERROR_IMAGE_INCORRECT
				data["msg"] = fmt.Sprintf("抱歉，系统忙碌中，请稍候重试.\n Sorry, The system is busy. Please try agian later.\n")
				tmpl.Execute(w, data)
				return
			}
		}
	case "sfpayqrcode":
		data = GenQRCode()
	default:
		sendToOldAPI(w, r)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("Handler Panic Error:%s: %s", err, debug.Stack())
			data["status"] = false
			data["code"] = ERROR_REQUEST_VERIFY
			data["error"] = "Request verify error"
			jsonParser(data, w)
			return
		}
	}()

	if len(data) != 0 {
		jsonParser(data, w)
		return
	}

	jsonParser(data, w)
}

func sendToOldAPI(w http.ResponseWriter, r *http.Request) {
	resp, err := SendCustomRequest(r)
	if err != nil {
		log.Println("sendToOldAPI Error:", err)
	} else {
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(resp.Body.([]byte))
		return
	}
}

func getInputQuery(r *http.Request) map[string]interface{} {
	query := r.URL.Query()
	input := make(map[string]interface{})
	for k, v := range query {
		if len(v) == 1 {
			input[k] = v[0]
		}
	}
	return input
}

func getInputForm(r *http.Request) (map[string]interface{}, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("[Handler Error][%d]%v\n", ERROR_REQUEST_FROM, err)
	}
	fmt.Println(r.Form) // print information on server side.
	input := make(map[string]interface{})

	for k, v := range r.Form {
		if reflect.TypeOf(v).String() == "[]string" {
			input[k] = v[0]
		}
	}
	return input, nil
}

func getInputJSON(r *http.Request) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		return nil, fmt.Errorf("[Handler Error][%d]%v\n", ERROR_HTTP_BODY, err)
	}
	defer r.Body.Close()
	var input map[string]interface{}
	err = json.Unmarshal(body, &input)
	if err != nil || input == nil {
		log.Printf("Body input json error:%v\n", err)
		return nil, fmt.Errorf("[Handler Error][%d]%v\n", ERROR_REQUEST_JSON, err)
	}
	return input, nil

}

func verifyToken(token string) bool {
	return false
}

func verifyFields(input map[string]interface{}, fields []string) bool {
	findNum := 0
	for _, v := range fields {
		if _, ok := input[v]; ok {
			findNum++
		}
	}
	if findNum == len(fields) {
		return true
	}
	return false
}

func jsonParser(data interface{}, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if data != nil {
		json, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(500)
			log.Println("Error generating json", err)
			fmt.Fprintln(w, "Could not generate JSON")
			return
		}
		fmt.Fprint(w, string(json))
	} else {
		w.WriteHeader(404)
		fmt.Fprint(w, "404 no data can be find.")
	}
}

func GetMissingFieldsError() map[string]interface{} {
	data := make(map[string]interface{})
	data["status"] = false
	data["code"] = ERROR_MISSING_FIELDS
	data["error"] = "Missing some fields"
	return data
}

func GetPersonSetting(agent string) CustomPerson {
	if v, ok := CONFIGS.Person.CustomList[agent]; ok {
		return v
	}
	return CustomPerson{Name: agent, AccountLockTime: CONFIGS.Person.AccountLockTime, ClientLockTime: CONFIGS.Person.ClientLockTime, Pay: CONFIGS.Person.Pay, SFPay: CONFIGS.SFPay}
}

func GetDevice(ua string) string {
	device := "desktop"
	if strings.Contains(ua, "iPhone") {
		device = "ios"
	} else if strings.Contains(ua, "Android") {
		device = "android"
	} else if strings.Contains(ua, "Windows Phone") {
		device = "ms"
	}
	return device
}

func getIP(req *http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
		return req.RemoteAddr
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return req.RemoteAddr
	}

	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	forward := req.Header.Get("X-Forwarded-For")
	log.Printf("Forwarded for: %s\n", forward)
	return ip
}
