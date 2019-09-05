package main

import (
	_ "crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var WebRouter *mux.Router

func WebServer() {
	WebRouter = mux.NewRouter()
	WebRouter.HandleFunc("/", HomeHandler)
	WebRouter.HandleFunc("/{cmd}", GowHandler)
	images := http.StripPrefix("/images/", http.FileServer(http.Dir("./images/")))
	WebRouter.PathPrefix("/images/").Handler(images)
	go func() {
		srv := &http.Server{
			Addr:         CONFIGS.HTTP,
			Handler:      WebRouter,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		log.Println("Web Service starting.", CONFIGS.HTTP)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal("Web Service start failed.", err)
		}
	}()
	http.HandleFunc("/ws", wsHandler)

	WebSocketInit()
	go func() {
		srv := &http.Server{
			Addr:         CONFIGS.WS,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		log.Println("Websocket Service starting.", CONFIGS.WS)
		if CONFIGS.WS != "" {
			err := srv.ListenAndServe()
			if err != nil {
				log.Fatal("Web Service start failed.", err)
			}
		}
	}()
	if CONFIGS.HTTPS != "" && CONFIGS.SSL.Key != "" && CONFIGS.SSL.Crt != "" {
		srv := &http.Server{
			Addr:         CONFIGS.HTTPS,
			Handler:      WebRouter,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		}
		log.Println("Web HTTPS Service starting.", CONFIGS.HTTPS)
		err := srv.ListenAndServeTLS(CONFIGS.SSL.Crt, CONFIGS.SSL.Key)
		if err != nil {
			fmt.Println("ListenAndServeTLS:", err)
			log.Fatal("Listen SSL Web Server Failed:", err)
		}
	}
}
