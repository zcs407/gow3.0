package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func SendGetRequest(r *http.Request) (APIResponse, error) {
	if CONFIGS.Debug {
		log.Printf("SendGetRequest Host:%s,Header:%v,URI:%v\n", r.Host, r.Header, r.URL.RequestURI())
	}
	var data APIResponse
	data.Type = "origin"
	client := &http.Client{}
	url := fmt.Sprintf("http://%s%s", CONFIGS.API.Server, r.URL.RequestURI())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Generate Request Failed:", err)
		return data, err
	}
	req.Header = r.Header
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Println("SendGetRequest Send Error:", err)
		return data, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		log.Println("SendGetRequest Body Error:", err)
		return data, err
	}
	data.Header = resp.Header
	data.Body = body
	data.Type = "origin"
	return data, nil
}

type APIResponse struct {
	Type   string
	Header http.Header
	Body   interface{}
}

func SendPostRequest(r *http.Request) (APIResponse, error) {
	if CONFIGS.Debug {
		log.Printf("SendPostRequest Host:%s,Header:%v,URI:%v\n", r.Host, r.Header, r.URL.RequestURI())
	}

	var data APIResponse
	data.Type = "origin"
	client := &http.Client{}
	url := fmt.Sprintf("http://%s%s", CONFIGS.API.Server, r.URL.RequestURI())
	req, err := http.NewRequest("POST", url, r.Body)
	if err != nil {
		log.Println("SendPostRequest Generate Request Failed:", err)
		return data, err
	}
	req.Header = r.Header
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Println("SendPostRequest Send Error:", err)
		return data, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		log.Println("SendPostRequest Body Error:", err)
		return data, err
	}
	data.Header = resp.Header
	data.Body = body
	data.Type = "origin"
	return data, nil
}

func SendCustomRequest(r *http.Request) (APIResponse, error) {
	if CONFIGS.Debug {
		log.Printf("SendCustomRequest Host:%s,Header:%v,URI:%v\n", r.Host, r.Header, r.URL.RequestURI())
	}

	var data APIResponse
	data.Type = "origin"
	client := &http.Client{}
	url := fmt.Sprintf("http://%s%s", CONFIGS.API.Server, r.URL.RequestURI())
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		log.Println("SendCustomRequest Generate Request Failed:", err)
		return data, err
	}
	req.Header = r.Header
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Println("SendCustomRequest Send Error:", err)
		return data, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		log.Println("SendCustomRequest Body Error:", err)
		return data, err
	}
	data.Header = resp.Header
	data.Body = body
	data.Type = "origin"
	return data, nil
}
