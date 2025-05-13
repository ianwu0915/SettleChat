package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ianwu0915/SettleChat/internal/chat"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// ChatMessageHandler 處理聊天消息
type ChatMessageHandler struct {
	store     *storage.PostgresStore
	publisher types.NATSPublisher
	topics    types.TopicFormatter
}

func NewChatMessageHandler(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter) *ChatMessageHandler {
	return &ChatMessageHandler{
		store:     store,
		publisher: publisher,
		topics:    topics,
	}
}

func (h *ChatMessageHandler) Handle(msg *nats.Msg) error {
	var chatMsg storage.ChatMessage
	if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
		return err
	}

	// 儲存消息到數據庫
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.store.SaveMessage(ctx, chatMsg); err != nil {
		log.Printf("Failed to save message to database: %v", err)
		return err
	}

	// 廣播消息給所有客戶端
	broadcastTopic := h.topics.GetBroadcastTopic(chatMsg.RoomID)
	if err := h.publisher.Publish(broadcastTopic, msg.Data); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
		return err
	}

	return nil
}

// HistoryHandler 處理歷史消息請求
type HistoryHandler struct {
	store     *storage.PostgresStore
	publisher types.NATSPublisher
	topics    types.TopicFormatter
	env       string
}

func NewHistoryHandler(store *storage.PostgresStore, publisher types.NATSPublisher, topics types.TopicFormatter, env string) *HistoryHandler {
	return &HistoryHandler{
		store:     store,
		publisher: publisher,
		topics:    topics,
		env:       env,
	}
}

func (h *HistoryHandler) Handle(msg *nats.Msg) error {
	var payload types.HistoryRequest
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal history request: %v", err)
		return err
	}

	log.Printf("Received history request for room %s from user %s", payload.RoomID, payload.UserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := h.store.GetRecentMessages(ctx, payload.RoomID, payload.Limit)
	if err != nil {
		log.Printf("Failed to get recent messages: %v", err)
		return err
	}

	log.Printf("Found %d messages for room %s", len(messages), payload.RoomID)

	response := types.HistoryResponse{
		RoomID:   payload.RoomID,
		Messages: messages,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal history response: %v", err)
		return err
	}

	// 使用特定用戶的響應主題
	replyTopic := h.topics.GetHistoryResponseTopic(payload.RoomID, payload.UserID)
	if err := h.publisher.Publish(replyTopic, responseData); err != nil {
		log.Printf("Failed to publish history response: %v", err)
		return err
	}

	log.Printf("Successfully sent history response to %s", replyTopic)
	return nil
}

// BroadcastHandler 處理廣播消息
type BroadcastHandler struct {
	hub *chat.Hub
}

func NewBroadcastHandler(hub *chat.Hub) *BroadcastHandler {
	return &BroadcastHandler{
		hub: hub,
	}
}

func (h *BroadcastHandler) Handle(msg *nats.Msg) error {
	var chatMsg storage.ChatMessage
	if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
		log.Printf("Failed to unmarshal broadcast message: %v", err)
		return err
	}

	// 獲取對應的房間
	room := h.hub.GetRoom(chatMsg.RoomID)
	if room == nil {
		log.Printf("Room not found: %s", chatMsg.RoomID)
		return fmt.Errorf("room not found: %s", chatMsg.RoomID)
	}

	// 發送消息給房間內的所有客戶端
	room.Mu.Lock()
	for _, client := range room.Clients {
		select {
		case client.Send <- chatMsg:
			log.Printf("Sent message to client %s", client.ID)
		default:
			log.Printf("Client %s send buffer full, message dropped", client.ID)
			close(client.Send)
			delete(room.Clients, client.ID)
		}
	}
	room.Mu.Unlock()

	return nil
}

// HistoryResponseHandler 處理歷史消息響應
type HistoryResponseHandler struct {
	hub *chat.Hub
}

func NewHistoryResponseHandler(hub *chat.Hub) *HistoryResponseHandler {
	return &HistoryResponseHandler{
		hub: hub,
	}
}

func (h *HistoryResponseHandler) Handle(msg *nats.Msg) error {
	var response types.HistoryResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		log.Printf("Failed to unmarshal history response: %v", err)
		return err
	}

	// 從主題中提取用戶ID (格式: settlechat.message.history.response.{roomID}.{userID})
	parts := strings.Split(msg.Subject, ".")
	if len(parts) < 6 {
		log.Printf("Invalid history response topic format: %s", msg.Subject)
		return fmt.Errorf("invalid history response topic format: %s", msg.Subject)
	}
	userID := parts[5]

	log.Printf("Received history response for room %s, user %s with %d messages", 
		response.RoomID, userID, len(response.Messages))

	// 查找對應的客戶端
	client, found := h.hub.FindClient(response.RoomID, userID)
	if !found {
		log.Printf("Client not found: room=%s, user=%s", response.RoomID, userID)
		return fmt.Errorf("client not found: room=%s, user=%s", response.RoomID, userID)
	}

	// 發送歷史消息（已經是按時間順序從舊到新排列）
	for _, message := range response.Messages {
		select {
		case client.Send <- message:
			log.Printf("Sent history message to client %s", client.ID)
		default:
			log.Printf("Client %s send buffer full, history message dropped", client.ID)
			return fmt.Errorf("client send buffer full")
		}
	}

	log.Printf("Successfully sent all history messages to client %s", client.ID)
	return nil
} 