package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

func SysDeposit(depositNumber string, platfrom string) map[string]interface{} {
	data := make(map[string]interface{})
	deposit, err := dbManager.GetDeposit(depositNumber, platfrom)
	if deposit == nil || err != nil {
		data["code"] = ERROR_DEPOSIT_NOT_EXISTS
		data["msg"] = "Deposit not exists."
		return data
	}
	platfromId, _ := strconv.Atoi(platfrom)
	agent, err := dbManager.GetAgentById(platfromId)
	if agent == nil || err != nil {
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "Platfrom not exists."
		return data
	}
	deposit.CallbackUrl = agent.CallbackUrl
	if deposit.Remark != "" {
		payer, _ := dbManager.GetPayer(deposit.WechatName, deposit.Remark, deposit.Amount)
		log.Println("[SysDeposit] payer:", payer)
		if payer != nil {
			deposit.CustomSign = payer.CustomSign
		}
		dbManager.UpdateMidpayRecord(deposit.WechatName, deposit.Remark, deposit.Amount, payer.RequestTime)
		log.Println("[SysDeposit] UpdateMidpayRecord:", payer)
	}
	log.Printf("SysDeposit:[%v] Platfrom:%v Type:%v CallbackUrl:%v Payer:%s CS:%v\n", deposit.DepositNumber, deposit.Platfrom, deposit.PayType, deposit.CallbackUrl, deposit.Remark, deposit.CustomSign)
	status := SendNotifcationToAgent(deposit)
	data["code"] = status
	if status != 200 {
		data["msg"] = fmt.Sprintf("sync deposit failed.Number:%s Platfrom:%s", deposit.DepositNumber, deposit.CallbackUrl)
	} else {
		data["msg"] = "success"
	}
	return data
}

func genAccountAmount(input map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	fields := []string{"account", "amount"}
	if !verifyFields(input, fields) {
		return GetMissingFieldsError()
	}
	accountInfo, err := dbManager.CheckAccountExists(input["account"].(string))
	if err != nil {
		log.Println("[GenAmountQRCode]Can't find account info.", input["account"])
		data["code"] = 400
		data["msg"] = "Can't find account info"
		return data
	}
	noteMiddle := String(2)
	amount, _ := strconv.ParseFloat(input["amount"].(string), 64)
	if amount <= 0 {
		data["code"] = 400
		data["msg"] = fmt.Sprintf("Amount can't be zero. Amount:%.2f", amount)
		return data
	}
	noteSign := fmt.Sprintf("%s-%s-%.2f", accountInfo.IP[2:], noteMiddle, amount)
	qrUrl := getPayAppUrl(accountInfo.Account, accountInfo.Type, noteSign, amount)
	if qrUrl != "" {
		exist, _ := dbManager.CheckSFPayNoteExists(accountInfo.Account, amount)
		if exist {
			err := dbManager.UpdateSFPay(accountInfo.Account, accountInfo.Platfrom, amount, noteSign, qrUrl)
			if err == nil {
				log.Printf("[GenAmountQRCode] Update QrCode: Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, qrUrl)
				data["code"] = 200
				data["msg"] = "success. Mode: Update"
			} else {
				log.Printf("[GenAmountQRCode] Update QrCode Error: Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, qrUrl)
				data["code"] = 200
				data["msg"] = "Update Failed. QRCode:" + err.Error()
			}
		} else {
			state, _ := dbManager.InsertSFPay(accountInfo, amount, noteSign, qrUrl)
			if state {
				log.Printf("[GenAmountQRCode] Insert QrCode:Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, qrUrl)
				data["code"] = 200
				data["msg"] = "success. Mode: Insert"
			} else {
				log.Printf("[GenAmountQRCode] Insert QrCode Error: Account:%s amount:%.2f qrcode:%s\n", accountInfo.Account, amount, qrUrl)
				data["code"] = 400
				data["msg"] = "Insert Failed. QRCode:" + qrUrl
			}
		}
	} else {
		data["code"] = 402
		data["msg"] = "Can't generate the QRCoce."
	}

	return data
}

func manualSaveDeposit(record map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	log.Println("manualSaveDeposit: The deposit not exsits. create new.")
	var depositRecord DepositRecord
	if record["createUser"] == nil {
		log.Println("createUser not exists.")
		data["code"] = ERROR_DEPOSIT_NUMBER_EXISTS
		data["msg"] = "createUser number exists."
		return data
	}
	if record["platfrom"] == nil {
		log.Println("platfrom not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "platfrom not exists."
		return data
	}
	if record["amount"] == nil {
		log.Println("Amount not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "Amount not exists."
		return data
	}
	if record["wechatName"] == nil {
		log.Println("account not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "account not exists."
		return data
	}
	if record["payAccount"] == nil {
		log.Println("payAccount not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "payAccount not exists."
		return data
	}
	if record["depositNumber"] == nil {
		log.Println("DepositNumber not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "DepositNumber not exists."
		return data
	}
	if record["note"] == nil {
		log.Println("Note not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "Note not exists."
		return data
	}
	if record["tranTime"] == nil {
		log.Println("transferTime not exists.")
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "transferTime not exists."
		return data
	}
	deposit, err := dbManager.GetDeposit(record["depositNumber"].(string), record["platfrom"].(string))
	if deposit != nil {
		data["code"] = ERROR_DEPOSIT_NUMBER_EXISTS
		data["msg"] = "deposit number exists."
		return data
	}
	accountInfo, err := dbManager.CheckAccountExists(record["wechatName"].(string))
	if err != nil || accountInfo.Account == "" {
		log.Println("account not exists.")
		data["code"] = ERROR_ACCOUNT_NOT_EXISTS
		data["msg"] = "Account not exists."
		return data
	}

	payerRecord, err := dbManager.GetPayerInfo(record["wechatName"].(string), record["amount"].(string), record["note"].(string))
	if err != nil {
		log.Println("manualSaveDeposit PayerRecord Error:", err)
	}
	record["sign"] = record["note"]
	if payerRecord.Payer != "" {
		record["sign"] = record["note"]
		record["note"] = payerRecord.Payer
		depositRecord.Remark = payerRecord.Payer
		depositRecord.CustomSign = payerRecord.CustomSign
	}
	amount, _ := strconv.ParseFloat(record["amount"].(string), 64)
	depositRecord.Amount = amount
	depositRecord.CreateUser = record["createUser"].(string)
	depositRecord.DepositNumber = record["depositNumber"].(string)
	depositRecord.Note = record["note"].(string)
	depositRecord.PayAccount = record["payAccount"].(string)
	depositRecord.Platfrom, _ = strconv.Atoi(record["platfrom"].(string))
	if depositRecord.Platfrom == 0 {
		depositRecord.Platfrom = accountInfo.Platfrom
		log.Println("ManualSaveDeposit: Can't find platform from client. Use account platform")
	}
	// Data init
	depositRecord.CreateTime = time.Now()
	depositRecord.TranTime = depositRecord.CreateTime
	depositRecord.ExcuteTime = depositRecord.CreateTime
	if record["tranTime"] == nil {
		record["tranTime"] = depositRecord.CreateTime.Format("2006-01-02 15:04:05")
	} else {
		i, err := strconv.ParseInt(record["tranTime"].(string), 10, 64)
		if err != nil {
			log.Println("transferTime Error:", err)
		}
		i = int64(i / 1000)
		tm := time.Unix(i, 0)
		record["tranTime"] = tm.Format("2006-01-02 15:04:05")
	}
	depositRecord.TransferTime = record["tranTime"].(string)
	depositRecord.State = STATE_PENDING
	depositRecord.Times = 0
	depositRecord.BillNo = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	depositRecord.WechatName = record["wechatName"].(string)
	record["deviceAccount"] = depositRecord.WechatName
	log.Println("manualSaveDeposit:", depositRecord)
	data = saveDeposit(&accountInfo, &depositRecord, record)
	log.Println("manualSaveDeposit Result:", data)
	return data
}
