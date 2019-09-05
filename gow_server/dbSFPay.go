package main

import (
	"fmt"
	"log"
	"time"
)

func (dbm *DBManager) UpdateSFPay(account string, platform int, amount float64, note string, qrurl string) error {
	log.Println("UpdateSFPay:", account, amount)
	result, err := dbm.DB.Exec("Update sfpay set qrUrl=?,platform=?, createTime=?,note=? where account=? and amount=?", qrurl, fmt.Sprintf("%d", platform), time.Now(), note, account, amount)
	if err != nil {
		log.Println("UpdateSFPay Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateSFPay Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update UpdateSFPay. Account:%s\n", account)
	}
	return nil
}

func (dbm *DBManager) UpdateSFPayStatus(account string, amount float64) error {
	log.Println("UpdateSFPayStatus:", account, amount)
	sfAccount, err := dbm.GetSFPayAccount(account, amount)
	if err != nil {
		log.Println("UpdateSFPayStatus GetSFPayAccount Error:", err, account, amount)
		return err
	}
	if sfAccount == nil {
		log.Println("UpdateSFPayStatus GetSFPayAccount Error: Can't find account", account, amount)
		return fmt.Errorf("UpdateSFPayStatus GetSFPayAccount Error: Can't find account:%s Amount:%f", account, amount)
	}
	if sfAccount.Status == 2 {
		log.Println("UpdateSFPayStatus SFAccount had Disaled.", account, amount)
		return fmt.Errorf("UpdateSFPayStatus SFAccount had Disaled.[%s]:%f", account, amount)
	}
	result, err := dbm.DB.Exec("Update sfpay set status='0' where account=? and amount=?", account, amount)
	if err != nil {
		log.Println("UpdateSFPayStatus Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateSFPayStatus Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update UpdateSFPay. Account:%s\n", account)
	}
	return nil
}

func (dbm *DBManager) InsertSFPay(accountInfo AccountInfo, amount float64, note string, qrurl string) (bool, error) {
	result, err := dbm.DB.Exec("INSERT INTO sfpay(account,useTime,createTime,status,amount,note,qrUrl,type,payer,platform) values(?,?,?,?,?,?,?,?,?,?)",
		accountInfo.Account, "1970-01-01 00:00:00", time.Now(), false, amount, note, qrurl, accountInfo.Type, "", fmt.Sprintf("%d", accountInfo.Platfrom))
	if err != nil {
		log.Println("InsertSFPay Error:", err)
		return false, err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertSFPay Affect Error:", err)
		return false, err
	}
	if affect == 0 {
		return false, fmt.Errorf("InsertSFPay failed")
	}
	return true, nil
}

func (dbm *DBManager) CheckSFPayNoteExists(account string, amount float64) (bool, error) {
	rows, err := dbm.DB.Query("SELECT * FROM sfpay WHERE account = ? AND amount = ?", account, amount)
	if err != nil {
		log.Printf("CheckSFPayExists Error:%s\n", err)
		return true, err
	}
	defer rows.Close()
	exist := false

	for rows.Next() {
		exist = true
	}
	return exist, nil
}

func (dbm *DBManager) CheckSFPayExists(account string) (bool, error) {
	rows, err := dbm.DB.Query("SELECT * FROM sfpay WHERE account = ?", account)
	if err != nil {
		log.Printf("CheckSFPayExists Error:%s\n", err)
		return true, err
	}
	defer rows.Close()
	exist := false

	for rows.Next() {
		exist = true
	}
	return exist, nil
}

func (dbm *DBManager) GetAvailableSFAccountBySQL(sqlCtx string, args ...interface{}) ([]PayPerson, error) {
	var personList []PayPerson
	rows, err := dbm.DB.Query(sqlCtx, args...)
	if err != nil {
		log.Printf("GetAvailableSFAccountBySQL Error:%s\n", err)
		return personList, err
	}
	defer rows.Close()
	for rows.Next() {
		//fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(account varchar(100),useTime datetime,createTime datetime, status boolean,amount int(8),note varchar(255),qrUrl varchar(255),type varchar(5),payer varchar(255),PRIMARY KEY (note))",
		var payPerson PayPerson
		if err := rows.Scan(&payPerson.Account, &payPerson.UseTime, &payPerson.CreateTime, &payPerson.Status, &payPerson.Amount, &payPerson.Note, &payPerson.QrUrl, &payPerson.Type, &payPerson.Payer, &payPerson.Platfrom); err != nil {
			log.Println("GetAvailableSFAccountBySQL Select Error:", err)
		}
		personList = append(personList, payPerson)
	}
	return personList, nil
}

func (dbm *DBManager) GetSFPayAccount(account string, amount float64) (*PayPerson, error) {
	personList, err := dbm.GetAvailableSFAccountBySQL("SELECT * FROM sfpay WHERE account=? and amount = ? ORDER BY useTime ASC", account, amount)
	if err != nil {
		log.Println("GetSFPayAccount SQL error:", err)
		return nil, err
	}
	if len(personList) >= 1 {
		return &personList[0], nil
	}
	return nil, nil
}

func (dbm *DBManager) UpdateSFPayInfo(account string, note string, payer string, platform string) error {
	_, err := dbm.DB.Exec("Update sfpay set platfrom=?,payer=? where account=? and note=?", platform, payer, account, note)
	if err != nil {
		log.Println("UpdateSFPayInfo Update Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) UpdateSFPayTime(payer string, account string, amount float64, note string) error {
	_, err := dbm.DB.Exec("Update sfpay set useTime=?,status=1,payer=? where account=? and amount=? and note=?", time.Now(), payer, account, amount, note)
	if err != nil {
		log.Println("UpdateSFPayTime Update Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) GetAvailableSFAccount(agent *Agent, payer string, personType string, payAmount float64) (*Person, error) {
	personSetting := GetPersonSetting(agent.Name)
	platfrom := fmt.Sprintf("%d", agent.Id)
	var personList []PayPerson
	var sqlError error
	personList, sqlError = dbm.GetAvailableSFAccountBySQL("SELECT * FROM sfpay WHERE type = ? AND (status=0 or useTime < ?) AND status <> 2 AND amount = ? AND platform = ? ORDER BY useTime ASC",
		personType, time.Now().Add(-time.Duration(personSetting.AccountLockTime)*time.Second), payAmount, platfrom)

	if sqlError != nil {
		log.Println("GetAvailableSFAccount SQL error:", sqlError)
	}

	update := false
	if len(personList) > 0 {
		update = true
	}
	var person *Person
	var useAccountInfo AccountInfo

	accountList, err := dbm.GetAllAccountsStatus()
	if err != nil {
		log.Println("GetAvailableSFAccount Error:", err)
	}
	log.Printf("GetAvailableSFAccount SFPAY Platform:%v Payer:%s PersonList:%d AccountList:%d\n", platfrom, payer, len(personList), len(accountList))
	if update {
		match := false
		qrUrl := ""
		Note := ""
		for _, payPerson := range personList {
			match = false
			for _, accountInfo := range accountList {
				if accountInfo.Account == payPerson.Account {
					if accountInfo.Amount+payAmount < accountInfo.DayLimit {
						qrUrl = payPerson.QrUrl
						Note = payPerson.Note
						useAccountInfo = accountInfo
						match = true
						break
					}
				}
			}
			if match {
				break
			}
		}
		if match {
			log.Println("GetAvailableSFAccount Find:", useAccountInfo)
			if v, ok := dbm.PersonList[useAccountInfo.Account]; ok {
				person = v
			} else {
				person, err = dbm.GetPayAccount(useAccountInfo.Account)
				if err != nil {
					return nil, fmt.Errorf("Person not exists in cache.")
				}
			}
			person.Note = Note
			person.QrUrl = qrUrl
			log.Println("GetAvailableSFAccount Person Info:", person)
		} else {
			log.Println("GetAvailableSFAccount can't found person. no match account.")
		}
	} else {
		log.Println("GetAvailableSFAccount can't found person.")
	}
	return person, nil
}

func (dbm *DBManager) CheckDisableSFAccount(payPerson *Person) {
	exist, _ := dbm.CheckSFPayExists(payPerson.Account)
	log.Printf("Delete account:%v exists:%v\n", payPerson.Account, exist)
	if exist {
		dbm.DisableSFAccount(payPerson.Account, "The Account had disabled.")
	} else {
		log.Println("Delete account failed. account not exsits.", payPerson.Account)
	}
	log.Println("Delete account from memory.", payPerson.Account)
}

func (dbm *DBManager) CheckEnableSFAccount(accountInfo AccountInfo) {
	//Check Enabled Account
	exist, _ := dbm.CheckSFPayExists(accountInfo.Account)
	if !exist {
		log.Printf("SFAccount The Account had enabled. Account:%s Platfrom:%d Type:%s\n", accountInfo.Account, accountInfo.Platfrom, accountInfo.Type)
		GenSingleQrCode(accountInfo.Account)
		log.Printf("SFAccount Add Account to memory. Account:%s\n", accountInfo.Account)
	} else {
		dbm.EnableSFAccount(accountInfo.Account, fmt.Sprintf("%d", accountInfo.Platfrom))
		log.Printf("SFAccount The Account had exsits. Account:%s Platfrom:%d Type:%s\n", accountInfo.Account, accountInfo.Platfrom, accountInfo.Type)
	}
}

func (dbm *DBManager) DisableSFAccount(account string, reason string) {
	//移除不能用的帳號
	tx, err := dbm.DB.Begin()
	if err != nil {
		log.Printf("DisableSFAccount Error:%v\n", err)
		tx.Rollback()
	}
	tx.Exec("UPDATE sfpay SET status='2' WHERE account=?", account)
	log.Printf("Delete Account from sfpay:%s Reason:reason:%s\n", account, reason)
	defer tx.Commit()
}

func (dbm *DBManager) EnableSFAccount(account string, platform string) {
	//移除不能用的帳號
	tx, err := dbm.DB.Begin()
	if err != nil {
		log.Printf("EnableSFAccount Error:%v\n", err)
		tx.Rollback()
	}
	tx.Exec("UPDATE sfpay SET status='0',platform=? WHERE account=?", platform, account)
	log.Printf("Enable Account from sfpay:%s\n", account)
	defer tx.Commit()
}
