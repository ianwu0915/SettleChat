package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (p *PostgresStore) Register(ctx context.Context, username, password string) (string, error) {
	// Check if the current username exist
	var exists bool
	err := p.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`, username).Scan(&exists)
	if err != nil {
		log.Printf("Check User Exists Failed: %s", err)
		return "", err
	}

	if exists {
		return "", errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // what is Cost?
	if err != nil {
		log.Println("Hashing Failed")
		return "", err
	}

	userID := uuid.NewString()
	_, err = p.DB.Exec(ctx, `
		INSERT INTO users(id, username, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, username, string(hash), time.Now().UTC())

	if err != nil {
		log.Println("Inserting Failed")
		return "", err
	}

	return userID, nil

}

func (p *PostgresStore) Login(ctx context.Context, username, password string) (string, error) {
	var userID, hash string
	err := p.DB.QueryRow(ctx, `SELECT id, password_hash FROM users WHERE username=$1`, username).Scan(&userID, &hash)
	if err != nil {
		log.Printf("Login failed for user %s: %v", username, err)
		return "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		log.Println("Wrong Password!")
		return "", errors.New("invalid password")
	}

	return userID, nil

}

func (p *PostgresStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	row := p.DB.QueryRow(ctx, `SELECT id, username, last_active, created_at FROM users WHERE id=$1`, userID)
	var u User
	if err := row.Scan(&u.ID, &u.UserName, &u.LastActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (p *PostgresStore) UpsertUser(ctx context.Context, user User) error {
	_, err := p.DB.Exec(ctx, `
		INSERT INTO users (id, username, last_active, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET username = EXCLUDED.username,
		    last_active = EXCLUDED.last_active
	`, user.ID, user.UserName, user.LastActive, user.CreatedAt)
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
