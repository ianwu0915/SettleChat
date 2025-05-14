package chat

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// Define Client Struct
// Represnet a Websocker connection with a user and a corresponding room
type Client struct {
	Hub      *Hub
	ID       string
	Username string
	Conn     *websocket.Conn
	Send     chan storage.ChatMessage // Message received from broadcast to the room
	RoomID   string
	EventBus *messaging.EventBus
}

func NewClient(hub *Hub, id, username string, conn *websocket.Conn, roomID string, eventBus *messaging.EventBus) *Client {
	return &Client{
		Hub:      hub,
		ID:       id,
		Username: username,
		Conn:     conn,
		Send:     make(chan storage.ChatMessage),
		RoomID:   roomID,
		EventBus: eventBus,
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 120 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

// Write the message recieved from the Send Channel into Websocket to the front-end to display
func (c *Client) WritePump() {

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.Conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 設定 寫入Websocket的超時時間 避免碰到死掉的websocket
			// 如果在 10 秒內沒有成功寫入，這次操作就會 fail，返回錯誤 → goroutine 可以結束，不會 hang 死
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok { // no more values and the channel is closed
				// Server主動要關掉連線時送這個
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.ID, err)
				return
			}
			log.Printf("Sent ping to client: %s", c.ID) // 可選：用於調試
		}
	}
}

// Read the message input from the front-end passed into Websocket and pass into Room.Broadcast
func (c *Client) ReadPump() {
	log.Printf("client connected: %s (%s) in room %s", c.Username, c.ID, c.RoomID)
	defer func() {
		c.Hub.UnRegister <- c
		c.Conn.Close()
	}()

	// 設置最大消息大小
	c.Conn.SetReadLimit(maxMessageSize)

	// 設置初始的讀取截止時間
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// 設置pong處理器，每當收到pong就延長截止時間
	c.Conn.SetPongHandler(func(string) error {
		// 重設讀取截止時間
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		log.Printf("Received pong from client: %s", c.ID) // 可選：用於調試
		return nil
	})

	for {
		var msg storage.ChatMessage
		log.Println("waiting for message...")

		// Read Message from WebSocket
		if err := c.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("unexpected close error:", err)
			} else {
				log.Println("client closed connection:", err)
			}
			break
		}

		// 處理前端發送的心跳消息
		if msg.Content == "" && msg.SenderID == "" {
			// 這可能是前端發送的心跳消息，重置超時並忽略它
			c.Conn.SetReadDeadline(time.Now().Add(pongWait))
			log.Printf("Received heartbeat from client: %s", c.ID)
			continue
		}

		log.Printf("[%s] %s: %s ", c.ID, c.Username, msg.Content)

		// 補上其他field
		// msg.RoomID = c.RoomID
		// msg.SenderID = c.ID
		// msg.Sender = c.Username
		// msg.Timestamp = time.Now()

		// 每次收到消息，重設讀取截止時間
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		// 發佈新消息傳送事件（使用eventbus)
		if c.EventBus != nil {
			if err := c.EventBus.PublishNewMessageEvent(c.RoomID, c.ID, c.Username, msg.Content); err != nil {
				log.Printf("Failed to publish New Message event: %v", err)
			} else {
				log.Printf("Published NewMessage event for %s in room %s", c.Username, c.RoomID)
			}
		} else {
			log.Printf("EventBus is nil!!")
		}

		// if err := c.Hub.Publisher.PublishMessage(msg); err != nil {
		// 	log.Printf("Failed to publish message: %v", err)
		// 	continue
		// }

	}
}
