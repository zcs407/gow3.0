package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

func (dbm *DBManager) UpdateAgentAmount(account string, amount float64) error {
	queryTx, err := dbm.DB.Begin()
	if err != nil {
		log.Println("UpdateAgentAmount Query Error:", err)
		return err
	}
	defer queryTx.Commit()
	log.Printf("UpdateAgentAmount:%s Amount:%f\n", account, amount)
	result, err := queryTx.Exec("Update agent set amount=amount+? where name=?", amount, account)
	if err != nil {
		log.Println("UpdateAgentAmount Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateAgentAmount Update Affect Error:", err)
		return err
	}

	if affect == 0 {
		return fmt.Errorf("Can't update account amount. Account:%s Amount:%f\n", account, amount)
	}
	return nil
}

func (dbm *DBManager) GetAgentByName(name string) (*Agent, error) {
	return dbm.getAgentBySQL("SELECT id,agent_name,amount,bank_card,bank_card_name,bank_card_type,callbackurl,creater_user,last_login_ip,lock_money,name,payfee,remark,sign,state,pay_safe,pay_secret,payqr,payword FROM agent WHERE name = ?", name)
}

func (dbm *DBManager) GetParentAgentByName(name string) ([]Agent, error) {
	return dbm.getAgentListBySQL("SELECT id,agent_name,amount,bank_card,bank_card_name,bank_card_type,callbackurl,creater_user,last_login_ip,lock_money,name,payfee,remark,sign,state,pay_safe,pay_secret,payqr,payword FROM agent WHERE agent_name = ?", name)
}

func (dbm *DBManager) GetAgentById(platfromId int) (*Agent, error) {
	agent, err := dbm.getAgentBySQL("SELECT id,agent_name,amount,bank_card,bank_card_name,bank_card_type,callbackurl,creater_user,last_login_ip,lock_money,name,payfee,remark,sign,state,pay_safe,pay_secret,payqr,payword FROM agent WHERE id = ?", platfromId)
	return agent, err
}

func (dbm *DBManager) getAgentBySQL(sqlString string, args ...interface{}) (*Agent, error) {
	rows, err := dbm.DB.Query(sqlString, args...)
	if err != nil {
		log.Printf("GetAgent Error:%s\n", err)
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetAgent Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	var agent Agent
	find := false
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("GetAgent Error:", err.Error())
			continue
		}
		var value string
		find = true
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "id":
				agent.Id, _ = strconv.Atoi(value)
			case "agent_name":
				agent.AgentName = value
			case "amount":
				agent.Amount, _ = strconv.ParseFloat(value, 64)
			case "bank_card":
				agent.BankCard = value
			case "bank_card_name":
				agent.BankCardName = value
			case "bank_card_type":
				agent.BankCardType = value
			case "callbackurl":
				agent.CallbackUrl = value
			case "crate_time":
				agent.CreateTime = value
			case "creater_user":
				agent.CreateUser = value
			case "last_login_ip":
				agent.LoginIP = value
			case "last_login_time":
				agent.LoignTime = value
			case "lock_money":
				agent.LockMoney, _ = strconv.ParseFloat(value, 64)
			case "name":
				agent.Name = value
			case "payfee":
				agent.PayFee, _ = strconv.ParseFloat(value, 64)
			case "remark":
				agent.Remark = value
			case "sign":
				agent.Sign = value
			case "state":
				agent.State = value
			case "pay_safe":
				agent.PaySafe = value
			case "pay_secret":
				agent.PaySecret = value
			case "payqr":
				agent.PayQRCode = value
			case "payword":
				agent.Payword = value
			}
		}
	}

	if !find {
		return nil, fmt.Errorf("Can't find th agent.\n")
	}
	return &agent, nil
}

func (dbm *DBManager) getAgentListBySQL(sqlString string, args ...interface{}) ([]Agent, error) {
	var agentList []Agent
	rows, err := dbm.DB.Query(sqlString, args...)
	if err != nil {
		log.Printf("GetAgent Error:%s\n", err)
		return agentList, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		log.Printf("GetAgent Error:%s\n", err.Error())
	}
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	find := false
	// Fetch rows
	for rows.Next() {
		var agent Agent
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println("GetAgentList Error:", err.Error())
			continue
		}
		var value string
		find = true
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = ""
			} else {
				value = string(col)
			}
			switch columns[i] {
			case "id":
				agent.Id, _ = strconv.Atoi(value)
			case "agent_name":
				agent.AgentName = value
			case "amount":
				agent.Amount, _ = strconv.ParseFloat(value, 64)
			case "bank_card":
				agent.BankCard = value
			case "bank_card_name":
				agent.BankCardName = value
			case "bank_card_type":
				agent.BankCardType = value
			case "callbackurl":
				agent.CallbackUrl = value
			case "crate_time":
				agent.CreateTime = value
			case "creater_user":
				agent.CreateUser = value
			case "last_login_ip":
				agent.LoginIP = value
			case "last_login_time":
				agent.LoignTime = value
			case "lock_money":
				agent.LockMoney, _ = strconv.ParseFloat(value, 64)
			case "name":
				agent.Name = value
			case "payfee":
				agent.PayFee, _ = strconv.ParseFloat(value, 64)
			case "remark":
				agent.Remark = value
			case "sign":
				agent.Sign = value
			case "state":
				agent.State = value
			case "pay_safe":
				agent.PaySafe = value
			case "pay_secret":
				agent.PaySecret = value
			case "payqr":
				agent.PayQRCode = value
			case "payword":
				agent.Payword = value
			}
		}
		agentList = append(agentList, agent)
	}

	if !find {
		return agentList, fmt.Errorf("Can't find th agent.\n")
	}
	return agentList, nil
}
