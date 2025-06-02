package event_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// UserJoinedHandler 處理用戶加入事件
type UserJoinedHandler struct {
	store     *storage.PostgresStore
	publisher types.NATSPublisher
	topics    types.TopicFormatter
	env       string
}

func NewUserJoinedHandler(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter, env string) *UserJoinedHandler {
	return &UserJoinedHandler{
		store:     store,
		publisher: publisher,
		topics:    topics,
		env:       env,
	}
}

func (h *UserJoinedHandler) Handle(msg *nats.Msg) error {
	var payload types.UserJoinedMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal user joined message: %v", err)
		return err
	}

	// 將用戶添加到房間
	if err := h.store.AddUserToRoom(context.Background(), payload.UserID, payload.RoomID); err != nil {
		log.Printf("Failed to add user to room in database: %v", err)
		return err
	}

	// 發布系統消息
	systemMsg := fmt.Sprintf("%s joined the room", payload.Username)
	systemTopic := h.topics.GetSystemMessageTopic(payload.RoomID)
	systemPayload := types.SystemMessage{
		RoomID:    payload.RoomID,
		Message:   systemMsg,
		Timestamp: time.Now(),
	}
	systemData, err := json.Marshal(systemPayload)
	if err != nil {
		log.Printf("Failed to marshal system message: %v", err)
		return err
	}
	if err := h.publisher.Publish(systemTopic, systemData); err != nil {
		log.Printf("Failed to publish system message: %v", err)
	}

	// 發布在線狀態更新
	presenceTopic := h.topics.GetPresenceTopic(payload.RoomID)
	presenceMsg := types.PresenceMessage{
		RoomID:   payload.RoomID,
		UserID:   payload.UserID,
		Username: payload.Username,
		IsOnline: true,
	}
	presenceData, err := json.Marshal(presenceMsg)
	if err != nil {
		return err
	}
	if err := h.publisher.Publish(presenceTopic, presenceData); err != nil {
		log.Printf("Failed to publish presence message: %v", err)
	}

	return nil
}

// UserLeftHandler 處理用戶離開事件
type UserLeftHandler struct {
	store     *storage.PostgresStore
	publisher types.NATSPublisher
	topics    types.TopicFormatter
	env       string
}

func NewUserLeftHandler(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter, env string) *UserLeftHandler {
	return &UserLeftHandler{
		store:     store,
		publisher: publisher,
		topics:    topics,
		env:       env,
	}
}

func (h *UserLeftHandler) Handle(msg *nats.Msg) error {
	var payload types.UserLeftMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal user left message: %v", err)
		return err
	}

	// // 從房間中移除用戶
	// if err := h.store.RemoveUserFromRoom(context.Background(), payload.UserID, payload.RoomID); err != nil {
	// 	log.Printf("Failed to remove user from room in database: %v", err)
	// 	return err
	// }

	// 發布系統消息
	systemMsg := fmt.Sprintf("%s left the room", payload.Username)
	systemTopic := h.topics.GetSystemMessageTopic(payload.RoomID)
	systemPayload := types.SystemMessage{
		RoomID:    payload.RoomID,
		Message:   systemMsg,
		Timestamp: time.Now(),
	}
	systemData, err := json.Marshal(systemPayload)
	if err != nil {
		log.Printf("Failed to marshal system message: %v", err)
		return err
	}
	if err := h.publisher.Publish(systemTopic, systemData); err != nil {
		log.Printf("Failed to publish system message: %v", err)
	}

	// 發布離線狀態更新
	presenceTopic := h.topics.GetPresenceTopic(payload.RoomID)
	presenceMsg := types.PresenceMessage{
		RoomID:   payload.RoomID,
		UserID:   payload.UserID,
		Username: payload.Username,
		IsOnline: false,
	}
	presenceData, err := json.Marshal(presenceMsg)
	if err != nil {
		return err
	}
	if err := h.publisher.Publish(presenceTopic, presenceData); err != nil {
		log.Printf("Failed to publish presence message: %v", err)
	}

	return nil
}

// PresenceHandler 處理用戶在線狀態
type PresenceHandler struct {
	store  *storage.PostgresStore
	topics types.TopicFormatter
	env    string
}

// NewPresenceHandler 創建新的 PresenceHandler
func NewPresenceHandler(store *storage.PostgresStore, topics types.TopicFormatter, env string) *PresenceHandler {
	return &PresenceHandler{
		store:  store,
		topics: topics,
		env:    env,
	}
}

// Handle 處理用戶在線狀態消息
func (h *PresenceHandler) Handle(msg *nats.Msg) error {
	var presence types.PresenceMessage
	if err := json.Unmarshal(msg.Data, &presence); err != nil {
		log.Printf("Failed to unmarshal presence message: %v", err)
		return err
	}

	// 更新數據庫中的在線狀態
	if err := h.store.UpdatePresence(context.Background(), presence.RoomID, presence.UserID, presence.IsOnline); err != nil {
		log.Printf("Failed to update presence in database: %v", err)
		return err
	}

	// 更新用戶的最後活動時間
	if err := h.store.UpdateLastActive(context.Background(), presence.UserID); err != nil {
		log.Printf("Failed to update user's last active time: %v", err)
		// 不返回錯誤，因為這不是關鍵操作
	}

	log.Printf("Updated presence for user %s (%s) in room %s: online=%v",
		presence.Username, presence.UserID, presence.RoomID, presence.IsOnline)

	return nil
}
