package messaging

import (
	"fmt"
	"log"
	"strings"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
	"github.com/nats-io/nats.go"
)

// Subscriber 管理 NATS 訂閱
type Subscriber struct {
	natsManager *NATSManager
	store       *storage.PostgresStore
	subs        []*nats.Subscription
	env         string
	handlers    map[string]types.MessageHandler
	Topics      types.TopicFormatter
}

func NewSubscriber(natsManager *NATSManager, store *storage.PostgresStore, env string, topics types.TopicFormatter) *Subscriber {
	log.Printf("Creating new subscriber for environment: %s", env)
	s := &Subscriber{
		natsManager: natsManager,
		store:       store,
		subs:        make([]*nats.Subscription, 0),
		env:         env,
		handlers:    make(map[string]types.MessageHandler),
		Topics:      topics,
	}
	log.Printf("Subscriber created successfully with env: %s", env)
	return s
}

// RegisterHandler 註冊消息處理器
func (s *Subscriber) RegisterHandler(category, action string, handler types.MessageHandler) {
	handlerKey := category + "." + action
	log.Printf("Registering handler for %s", handlerKey)
	s.handlers[handlerKey] = handler
	log.Printf("Handler registered successfully for %s", handlerKey)
}

// SubscribeToRoom 訂閱特定房間的所有相關主題
func (s *Subscriber) SubscribeToRoom(roomID string) error {
	log.Printf("Starting subscription process for room: %s", roomID)

	// 訂閱用戶加入相關事件
	userJoinedTopic := s.Topics.GetUserJoinedTopic(roomID)
	log.Printf("Subscribing to user joined topic: %s", userJoinedTopic)
	if err := s.SubscribeTopic(userJoinedTopic); err != nil {
		log.Printf("Failed to subscribe to user joined topic: %v", err)
		return err
	}

	// 訂閱用戶離開相關事件
	userLeftTopic := s.Topics.GetUserLeftTopic(roomID)
	log.Printf("Subscribing to user left topic: %s", userLeftTopic)
	if err := s.SubscribeTopic(userLeftTopic); err != nil {
		log.Printf("Failed to subscribe to user left topic: %v", err)
		return err
	}

	// 訂閱訊息相關事件
	messageTopic := s.Topics.GetMessageTopic(roomID)
	log.Printf("Subscribing to message topic: %s", messageTopic)
	if err := s.SubscribeTopic(messageTopic); err != nil {
		log.Printf("Failed to subscribe to message topic: %v", err)
		return err
	}

	// 訂閱廣播消息事件
	broadcastTopic := s.Topics.GetBroadcastTopic(roomID)
	log.Printf("Subscribing to broadcast topic: %s", broadcastTopic)
	if err := s.SubscribeTopic(broadcastTopic); err != nil {
		log.Printf("Failed to subscribe to broadcast topic: %v", err)
		return err
	}

	// 訂閱系統消息事件
	systemTopic := s.Topics.GetSystemMessageTopic(roomID)
	log.Printf("Subscribing to system topic: %s", systemTopic)
	if err := s.SubscribeTopic(systemTopic); err != nil {
		log.Printf("Failed to subscribe to system topic: %v", err)
		return err
	}

	// 訂閱用戶上線狀態事件
	userPresenceTopic := s.Topics.GetPresenceTopic(roomID)
	log.Printf("Subscribing to presence topic: %s", userPresenceTopic)
	if err := s.SubscribeTopic(userPresenceTopic); err != nil {
		log.Printf("Failed to subscribe to presence topic: %v", err)
		return err
	}

	// 訂閱聊天室歷史訊息
	historyMessageRequestTopic := s.Topics.GetHistoryRequestTopic(roomID)
	log.Printf("Subscribing to history message topic: %s", historyMessageRequestTopic)
	if err := s.SubscribeTopic(historyMessageRequestTopic); err != nil {
		log.Printf("Failed to subscribe to history message topic: %v", err)
		return err
	}

	// 訂閱連接事件
	connectionTopic := s.Topics.GetConnectionTopic(roomID)
	log.Printf("Subscribing to connection event topic: %s", connectionTopic)
	if err := s.SubscribeTopic(connectionTopic); err != nil {
		log.Printf("Failed to subscribe to connection event topic: %v", err)
		return err
	}

	log.Printf("Successfully subscribed to all topics for room: %s", roomID)
	return nil
}

// SubscribeTopic 訂閱特定主題
func (s *Subscriber) SubscribeTopic(topic string) error {
	log.Printf("Attempting to subscribe to topic: %s", topic)

	sub, err := s.natsManager.Subscribe(topic, func(msg *nats.Msg) {
		log.Printf("Received message on topic: %s", msg.Subject)

		// 從主題中提取類別和動作
		parts := parseTopic(msg.Subject)
		if len(parts) < 4 {
			log.Printf("Error: Invalid topic format: %s (expected at least 4 parts)", msg.Subject)
			return
		}

		handlerKey := parts[1] + "." + parts[2]
		log.Printf("Looking for handler with key: %s", handlerKey)

		handler, exists := s.handlers[handlerKey]
		if !exists {
			log.Printf("Error: No handler found for topic: %s (key: %s)", msg.Subject, handlerKey)
			return
		}

		log.Printf("Processing message with handler for key: %s", handlerKey)
		if err := handler.Handle(msg); err != nil {
			log.Printf("Error: Failed to handle message for topic %s: %v", msg.Subject, err)
		} else {
			log.Printf("Successfully processed message for topic: %s", msg.Subject)
		}
	})

	if err != nil {
		log.Printf("Error: Failed to subscribe to topic %s: %v", topic, err)
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	s.subs = append(s.subs, sub)
	log.Printf("Successfully subscribed to topic: %s", topic)
	return nil
}

// parseTopic 解析主題字符串
func parseTopic(topic string) []string {
	parts := strings.Split(topic, ".")
	log.Printf("Parsed topic %s into %d parts", topic, len(parts))

	// 如果是歷史消息相關主題，需要特殊處理
	if len(parts) >= 4 && parts[1] == "message" {
		if parts[2] == "history" && parts[3] == "request" {
			// 將 history.request 作為一個完整的 action
			newParts := make([]string, 0)
			newParts = append(newParts, parts[0])          // settlechat
			newParts = append(newParts, parts[1])          // message
			newParts = append(newParts, "history.request") // 完整的 action
			newParts = append(newParts, parts[4:]...)      // roomID 等其他部分
			log.Printf("Reformatted history request topic parts: %v", newParts)
			return newParts
		} else if parts[2] == "history" && parts[3] == "response" {
			// 將 history.response 作為一個完整的 action
			newParts := make([]string, 0)
			newParts = append(newParts, parts[0])           // settlechat
			newParts = append(newParts, parts[1])           // message
			newParts = append(newParts, "history.response") // 完整的 action
			newParts = append(newParts, parts[4:]...)       // roomID 和 userID
			log.Printf("Reformatted history response topic parts: %v", newParts)
			return newParts
		}
	}

	return parts
}

// Unsubscribe 取消所有訂閱
func (s *Subscriber) Unsubscribe() {
	log.Printf("Starting unsubscribe process for %d subscriptions", len(s.subs))
	for i, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing from subscription %d: %v", i+1, err)
		} else {
			log.Printf("Successfully unsubscribed from subscription %d", i+1)
		}
	}
	s.subs = nil
	log.Println("Completed unsubscribe process for all subscriptions")
}

// Close 清理訂閱並關閉資源
func (s *Subscriber) Close() error {
	log.Println("Closing subscriber and cleaning up resources")
	s.Unsubscribe()
	log.Println("Subscriber closed successfully")
	return nil
}

// UnsubscribeTopic 取消訂閱特定主題
func (s *Subscriber) UnsubscribeTopic(topic string) error {
	log.Printf("Attempting to unsubscribe from topic: %s", topic)
	for i, sub := range s.subs {
		if sub.Subject == topic {
			if err := sub.Unsubscribe(); err != nil {
				log.Printf("Error unsubscribing from topic %s: %v", topic, err)
				return err
			}
			// 移除已取消訂閱的主題
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			log.Printf("Successfully unsubscribed from topic: %s", topic)
			return nil
		}
	}
	log.Printf("No subscription found for topic: %s", topic)
	return fmt.Errorf("no subscription found for topic: %s", topic)
}
