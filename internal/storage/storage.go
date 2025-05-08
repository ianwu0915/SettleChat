package storage

import (
	"context"
	"time"
)

type ChatMessage struct {
	ID        int
	RoomID    string
	SenderID  string
	Sender    string
	Content   string
	Timestamp time.Time
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

