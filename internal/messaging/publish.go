package messaging

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ianwu0915/SettleChat/internal/types"
)

// Publish publish message to NATS
type NATSPublisher struct {
	natsManager *NATSManager
	env         string
	topics      types.TopicFormatter
}

func NewPublisher(natsManager *NATSManager, env string, topics types.TopicFormatter) *NATSPublisher {
	log.Printf("Creating new publisher for environment: %s", env)
	p := &NATSPublisher{
		natsManager: natsManager,
		env:         env,
		topics:      topics,
	}
	log.Printf("Publisher created successfully with env: %s", env)
	return p
}

// Publish implements the types.NATSPublisher interface
func (p *NATSPublisher) Publish(topic string, data []byte) error {
	return p.natsManager.Publish(topic, data)
}

// 定義各種可以Publish的事件，帶著對應的payload傳送至NATS
// PublishUserJoined 發布用戶加入事件
func (p *NATSPublisher) PublishUserJoined(roomID, userID, username string) error {
	msg := types.UserJoinedMessage{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		JoinedAt: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetUserJoinedTopic(roomID)
	return p.natsManager.Publish(topic, data)
}

// PublishUserLeft 發布用戶離開事件
func (p *NATSPublisher) PublishUserLeft(roomID, userID, username string) error {
	msg := types.UserLeftMessage{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		LeftAt:   time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetUserLeftTopic(roomID)
	return p.natsManager.Publish(topic, data)
}

// PublishMessage 發布聊天消息
func (p *NATSPublisher) PublishMessage(msg types.ChatMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetMessageTopic(msg.RoomID)
	return p.natsManager.Publish(topic, data)
}

// PublishSystemMessage 發布系統消息
func (p *NATSPublisher) PublishSystemMessage(roomID, message string) error {
	msg := types.SystemMessage{
		RoomID:    roomID,
		Message:   message,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetSystemMessageTopic(roomID)
	return p.natsManager.Publish(topic, data)
}

// PublishUserPresence 發布用戶在線狀態
func (p *NATSPublisher) PublishUserPresence(roomID, userID, username string, isOnline bool) error {
	msg := types.PresenceMessage{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		IsOnline: isOnline,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetPresenceTopic(roomID)
	return p.natsManager.Publish(topic, data)
}

// RequestHistory 請求歷史消息
func (p *NATSPublisher) RequestHistory(roomID string, userID string, limit int) error {
	msg := types.HistoryRequest{
		RoomID: roomID,
		UserID: userID,
		Limit:  limit,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := p.topics.GetHistoryRequestTopic(roomID)
	return p.natsManager.Publish(topic, data)
}