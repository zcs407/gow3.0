package main

import "fmt"

func getDepositStatus(input map[string]interface{}) map[string]interface{} {
	fields := []string{"depositNumber", "MerchaantNo", "payer", "sign"}
	if !verifyFields(input, fields) {
		return GetMissingFieldsError()
	}
	depositNumber := input["depositNumber"].(string)
	merchanntNo := input["MerchaantNo"].(string)
	payer := input["payer"].(string)
	userSign := input["sign"].(string)
	data := make(map[string]interface{})
	agent, err := dbManager.GetAgentByName(merchanntNo)
	if agent == nil || err != nil {
		data["code"] = ERROR_PLATFROM_NOT_EXISTS
		data["msg"] = "Platfrom not exists."
		return data
	}
	sign := fmt.Sprintf("%s%s%s%s", depositNumber, merchanntNo, payer, agent.Sign)
	serviceSign := getMD5(sign)
	if serviceSign != userSign {
		data["code"] = 400
		data["msg"] = "Sign Verify Error."
		return data
	}
	if len(depositNumber) > 0 {
		deposit, err := dbManager.GetDeposit(depositNumber, fmt.Sprintf("%d", agent.Id))
		if err != nil {
			data["code"] = ERROR_DEPOSIT_NOT_EXISTS
			data["msg"] = "get deposit failed. Payer not exists."
			return data
		}
		data["code"] = 200
		data["msg"] = "success"
		data["data"] = deposit
	} else if len(payer) > 0 {
		deposit, err := dbManager.GetDepositStatus(payer, fmt.Sprintf("%d", agent.Id))
		if err != nil {
			data["code"] = ERROR_DEPOSIT_NOT_EXISTS
			data["msg"] = "GetDeposit failed."
			return data
		}
		data["code"] = 200
		data["msg"] = "success"
		data["data"] = deposit
	} else {
		data["code"] = 400
		data["msg"] = "depositNumber and payer has empty."
	}

	return data
}
