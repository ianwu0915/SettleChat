package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// ConnectionEventHandler 處理連接相關事件
type ConnectionEventHandler struct {
	store     *storage.PostgresStore
	publisher types.NATSPublisher
	topics    types.TopicFormatter
}

// NewConnectionEventHandler 創建一個新的連接事件處理器
func NewConnectionEventHandler(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter) *ConnectionEventHandler {
	return &ConnectionEventHandler{
		store:     store,
		publisher: publisher,
		topics:    topics,
	}
}

// HandleConnection 處理客戶端連接事件
func (h *ConnectionEventHandler) Handle(msg *nats.Msg) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal connection event: %v", err)
		return err
	}

	roomID, ok := payload["room_id"].(string)
	if !ok {
		log.Printf("Invalid room_id in connection event")
		return nil
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		log.Printf("Invalid user_id in connection event")
		return nil
	}

	username, ok := payload["username"].(string)
	if !ok {
		log.Printf("Invalid username in connection event")
		return nil
	}

	// 檢查是連接還是斷開事件
	isConnection := true
	if eventType, ok := payload["event_type"].(string); ok && eventType == "disconnect" {
		isConnection = false
	}

	// 更新用戶的最後活動時間
	if err := h.store.UpdateLastActive(context.Background(), userID); err != nil {
		log.Printf("Failed to update user's last active time: %v", err)
	}

	log.Printf("Processed connection event for user %s in room %s: connected=%v", username, roomID, isConnection)
	return nil
}
