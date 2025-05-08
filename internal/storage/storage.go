package storage

import (
	"context"
	"time"
)

type ChatMessage struct {
	ID        int       `json:"-"` // 不 expose 給前端
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type User struct {
	ID         string
	Username   string
	CreatedAt  time.Time
	LastActive time.Time
}

// Leverage the Go interface for better decoupling, unit-testing


type MessageStore interface {
	SaveMessage(ctx context.Context, msg ChatMessage) error 
	GetRecentMessage(ctx context.Context, roomID string, limit int) ([]ChatMessage, error)
}

type UserStore interface {
	UpsertUser(ctx context.Context, user User) error // insert or update
	UpdateLastActive(ctx context.Context, userID string) error
	GetUser(ctx context.Context, userId string) (*User, error)
}

