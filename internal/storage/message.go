package storage

import (
	"context"
)

func (p *PostgresStore) SaveMessage(ctx context.Context, msg ChatMessage) error {
	_, err := p.DB.Exec(ctx, `
		INSERT INTO messages (room_id, sender_id, sender, content, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`, msg.RoomID, msg.SenderID, msg.Sender, msg.Content, msg.Timestamp)
	return err
}

func (p *PostgresStore) GetRecentMessages(ctx context.Context, roomId string, limit int) ([]ChatMessage, error) {
	rows, err := p.DB.Query(ctx, `
		SELECT id, room_id, sender_id, sender, content, timestamp 
		FROM messages 
		WHERE room_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, roomId, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Sender, &msg.Content, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
