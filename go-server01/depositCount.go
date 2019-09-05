package main

import "log"

func ResetDepositCount() {
	accounts, err := dbManager.GetAllAccountsStatus()
	if err != nil {
		log.Println("[ResetDepositCount]Error:", err)
		return
	}
	success := 0
	for _, v := range accounts {
		err := dbManager.UpdateDepositCountReset(v.Account)
		if err == nil {
			success++
		} else {
			log.Println("[UpdateDepositCountReset]Error:", err)
		}
	}
	log.Printf("[ResetDepositCount]: Accounts:%d Updated:%d\n", len(accounts), success)
}
