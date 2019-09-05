package main

import (
	"database/sql"
	"fmt"
	"log"
)

func (dbm *DBManager) UpdateDepositCount(account string) error {
	result, err := dbm.DB.Exec("Update wechat set deposit_count=deposit_count+? where wechat_name=?", 1, account)
	if err != nil {
		log.Println("UpdateDepositCount Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateDepositCount Update Affect Error:", err)
		return err
	}
	log.Printf("[UpdateIncreaseDepositCount]:Account:%s updated:%v\n", account, affect)
	if affect == 0 {
		return fmt.Errorf("Can't update account deposit count. Account:%s\n", account)
	}
	return nil
}

func (dbm *DBManager) UpdateDepositCountReset(account string) error {
	result, err := dbm.DB.Exec("Update wechat set deposit_count=? where wechat_name=?", 0, account)
	if err != nil {
		log.Println("[UpdateDepositCountReset] Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateDepositCountReset Update Affect Error:", err)
		return err
	}
	log.Printf("[UpdateDepositCountReset]Account:%s Updated:%v\n", account, affect)
	if affect == 0 {
		log.Printf("Can't reset account deposit count. Account:%s affect is zero\n", account)
		return nil
	}
	return nil
}

func (dbm *DBManager) GetDepositCount(account string) (int, error) {
	rows, err := dbm.DB.Query("SELECT deposit_count FROM wechat WHERE wechat_name=?", account)
	if err != nil {
		log.Printf("GetDepositCount Error:%s\n", err)
		return CONFIGS.Deposit.Count, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetDepositCount Error:%s\n", err.Error())
		return CONFIGS.Deposit.Count, err
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	count := CONFIGS.Deposit.Count
	// Fetch rows
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Println("GetDepositCount Error:", err.Error())
		}
	}
	return count, nil
}
