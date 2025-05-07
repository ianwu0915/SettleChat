package main

import (
	"log"
	"net/http"

	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/ws"
)

func main() {
	// Creat a hub
	hub := chat.NewHub()
	go hub.Run()

	// Register WsHandler
	http.HandleFunc("/ws", ws.WebsocketHandler(hub))

	// Acticate the Sercer
	addr := ":8080"
	log.Println("WebSocket server listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServer:", err)
	}

}
