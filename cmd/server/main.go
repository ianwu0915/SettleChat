package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ianwu0915/SettleChat/cmd/server/handler"
	"github.com/ianwu0915/SettleChat/internal/chat"
	handlers "github.com/ianwu0915/SettleChat/internal/event_handlers"
	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	// 1. 加載環境變量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "dev"
	}

	// 2. 初始化數據庫連接
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	store, err := storage.NewPostgresStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer store.Close()

	// 3. 初始化 NATS 連接
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	natsManager := messaging.NewNATSManager(natsURL, true)
	if err := natsManager.Connect(); err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsManager.Disconnect()

	// 4. 創建主題格式化器
	topics := messaging.NewTopicFormatter("")

	// 5. 創建發布器
	publisher := messaging.NewPublisher(natsManager, env, topics)

	// 5.1 創建事件總線
	eventBus := messaging.NewEventBus(natsManager, topics)

	// 6. 創建 Hub
	hub := chat.NewHub(store, publisher, nil, topics, eventBus)
	go hub.Run()

	// 7. 創建消息處理器
	messageHandlers := initializeHandlers(store, publisher, topics, env, hub)

	// 8. 創建並初始化訂閱器
	subscriber := messaging.NewSubscriber(natsManager, store, env, topics)
	registerHandlers(subscriber, messageHandlers)

	// 設置 Hub 的訂閱器
	hub.Subscriber = subscriber

	// 9. 創建 HTTP 處理器
	authHandler := handler.NewAuthHandler(store)
	roomHandler := handler.NewRoomHandler(store, publisher, env, eventBus)

	// 10. 設置路由
	mux := http.NewServeMux()
	setupRoutes(mux, hub, authHandler, roomHandler)

	// 11. 創建 HTTP 服務器
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// 12. 設置優雅關閉
	go gracefulShutdown(server, hub, subscriber)

	// 13. 啟動服務器
	log.Printf("Server starting on %s in %s environment", server.Addr, env)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// initializeHandlers 初始化所有消息處理器
func initializeHandlers(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter, env string, hub *chat.Hub) map[string]types.MessageHandler {

	handlers_list := make(map[string]types.MessageHandler)
	handlers_list["user.joined"] = handlers.NewUserJoinedHandler(store, publisher, topics, env)
	handlers_list["user.left"] = handlers.NewUserLeftHandler(store, publisher, topics, env)
	handlers_list["user.presence"] = handlers.NewPresenceHandler(store, topics, env)
	handlers_list["message.chat"] = handlers.NewChatMessageHandler(store, publisher, topics)
	handlers_list["message.history.request"] = handlers.NewHistoryHandler(store, publisher, topics, env)
	handlers_list["message.history.response"] = handlers.NewHistoryResponseHandler(hub)
	handlers_list["message.broadcast"] = handlers.NewBroadcastHandler(hub)
	handlers_list["system.message"] = handlers.NewSystemMessageHandler(publisher, topics, env)
	handlers_list["connection.event"] = handlers.NewConnectionEventHandler(store, publisher, topics)

	return handlers_list
}

// registerHandlers 註冊所有處理器到訂閱器
func registerHandlers(subscriber *messaging.Subscriber, handlers map[string]types.MessageHandler) {
	for topic, handler := range handlers {
		parts := strings.Split(topic, ".")
		if len(parts) >= 2 {
			subscriber.RegisterHandler(parts[0], strings.Join(parts[1:], "."), handler)
		}
	}
}

// setupRoutes 設置 HTTP 路由
func setupRoutes(mux *http.ServeMux, hub *chat.Hub, auth *handler.AuthHandler, room *handler.RoomHandler) {
	mux.HandleFunc("/ws", handler.WebsocketHandler(hub))
	mux.Handle("/register", http.HandlerFunc(auth.Register))
	mux.Handle("/login", http.HandlerFunc(auth.Login))
	mux.Handle("/user", http.HandlerFunc(auth.GetUserByID))
	mux.Handle("/rooms/create", http.HandlerFunc(room.CreateRoom))
	mux.Handle("/rooms/join", http.HandlerFunc(room.JoinRoom))
	mux.Handle("/rooms/leave", http.HandlerFunc(room.LeaveRoom))
	mux.Handle("/rooms", http.HandlerFunc(room.GetUserRooms))
	mux.Handle("/", http.FileServer(http.Dir("./web")))
}

// gracefulShutdown 處理優雅關閉
func gracefulShutdown(server *http.Server, hub *chat.Hub, subscriber *messaging.Subscriber) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	// 1. 停止接受新的 HTTP 請求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// 2. 關閉 Hub（這會關閉所有 WebSocket 連接）
	hub.Close()

	// 3. 取消 NATS 訂閱
	subscriber.Close()

	log.Println("Server gracefully stopped")
}
