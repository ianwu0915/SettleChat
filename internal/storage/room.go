package storage

import (
	"context"
	// "fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func (p *PostgresStore) CreateRoom(ctx context.Context, name, createdBy string) (string, error) {

	// Check if the room exist or not
	var existingID string

	err := p.DB.QueryRow(ctx, `SELECT id FROM rooms WHERE roomname = $1`, name).Scan(&existingID)
	if err == nil {
		log.Printf("Room already exists: %s", existingID)
		return existingID, nil
	}

	roomID := uuid.NewString()
	log.Printf("Creating room: ID=%s, Name=%s, CreatedBy=%s", roomID, name, createdBy)

	_, err = p.DB.Exec(ctx, `
		INSERT INTO rooms (id, roomname, created_by, created_at)
		VALUES ($1, $2, $3, $4)
	`, roomID, name, createdBy, time.Now().UTC())
	if err != nil {
		log.Printf("Error creating room: %v", err)
		return "", err
	}
	log.Println("Room created successfully")
	return roomID, nil
}

func (p *PostgresStore) AddUserToRoom(ctx context.Context, userID, roomID string) error {
	log.Printf("Adding user %s to room %s", userID, roomID)

	_, err := p.DB.Exec(ctx, `
		INSERT INTO room_members (user_id, room_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, userID, roomID, time.Now().UTC())
	if err != nil {
		log.Printf("Error adding user to room: %v", err)
		return err
	}
	log.Println("User added to room successfully (or already a member)")
	return nil
}

func (p *PostgresStore) GetUserRooms(ctx context.Context, userID string) ([]Room, error) {
	log.Printf("Fetching rooms for user: %s", userID)

	rows, err := p.DB.Query(ctx, `
		SELECT r.id, r.roomname, r.created_by, r.created_at
		FROM rooms r
		JOIN room_members m ON r.id = m.room_id
		WHERE m.user_id = $1
	`, userID)
	if err != nil {
		log.Printf("Error querying user rooms: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.RoomName, &room.CreatedBy, &room.CreatedAt); err != nil {
			log.Printf("Error scanning room row: %v", err)
			return nil, err
		}
		log.Printf("Found room: %+v", room)
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return nil, err
	}
	log.Printf("Total rooms found: %d", len(rooms))
	return rooms, nil
}

// RemoveUserFromRoom 從房間中移除用戶
func (p *PostgresStore) RemoveUserFromRoom(ctx context.Context, userID, roomID string) error {
	// 從 room_members 表中刪除記錄
	// _, err := p.DB.Exec(ctx, `
	// 	DELETE FROM room_members
	// 	WHERE user_id = $1 AND room_id = $2
	// `, userID, roomID)
	// if err != nil {
	// 	return fmt.Errorf("failed to remove user from room: %w", err)
	// }

	// 更新用戶在線狀態為離線
	_, err := p.DB.Exec(ctx, `
		UPDATE user_presence 
		SET is_online = false, last_seen = NOW()
		WHERE user_id = $1 AND room_id = $2
	`, userID, roomID)
	if err != nil {
		log.Printf("Failed to update user presence on room leave: %v", err)
		// 不返回錯誤，因為這是次要操作
	}

	return nil
}
