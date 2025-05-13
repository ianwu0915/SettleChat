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
	// 先用子查詢按時間倒序選擇最近的消息，然後在外層查詢中按時間正序排列
	// 排除 sender_id = 'system' 的消息
	rows, err := p.DB.Query(ctx, `
		WITH recent_messages AS (
			SELECT id, room_id, sender_id, sender, content, timestamp 
			FROM messages 
			WHERE room_id = $1 AND sender_id != 'system'
			ORDER BY timestamp DESC
			LIMIT $2
		)
		SELECT id, room_id, sender_id, sender, content, timestamp 
		FROM recent_messages
		ORDER BY timestamp ASC
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
