package types

import (
	"time"
)

// Event 是所有事件的基本介面
type Event interface {
	// GetType 返回事件類型
	GetType() string

	// GetTimestamp 返回事件發生時間
	GetTimestamp() time.Time
}

// BaseEvent 提供事件的基本實現
type BaseEvent struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// GetType 實現 Event 介面
func (e BaseEvent) GetType() string {
	return e.Type
}

// GetTimestamp 實現 Event 介面
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// NewBaseEvent 創建一個基本事件
func NewBaseEvent(eventType string) BaseEvent {
	return BaseEvent{
		Type:      eventType,
		Timestamp: time.Now(),
	}
}

// 定義標準事件類型常量
const (
	// 連接相關事件
	EventTypeConnect    = "connection.connect"
	EventTypeDisconnect = "connection.disconnect"

	// 用戶相關事件
	EventTypeUserJoined   = "user.joined"
	EventTypeUserLeft     = "user.left"
	EventTypeUserPresence = "user.presence"

	// 消息相關事件
	// 傳送訊息
	EventTypeNewMessage     = "message.new"
	EventTypeBroadcastMsg   = "message.broadcast"
	EventTypeMessageHistory = "message.history"
)

// ConnectionEvent 連接事件
type ConnectionEvent struct {
	BaseEvent
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// NewConnectEvent 創建連接事件
func NewConnectEvent(roomID, userID, username string) ConnectionEvent {
	return ConnectionEvent{
		BaseEvent: NewBaseEvent(EventTypeConnect),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
	}
}

// NewDisconnectEvent 創建斷開連接事件
func NewDisconnectEvent(roomID, userID, username string) ConnectionEvent {
	return ConnectionEvent{
		BaseEvent: NewBaseEvent(EventTypeDisconnect),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
	}
}

// UserEvent 用戶事件
type UserEvent struct {
	BaseEvent
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// NewUserJoinedEvent 創建用戶加入事件
func NewUserJoinedEvent(roomID, userID, username string) UserEvent {
	return UserEvent{
		BaseEvent: NewBaseEvent(EventTypeUserJoined),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
	}
}

// NewUserLeftEvent 創建用戶離開事件
func NewUserLeftEvent(roomID, userID, username string) UserEvent {
	return UserEvent{
		BaseEvent: NewBaseEvent(EventTypeUserLeft),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
	}
}

// PresenceEvent 在線狀態事件
type PresenceEvent struct {
	BaseEvent
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"`
}

// NewPresenceEvent 創建在線狀態事件
func NewPresenceEvent(roomID, userID, username string, isOnline bool) PresenceEvent {
	return PresenceEvent{
		BaseEvent: NewBaseEvent(EventTypeUserPresence),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		IsOnline:  isOnline,
	}
}

type ChatMessageEvent struct {
	BaseEvent
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewChatMessageEvent 創建聊天消息事件
func NewChatMessageEvent(roomID, userID, username, content string) ChatMessageEvent {
	now := time.Now()
	return ChatMessageEvent{
		BaseEvent: NewBaseEvent(EventTypeNewMessage),
		RoomID:    roomID,
		SenderID:  userID,
		Sender:    username,
		Content:   content,
		Timestamp: now,
	}
}

// HistoryRequestEvent 歷史消息請求事件
type HistoryRequestEvent struct {
	BaseEvent
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Limit  int    `json:"limit"`
}

// NewHistoryRequestEvent 創建歷史消息請求事件
func NewHistoryRequestEvent(roomID, userID string, limit int) HistoryRequestEvent {
	return HistoryRequestEvent{
		BaseEvent: NewBaseEvent(EventTypeMessageHistory + ".request"),
		RoomID:    roomID,
		UserID:    userID,
		Limit:     limit,
	}
}
