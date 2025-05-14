package types

import (
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

// UserJoinedMessage 用戶加入消息
type UserJoinedMessage struct {
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

// UserLeftMessage 用戶離開消息
type UserLeftMessage struct {
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	LeftAt   time.Time `json:"left_at"`
}

// PresenceMessage 在線狀態消息
type PresenceMessage struct {
	RoomID   string `json:"room_id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsOnline bool   `json:"is_online"`
}

// SystemMessage 系統消息
type SystemMessage struct {
	RoomID    string    `json:"room_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// HistoryRequest 歷史消息請求
type HistoryRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Limit  int    `json:"limit"`
}

// HistoryResponse 歷史消息響應
type HistoryResponse struct {
	RoomID   string                `json:"room_id"`
	Messages []storage.ChatMessage `json:"messages"`
}
