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

// 在這邊確認是否是

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

// 可以在這邊攔截Message是否是AI command
// 是的話 Publish AICommandEvent到Eventbus 然後處理
func (h *ChatMessageHandler) Handle(msg *nats.Msg) error {
	// 先嘗試解析為 types.ChatMessageEvent
	var event types.ChatMessageEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("嘗試舊的方式直接解析為 storage.ChatMessage")
		// 嘗試舊的方式直接解析為 storage.ChatMessage
		
		var chatMsg storage.ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("舊方式解析失敗: %v", err)
			return err
		}
		
		// 確保必要的字段存在
		if chatMsg.SenderID == "" {
			log.Printf("Warning: SenderID is empty in the message")
		}
		if chatMsg.Sender == "" {
			log.Printf("Warning: Sender is empty in the message")
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
	
	// 成功解析為 ChatMessageEvent，轉換為 storage.ChatMessage
	log.Printf("成功解析為 ChatMessageEvent，轉換為 storage.ChatMessage")
	chatMsg := storage.ChatMessage{
		RoomID:    event.RoomID,
		SenderID:  event.SenderID,
		Sender:    event.Sender,
		Content:   event.Content,
		Timestamp: event.Timestamp,
	}
	
	// 儲存消息到數據庫
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.store.SaveMessage(ctx, chatMsg); err != nil {
		log.Printf("Failed to save message to database: %v", err)
		return err
	}

	// 將 chatMsg 重新序列化以便廣播
	broadcastData, err := json.Marshal(chatMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message for broadcast: %v", err)
		return err
	}

	// 廣播消息給所有客戶端
	broadcastTopic := h.topics.GetBroadcastTopic(chatMsg.RoomID)
	if err := h.publisher.Publish(broadcastTopic, broadcastData); err != nil {
		log.Printf("Failed to broadcast message: %v", err)
		return err
	}

	log.Printf("Message from %s processed and broadcasted successfully", chatMsg.Sender)
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
	
	// 先嘗試解析為 types.ChatMessageEvent
	var event types.ChatMessageEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		// 嘗試舊的方式直接解析為 storage.ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("Failed to unmarshal broadcast message: %v", err)
			return err
		}
	} else {
		// 從 event 轉換為 storage.ChatMessage
		chatMsg = storage.ChatMessage{
			RoomID:    event.RoomID,
			SenderID:  event.SenderID,
			Sender:    event.Sender,
			Content:   event.Content,
			Timestamp: event.Timestamp,
		}
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

// HistoryResponseHandler 處理歷史消息回應
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

	// 計算消息總數，以便顯示進度
	totalMessages := len(response.Messages)
	if totalMessages == 0 {
		log.Printf("No history messages to send for client %s", client.ID)
		return nil
	}

	// 發送歷史消息（已經是按時間順序從舊到新排列）
	// 使用批量發送策略，每批次發送5條消息，並在批次之間添加短暫延遲
	const batchSize = 10
	const delayBetweenBatches = 50 * time.Millisecond
	
	for i := 0; i < totalMessages; i += batchSize {
		// 計算當前批次的結束位置
		end := i + batchSize
		if end > totalMessages {
			end = totalMessages
		}
		
		// 處理當前批次的消息
		for j := i; j < end; j++ {
			message := response.Messages[j]
			select {
			case client.Send <- message:
				log.Printf("Sent history message %d/%d to client %s", j+1, totalMessages, client.ID)
			default:
				log.Printf("Client %s send buffer full at message %d/%d, waiting before retry...", 
					client.ID, j+1, totalMessages)
				
				// 如果發送通道已滿，等待一段時間後再嘗試
				time.Sleep(100 * time.Millisecond)
				
				// 重試發送，如果仍然失敗則報錯
				select {
				case client.Send <- message:
					log.Printf("Retry success: Sent history message %d/%d to client %s", 
						j+1, totalMessages, client.ID)
				default:
					log.Printf("Client %s send buffer still full after retry, message %d/%d dropped", 
						client.ID, j+1, totalMessages)
					return fmt.Errorf("client send buffer full after retry")
				}
			}
		}
		
		// 批次之間添加延遲，給客戶端處理消息的時間
		if end < totalMessages {
			time.Sleep(delayBetweenBatches)
		}
	}

	log.Printf("Successfully sent all %d history messages to client %s", totalMessages, client.ID)
	return nil
}
