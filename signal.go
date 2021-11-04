package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { 
		return true 
	},
}

func serveSignal(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	doneChan := make(chan struct{})
	defer ws.Close()
	go hub(ws, doneChan)
	// go ping(ws, doneChan)
	<-doneChan
}
