package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // dev, prod -> origin whitelist
	},
}

// 從路由參數取得 roomID
// 升級 HTTP → WebSocket
// 建立 Client 實例（包含：userID、username、roomID、conn、send chan）
// 把這個 client 註冊進 Hub.Register
// 啟動這個 client 的 ReadPump() + WritePump() goroutines
func WebsocketHandler(hub *chat.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading", err)
			return
		}

		// Get all the info :roomId, userId, username from from the URL query parameters
		roomID := r.URL.Query().Get("room")
		userID := r.URL.Query().Get("user_id")
		username := r.URL.Query().Get("username")

		if roomID == "" || userID == "" || username == "" {
			log.Println("Missing query parameters")
			http.Error(w, "Missing room/user_id/username", http.StatusBadRequest)
			return
		}

		// Construct Client
		client := &chat.Client{
			Hub:      hub,
			ID:       userID,
			Username: username,
			Conn:     conn,
			Send:     make(chan storage.ChatMessage),
			RoomID:   roomID,
		}

		// Register the client into the room
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump() //handle conn close

	}

}
