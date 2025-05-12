package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ianwu0915/SettleChat/cmd/server/handler"
	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/ws"
	"github.com/nats-io/nats.go"

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

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	} 

	natsManager := messaging.NewNATSManger(natsURL, true)
	if err := natsManager.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsManager.Disconnect()

	publisher := messaging.NewPublisher(natsManager)
	subscriber := messaging.NewSubscriber(natsManager, store)

	if err := subscriber.SubscribeForStorage(); err != nil {
		log.Fatalf("Failed to setup storage subscriber: %v", err)
	}
	

	// Creat a hub
	hub := chat.NewHub(store, publisher, subscriber)
	go hub.Run()

	authHandler := handler.NewAuthHandler(store)
	roomHandler := handler.NewRoomHandler(store)

	mux := http.NewServeMux()
	// Register WsHandler
	mux.HandleFunc("/ws", ws.WebsocketHandler(hub))
	mux.Handle("/register", http.HandlerFunc(authHandler.Register))
	mux.Handle("/login", http.HandlerFunc(authHandler.Login))
	mux.Handle("/user", http.HandlerFunc(authHandler.GetUserByID))
	mux.Handle("/rooms/create", http.HandlerFunc(roomHandler.CreateRoom))
	mux.Handle("/rooms/join", http.HandlerFunc(roomHandler.JoinRoom))
	mux.Handle("/rooms", http.HandlerFunc(roomHandler.GetUserRooms))
	mux.Handle("/", http.FileServer(http.Dir("./web")))


	// Acticate the Sercer
	addr := ":8080"
	log.Println("WebSocket server listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
