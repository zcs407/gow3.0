package main
import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (dbm *DBManager) UpdateMidpay(account string) error {
	log.Println("UpdateMidpay:", account)
	result, err := dbm.DB.Exec("Update midpay set status=? where account=?", 0, account)
	if err != nil {
		log.Println("UpdateMidpay Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateMidpay Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update UpdateMidpay. Account:%s\n", account)
	}
	return nil
}

func (dbm *DBManager) UpdateMidpayStatusRecord(account string, ts string, payer string, amount string, status string) error {
	sec, _ := strconv.ParseInt(ts, 10, 64)
	if sec == 0 {
		return fmt.Errorf("Timestamp format incorrect. ts:%s\n", ts)
	}
	requestTime := time.Unix(sec, 0)
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateMidpayRecord Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	log.Println("[UpdateMidpayStatusRecord]:", account, payer, amount, status)
	result, err := queryTx.Exec("Update midpay_records set status=? where account=? and payer = ? and amount=? and request_time=? and status <> ?", status, account, payer, amount, requestTime, STATE_EXECUTED)
	if err != nil {
		log.Println("[UpdateMidpayStatusRecord] Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("[UpdateMidpayStatusRecord] Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update UpdateMidpayStatusRecord. Account:%s Payer:%s\n", account, payer)
	}

	return nil
}

func (dbm *DBManager) UpdateMidpayRecord(account string, payer string, amount float64, requestTime time.Time) error {
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateMidpayRecord Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	log.Println("UpdateMidpayRecord:", account, payer, requestTime)
	result, err := queryTx.Exec("Update midpay_records set status=?,excute_time=? where request_time=? and account=? and payer = ? and amount=?", STATE_EXECUTED, time.Now(), requestTime, account, payer, amount)
	if err != nil {
		log.Println("UpdateMidpayRecord Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateMidpayRecord Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update UpdateMidpayRecord. Account:%s Payer:%s RequestTime:%v\n", account, payer, requestTime)
	}

	return nil
}

func (dbm *DBManager) InsertMidpayRecord(account string, amount float64, payer string, insertTime time.Time, custom_sign string, platform string, api_type string, qrurl string, device string, state string) (bool, error) {
	result, err := dbm.DB.Exec("INSERT INTO midpay_records(account,request_time,payer,amount,status,custom_sign,platform,api_type,qrurl,device) values(?,?,?,?,?,?,?,?,?,?)", account, insertTime, payer, amount, state, custom_sign, platform, api_type, qrurl, device)
	if err != nil {
		log.Println("InsertMidpayRecord Update Error:", err)
		return false, err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertMidpayRecord Update Affect Error:", err)
		return false, err
	}
	if affect == 0 {
		return false, fmt.Errorf("InsertMidpayRecord Update failed")
	}
	return true, nil
}

func (dbm *DBManager) CheckMidpayExists(account string) (bool, error) {
	rows, err := dbm.DB.Query("SELECT * FROM midpay WHERE account= ?", account)
	if err != nil {
		log.Printf("CheckMidpayExists Error:%s\n", err)
		return true, err
	}
	defer rows.Close()
	exist := false

	for rows.Next() {
		exist = true
	}
	return exist, nil
}

func (dbm *DBManager) GetAvailableAccountBySQL(sqlCtx string, args ...interface{}) ([]PayPerson, error) {
	var personList []PayPerson
	rows, err := dbm.DB.Query(sqlCtx, args...)
	if err != nil {
		log.Printf("GetAvailableAccountBySQL Error:%s\n", err)
		return personList, err
	}
	defer rows.Close()
	for rows.Next() {
		var payPerson PayPerson
		if err := rows.Scan(&payPerson.Account, &payPerson.UseTime, &payPerson.Status, &payPerson.Platfrom, &payPerson.Type, &payPerson.Payer, &payPerson.Lock, &payPerson.LockTime); err != nil {
			log.Println("GetAvailableAccount Select Error:", err)
		}
		personList = append(personList, payPerson)
	}
	return personList, nil
}

func (dbm *DBManager) UpdateLockedAccount(account string) error {
	rows, err := dbm.DB.Query("SELECT account_lock_time FROM midpay WHERE account= ?", account)
	if err != nil {
		log.Printf("UpdateLockedAccount Error:%s\n", err)
		return err
	}
	defer rows.Close()
	var lockTIme time.Time
	for rows.Next() {
		if err := rows.Scan(&lockTIme); err != nil {
			log.Println("UpdateLockedAccount Select Error:", err)
		}
	}
	if time.Now().Unix()-lockTIme.Unix() > 24*60*60 {
		log.Printf("UpdateLockedAccount:%s Over 24 hour.\n", account)
		_, err := dbm.DB.Exec("Update midpay set account_lock='1',account_lock_time = ? where account=?", time.Now(), account)
		if err != nil {
			log.Println("UpdateLockedAccount Update Error:", err)
			return err
		}
	}
	return nil
}

func (dbm *DBManager) UpdateUnLockedAccount(account string) error {
	log.Printf("UpdateUnLockedAccount:%s\n", account)
	_, err := dbm.DB.Exec("Update midpay set account_lock='0' where account=?", account)
	if err != nil {
		log.Println("UpdateUnLockedAccount Update Error:", err)
		return err
	}

	return nil
}
func (dbm *DBManager) GetMidpayAccount(account string) (*PayPerson, error) {
	personList, err := dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE account=? ORDER BY useTime ASC", account)
	if err != nil {
		log.Println("GetMidpayAccount SQL error:", err)
		return nil, err
	}
	if len(personList) >= 1 {
		return &personList[0], nil
	}
	return nil, nil
}
func (dbm *DBManager) GetAvailableAccount(agent *Agent, payer string, personType string, payAmount float64) (*Person, error) {
	personSetting := GetPersonSetting(agent.Name)
	platfrom := fmt.Sprintf("%d", agent.Id)
	var personList []PayPerson
	var sqlError error
	log.Println("personSetting:", personSetting)
	if personSetting.SFPay {
		personList, sqlError = dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE (platfrom = ? and type = ?) AND account_lock='0' ORDER BY useTime ASC", platfrom, personType)
		if sqlError != nil {
			log.Println("GetAvailableAccount SQL error:", sqlError)
		}
	} else {
		personList, sqlError = dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE (platfrom = ? AND type = ?) AND (status=0 or useTime < ?) AND account_lock = '0' ORDER BY useTime ASC",
			platfrom, personType, time.Now().Add(-time.Duration(personSetting.AccountLockTime)*time.Second))
		if sqlError != nil {
			log.Println("GetAvailableAccount SQL error:", sqlError)
		}
	}
	update := false
	if len(personList) > 0 {
		update = true
	}
	var person *Person
	var useAccountInfo AccountInfo

	accountList, err := dbm.GetAccountsStatus(platfrom, personType)
	if err != nil {
		log.Println("GetAccountsStatus Error:", err)
	}
	log.Printf("[GetAvailableAccount]Update:%v PersonList:%v\n AccountList:%v\n", update, personList, accountList)
	if update {
		match := false
		for _, payPerson := range personList {
			match = false
			for _, accountInfo := range accountList {
				if accountInfo.Account == payPerson.Account {
					if accountInfo.Amount+payAmount < accountInfo.DayLimit {
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
			log.Println("[GetAvailableAccount]Find Account:", useAccountInfo)
			if v, ok := dbm.PersonList[useAccountInfo.Account]; ok {
				person = v
			} else {
				person, err = dbm.GetPayAccount(useAccountInfo.Account)
				if err != nil {
					return nil, fmt.Errorf("Person not exists in cache.")
				}
			}
		} else {
			log.Println("[GetAvailableAccount] can't found person.")
		}
	} else {
		log.Println("GetAvailableAccount can't found person.")
	}
	return person, nil
}

func (dbm *DBManager) UpdateMidpayInfo(account string, platform string) error {
	_, err := dbm.DB.Exec("Update midpay set platfrom=? where account=?", platform, account)
	if err != nil {
		log.Println("UpdateMidpayInfo Update Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) UpdateMidpayTime(payer string, account string) error {
	_, err := dbm.DB.Exec("Update midpay set useTime=?,status=1,payer=? where account=?", time.Now(), payer, account)
	if err != nil {
		log.Println("UpdateMidpayTime Update Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) GetAvailableUnLockedAccount(agent *Agent, payer string, personType string, payAmount float64) (*Person, error) {
	personSetting := GetPersonSetting(agent.Name)
	platfrom := fmt.Sprintf("%d", agent.Id)
	var personList []PayPerson
	var sqlError error
	personList, sqlError = dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE (platfrom = ? AND type = ?) AND (status=0 or useTime < ?) AND account_lock_time < ? ORDER BY useTime ASC",
		platfrom, personType, time.Now().Add(-time.Duration(personSetting.AccountLockTime)*time.Second), time.Now().Add(-24*time.Hour))
	if sqlError != nil {
		log.Println("GetAvailableUnLockedAccount SQL error:", sqlError)
	}

	update := false
	if len(personList) > 0 {
		update = true
	}
	var person *Person
	var useAccountInfo AccountInfo

	accountList, err := dbm.GetAccountsStatus(platfrom, personType)
	if err != nil {
		log.Println("GetAvailableUnLockedAccount Error:", err)
	}
	log.Printf("GetAvailableUnLockedAccount PersonList:%d AccountList:%d\n", len(personList), len(accountList))
	if update {
		match := false
		for _, payPerson := range personList {
			match = false
			for _, accountInfo := range accountList {
				if accountInfo.Account == payPerson.Account {
					if accountInfo.Amount+payAmount < accountInfo.DayLimit {
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
			log.Println("GetAvailableUnLockedAccount Find:", useAccountInfo)
			if v, ok := dbm.PersonList[useAccountInfo.Account]; ok {
				person = v
			} else {
				person, err = dbm.GetPayAccount(useAccountInfo.Account)
				if err != nil {
					return nil, fmt.Errorf("Person not exists in cache.")
				}
			}
		} else {
			log.Println("GetAvailableUnLockedAccount can't found person.")
		}
	} else {
		log.Println("GetAvailableUnLockedAccount can't found person.")
	}
	return person, nil
}

func (dbm *DBManager) GetAvailableLockedAccount(agent *Agent, payer string, personType string, payAmount float64) (*Person, error) {
	personSetting := GetPersonSetting(agent.Name)
	platfrom := fmt.Sprintf("%d", agent.Id)
	var personList []PayPerson
	var sqlError error
	if CONFIGS.HB {
		personList, sqlError = dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE (platfrom = ? AND type = ?) AND (status=0 or useTime < ?) ORDER BY useTime ASC",
			platfrom, personType, time.Now().Add(-time.Duration(personSetting.AccountLockTime)*time.Second))
	} else {
		personList, sqlError = dbm.GetAvailableAccountBySQL("SELECT * FROM midpay WHERE (platfrom = ? AND type = ?) AND (status=0 or useTime < ?) AND account_lock='1' ORDER BY useTime ASC",
			platfrom, personType, time.Now().Add(-time.Duration(personSetting.AccountLockTime)*time.Second))
	}
	if sqlError != nil {
		log.Println("GetAvailableLockedAccount SQL error:", sqlError)
	}

	update := false
	if len(personList) > 0 {
		update = true
	}
	var person *Person
	var useAccountInfo AccountInfo

	accountList, err := dbm.GetAccountsStatus(platfrom, personType)
	if err != nil {
		log.Println("GetAvailableLockedAccount Error:", err)
	}
	log.Printf("GetAvailableLockedAccount Midpay Platform:%v Payer:%s PersonList:%d AccountList:%d\n", platfrom, payer, len(personList), len(accountList))
	if update {
		match := false
		for _, payPerson := range personList {
			match = false
			for _, accountInfo := range accountList {
				if accountInfo.Account == payPerson.Account {
					if accountInfo.Amount+payAmount < accountInfo.DayLimit {
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
			log.Println("GetAvailableLockedAccount Find:", useAccountInfo)
			person, err = dbm.GetPayAccount(useAccountInfo.Account)
			if err != nil {
				return nil, fmt.Errorf("Person not exists in cache.")
			}
		} else {
			log.Println("GetAvailableLockedAccount can't found person.")
		}
	} else {
		log.Println("GetAvailableLockedAccount can't found person.")
	}
	return person, nil
}

func CheckMidpayAccount() {
	if dbManager == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("CheckMidpayAccount Error Recover System:", r)
		}
	}()
	accountList, err := dbManager.GetAllAccountsStatus()
	if err != nil {
		log.Println("GetAllAccountsStatus Error:", err)
		return
	}
	//Remove disable account
	dbManager.CheckDisableAccount(accountList)
	//Add Normal Account
	dbManager.CheckEnableAccount(accountList)
}

func (dbm *DBManager) CheckDisableAccount(accountList []AccountInfo) {
	//Check Disable Account
	find := false
	for _, payPerson := range dbm.PersonList {
		find = false
		for _, accountInfo := range accountList {
			if accountInfo.Account == payPerson.Account {
				find = true
				break
			}
		}
		if !find {
			exist, _ := dbm.CheckMidpayExists(payPerson.Account)
			log.Printf("Delete account:%v exists:%v\n", payPerson.Account, exist)
			if exist {
				dbm.DeleteAccount(payPerson.Account, "The Account had disabled.")
			} else {
				log.Println("Delete account failed. account not exsits.", payPerson.Account)
			}
			log.Println("Delete account from memory.", payPerson.Account)
			delete(dbm.PersonList, payPerson.Account)
			go dbm.CheckDisableSFAccount(payPerson)
		}
	}
}

func (dbm *DBManager) CheckEnableAccount(accountList []AccountInfo) {
	//Check Enabled Account
	log.Printf("CheckEnableAccount person size:%d account size:%d\n", len(dbm.PersonList), len(accountList))
	find := false
	for _, accountInfo := range accountList {
		find = false
		for _, payPerson := range dbm.PersonList {
			if accountInfo.Account == payPerson.Account && fmt.Sprintf("%d", accountInfo.Platfrom) == payPerson.Platfrom {
				find = true
				break
			}
		}
		if !find {
			var person Person
			person.Account = accountInfo.Account
			person.State = accountInfo.State
			person.NickName = accountInfo.NickName
			person.QrUrl = accountInfo.QrUrl
			person.RealName = accountInfo.RealName
			person.Platfrom = fmt.Sprintf("%d", accountInfo.Platfrom)
			person.Type = accountInfo.Type
			dbm.PersonList[accountInfo.Account] = &person
			exist, _ := dbm.CheckMidpayExists(accountInfo.Account)
			if !exist {
				log.Printf("The Account had enabled. Account:%s Platfrom:%d Type:%s\n", accountInfo.Account, accountInfo.Platfrom, accountInfo.Type)
				dbm.InsertMidpayAccount(accountInfo.Account, fmt.Sprintf("%d", accountInfo.Platfrom), accountInfo.Type)

				log.Printf("Add Account to memory. Account:%s Person:%v\n", accountInfo.Account, person)
			} else {
				dbm.UpdateMidpayInfo(accountInfo.Account, fmt.Sprintf("%d", accountInfo.Platfrom))
				log.Printf("The Account had exsits. Account:%s Platfrom:%d Type:%s\n", accountInfo.Account, accountInfo.Platfrom, accountInfo.Type)
			}
			go dbm.CheckEnableSFAccount(accountInfo)
		}
	}
}

func (dbm *DBManager) GetMidPayRecord(account string, amount float64, payer string, requestTime time.Time) ([]MidPayRecord, error) {
	log.Println("GetMidPayRecord", account, amount, payer, requestTime)
	return dbm.CheckDepoistAccountByNormalSql("SELECT account,request_time,payer,amount,status,custom_sign,platform,api_type,qrurl FROM midpay_records WHERE account = ? AND amount = ? AND status <> ? AND payer = ? AND request_time = ? Order BY request_time DESC limit 1", account, amount, STATE_EXECUTED, payer, requestTime)
}

func (dbm *DBManager) CheckDepoistAccountRequest(personSetting CustomPerson, account string, deposit *DepositRecord) ([]MidPayRecord, error) {
	log.Println("CheckDepoistAccountRequest", personSetting, account, deposit.Amount, time.Now().Unix())
	t, err := time.Parse("2006-01-02 15:04:05", deposit.TransferTime)
	if err != nil {
		t = time.Now()
		log.Println("[CheckDepoistAccountRequest]Parse Time Error:", err, deposit.TransferTime)
	}
	return dbm.CheckDepoistAccountByNormalSql("SELECT account,request_time,payer,amount,status,custom_sign,platform,api_type,qrurl FROM midpay_records WHERE request_time BETWEEN ? AND ? AND account = ? AND amount = ? AND status <> ? Order BY request_time DESC", t.Add(time.Duration(-personSetting.AccountLockTime)*time.Second), t, account, deposit.Amount, STATE_EXECUTED)
}

func (dbm *DBManager) CheckDepoistAccountOverTimeRequest(personSetting CustomPerson, account string, amount float64, depositTime time.Time) ([]MidPayRecord, error) {
	log.Println("[CheckDepoistAccountOverTimeRequest]:", personSetting, account, amount, depositTime.Add(time.Duration(-personSetting.AccountLockTime)*time.Second), depositTime)
	return dbm.CheckDepoistAccountByNormalSql("SELECT account,request_time,payer,amount,status,custom_sign,platform,api_type,qrurl FROM midpay_records WHERE request_time BETWEEN ? AND ? AND account = ? AND amount = ? AND status<> ? Order BY request_time DESC", depositTime.Add(time.Duration(-personSetting.AccountLockTime)*time.Second), depositTime, account, amount, STATE_EXECUTED)
}

func (dbm *DBManager) CheckDepoistAccountByNormalSql(sqlCtx string, args ...interface{}) ([]MidPayRecord, error) {
	var payer_records []MidPayRecord
	//log.Println("CheckDepoistAccountBySql Query:", sqlCtx, args)
	rows, err := dbm.DB.Query(sqlCtx, args...)
	if err != nil {
		log.Printf("CheckDepoistAccountBySql Error:%s\n", err)
		return payer_records, err
	}
	defer rows.Close()
	for rows.Next() {
		var payer MidPayRecord
		if err := rows.Scan(&payer.Account, &payer.RequestTime, &payer.Payer, &payer.Amount, &payer.Status, &payer.CustomSign, &payer.Platform, &payer.APIType, &payer.QrUrl); err != nil {
			log.Println("CheckDepoistAccountBySql Select Error:", err)
			return payer_records, err
		}
		log.Println("CheckDepoistAccountBySql Find Record:", payer)
		payer_records = append(payer_records, payer)
	}
	return payer_records, nil
}

func (dbm *DBManager) CheckDepoistAccountBySql(sqlCtx string, args ...interface{}) ([]MidPayRecord, error) {
	var payer_records []MidPayRecord
	//log.Println("CheckDepoistAccountBySql Query:", sqlCtx, args)
	rows, err := dbm.DB.Query(sqlCtx, args...)
	if err != nil {
		log.Printf("CheckDepoistAccountBySql Error:%s\n", err)
		return payer_records, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("CheckDepoistAccountBySql Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	// Fetch rows
	for rows.Next() {
		var payer MidPayRecord
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("CheckDepoistAccountBySql Error:", err.Error())
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
			case "account":
				payer.Account = value
			case "payer":
				payer.Payer = value
			case "custom_sign":
				payer.CustomSign = value
			case "api_type":
				payer.APIType = value
			case "request_time":
				payer.RequestTime, err = time.Parse("2006-01-02T15:04:05+08:00", value)
				if err != nil {
					log.Println("time Parse Error:", err)
				}
			case "status":
				payer.Status = value
			case "platform":
				payer.Platform = value
			case "amount":
				payer.Amount, _ = strconv.ParseFloat(value, 64)
			}
		}
		payer_records = append(payer_records, payer)
	}
	return payer_records, nil
}

func (dbm *DBManager) CheckPersonAccountRequest(Person CustomPerson, payer string, amount float64) ([]MidPayRecordObject, error) {
	var accounts []MidPayRecordObject
	rows, err := dbm.DB.Query("SELECT account,request_time,amount,api_type,qrurl FROM midpay_records WHERE request_time >= ? AND payer = ? AND amount = ? AND status <> ?", time.Now().Add(time.Duration(-Person.AccountLockTime)*time.Second), payer, amount, STATE_EXECUTED)
	if err != nil {
		log.Printf("CheckPersonAccountRequest Error:%s\n", err)
		return accounts, err
	}
	defer rows.Close()

	if rows.Next() {
		var obj MidPayRecordObject
		if err := rows.Scan(&obj.Account, &obj.RequestTime, &obj.Amount, &obj.Mode, &obj.QrUrl); err != nil {
			log.Println("CheckPersonAccountRequest Select Error:", err)
			return accounts, err
		}
		accounts = append(accounts, obj)
	}
	return accounts, nil
}

func (dbm *DBManager) DeleteAccount(account string, reason string) {
	//移除不能用的帳號
	tx, err := dbm.DB.Begin()
	if err != nil {
		log.Printf("DeleteAccount Error:%v\n", err)
		tx.Rollback()
	}
	tx.Exec("DELETE FROM midpay WHERE account=?", account)
	log.Printf("Delete Account from midpay:%s Reason:reason:%s\n", account, reason)
	defer tx.Commit()
}
