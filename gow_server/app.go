package main

import (
	"log"
	"time"
)

type PayApp struct {
	Command     string
	Timestamp   int64
	Payer       string
	ResultState bool
	AccountType string
	Account     string
	Amount      float64
	Data        PayAppResult
}

type PayAppResult struct {
	Account string
	Url     string
	Code    int
	Message string
	Data    map[string]interface{}
}

func GetPayApp(cmd string, account string, accountType string, payer string, amount float64) *PayApp {
	var pa PayApp
	pa.Command = cmd
	switch accountType {
	case "1":
		pa.AccountType = "alipay"
	case "2":
		pa.AccountType = "wechat"
	case "3":
		pa.AccountType = "qq"
	}
	pa.Account = account
	pa.Amount = amount
	pa.Payer = payer
	pa.Timestamp = time.Now().Unix()
	return &pa
}

func (pa *PayApp) GetPay() {
	err := GetAppPay(pa)
	if err != nil {
		log.Println("GetPay Error", err)
		return
	}
	for !pa.ResultState {
		if pa.ResultState {
			break
		}
		if time.Now().Unix()-pa.Timestamp > 30 {
			log.Println("Request App timeout")
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (pa *PayApp) SetPay(input PayAppResult) {
	pa.Data = input
	pa.ResultState = true
}
