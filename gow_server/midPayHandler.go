package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func midPay(input map[string]interface{}) map[string]interface{} {
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
	availableAccountsSize := len(dbManager.GetAvailableAccounts(agent.Id))
	log.Printf("merchanntNo:%s AccountSize:%d personSetting:%v\n", agent.Name, availableAccountsSize, personSetting)
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
	realUrl := ""
	payerSign := ""
	insertTime := time.Now()
	if len(accounts) == 1 {
		person, err = dbManager.GetPayAccount(accounts[0].Account)
		if err != nil {
			err = fmt.Errorf("Can't find any account use.")
		}
		if person != nil {
			url = fmt.Sprintf("%s/%s.jpg", CONFIGS.Person.ImageServer, person.Account)
		}
	} else {
		for {
			userType := "1"
			switch midpayType {
			case "wxapi":
				userType = "0"
			case "aliapi":
				userType = "1"
			case "qqapi":
				userType = "2"
			}
			if CONFIGS.HB {
				person, err = dbManager.GetAvailableLockedAccount(agent, payer, userType, amount)
				if err != nil {
					log.Println("GetAvailableUnLockedAccount Failed.", err)
					break
				}
			} else {
				person, err = dbManager.GetAvailableAccount(agent, payer, userType, amount)
				if err != nil {
					log.Println("GetAvailableAccount Failed.", err)
					break
				}
			}

			if person != nil {
				payerSign = getMD5(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
				if !CONFIGS.HB {
					realUrl = getPayAppUrl(person.Account, person.Type, payerSign, amount)
					if realUrl == "" {
						dbManager.UpdateLockedAccount(person.Account)
					}
				} else {
					realUrl = person.QrUrl
					err = nil
				}
			}
			if realUrl == "" {
				person, err = dbManager.GetAvailableUnLockedAccount(agent, payer, userType, amount)
				if err != nil {
					log.Println("GetAvailableUnLockedAccount Failed.", err)
					break
				}
				query := false
				if person != nil {
					payerSign = getMD5(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
					realUrl = getPayAppUrl(person.Account, person.Type, payerSign, amount)
					if realUrl == "" {
						query = true
					} else {
						dbManager.UpdateUnLockedAccount(person.Account)
						dbManager.UpdateMidpayTime(payer, person.Account)
					}
				} else {
					err = fmt.Errorf("Can't find any account use.[2]")
					query = true
				}

				if query {
					person, err = dbManager.GetAvailableLockedAccount(agent, payer, userType, amount)
					if err != nil {
						log.Println("GetAvailableUnLockedAccount Failed.", err)
						break
					}
					if person != nil {
						realUrl = person.QrUrl
						err = nil
						dbManager.UpdateMidpayTime(payer, person.Account)
						log.Println("Find locked account.", realUrl)
					} else {
						err = fmt.Errorf("Can't find any account use.[3]")
						break
					}
				}
			} else {
				dbManager.UpdateMidpayTime(payer, person.Account)
			}
			break
		}
	}

	if err != nil {
		data["code"] = 501
		data["msg"] = err.Error()
		data["data"] = nil
		return data
	}
	if amount == 0 {
		data["code"] = 500
		data["msg"] = "Amount cant't be zero."
		data["data"] = nil
	} else {
		//代表這個payer值沒被用過,可以用
		if len(accounts) == 0 {
			status, err := dbManager.InsertMidpayRecord(person.Account, amount, payer, insertTime, customSign, fmt.Sprintf("%d", agent.Id), "midpay", realUrl, input["device"].(string), STATE_DEPOSIT_CHECKING)
			if !status || err != nil {
				log.Println("Insert to midpay record failed.", err)
			}
			payerSign := getMD5Slim(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
			status, err = dbManager.InsertPayerRecord(person.Account, amount, payer, insertTime, payerSign, customSign)
			if !status || err != nil {
				log.Println("Insert to payer record failed.", err)
			}
		} else if len(accounts) == 1 {
			insertTime = accounts[0].RequestTime
		}
		//Set the output data in data object
		var midPayObject MidPayObject
		midPayObject.Mode = "midpay"
		midPayObject.Account = person.Account
		midPayObject.Amount = amount
		midPayObject.NickName = person.NickName
		midPayObject.OverTime = personSetting.ClientLockTime
		midPayObject.Payer = payer
		midPayObject.QRUrl = person.QrUrl
		midPayObject.Timestamp = insertTime.Unix()
		midPayObject.RealName = person.RealName
		payerSign := getMD5(fmt.Sprintf("%s%s%d%.2f%s", person.Account, payer, insertTime.Unix(), amount, agent.Sign))
		url = fmt.Sprintf("%s/qrcode?account=%s&ts=%d&payer=%s&platfrom=%d&amount=%.2f&sign=%s", CONFIGS.Person.OutputImageServer, person.Account, insertTime.Unix(), payer, agent.Id, amount, payerSign)
		midPayObject.Url = url
		midPayObject.RefreshTime = personSetting.AccountLockTime
		data["data"] = midPayObject
	}

	return data
}

func checkImageUrlExists(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("checkImageUrlExists Error:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func getPayAppUrl(account string, accountType string, userSign string, amount float64) string {
	if len(userSign) > 20 {
		userSign = userSign[:20]
	}
	payApp := GetPayApp("getpay", account, accountType, userSign, amount)
	payApp.GetPay()
	realUrl := ""
	if payApp.Data.Code == 200 {
		realUrl = payApp.Data.Url
	} else {
		if CONFIGS.HB {
			//realUrl = readQrCode(account)
			//log.Println("QR Scanner has disabled.")
		} else {
			log.Printf("get pay app error[%v]: %v.\n", account, payApp.Data.Message)
		}
	}
	log.Printf("getPayAppUrl Account:%s Url:%s\n", account, realUrl)
	return realUrl
}
