package events

import (
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// Event 定義基本事件介面
type Event interface {
	Type() string
	Payload() interface{}
}

// EventHandler 定義事件處理器介面
type EventHandler interface {
	Handle(Event) error
}

// EventBus 定義事件總線介面
type EventBus interface {
	Publish(Event) error
	Subscribe(eventType string, handler EventHandler)
	Unsubscribe(eventType string, handler EventHandler)
}

// 定義事件類型常量
const (
	UserJoinedEvent    = "user.joined"
	UserLeftEvent      = "user.left"
	MessageSentEvent   = "message.sent"
	SystemMessageEvent = "system.message"
	PresenceEvent      = "user.presence"
	HistoryMessageEvent = "message.history"
)

// BaseEvent 提供基本事件實現
type BaseEvent struct {
	eventType string
	payload   interface{}
}

func (e *BaseEvent) Type() string {
	return e.eventType
}

func (e *BaseEvent) Payload() interface{} {
	return e.payload
}

// NewEvent 創建新的事件
func NewEvent(eventType string, payload interface{}) Event {
	return &BaseEvent{
		eventType: eventType,
		payload:   payload,
	}
}

// 具體事件類型定義
type UserJoinedPayload struct {
	RoomID   string
	UserID   string
	Username string
}

type UserLeftPayload struct {
	RoomID   string
	UserID   string
	Username string
}

type MessageSentPayload struct {
	Message storage.ChatMessage
}

type SystemMessagePayload struct {
	RoomID  string
	Message string
}

type PresencePayload struct {
	RoomID   string
	UserID   string
	Username string
	IsOnline bool
}

type HistoryMessagePayload struct {
	RoomID   string
	UserID   string
	Username string
	Messages []storage.ChatMessage
} 