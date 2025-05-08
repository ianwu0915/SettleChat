package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/ws"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load(".env")
	dbURL := os.Getenv("DATABASE_URL")


	// Creat a hub
	hub := chat.NewHub()
	go hub.Run()

	// // Connected to Postgres
	// store, err := storage.NewPostgresStore(dbURL)


	// Register WsHandler
	http.HandleFunc("/ws", ws.WebsocketHandler(hub))

	// Acticate the Sercer
	addr := ":8080"
	log.Println("WebSocket server listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
