package main

import (
	"fmt"
	"time"
)

type Configs struct {
	Debug bool   `json:"debug"`
	Env   string `json:"env"`
	HTTP  string `json:"http"`
	WS    string `json:"ws"`
	HTTPS string `json:"https"`
	HB    bool   `json:"hb"`
	SFPay bool   `json:"sfpay"`
	DB    struct {
		Addr     string `json:"addr"`
		DbName   string `json:"db"`
		User     string `json:"user"`
		Password string `json:"password"`
		Cron     string `json:"cron"`
	} `json:"db"`
	SSL struct {
		Key string `json:"key"`
		Crt string `json:"crt"`
	} `json:"ssl"`
	Person struct {
		AccountLockTime   int    `json:"accountLockTime"`
		ImageServer       string `json:"imageServer"`
		OutputImageServer string `json:"outputImageServer"`
		Pay               struct {
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"pay"`
		ClientLockTime int                     `json:"clientLockTime"`
		CustomList     map[string]CustomPerson `json:"customList"`
	} `json:"person"`

	API struct {
		Server string `json:"server"`
	} `json:"api"`
	Deposit struct {
		Cron  string `json:"cronjob"`
		Reset string `json:"resetjob"`
		Times int    `json:"times"`
		Hours int    `json:"hours"`
		Count int    `json:"count"`
	} `json:"deposit"`
	QrCode struct {
		GenPlatform int   `json:"platform"`
		AmountList  []int `json:"amountList"`
	} `json:"qrcode"`
	Biz struct {
		Acccounts []string `json:"accounts"`
		Agents    []string `json:"agents"`
	} `json:"biz"`
}

type CustomPerson struct {
	SFPay           bool `json:"sfpay"`
	AccountLockTime int  `json:"accountLockTime"`
	ClientLockTime  int  `json:"clientLockTime"`
	Pay             struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"pay"`
	Name string `json:"name"`
}

type Person struct {
	Account  string `json:"account"`
	State    string `json:"state"`
	Url      string `json:"url"`
	NickName string `json:"nickname"`
	RealName string `json:"realname"`
	QrUrl    string `json:"qrurl"`
	Note     string `json:"note"`
	Platfrom string `json:"platfrom"`
	Type     string `json:"type"`
}

type PayPerson struct {
	Account    string
	UseTime    time.Time
	CreateTime time.Time
	Status     int
	Amount     float64
	Platfrom   string
	Type       string
	Payer      string
	QrUrl      string
	Note       string
	Lock       string
	LockTime   time.Time
}

type AccountInfo struct {
	Account   string  `json:"account"`
	Type      string  `json:"type"`
	IP        string  `json:"ip"`
	NickName  string  `json:"nickname"`
	Amount    float64 `json:"amount"`
	DayAmount float64 `json:"dayamount"`
	DayLimit  float64 `json:"daylimit"`
	Platfrom  int     `json:"platfrom"`
	QrUrl     string  `json:"qrurl"`
	RealName  string  `json:realname`
	State     string  `json:"state"`
}

type DepositRecord struct {
	Id            int       `json:"id"`
	Amount        float64   `json:"amount"`
	CreateUser    string    `json:"createUser"`
	DepositNumber string    `json:"depositNumber"`
	Note          string    `json:"note"`
	PayAccount    string    `json:"payAccount"`
	Platfrom      int       `json:"platfrom"`
	TransferTime  string    `json:"transferTime"`
	WechatName    string    `json:"wechatName"`
	TranTime      time.Time `json:"tranTime"`
	CreateTime    time.Time `json:"createTime"`
	ExcuteTime    time.Time `json:"excuteTime"`
	TranFee       float64   `json:"tranfee"`
	NickName      string    `json:"nickName"`
	State         string    `json:"state"`
	BillNo        string    `json:"billNo"`
	Times         int       `json:"times"`
	IP            string    `json:"ip"`
	Remark        string    `json:"userRemark"`
	PayType       string    `json:"payType"`
	CallbackUrl   string    `json:"callUrl"`
	Sign          string    `json:"sign"`
	CustomSign    string    `json:"customSign"`
}

type Agent struct {
	Id           int
	AgentName    string
	Amount       float64
	BankCard     string
	BankCardName string
	BankCardType string
	CallbackUrl  string
	CreateTime   string
	CreateUser   string
	LoginIP      string
	LoignTime    string
	LockMoney    float64
	Name         string
	PayFee       float64
	Remark       string
	Sign         string
	State        string
	PaySafe      string
	PaySecret    string
	PayQRCode    string
	Payword      string
}

func (a *Agent) Dump() {
	fmt.Printf("Id:%v AgentName:%v CallbackUrl:%v Name:%v Sign:%v\n", a.Id, a.AgentName, a.CallbackUrl, a.Name, a.Sign)
}

type UserReport struct {
	Id           int
	Platfrom     string
	Account      string
	AccountType  string
	Type         string
	CreateUser   string
	CreateTime   time.Time
	ChangeMoney  float64
	BeforeMoney  float64
	NowMoney     float64
	NickName     string
	UserName     string
	Remark       string
	State        string
	DestBankCard string
	IP           string
	DestIP       string
}

type MidPayRecord struct {
	Account     string
	RequestTime time.Time
	ExcuteTime  time.Time
	Payer       string
	Amount      float64
	Status      string
	CustomSign  string
	Platform    string
	APIType     string
	QrUrl       string
}

type MidPayObject struct {
	Url         string  `json:"url"`
	Amount      float64 `json:"amount"`
	Payer       string  `json:"username"`
	NickName    string  `json:"nickname"`
	RealName    string  `json:"realname"`
	Account     string  `json:"account"`
	OverTime    int     `json:"overtime"`
	QRUrl       string  `json:"qrurl"`
	RefreshTime int     `json:"refreshtime"`
	Timestamp   int64   `json:"ts"`
	Mode        string  `json:"mode"`
}

type MidPayRecordObject struct {
	Account     string
	RequestTime time.Time
	Amount      float64
	Mode        string `json:"mode"`
	QrUrl       string
}
