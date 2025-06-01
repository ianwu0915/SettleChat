package types

import (
	"time"

	"github.com/nats-io/nats.go"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

// MessageHandler 定義消息處理器接口
type MessageHandler interface {
	Handle(msg *nats.Msg) error
}

// TopicFormatter 定義主題格式化接口
type TopicFormatter interface {
	GetMessageTopic(roomID string) string
	GetPresenceTopic(roomID string) string
	GetHistoryRequestTopic(roomID string) string
	GetHistoryResponseTopic(roomID, userID string) string
	GetSystemMessageTopic(roomID string) string
	GetUserJoinedTopic(roomID string) string
	GetUserLeftTopic(roomID string) string
	GetBroadcastTopic(roomID string) string
	GetConnectionTopic(roomID string) string
	GetAICommandTopic(roomID string) string
}

// ChatMessageEvent 聊天消息事件
type ChatMessageEvent struct {
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// AICommandEvent AI 命令事件
type AICommandEvent struct {
	storage.ChatMessage
	Timestamp time.Time `json:"timestamp"`
}
