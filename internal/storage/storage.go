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
	ID         string `json:"user_id"`
	UserName   string `json:"user_name"`
	CreatedAt  time.Time `json:"created_At"`
	LastActive time.Time `json:"last_active"`
}

type Room struct {
	ID        string    `json:"room_id"`
	RoomName      string    `json:"room_name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}


// Leverage the Go interface for better decoupling, unit-testing
type MessageStore interface {
	SaveMessage(ctx context.Context, msg ChatMessage) error 
	GetRecentMessages(ctx context.Context, roomID string, limit int) ([]ChatMessage, error)
}

type UserStore interface {
	// Regiser and Login
	UpsertUser(ctx context.Context, user User) error // insert or update
	UpdateLastActive(ctx context.Context, userID string) error
	GetUser(ctx context.Context, userId string) (*User, error)
}

type RoomStore interface {
	CreateRoom(ctx context.Context, name, createdBy string) (string, error)
	GetUserRooms(ctx context.Context, userID string) ([]Room, error)
	AddUserToRoom(ctx context.Context, userID, roomID string) error
}

