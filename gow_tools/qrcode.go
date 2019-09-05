package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	barcode "github.com/bieber/barcode"
	_ "github.com/go-sql-driver/mysql"
)

type Configs struct {
	DB struct {
		Addr     string `json:"addr"`
		DbName   string `json:"db"`
		User     string `json:"user"`
		Password string `json:"password"`
		Cron     string `json:"cron"`
	} `json:"db"`
	List []string `json:"list"`
}

var (
	DB *sql.DB
)

func main() {
	configs, _ := loadConfigs("qrcode.json")
	db, err := Connect(configs.DB.Addr, configs.DB.DbName, configs.DB.User, configs.DB.Password)
	if err != nil {
		return
	}
	DB = db
	for k, v := range configs.List {
		qrcode := readQrCode(v)
		log.Printf("qrcode[%d]:%s url:%s\n", k, v, qrcode)
		updateQrCode(v, qrcode)
	}
}

func updateQrCode(account string, qrurl string) error {
	result, err := DB.Exec("Update wechat set qrurl=? where wechat_name=?", qrurl, account)
	if err != nil {
		log.Println("updateQrCode Update Error:", err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Println("UpdateAmount Update Affect Error:", err)
		return err
	}
	log.Printf("[%s]Update:%v\n", account, affect)
	return nil
}

func loadConfigs(fileName string) (Configs, error) {
	file, e := ioutil.ReadFile(fileName)
	if e != nil {
		log.Printf("Load config file error: %v\n", e)
		os.Exit(1)
	}

	var config Configs
	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Config load error:%v \n", err)
		return config, err
	}
	return config, nil
}

func Connect(host string, database string, user string, password string) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local", user, password, host, database))
	if err != nil {
		log.Fatalf("DB Connect Failed. Error:%v\n", err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("DB Connect Failed. Error:%v\n", err)
		return nil, err
	}
	return db, nil
}

func readQrCode(account string) string {
	fileName := fmt.Sprintf("qrcode/%s.jpg", account)
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Println("File not exists.")
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Println("DownloadFile Error Failed:", err)
		return ""
	}
	fin, err := os.Open(fileName)
	if err != nil {
		log.Println("Can't open qrcode image", err)
		return ""
	}
	defer fin.Close()
	src, err := jpeg.Decode(fin)
	if err != nil {
		log.Println("jpeg decode qrcode image", err)
		src, err = png.Decode(fin)
		if err != nil {
			log.Println("png decode qrcode image", err)
			return ""
		}
	}

	img := barcode.NewImage(src)
	scanner := barcode.NewScanner().SetEnabledAll(true)

	symbols, err := scanner.ScanImage(img)
	if err != nil {
		log.Println("ScanImage error", err)
		return ""
	}
	data := ""
	for _, s := range symbols {
		data = s.Data
		fmt.Printf("Name:%s Data:%s Quality:%v Boundary:%v\n", s.Type.Name(), s.Data, s.Quality, s.Boundary)
	}
	return data
}
