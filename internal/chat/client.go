package chat

import (
	"log"
	"time"
	"github.com/gorilla/websocket"
)

// Define Client Struct

type Client struct {
	hub *Hub
	ID       string
	Username string
	Conn     *websocket.Conn
	Send     chan ChatMessage // Message Received from broadcast
	Room     *Room
}

// Define Message Struct
type ChatMessage struct {
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
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
				return
			}

		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Read the message input from the front-end passed into Websocket and pass into Room.Broadcast
func (c *Client) ReadPump() {

	defer func() {
		c.hub.UnRegister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var msg ChatMessage
		// 會接收content
		if err := c.Conn.ReadJSON(&msg); err != nil {
			log.Println("read error: ", err)
			break
		}
		// 補上其他field
		msg.RoomID = c.Room.ID
		msg.SenderID = c.ID
		msg.Sender = c.Username
		msg.Timestamp = time.Now()

		c.Room.Broadcast <- msg

	}
}
