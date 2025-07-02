package event_handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ianwu0915/SettleChat/internal/ai"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// AICommandHandler 處理 AI 命令
type AICommandHandler struct {
	publisher types.NATSPublisher
	topics    types.TopicFormatter
	env       string
	manager   *ai.Manager
}

func NewAICommandHandler(publisher types.NATSPublisher, topics types.TopicFormatter, env string, manager *ai.Manager) *AICommandHandler {
	return &AICommandHandler{
		publisher: publisher,
		topics:    topics,
		env:       env,
		manager:   manager,
	}
}

func (h *AICommandHandler) Handle(msg *nats.Msg) error {
	// 1. 解析 AI 命令事件
	var event types.AICommandEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal AI command event: %v", err)
		return err
	}

	// 2. 創建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 3. 處理 AI 命令
	isCommand, response, err := h.manager.HandleAIMessage(ctx, *event.Message)
	if err != nil {
		log.Printf("Failed to handle AI command: %v", err)
		// 發布錯誤消息
		errorMsg := storage.ChatMessage{
			RoomID:    event.Message.RoomID,
			SenderID:  "system",
			Sender:    "System",
			Content:   "Sorry, I encountered an error processing your command.",
			Timestamp: time.Now(),
		}
		return h.publishResponse(errorMsg)
	}

	log.Println(response)

	// 如果不是命令，直接返回
	if !isCommand {
		return nil
	}

	// 4. 發布 AI 回應
	responseMsg := storage.ChatMessage{
		RoomID:    event.Message.RoomID,
		SenderID:  "ai",
		Sender:    "AI Assistant",
		Content:   response,
		Timestamp: time.Now(),
	}

	return h.publishResponse(responseMsg)
}

// publishResponse 發布 AI 回應到聊天室
func (h *AICommandHandler) publishResponse(msg storage.ChatMessage) error {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal AI response: %v", err)
		return err
	}

	// 發布到廣播主題
	broadcastTopic := h.topics.GetBroadcastTopic(msg.RoomID)
	if err := h.publisher.Publish(broadcastTopic, data); err != nil {
		log.Printf("Failed to publish AI response: %v", err)
		return err
	}

	log.Printf("Successfully published AI response to room %s with sender id %s", msg.RoomID, msg.SenderID)
	return nil
}

