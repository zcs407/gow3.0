package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/robfig/cron"
)

const (
	programName string = "GOW API"
	version     string = "0.0.91"
)

var (
	CONFIGS    Configs
	modTime    time.Time
	configPath string
	dbManager  *DBManager
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Printf("[%s] Version:%s\n", programName, version)
	//get config file from command line
	flag.StringVar(&configPath, "c", "configs.json", "config file path")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s version[%s]\r\nUsage: %s [OPTIONS]\r\n", programName, version, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	log.Printf("[Config File]:%s\n", configPath)
	//Load config file
	config, err := loadConfigs(configPath)
	if err != nil {
		log.Println("Load config file error:", err)
		os.Exit(1)
	}
	CONFIGS = config

	SetUlimit(1002000)
	//Set Using CPUs
	useCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(useCPU)
	//Start Web Server
	go WebServer()
	//Handle the Singal
	go signalHandler()
	// cronjob for checkFailDeposit
	startCheckFailDeposit()
	//cronjob for account status
	startCheckMidpayAccount()
	//cronjob for reset deposit count
	startCheckResetDepositCount()
	//Database DBManager
	dbManager = NewDBManager()
	dbManager.Connect(CONFIGS.DB.Addr, CONFIGS.DB.DbName, CONFIGS.DB.User, CONFIGS.DB.Password)
	for {
		select {
		case <-time.After(1 * time.Second):
			configWatcher()
			if time.Now().Second()%30 == 0 {
				if CONFIGS.Debug {
					log.Printf("[System Status] CPU:%d Goroutines:%d\n", useCPU, runtime.NumGoroutine())
				}
			}
		}
	}
}

func startCheckFailDeposit() {
	c := cron.New()
	log.Println("CheckFailDeposit starting.", CONFIGS.Deposit.Cron)
	c.AddFunc(CONFIGS.Deposit.Cron, checkFailDeposit)
	c.Start()
}

func startCheckMidpayAccount() {
	c := cron.New()
	log.Println("startCheckMidpayAccount starting.", CONFIGS.Deposit.Cron)
	c.AddFunc(CONFIGS.Deposit.Cron, CheckMidpayAccount)
	c.Start()
}

func startCheckResetDepositCount() {
	c := cron.New()
	log.Println("startCheckResetDepositCount starting.", CONFIGS.Deposit.Reset)
	c.AddFunc(CONFIGS.Deposit.Reset, ResetDepositCount)
	c.Start()
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	// Passing no signals to Notify means that
	// all signals will be sent to the channel.
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM)

	// Block until any signal is received.
	s := <-c
	dbManager.Close()
	log.Fatalf("OS Signal:%v\n", s)
}

func configWatcher() {
	file, err := os.Open(configPath) // For read access.
	if err != nil {
		log.Println("configWatcher error:", err)
	}
	info, err := file.Stat()
	if err != nil {
		log.Println("configWatcher error:", err)
	}
	if modTime.Unix() == -62135596800 {
		modTime = info.ModTime()
	}

	if info.ModTime() != modTime {
		log.Printf("Config file changed. Reolad config file.\n")
		modTime = info.ModTime()
		CONFIGS, err = loadConfigs(configPath)
		if err != nil {
			log.Printf("configWatcher error:%v\n", err)
		}
	}
	defer file.Close()
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
	switch config.Env {
	case "prd":
	case "dev":
	case "test":
	}
	return config, nil
}
