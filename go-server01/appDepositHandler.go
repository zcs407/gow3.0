package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func depositList(input map[string]interface{}) map[string]interface{} {
	fields := []string{"alipayAccount", "depositRecords"}
	if !verifyFields(input, fields) {
		return GetMissingFieldsError()
	}
	data := make(map[string]interface{})
	log.Println("alipayAccount", input["alipayAccount"])
	alipayAccount := input["alipayAccount"].(string)
	log.Println("alipayAccount:", alipayAccount)
	depositRecords := input["depositRecords"].([]interface{})
	accountInfo, err := dbManager.CheckAccountExists(alipayAccount)
	if err != nil || accountInfo.Account == "" {
		data["code"] = 500
		data["msg"] = "Account not exists."
		return data
	}
	var depositList []DepositRecord
	for _, raw := range depositRecords {
		record := raw.(map[string]interface{})
		var depositRecord DepositRecord

		if record["amount"] == nil {
			data["code"] = ERROR_MISSING_FIELDS
			data["msg"] = "Amount not exists."
			return data
		}
		if record["wechatName"] == nil {
			data["code"] = ERROR_MISSING_FIELDS
			data["msg"] = "WechatName not exists."
			return data
		}
		if record["depositNumber"] == nil {
			data["code"] = ERROR_MISSING_FIELDS
			data["msg"] = "DepositNumber not exists."
			return data
		}
		if record["note"] == nil {
			if accountInfo.Type == "1" {
				data["code"] = ERROR_MISSING_FIELDS
				data["msg"] = "Note not exists."
				return data
			}
			record["note"] = ""
		}
		depositRecord.Amount = record["amount"].(float64)
		depositRecord.CreateUser = record["createUser"].(string)
		depositRecord.DepositNumber = record["depositNumber"].(string)
		depositRecord.Note = record["note"].(string)
		depositRecord.PayAccount = record["payAccount"].(string)
		depositRecord.Platfrom, _ = strconv.Atoi(record["platfrom"].(string))
		// Data init
		depositRecord.CreateTime = time.Now()
		depositRecord.TranTime = depositRecord.CreateTime
		depositRecord.ExcuteTime = depositRecord.CreateTime
		if record["transferTime"] == nil {
			record["transferTime"] = depositRecord.CreateTime.Format("2006-01-02 15:04:05")
		}
		depositRecord.TransferTime = record["transferTime"].(string)
		depositRecord.State = STATE_PENDING
		depositRecord.Times = 0
		depositRecord.BillNo = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
		depositRecord.WechatName = record["wechatName"].(string)
		depositList = append(depositList, depositRecord)
	}

	switch accountInfo.Type {
	case "0": //微信
		for _, deposit := range depositList {
			data = saveDeposit(&accountInfo, &deposit, nil)
			if data["code"].(int) != 200 {
				return data
			}
		}
	case "1": //支付寶
		if len(depositList) == 1 {
			log.Println("saveDeposit APP[Origin]:", depositList)
			for _, deposit := range depositList {
				data = saveDeposit(&accountInfo, &deposit, nil)
				if data["code"].(int) != 200 {
					return data
				}
			}
		} else {
			log.Println("saveDepositRecord:", depositList)
			for _, deposit := range depositList {
				result := saveDepositRecord(&accountInfo, &deposit)
				if result["code"] != 200 {
					log.Printf("saveDepositRecord Failed:%s Deposit:%v\n", result["msg"], deposit)
				}
			}
		}
	}
	subData := make(map[string]interface{})

	data["code"] = 200
	data["msg"] = "success"
	data["data"] = nil
	data["totalnumber"] = nil
	data["totalamount"] = nil
	data["pageamount"] = nil
	subData["code"] = 200
	subData["msg"] = "success"
	subData["data"] = nil
	subData["totalnumber"] = nil
	subData["totalamount"] = nil
	subData["pageamount"] = nil
	data["data"] = subData
	return data
}

func saveDepositRecord(account *AccountInfo, deposit *DepositRecord) map[string]interface{} {
	data := make(map[string]interface{})
	data["code"] = 200
	data["msg"] = "success"
	if account == nil {
		data["code"] = ERROR_MISSING_FIELDS
		data["msg"] = "Can't find account information."
		return data
	}

	exist, _ := dbManager.CheckDepositNumberExists(deposit.DepositNumber)
	if exist {
		data["code"] = ERROR_DEPOSIT_NUMBER_EXISTS
		data["msg"] = "Deposit number had exists."
		return data
	}
	agent, err := dbManager.GetAgentById(deposit.Platfrom)
	if agent == nil || err != nil {
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "platfrom not exists."
		return data
	}
	deposit.CallbackUrl = agent.CallbackUrl
	deposit.Note = strings.Replace(deposit.Note, "商品-", "", -1)
	deposit.IP = account.IP
	deposit.PayType = account.Type
	deposit.NickName = account.NickName
	deposit.TranFee = -deposit.Amount * agent.PayFee
	deposit.Sign = getMD5(fmt.Sprintf("%s%.f%s%s", deposit.DepositNumber, deposit.Amount, deposit.Note, agent.Sign))
	ok, err := dbManager.InsertDepositRecord(deposit)
	if !ok || err != nil {
		data["code"] = ERROR_DATABASE_INSERT
		data["msg"] = "Insert the deposit failed."
		return data
	}
	return data
}

func saveDeposit(account *AccountInfo, deposit *DepositRecord, args map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	data["code"] = 200
	data["msg"] = "success"
	if account == nil {
		data["code"] = ERROR_MISSING_FIELDS
		data["msg"] = "Can't find account information."
		return data
	}

	exist, _ := dbManager.CheckDepositNumberExists(deposit.DepositNumber)
	if exist {
		data["code"] = ERROR_DEPOSIT_NUMBER_EXISTS
		data["msg"] = "Deposit number had exists."
		return data
	}
	accountInfo, err := dbManager.CheckAccountExists(deposit.WechatName)
	if accountInfo.Account == "" || err != nil {
		data["code"] = ERROR_ACCOUNT_NOT_EXISTS
		data["msg"] = "Account not exists."
		return data
	}
	if accountInfo.Platfrom == 0 {
		accountInfo.Platfrom = deposit.Platfrom
	}
	if accountInfo.Platfrom != deposit.Platfrom {
		log.Printf("Warning Deposit Platfrom Error: Account Platfrom:%d Deposit:%d\n", accountInfo.Platfrom, deposit.Platfrom)
		//deposit.Platfrom = accountInfo.Platfrom
	}
	agent, err := dbManager.GetAgentById(deposit.Platfrom)
	if agent == nil || err != nil {
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "platfrom not exists."
		return data
	}
	deposit.CallbackUrl = agent.CallbackUrl
	deposit.Note = strings.Replace(deposit.Note, "商品-", "", -1)
	deposit.IP = account.IP
	deposit.PayType = account.Type
	//跳過舊的API 檢查機制
	deposit.Times = 3
	deposit.NickName = account.NickName
	deposit.TranFee = -deposit.Amount * agent.PayFee
	deposit.Sign = getMD5(fmt.Sprintf("%s%d%s%s", deposit.DepositNumber, int64(deposit.Amount), deposit.Note, agent.Sign))
	personSetting := GetPersonSetting(agent.Name)
	var payers []MidPayRecord
	if personSetting.SFPay && args != nil {
		if _, ok := args["deviceAccount"]; !ok {
			data["code"] = 400
			data["msg"] = "deviceAccount not exists."
			return data
		}
		if _, ok := args["amount"]; !ok {
			data["code"] = 400
			data["msg"] = "amount not exists."
			return data
		}
		if _, ok := args["sign"]; !ok {
			data["code"] = 400
			data["msg"] = "sign not exists."
			return data
		}
		payerRecord, err := dbManager.GetPayerInfo(args["deviceAccount"].(string), args["amount"].(string), args["sign"].(string))
		if err != nil {
			log.Println("SFDeposit PayerRecord Error:", err)
		}
		log.Println("Use SFPay PayerRecord:", payerRecord)
		if payerRecord.Payer != "" {
			payers = append(payers, payerRecord)
		} else {
			payers, _ = dbManager.CheckDepoistAccountRequest(personSetting, account.Account, deposit)
		}
	} else {
		payers, _ = dbManager.CheckDepoistAccountRequest(personSetting, account.Account, deposit)
	}
	if len(payers) == 0 {
		layout := "2006-01-02 15:04:05"
		depositTime, err := time.Parse(layout, deposit.TransferTime)
		if err != nil {
			log.Println("Deposit Time Error:", err)
		} else {
			localTime := depositTime.Local()
			_, zoneDiff := localTime.Zone()
			localTime = localTime.Add(time.Duration(-int64(zoneDiff)) * time.Second)
			payers, _ = dbManager.CheckDepoistAccountOverTimeRequest(personSetting, account.Account, deposit.Amount, localTime)
			if len(payers) == 1 {
				//payers[0].RequestTime = localTime
				log.Println("[SaveDeposit]Find Over Time Records:", account.Account, deposit.Amount, deposit.TransferTime)
			}
			//log.Println("SaveDeposit Search OverTime:", account.Account, deposit.Amount, deposit.TransferTime, localTime)
		}
	}

	log.Println("[Payers]:", account.Account, payers)
	apiType := ""

	var payer *MidPayRecord
	if len(payers) > 0 {
		if len(payers) > 1 {
			for _, v := range payers {
				if deposit.Note == "商品" && v.APIType == "midpay" {
					payer = &v
					log.Println("Find Duplicate Records. It's Midpay.", payer)
					break
				}
			}
		} else if len(payers) == 1 {
			payer = &payers[0]
		}
	}
	if payer != nil {
		midpayRecords, _ := dbManager.GetMidPayRecord(account.Account, deposit.Amount, payer.Payer, payer.RequestTime)
		payerPlatform := 0
		if len(midpayRecords) == 1 {
			payerPlatform, _ = strconv.Atoi(midpayRecords[0].Platform)
			apiType = midpayRecords[0].APIType
			if payerPlatform != 0 {
				log.Println("Set Deposit Platform Id:", account.Account, deposit.Amount, payer.Payer, payerPlatform)
				deposit.Platfrom = payerPlatform
			}
		}
		deposit.Note = payer.Payer
		deposit.Remark = payer.Payer
		deposit.CustomSign = payer.CustomSign
		deposit.Sign = getMD5(fmt.Sprintf("%s%d%s%s", deposit.DepositNumber, int64(deposit.Amount), deposit.Note, agent.Sign))
		log.Printf("Find Payer[%s]:Set Payer info to deposit:%v\n", apiType, deposit)
		switch apiType {
		case "midpay":
			dbManager.UpdateMidpay(payers[0].Account)
		case "sfpay":
			dbManager.UpdateSFPayStatus(payers[0].Account, deposit.Amount)
		default:
			log.Println("API type not found.")
		}
		dbManager.UpdateMidpayRecord(payers[0].Account, payers[0].Payer, deposit.Amount, payers[0].RequestTime)
	}
	ok, err := dbManager.InsertDepositRecord(deposit)
	if !ok || err != nil {
		data["code"] = ERROR_DATABASE_INSERT
		data["msg"] = "Insert the deposit failed."
		return data
	}
	if apiType == "sfpay" {
		log.Println("[UpdateDepositCount]:", payer.Account, apiType)
		dbManager.UpdateDepositCount(payer.Account)
	}
	if ok {
		go SendNotifcationToAgent(deposit)
	}
	log.Println("Save Deposit Success", deposit)
	state := STATE_NORMAL
	if accountInfo.DayAmount+deposit.Amount > accountInfo.DayLimit {
		state = STATE_DISABLED
		log.Printf("[Account DAY Limit]:%s Amount:%.f DayLimit:%.f\n", deposit.WechatName, accountInfo.DayAmount+deposit.Amount, accountInfo.DayLimit)
	}
	accountInfo, _ = dbManager.CheckAccountExists(deposit.WechatName)
	dbManager.UpdateAmount(deposit.WechatName, accountInfo.DayAmount+deposit.Amount, deposit.Amount, state)
	nowAmount, err := dbManager.GetAccountAmount(deposit.WechatName)
	if err != nil {
		nowAmount = accountInfo.Amount + deposit.Amount
	}
	//Normal Report
	var report UserReport
	report.Account = deposit.WechatName
	report.BeforeMoney = accountInfo.Amount
	report.IP = deposit.IP
	report.NowMoney = nowAmount
	report.ChangeMoney = deposit.Amount
	report.NickName = deposit.NickName
	report.UserName = deposit.Note + "-" + deposit.Remark
	report.Remark = deposit.DepositNumber
	report.Type = COMINPUT
	report.CreateTime = time.Now()
	report.Platfrom = fmt.Sprintf("%d", deposit.Platfrom)
	report.CreateUser = deposit.CreateUser
	report.AccountType = accountInfo.Type
	err = dbManager.InsertReport(report)
	if err != nil {
		log.Println("InsertReport Error:", err)
	}
	//Fee Report
	var reportFee UserReport
	reportFee.Account = deposit.WechatName
	reportFee.BeforeMoney = accountInfo.Amount + deposit.Amount
	reportFee.IP = deposit.IP
	reportFee.NowMoney = accountInfo.Amount + deposit.Amount
	reportFee.ChangeMoney = -deposit.Amount * agent.PayFee
	reportFee.NickName = deposit.NickName
	reportFee.UserName = deposit.Note + "-" + deposit.Remark
	reportFee.Remark = deposit.DepositNumber
	reportFee.Type = COMINPUTFEE
	reportFee.CreateTime = time.Now()
	reportFee.Platfrom = fmt.Sprintf("%d", deposit.Platfrom)
	reportFee.CreateUser = deposit.CreateUser
	reportFee.AccountType = accountInfo.Type
	dbManager.InsertReport(reportFee)
	agent, err = dbManager.GetAgentById(deposit.Platfrom)
	if agent == nil || err != nil {
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "platfrom not exists."
		return data
	}
	dbManager.UpdateAgentAmount(agent.Name, (deposit.Amount + reportFee.ChangeMoney))
	//sent fee to parent agent.
	agentList, err := dbManager.GetParentAgentByName(agent.Name)
	if err != nil {
		log.Printf("GetParentAgentByName Error:%v\n", err)
	}
	for _, pa := range agentList {
		fee := deposit.Amount * pa.PayFee
		var reportFee UserReport
		reportFee.Account = deposit.WechatName
		reportFee.BeforeMoney = pa.Amount
		reportFee.IP = deposit.IP
		reportFee.NowMoney = pa.Amount + fee
		reportFee.ChangeMoney = fee
		reportFee.NickName = deposit.NickName
		reportFee.UserName = deposit.Note
		reportFee.Remark = deposit.DepositNumber + " Payer:" + deposit.Remark
		reportFee.Type = COMINPUT
		reportFee.CreateTime = time.Now()
		reportFee.Platfrom = fmt.Sprintf("%d", pa.Id)
		reportFee.CreateUser = deposit.CreateUser
		reportFee.AccountType = accountInfo.Type
		dbManager.InsertReport(reportFee)
		log.Printf("SentFee[%s][%s] Deposit:%f Fee:%f Agent Amount:%f\n", agent.Name, pa.Name, deposit.Amount, fee, reportFee.NowMoney)
		dbManager.UpdateAgentAmount(pa.Name, (fee))
	}

	data["code"] = 200
	data["msg"] = "success"
	return data
}

func checkFailDeposit() {
	if dbManager == nil {
		return
	}
	failedList, err := dbManager.GetDepositFailedStatus()
	if err != nil {
		log.Println("checkFailDeposit Error:", err)
		return
	}
	if len(failedList) == 0 {
		//log.Println("checkFailDeposit can find failed deposit.")
		return
	}
	for _, deposit := range failedList {
		go SendNotifcationToAgent(&deposit)
	}
}

func SendNotifcationToAgent(deposit *DepositRecord) int {
	url := deposit.CallbackUrl
	if url == "" {
		agent, err := dbManager.GetAgentById(deposit.Platfrom)
		if agent != nil && err == nil {
			deposit.CallbackUrl = agent.CallbackUrl
			log.Println("SendNotifcationToAgent: Deposit callback Url is empty. reset it.", deposit)
			url = deposit.CallbackUrl
		}
	}
	fmt.Println("URL:>", url)
	jsonBytes, err := json.Marshal(deposit)
	if err != nil {
		log.Println("SendNotifcationToAgent JSON Error:", err)
		return 500
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("SendNotifcationToAgent Error:", err)
		return 500
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if resp.StatusCode == 200 {
		dbManager.UpdateDepositState(deposit.DepositNumber, STATE_EXECUTED)
	} else {
		deposit.Times++
		dbManager.UpdateDepositStateFailed(deposit.DepositNumber, STATE_PENDING, deposit.Times)
	}
	return resp.StatusCode
}
