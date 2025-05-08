package storage

import (
	"context"
	"time"
)

func (p *PostgresStore) UpsertUser(ctx context.Context, user User) error {
	_, err := p.DB.Exec(ctx, `
		INSERT INTO users (id, username, last_active, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET username = EXCLUDED.username,
		    last_active = EXCLUDED.last_active
	`, user.ID, user.Username, user.LastActive, user.CreatedAt)
	return err
}

func (p *PostgresStore) UpdateLastActive(ctx context.Context, userID string) error {
	_, err := p.DB.Exec(ctx, `
		UPDATE users
		SET last_active = $1
		WHERE id = $2
	`, time.Now().UTC(), userID)

	return err 
}