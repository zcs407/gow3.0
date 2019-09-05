package main

import (
	"fmt"
	"log"
)

func (dbm *DBManager) InsertUserReport(report UserReport) error {
	result, err := dbm.DB.Exec("INSERT INTO user_report_list(befroe_money,change_money,create_time,create_user,now_money,platfrom,remark,type) values(?,?,?,?,?,?,?,?)",
		report.BeforeMoney, report.ChangeMoney, report.CreateTime, report.CreateUser, report.NowMoney, report.Platfrom, report.Remark, report.Type)
	if err != nil {
		log.Println("InsertUserReport Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertUserReport Update Affect Error:", err)
		return err
	}
	if affect == 0 {
		return fmt.Errorf("InsertUserReport Update failed")
	}
	return nil
}

func (dbm *DBManager) InsertReport(report UserReport) error {
	result, err := dbm.DB.Exec("INSERT INTO report_list(account,befroe_money,change_money,create_time,create_user,dest_bankcard,ip,nick_name,now_money,platfrom,remark,state,type,username,destip,account_type) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		report.Account, report.BeforeMoney, report.ChangeMoney, report.CreateTime, report.CreateUser, report.DestBankCard, report.IP, report.NickName,
		report.NowMoney, report.Platfrom, report.Remark, report.State, report.Type, report.UserName, report.DestIP, report.AccountType)
	if err != nil {
		log.Println("InsertReport Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("InsertReport Update Affect Error:", err)
		return err
	}
	if affect == 0 {
		return fmt.Errorf("InsertReport Update failed")
	}
	return nil
}
