package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ianwu0915/SettleChat/cmd/server/handler"
	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/ws"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load(".env")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	// Connected to Postgres
	store, err := storage.NewPostgresStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}


	// Creat a hub
	hub := chat.NewHub(store)
	go hub.Run()

	authHandler := handler.NewAuthHandler(store)

	mux := http.NewServeMux()
	// Register WsHandler
	mux.HandleFunc("/ws", ws.WebsocketHandler(hub))
	mux.Handle("/register", http.HandlerFunc(authHandler.Register))
	mux.Handle("/login", http.HandlerFunc(authHandler.Login))

	// Acticate the Sercer
	addr := ":8080"
	log.Println("WebSocket server listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
