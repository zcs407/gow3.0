package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

type DBManager struct {
	DB         *sql.DB
	PersonList map[string]*Person
}

func (dbm *DBManager) Close() {
	if dbm.DB != nil {
		dbm.DB.Close()
	}
}

func (dbm *DBManager) Connect(host string, database string, user string, password string) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local", user, password, host, database))
	if err != nil {
		log.Fatalf("DB Connect Failed. Error:%v\n", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("DB Connect Failed. Error:%v\n", err)
	}
	rows, err := db.Query("SELECT current_timestamp()")
	log.Println("Server Time:", time.Now())
	var myTime time.Time
	if rows.Next() {
		if err = rows.Scan(&myTime); err != nil {
			log.Println("Error Time Incorrect.", err)
		}
	}
	log.Println("Sql Time:", myTime)
	dbm.DB = db
	dbm.PersonList = make(map[string]*Person)
	log.Printf("Connect to Database successful. [%s]:%s\n", host, database)
	log.Println("DB Init Starting.")
	dbm.tableInit()
}

func (dbm *DBManager) tableInit() {
	dbm.IsMidPayTableExists()
	dbm.IsMidPayLogTableExists()
	dbm.TransferPayAccounts()
	dbm.IsMidPayerLogTableExists()
	dbm.IsSFPayTableExists()
	count, _ := dbm.IsMidPayerLogTableSignExists()
	if count == 0 {
		dbm.AlterMidPayerLogTableSign()
	}
	count, _ = dbm.IsMidPayDeviceExists()
	if count == 0 {
		dbm.AlterMidPayDevice()
	}
	count, _ = dbm.IsMidpayQrUrlExist()
	if count == 0 {
		dbm.AlterMidpayQrUrl()
	}

	log.Println("DB Init Successful.")
}

func NewDBManager() *DBManager {
	var dbm DBManager
	return &dbm
}

func (dbm *DBManager) GetAvailableAccounts(platfrom int) []*Person {
	var accounts []*Person
	platfromId := fmt.Sprintf("%d", platfrom)
	for _, v := range dbm.PersonList {
		//fmt.Printf("GetAvailableAccounts[%s]:%s Platfrom:%s\n", v.Account, v.State, v.Platfrom)
		if v.State == "NORMAL" && v.Platfrom == platfromId {
			accounts = append(accounts, v)
		}
	}
	log.Println("GetAvailableAccounts: Platform:", platfromId, "accounts:", len(accounts))
	return accounts
}

func (dbm *DBManager) IsMidPayTableExists() error {
	_, err := dbm.DB.Exec(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(account varchar(100),useTime datetime, status boolean,platfrom varchar(255),type varchar(5),payer varchar(255))",
			CONFIGS.DB.DbName,
			"midpay",
		),
	)
	if err != nil {
		log.Println("IsMidPayTableExists Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) IsSFPayTableExists() error {
	_, err := dbm.DB.Exec(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(account varchar(100),useTime datetime,createTime datetime, status boolean,amount int(8),note varchar(255),qrUrl varchar(255),type varchar(5),payer varchar(255),platform varchar(255),PRIMARY KEY (note))",
			CONFIGS.DB.DbName,
			"sfpay",
		),
	)
	if err != nil {
		log.Println("IsSFPayTableExists Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) IsMidPayerLogTableExists() error {
	_, err := dbm.DB.Exec(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(account varchar(100),request_time datetime,sign varchar(100),payer varchar(255),amount DOUBLE, status varchar(255))",
			CONFIGS.DB.DbName,
			"payer_records",
		),
	)
	if err != nil {
		log.Println("IsMidPayLogTableExists Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) IsMidPayerLogTableSignExists() (int, error) {
	var count int
	rows, err := dbm.DB.Query("SELECT count(*) FROM information_schema.columns WHERE table_schema=? AND table_name = ? AND column_name = ?", CONFIGS.DB.DbName, "payer_records", "custom_sign")
	if err != nil {
		log.Printf("IsMidPayerLogTableSignExists Error:%s\n", err)
		return count, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println("IsMidPayerLogTableSignExists Select Error:", err)
			return count, err
		}
	}
	return count, nil
}

func (dbm *DBManager) AlterMidPayerLogTableSign() error {
	_, err := dbm.DB.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN custom_sign varchar(255)", "payer_records"))
	if err != nil {
		log.Println("AlterMidPayerLogTableSign Error:", err)
		return err
	}
	log.Println("AlterMidPayerLogTableSign add column custom_sign successful. ")
	return nil
}

func (dbm *DBManager) IsMidPayLogTableExists() error {
	_, err := dbm.DB.Exec(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(account varchar(100),request_time datetime,excute_time datetime,payer varchar(255),amount DOUBLE, status varchar(255))",
			CONFIGS.DB.DbName,
			"midpay_records",
		),
	)
	if err != nil {
		log.Println("IsMidPayLogTableExists Error:", err)
		return err
	}
	return nil
}

func (dbm *DBManager) IsMidPayDeviceExists() (int, error) {
	var count int
	rows, err := dbm.DB.Query("SELECT count(*) FROM information_schema.columns WHERE table_schema=? AND table_name = ? AND column_name = ?", CONFIGS.DB.DbName, "midpay_records", "device")
	if err != nil {
		log.Printf("IsMidPayDeviceExists Error:%s\n", err)
		return count, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println("IsMidPayDeviceExists Select Error:", err)
			return count, err
		}
	}
	return count, nil
}

func (dbm *DBManager) AlterMidPayDevice() error {
	_, err := dbm.DB.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN device varchar(255)", "midpay_records"))
	if err != nil {
		log.Println("AlterMidPayDevice Error:", err)
		return err
	}
	log.Println("AlterMidPayDevice add column device successful. ")
	return nil
}

func (dbm *DBManager) IsMidpayQrUrlExist() (int, error) {
	var count int
	rows, err := dbm.DB.Query("SELECT count(*) FROM information_schema.columns WHERE table_schema=? AND table_name = ? AND column_name = ?", CONFIGS.DB.DbName, "midpay_records", "qrurl")
	if err != nil {
		log.Printf("IsMidpayQrUrlExist Error:%s\n", err)
		return count, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Println("IsMidpayQrUrlExist Select Error:", err)
			return count, err
		}
	}
	return count, nil
}

func (dbm *DBManager) AlterMidpayQrUrl() error {
	_, err := dbm.DB.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN qrurl varchar(255) DEFAULT '%s'", "midpay_records", "sfpay"))
	if err != nil {
		log.Println("AlterMidpayQrUrl Error:", err)
		return err
	}
	log.Println("AlterMidpayQrUrl add column qrurl successful.")
	return nil
}
