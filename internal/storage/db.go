package storage

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

type PostgresStore struct {
	DB *pgxpool.Pool
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Println("Failed initializing connection pool")
		return nil, err
	}

	store := &PostgresStore{DB: db}
	if err := store.migrate(ctx); err != nil {
		return nil, err
	}
	log.Println("Connected to PostgresSQL")
	return store, nil
}

// Create index based on room_id and timestamp for fetching history message based on room and timestamp
// So that such query will be much faster:
// SELECT * FROM messages
// WHERE room_id = 'food'
// ORDER BY timestamp DESC
// LIMIT 50;


func (p *PostgresStore) migrate(ctx context.Context) error {
	_, err := p.DB.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		room_id TEXT NOT NULL,
		sender_id TEXT NOT NULL,
		sender TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_messages_room_time ON messages (room_id, timestamp);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
		last_active TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	);

	CREATE INDEX IF NOT EXISTS idx_users_last_active ON users (last_active);
	`)

	return err
}
