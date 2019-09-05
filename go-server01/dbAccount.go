package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

func (dbm *DBManager) GetAccountAmount(account string) (float64, error) {
	var amount float64
	queryTx, err := dbm.DB.Begin()
	rows, err := queryTx.Query("SELECT amount FROM wechat WHERE wechat_name=?", account)
	if err != nil {
		log.Printf("GetAccountAmount Error:%s\n", err)
		return 0, err
	}
	defer queryTx.Commit()
	if rows.Next() {
		if err = rows.Scan(&amount); err != nil {
			log.Println("Error GetAccountAmount Incorrect.", err)
			rows.Close()
			return amount, err
		}
	}
	rows.Close()
	return amount, nil
}

func (dbm *DBManager) GetAccountCookie(account string) (string, error) {
	var cookie string
	queryTx, err := dbm.DB.Begin()
	rows, err := queryTx.Query("SELECT url FROM wechat WHERE wechat_name=?", account)
	if err != nil {
		log.Printf("GetAccountCookie Error:%s\n", err)
		return cookie, err
	}
	defer queryTx.Commit()
	if rows.Next() {
		if err = rows.Scan(&cookie); err != nil {
			log.Println("Error GetAccountCookie Incorrect.", err)
			rows.Close()
			return cookie, err
		}
	}
	rows.Close()
	return cookie, nil
}

func (dbm *DBManager) UpdateAmount(account string, dayAmount float64, Amount float64, state string) error {
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateAmount Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	result, err := queryTx.Exec("Update wechat set amount=amount+?,dayamount=?,state=? where wechat_name=?", Amount, dayAmount, state, account)
	if err != nil {
		log.Println("UpdateAmount Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateAmount Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update account amount. Account:%s Amount:%f DayAmount:%f\n", account, Amount, dayAmount)
	}
	return nil
}

func (dbm *DBManager) TransferPayAccounts() {
	dbm.GetPayAccounts()
	dbm.VerifyPayAccounts()
}

//取得全部的帳號
func (dbm *DBManager) GetPayAccounts() {
	rows, err := dbm.DB.Query("SELECT wechat_name,state,url,nick_name,real_name,qrurl,plaftfrom,type FROM wechat WHERE state='NORMAL'")
	if err != nil {
		log.Printf("GetAvailableAccount Error:%s\n", err)
		return
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetPayAccounts Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		var account string
		var state string
		var url string
		var nickName string
		var realName string
		var qrUrl string
		var personType string
		var platfrom string
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "wechat_name":
				account = value
			case "type":
				personType = value
			case "plaftfrom":
				platfrom = value
			case "state":
				state = value
			case "url":
				url = value
			case "nick_name":
				nickName = value
			case "real_name":
				realName = value
			case "qrurl":
				qrUrl = value
			}
		}
		if _, ok := dbm.PersonList[account]; ok {
			dbm.PersonList[account].State = state
		} else {
			var person Person
			person.Account = account
			person.State = state
			person.NickName = nickName
			person.QrUrl = qrUrl
			person.RealName = realName
			person.Url = url
			person.Platfrom = platfrom
			person.Type = personType
			dbm.PersonList[account] = &person
		}
	}
}

func (dbm *DBManager) GetPayAccount(personAccount string) (*Person, error) {
	rows, err := dbm.DB.Query("SELECT wechat_name,state,url,nick_name,real_name,qrurl,plaftfrom,type FROM wechat WHERE wechat_name=?", personAccount)
	if err != nil {
		log.Printf("GetPayAccount Error:%s\n", err)
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetPayAccount Error:%s\n", err.Error())
		return nil, err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	var person Person
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		var account string
		var state string
		var url string
		var nickName string
		var realName string
		var qrUrl string
		var personType string
		var platfrom string
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				log.Printf("GetPayAccount Error[%s]:is nil.", columns[i])
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "wechat_name":
				account = value
			case "type":
				personType = value
			case "plaftfrom":
				platfrom = value
			case "state":
				state = value
			case "url":
				url = value
			case "nick_name":
				nickName = value
			case "real_name":
				realName = value
			case "qrurl":
				qrUrl = value
			}
		}
		person.Account = account
		person.State = state
		person.NickName = nickName
		person.QrUrl = qrUrl
		person.RealName = realName
		person.Url = url
		person.Platfrom = platfrom
		person.Type = personType
		dbm.PersonList[personAccount] = &person
	}
	return &person, nil
}

func (dbm *DBManager) MidPayAccountExists(account string) bool {
	exists := false
	rows, err := dbm.DB.Query("SELECT account FROM midpay WHERE account=?", account)
	if err != nil {
		log.Printf("MidPayAccountExists SELECT Error:%s\n", err)
		return exists
	}

	if rows.Next() {
		log.Printf("MidPayAccountExists Query Found:%s\n", account)
		exists = true
	}
	defer rows.Close()
	return exists
}

//取得全部的帳號
func (dbm *DBManager) VerifyPayAccounts() {
	//先移除不能用的帳號
	for account, person := range dbm.PersonList {
		if person.State != "NORMAL" && dbm.MidPayAccountExists(account) {
			dbm.DB.Exec("DELETE FROM midpay WHERE account=?", account)
			log.Printf("[VerifyPayAccounts] Delete Disabled Account:%s Platfrom:%s Type:%s\n", account, person.Platfrom, person.Type)
		}
	}
	//確認還沒加進去的帳號名單
	inDBList := make(map[string]*Person)
	for account, person := range dbm.PersonList {
		if person.State == "NORMAL" {
			rows, err := dbm.DB.Query("SELECT account FROM midpay WHERE account=?", account)
			if err != nil {
				log.Printf("VerifyPayAccounts SELECT Error:%s\n", err)
				inDBList[account] = person
			}
			if !rows.Next() {
				log.Printf("VerifyPayAccounts Query Not Found:%s\n", account)
				inDBList[account] = person
			}
			defer rows.Close()
		}
	}
	//加入的帳號加入
	for account, person := range inDBList {
		_, err := dbm.DB.Exec("INSERT INTO midpay(account,useTime,status,platfrom,type,payer) values(?,?,?,?,?,?)", account, "1970-01-01 00:00:00", 0, person.Platfrom, person.Type, "")
		log.Println("Add To Midpay:", person.Account)
		if err != nil {
			log.Printf("Add To Midpay Error:%v\n", err)
		}
	}
}

func (dbm *DBManager) InsertMidpayAccount(account string, platfrom string, personType string) {
	//把還沒加入的帳號加入
	_, err := dbm.DB.Exec("INSERT INTO midpay(account,useTime,status,platfrom,type,payer) values(?,?,?,?,?,?)", account, "1970-01-01 00:00:00", 0, platfrom, personType, "")
	log.Println("Add To Midpay:", account)
	if err != nil {
		log.Println("InsertMidpayAccount Insert Error:", err)
	}
}

func (dbm *DBManager) GetAllAccountsStatus() ([]AccountInfo, error) {
	var accountList []AccountInfo
	if dbManager == nil || dbManager.DB == nil {
		log.Println("Wait DB init.")
		return accountList, fmt.Errorf("GetAllAccountsStatus Error:Wait DB init")
	}
	rows, err := dbManager.DB.Query("SELECT id,amount,dayamount,daylimit,ip,nick_name,plaftfrom,real_name,state,type,qrurl,wechat_name FROM wechat WHERE state = ?", "NORMAL")
	if err != nil {
		log.Printf("GetAllAccountsStatus Error:%s\n", err)
		return accountList, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetAllAccountsStatus Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		var accountInfo AccountInfo
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("GetAllAccountsStatus Error:", err.Error())
		}
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "wechat_name":
				accountInfo.Account = value
			case "type":
				accountInfo.Type = value
			case "ip":
				accountInfo.IP = value
			case "nick_name":
				accountInfo.NickName = value
			case "real_name":
				accountInfo.RealName = value
			case "state":
				accountInfo.State = value
			case "qrurl":
				accountInfo.QrUrl = value
			case "plaftfrom":
				accountInfo.Platfrom, _ = strconv.Atoi(value)
			case "amount":
				accountInfo.Amount, _ = strconv.ParseFloat(value, 64)
			case "dayamount":
				accountInfo.DayAmount, _ = strconv.ParseFloat(value, 64)
			case "daylimit":
				accountInfo.DayLimit, _ = strconv.ParseFloat(value, 64)
			}
		}
		accountList = append(accountList, accountInfo)
	}
	return accountList, nil
}

func (dbm *DBManager) GetAccountsStatus(platfrom string, personType string) ([]AccountInfo, error) {
	var accountList []AccountInfo
	rows, err := dbm.DB.Query("SELECT id,amount,dayamount,daylimit,ip,nick_name,plaftfrom,real_name,state,type,wechat_name FROM wechat WHERE plaftfrom = ? and type = ? and state = ?", platfrom, personType, "NORMAL")
	log.Printf("GetAccountsStatus: Platfrom:%v AccountType:%s\n", platfrom, personType)
	if err != nil {
		log.Printf("GetAccountStatus Error:%s\n", err)
		return accountList, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetAccountStatus Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		var accountInfo AccountInfo
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("GetAccountStatus Error:", err.Error())
		}
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "wechat_name":
				accountInfo.Account = value
			case "type":
				accountInfo.Type = value
			case "ip":
				accountInfo.IP = value
			case "nick_name":
				accountInfo.NickName = value
			case "amount":
				accountInfo.Amount, _ = strconv.ParseFloat(value, 64)
			case "dayamount":
				accountInfo.DayAmount, _ = strconv.ParseFloat(value, 64)
			case "daylimit":
				accountInfo.DayLimit, _ = strconv.ParseFloat(value, 64)
			}
		}
		accountList = append(accountList, accountInfo)
	}
	return accountList, nil
}

func (dbm *DBManager) CheckAccountExists(account string) (AccountInfo, error) {
	var accountInfo AccountInfo
	rows, err := dbm.DB.Query("SELECT id,amount,dayamount,daylimit,ip,nick_name,plaftfrom,real_name,state,type,wechat_name FROM wechat WHERE wechat_name = ?", account)
	if err != nil {
		log.Printf("CheckAccountExists Error:%s\n", err)
		return accountInfo, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetPayAccounts Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("CheckAccountExists Error:", err.Error())
		}
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "wechat_name":
				accountInfo.Account = value
			case "plaftfrom":
				accountInfo.Platfrom, _ = strconv.Atoi(value)
			case "type":
				accountInfo.Type = value
			case "ip":
				accountInfo.IP = value
			case "nick_name":
				accountInfo.NickName = value
			case "amount":
				accountInfo.Amount, _ = strconv.ParseFloat(value, 64)
			case "dayamount":
				accountInfo.DayAmount, _ = strconv.ParseFloat(value, 64)
			case "daylimit":
				accountInfo.DayLimit, _ = strconv.ParseFloat(value, 64)
			}
		}
	}
	return accountInfo, nil
}
