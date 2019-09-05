package main

import (
	"fmt"
	"log"
	"time"
)

func (dbm *DBManager) GetDepositFailedStatus() ([]DepositRecord, error) {
	scanTime := time.Now().Add(-time.Duration(CONFIGS.Deposit.Hours) * time.Hour)
	var records []DepositRecord
	if dbManager == nil {
		log.Println("GetDepositFailedStatus:Wait DB Init")
		return records, fmt.Errorf("wait db init")
	}
	rows, err := dbManager.DB.Query("SELECT * FROM deposit WHERE state = ? and tran_time >= ? and times < ?", STATE_PENDING, scanTime, CONFIGS.Deposit.Times)
	if err != nil {
		log.Printf("GetDepositFailedStatus Error:%s\n", err)
		return records, err
	}
	defer rows.Close()

	if rows.Next() {
		var record DepositRecord
		//id,amount,bill_no,creat_time,create_user,deposit_number,excute_time,ip,nick_name,note,pay_account,platfrom,state,times,tran_time,tranfee,transfer_time,wechat_name,user_remark,pay_type,call_url,sign
		err = rows.Scan(&record.Id, &record.Amount, &record.BillNo, &record.CreateTime, &record.CreateUser, &record.DepositNumber,
			&record.ExcuteTime, &record.IP, &record.NickName, &record.Note, &record.PayAccount, &record.Platfrom, &record.State, &record.Times,
			&record.TranTime, &record.TranFee, &record.TransferTime, &record.WechatName, &record.Remark, &record.PayType, &record.CallbackUrl, &record.Sign)
		if err != nil {
			log.Println("GetDepositFailedStatus Error:", err.Error())
		}
		records = append(records, record)
	}
	if len(records) > 0 {
		log.Printf("GetDepositFailedStatus:Now: %s Start Time:%s", time.Now(), scanTime)
		log.Println("GetDepositFailedStatus have fail records:", len(records))
	}
	return records, nil
}

func (dbm *DBManager) UpdateDepositState(depositNumber string, state string) error {
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateDepositState Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	result, err := queryTx.Exec("Update deposit set state=? where deposit_number=?", state, depositNumber)
	if err != nil {
		log.Println("UpdateDepositState Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateDepositState Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update deposit state. depositNumber:%s state:%s\n", depositNumber, state)
	}
	return nil
}

func (dbm *DBManager) UpdateDepositStateFailed(depositNumber string, state string, times int) error {
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateDepositState Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	result, err := queryTx.Exec("Update deposit set state=?, times=? where deposit_number=?", state, times, depositNumber)
	if err != nil {
		log.Println("UpdateDepositState Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateDepositState Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update deposit state. depositNumber:%s state:%s\n", depositNumber, state)
	}
	return nil
}

func (dbm *DBManager) GetDepositStatus(payer string, platfrom string) (*DepositRecord, error) {
	log.Println("GetDepositStatus:", payer, platfrom)
	rows, err := dbm.DB.Query("SELECT * FROM deposit WHERE user_remark = ? and platfrom = ? order by creat_time DESC limit 1", payer, platfrom)
	if err != nil {
		log.Printf("GetDeposit Error:%s\n", err)
		return nil, err
	}
	defer rows.Close()
	var record DepositRecord
	find := false
	if rows.Next() {
		find = true
		//id,amount,bill_no,creat_time,create_user,deposit_number,excute_time,ip,nick_name,note,pay_account,platfrom,state,times,tran_time,tranfee,transfer_time,wechat_name,user_remark,pay_type,call_url,sign
		err = rows.Scan(&record.Id, &record.Amount, &record.BillNo, &record.CreateTime, &record.CreateUser, &record.DepositNumber,
			&record.ExcuteTime, &record.IP, &record.NickName, &record.Note, &record.PayAccount, &record.Platfrom, &record.State, &record.Times,
			&record.TranTime, &record.TranFee, &record.TransferTime, &record.WechatName, &record.Remark, &record.PayType, &record.CallbackUrl, &record.Sign)
		if err != nil {
			log.Println("GetDeposit Error:", err.Error())
		}
	}
	if !find {
		return nil, fmt.Errorf("Can't find deposit record.")
	}
	return &record, nil
}

func (dbm *DBManager) GetDeposit(depositNumber string, platfrom string) (*DepositRecord, error) {
	log.Println("GetDeposit:", depositNumber, platfrom)
	rows, err := dbm.DB.Query("SELECT * FROM deposit WHERE deposit_number = ? and platfrom = ?", depositNumber, platfrom)
	if err != nil {
		log.Printf("GetDeposit Error:%s\n", err)
		return nil, err
	}
	defer rows.Close()
	var record DepositRecord
	find := false
	if rows.Next() {
		find = true
		//id,amount,bill_no,creat_time,create_user,deposit_number,excute_time,ip,nick_name,note,pay_account,platfrom,state,times,tran_time,tranfee,transfer_time,wechat_name,user_remark,pay_type,call_url,sign
		err = rows.Scan(&record.Id, &record.Amount, &record.BillNo, &record.CreateTime, &record.CreateUser, &record.DepositNumber,
			&record.ExcuteTime, &record.IP, &record.NickName, &record.Note, &record.PayAccount, &record.Platfrom, &record.State, &record.Times,
			&record.TranTime, &record.TranFee, &record.TransferTime, &record.WechatName, &record.Remark, &record.PayType, &record.CallbackUrl, &record.Sign)
		if err != nil {
			log.Println("GetDeposit Error:", err.Error())
		}
	}
	if !find {
		return nil, fmt.Errorf("Can't find deposit record.")
	}
	return &record, nil
}

func (dbm *DBManager) CheckDepositNumberExists(number string) (bool, error) {
	rows, err := dbm.DB.Query("SELECT deposit_number FROM deposit WHERE deposit_number = ?", number)
	if err != nil {
		log.Printf("CheckDepositNumberExists Error:%s\n", err)
		return false, err
	}
	defer rows.Close()
	exist := false
	if rows.Next() {
		exist = true
	}
	return exist, nil
}

func (dbm *DBManager) InsertDepositRecord(record *DepositRecord) (bool, error) {
	result, err := dbm.DB.Exec("INSERT INTO deposit(amount,bill_no,creat_time,create_user,deposit_number,excute_time,ip,nick_name,note,pay_account,platfrom,state,times,tran_time,tranfee,transfer_time,wechat_name,user_remark,pay_type,call_url,sign) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		record.Amount, record.BillNo, record.CreateTime, record.CreateUser, record.DepositNumber, record.ExcuteTime, record.IP, record.NickName, record.Note,
		record.PayAccount, record.Platfrom, record.State, record.Times, record.TranTime, record.TranFee, record.TransferTime, record.WechatName, record.Remark,
		record.PayType, record.CallbackUrl, record.Sign)
	if err != nil {
		log.Println("InsertDepositRecord Update Error:", err)
		return false, err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertDepositRecord Update Affect Error:", err)
		return false, err
	}
	if affect == 0 {
		return false, fmt.Errorf("InsertDepositRecord Update failed")
	}
	return true, nil
}
