package types

import (
	"github.com/nats-io/nats.go"
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

// // ChatMessageEvent 聊天消息事件
// type ChatMessageEvent struct {
// 	RoomID    string    `json:"room_id"`
// 	SenderID  string    `json:"sender_id"`
// 	Sender    string    `json:"sender"`
// 	Content   string    `json:"content"`
// 	Timestamp time.Time `json:"timestamp"`
// }
