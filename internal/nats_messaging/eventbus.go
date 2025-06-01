package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
)

// EventBus 提供統一的事件發布機制
type EventBus struct {
	natsManager *NATSManager
	topics      types.TopicFormatter
}

// NewEventBus 創建一個新的事件總線
func NewEventBus(natsManager *NATSManager, topics types.TopicFormatter) *EventBus {
	return &EventBus{
		natsManager: natsManager,
		topics:      topics,
	}
}

// PublishEvent 發布事件到相應的主題
// 根據event得到對應的NATS topics 並透過natsManage發布
func (eb *EventBus) PublishEvent(event types.Event, roomID string) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event error: %w", err)
	}

	// Getting Topic
	topic := eb.getTopicForEvent(event, roomID)
	log.Printf("Publishing event type [%s] to topic: %s", event.GetType(), topic)

	// NatsManager發布到對應topic
	if err := eb.natsManager.Publish(topic, data); err != nil {
		return fmt.Errorf("publish event error: %w", err)
	}

	return nil
}

// getTopicForEvent 根據事件類型獲取對應的主題
func (eb *EventBus) getTopicForEvent(event types.Event, roomID string) string {
	eventType := event.GetType()

	// 根據事件類型前綴確定主題
	if strings.HasPrefix(eventType, "connection.") {
		return eb.topics.GetConnectionTopic(roomID)
	}

	if eventType == types.EventTypeUserJoined {
		return eb.topics.GetUserJoinedTopic(roomID)
	}

	if eventType == types.EventTypeUserLeft {
		return eb.topics.GetUserLeftTopic(roomID)
	}

	if eventType == types.EventTypeUserPresence {
		return eb.topics.GetPresenceTopic(roomID)
	}

	if eventType == types.EventTypeNewMessage {
		return eb.topics.GetMessageTopic(roomID)
	}

	if eventType == types.EventTypeBroadcastMsg {
		return eb.topics.GetBroadcastTopic(roomID)
	}

	if eventType == types.EventTypeNewAICommand {
		return eb.topics.GetAICommandTopic(roomID)
	}

	// HistoryRequest + HistoryResponse
	if strings.HasPrefix(eventType, "message.history") {
		// 歷史消息請求和響應使用不同的主題
		if strings.Contains(eventType, "request") {
			return eb.topics.GetHistoryRequestTopic(roomID)
		}

		// 對於響應，我們假設使用的是歷史消息響應事件類型
		// 這部分需要根據實際的歷史消息響應事件結構來調整
		// 暫時使用固定的用戶ID "system"
		return eb.topics.GetHistoryResponseTopic(roomID, "system")
	}

	// 默認使用系統消息主題
	return eb.topics.GetSystemMessageTopic(roomID)
}

// PublishConnectEvent 發布連接事件
func (eb *EventBus) PublishConnectEvent(roomID, userID, username string) error {
	event := types.NewConnectEvent(roomID, userID, username)
	return eb.PublishEvent(event, roomID)
}

// PublishDisconnectEvent 發布斷開連接事件
func (eb *EventBus) PublishDisconnectEvent(roomID, userID, username string) error {
	event := types.NewDisconnectEvent(roomID, userID, username)
	return eb.PublishEvent(event, roomID)
}

// PublishUserJoinedEvent 發布用戶加入事件
func (eb *EventBus) PublishUserJoinedEvent(roomID, userID, username string) error {
	event := types.NewUserJoinedEvent(roomID, userID, username)
	return eb.PublishEvent(event, roomID)
}

// PublishUserLeftEvent 發布用戶離開事件
func (eb *EventBus) PublishUserLeftEvent(roomID, userID, username string) error {
	event := types.NewUserLeftEvent(roomID, userID, username)
	return eb.PublishEvent(event, roomID)
}

// PublishPresenceEvent 發布在線狀態事件
func (eb *EventBus) PublishPresenceEvent(roomID, userID, username string, isOnline bool) error {
	event := types.NewPresenceEvent(roomID, userID, username, isOnline)
	return eb.PublishEvent(event, roomID)
}

// PublishNewMessageEvent 發布新訊息事件
func (eb *EventBus) PublishNewMessageEvent(roomID, senderID, sender, content string) error {
	event := types.ChatMessageEvent{
		RoomID:    roomID,
		SenderID:  senderID,
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal message event: %w", err)
	}

	topic := eb.topics.GetMessageTopic(roomID)
	return eb.natsManager.Publish(topic, data)
}

// PublishHistoryRequestEvent 發布歷史消息請求事件
func (eb *EventBus) PublishHistoryRequestEvent(roomID, userID string, limit int) error {
	event := types.NewHistoryRequestEvent(roomID, userID, limit)
	return eb.PublishEvent(event, roomID)
}

// PublishAICommandEvent 發布 AI 命令事件
func (eb *EventBus) PublishAICommandEvent(msg storage.ChatMessage) error {
	// 創建 AI 命令事件
	event := types.NewAICommandEvent(&msg)
	return eb.PublishEvent(event, msg.RoomID)
	
}
 