package main

import (
	"fmt"
	"log"
	"time"
)

func (dbm *DBManager) GetPayer(account string, payer string, amount float64) (*MidPayRecord, error) {
	var payer_record MidPayRecord
	log.Println("GetPayer", payer, amount)
	rows, err := dbm.DB.Query("SELECT account,request_time,payer,amount,status,custom_sign FROM payer_records WHERE account =? and payer=? and amount=? order by request_time DESC", account, payer, amount)

	if err != nil {
		log.Printf("GetPayer Error:%s\n", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var payer MidPayRecord
		if err := rows.Scan(&payer.Account, &payer.RequestTime, &payer.Payer, &payer.Amount, &payer.Status, &payer.CustomSign); err != nil {
			log.Println("GetPayer Select Error:", err)
			return &payer, err
		}
		payer_record = payer
	}
	return &payer_record, nil
}

func (dbm *DBManager) GetPayerInfo(account string, amount string, sign string) (MidPayRecord, error) {
	var payer_record MidPayRecord
	log.Println("getPayerInfo", account, amount, sign)
	//不使用固定碼
	if len(sign) > 20 {
		sign = sign[:20]
	}
	if sign == "商品" {
		log.Println("GetPayerInfo: sign incorrect.")
		return payer_record, fmt.Errorf("sign incorrect")
	}
	rows, err := dbm.DB.Query("SELECT account,request_time,payer,amount,status,custom_sign FROM payer_records WHERE account=? and amount=? and sign = ?", account, amount, sign)

	if err != nil {
		log.Printf("getPayerInfo Error:%s\n", err)
		return payer_record, err
	}
	defer rows.Close()
	for rows.Next() {
		var payer MidPayRecord
		if err := rows.Scan(&payer.Account, &payer.RequestTime, &payer.Payer, &payer.Amount, &payer.Status, &payer.CustomSign); err != nil {
			log.Println("getPayerInfo Select Error:", err)
			return payer_record, err
		}
		payer_record = payer
	}
	return payer_record, nil
}

func (dbm *DBManager) InsertPayerRecord(account string, amount float64, payer string, insertTime time.Time, sign string, custom_sign string) (bool, error) {
	if len(sign) > 20 {
		sign = sign[:20]
	}
	result, err := dbm.DB.Exec("INSERT INTO payer_records(account,request_time,payer,amount,status,sign,custom_sign) values(?,?,?,?,?,?,?)", account, insertTime, payer, amount, STATE_DEPOSIT_CHECKING, sign, custom_sign)
	if err != nil {
		log.Println("InsertPayerRecord Update Error:", err)
		return false, err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertPayerRecord Update Affect Error:", err)
		return false, err
	}
	if affect == 0 {
		return false, fmt.Errorf("InsertPayerRecord Update failed")
	}
	return true, nil
}
