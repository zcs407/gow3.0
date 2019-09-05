package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"
)

func getMD5(content string) string {
	h := md5.New()
	io.WriteString(h, content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getMD5Slim(content string) string {
	h := md5.New()
	io.WriteString(h, content)
	str := fmt.Sprintf("%x", h.Sum(nil))
	return str[:20]
}

func IPv4Verify(ip string) (bool, error) {
	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		return false, fmt.Errorf("%v is not an IPv4 address\n", trial)
	}
	return true, nil
}

func WriteToFile(filename string, data []byte) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return err
	}
	return nil
}

func SetUlimit(number uint64) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("[Error]: Getting Rlimit ", err)
	}
	rLimit.Max = number
	rLimit.Cur = number
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("[Error]: Setting Rlimit ", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("[Error]: Getting Rlimit ", err)
	}
	log.Println("set file limit done:", rLimit)
}

func SpliceArray(res []string, idx int) []string {
	var rep []string
	if len(res) < idx {
		return res
	}

	if idx > 0 {
		rep = res[:idx]
	}
	if idx+1 < len(res) {
		rep = append(rep, res[idx+1:]...)
	}
	return rep
}
