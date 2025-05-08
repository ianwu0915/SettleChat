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
	roomHandler := handler.NewRoomHandler(store)

	mux := http.NewServeMux()
	// Register WsHandler
	mux.HandleFunc("/ws", ws.WebsocketHandler(hub))
	mux.Handle("POST /register", http.HandlerFunc(authHandler.Register))
	mux.Handle("POST /login", http.HandlerFunc(authHandler.Login))
	mux.Handle("GET /user", http.HandlerFunc(authHandler.GetUserByID))
	mux.Handle("POST /rooms/create", http.HandlerFunc(roomHandler.CreateRoom))
	mux.Handle("POST /rooms/join", http.HandlerFunc(roomHandler.JoinRoom))
	mux.Handle("GET /rooms", http.HandlerFunc(roomHandler.GetUserRooms))

	// Acticate the Sercer
	addr := ":8080"
	log.Println("WebSocket server listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
