package storage
import (
	"time"
	"context"
	"github.com/google/uuid"
)

func (p *PostgresStore) CreateRoom(ctx context.Context, name, createdBy string) (string, error) {
	roomID := uuid.NewString()
	_, err := p.DB.Exec(ctx, `
		INSERT INTO rooms (id, roomname, created_by, created_at)
		VALUES ($1, $2, $3, $4)
	`, roomID, name, createdBy, time.Now().UTC())
	if err != nil {
		return "", err
	}
	return roomID, nil
}

func (p *PostgresStore) AddUserToRoom(ctx context.Context, userID, roomID string) error {
	_, err := p.DB.Exec(ctx, `
		INSERT INTO room_members (user_id, room_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`, userID, roomID, time.Now().UTC())
	return err
}

func (p *PostgresStore) GetUserRooms(ctx context.Context, userID string) ([]Room, error) {
	rows, err := p.DB.Query(ctx, `
		SELECT p.id, p.roomname, p.created_by, p.created_at
		FROM rooms r
		JOIN room_members m ON p.id = m.room_id
		WHERE m.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.RoomName, &room.CreatedBy, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}
