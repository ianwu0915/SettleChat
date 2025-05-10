package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

// Publish publish message to NATS
type Publisher struct {
	natsManager *NATSManager
}

func NewPublisher(NATSManager *NATSManager) *Publisher {
	return &Publisher {
		natsManager: NATSManager,
	}
}

// Publish Chatmessage to the target room subject
// subject: chat.room.roomId
func (p *Publisher) PublishChatMessage(msg storage.ChatMessage) error {
	if msg.RoomID == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	// Serialize the message (Paylod)
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	subject := fmt.Sprintf("chat.room.%s", msg.RoomID)

	err = p.natsManager.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to %s: %s", subject, msg.Content)
	return nil

}

// PublishSystemMessage 發布系統消息到指定房間
func (p *Publisher) PublishSystemMessage(roomID, content string) error {
	// 創建系統消息
	msg := storage.ChatMessage{
		RoomID:   roomID,
		SenderID: "system",
		Sender:   "System",
		Content:  content,
	}

	return p.PublishChatMessage(msg)
}

// PublishUserPresence 發布用戶在線狀態變更消息
func (p *Publisher) PublishUserPresence(roomID, userID, username string, isOnline bool) error {
	// 創建狀態消息
	presenceMsg := storage.PresenceMessage {
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		IsOnline: isOnline,
	}

	// 序列化消息
	data, err := json.Marshal(presenceMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal presence message: %w", err)
	}

	// 構建主題名稱
	subject := fmt.Sprintf("chat.presence.%s", roomID)

	// 發布消息
	err = p.natsManager.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish presence message: %w", err)
	}

	status := "joined"
	if !isOnline {
		status = "left"
	}
	log.Printf("Published presence update: %s %s room %s", username, status, roomID)
	return nil
}