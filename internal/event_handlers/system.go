package event_handlers

import (
	"encoding/json"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// SystemMessageHandler 處理系統消息
type SystemMessageHandler struct {
	publisher types.NATSPublisher
	topics    types.TopicFormatter
	env       string
}

func NewSystemMessageHandler(publisher types.NATSPublisher, topics types.TopicFormatter, env string) *SystemMessageHandler {
	return &SystemMessageHandler{
		publisher: publisher,
		topics:    topics,
		env:       env,
	}
}

func (h *SystemMessageHandler) Handle(msg *nats.Msg) error {
	var payload types.SystemMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		return err
	}

	// 廣播系統消息
	messageTopic := h.topics.GetMessageTopic(payload.RoomID)
	chatMsg := storage.ChatMessage{
		RoomID:    payload.RoomID,
		SenderID:  "system",
		Content:   payload.Message,
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(chatMsg)
	if err != nil {
		return err
	}

	return h.publisher.Publish(messageTopic, messageData)
}
