package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/nats-io/nats.go"
)

// 接口
// MessageHandler 定義處理接收消息的函數類型
// 實現者需要處理如何將消息轉發給客戶端或進行其他處理
// 例如：Room可以實現一個handler將消息發送給所有連接的客戶端
type MessageHandler func (msg storage.ChatMessage) error 

// PresenceHandler 定義處理用戶狀態變化的函數類型
// 實現者需要處理用戶上線/下線事件，例如更新在線用戶列表
// 或者向房間內的其他用戶通知狀態變化
type PresenceHandler func (roomID, userID, username string, isOnline bool) error 


type Subscriber struct {
	natsManager *NATSManager
	store *storage.PostgresStore
	subs []*nats.Subscription
}

func NewSubscriber(natsManager *NATSManager, store *storage.PostgresStore) *Subscriber {
	return &Subscriber{
		natsManager: natsManager,
		store: store,
		subs: make([]*nats.Subscription, 0),
	}
}

// Subscribe to a Single Room given a roomID
func (s *Subscriber) SubscribeToRoom(roomID string, handler MessageHandler) error {

	if roomID == "" {
		return fmt.Errorf("room ID cannot be empty")
	}

	subject := fmt.Sprintf("chat.room.%s", roomID)

	sub, err := s.natsManager.Subscribe(subject, func(msg *nats.Msg) {
		var chatMsg storage.ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		if err := handler(chatMsg); err != nil {
			log.Printf("Error handling Message: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to room %s: %w", roomID, err)
	}

	s.subs = append(s.subs, sub)
	log.Printf("Subscribed to room: %s", roomID)
	return nil
}

// SubscribeToAllRooms 訂閱所有房間的消息
func (s *Subscriber) SubscribeToAllRooms(handler MessageHandler) error {
	subject := "chat.room.*"
	
	sub, err := s.natsManager.Subscribe(subject, func(msg *nats.Msg) {
		var chatMsg storage.ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}
		
		if err := handler(chatMsg); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to all rooms: %w", err)
	}
	
	s.subs = append(s.subs, sub)
	log.Printf("Subscribed to all rooms")
	return nil
}

// SubscribeForStorage: used for saving messages into database
func (s *Subscriber) SubscribeForStorage() error {
	subject := "chat.room.*"

	sub, err := s.natsManager.Subscribe(subject, func(msg *nats.Msg) {
		var chatMsg storage.ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("Error Unmarshalling message for storege: %v", err)
			return 
		}

		if chatMsg.SenderID == "system" || chatMsg.Content == "" {
			return 
		}
		
		// Save the message to the database
		go func(msg storage.ChatMessage) {
			// Handle timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := s.store.SaveMessage(ctx, msg); err != nil {
				log.Printf("failed to save message to DB: %v", err)
				// Maybe Retry?
			}
		}(chatMsg)
	})

	if err != nil {
		return fmt.Errorf("failed to create storage subscriber: %w", err)
	}
	
	s.subs = append(s.subs, sub)
	log.Printf("Started message storage subscriber")
	return nil

}

// SubscribeToPresence 訂閱用戶在線狀態變化
func (s *Subscriber) SubscribeToPresence(roomID string, handler PresenceHandler) error {
	subject := fmt.Sprintf("chat.presence.%s", roomID)
	
	sub, err := s.natsManager.Subscribe(subject, func(msg *nats.Msg) {
		var presenceMsg storage.PresenceMessage
		
		if err := json.Unmarshal(msg.Data, &presenceMsg); err != nil {
			log.Printf("Error unmarshaling presence message: %v", err)
			return
		}
		
		if err := handler(presenceMsg.RoomID, presenceMsg.UserID, presenceMsg.Username, presenceMsg.IsOnline); err != nil {
			log.Printf("Error handling presence message: %v", err)
		}
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to presence for room %s: %w", roomID, err)
	}
	
	s.subs = append(s.subs, sub)
	log.Printf("Subscribed to presence updates for room: %s", roomID)
	return nil
}

// Unsubscribe 取消所有訂閱
func (s *Subscriber) Unsubscribe() {
	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing: %v", err)
		}
	}
	s.subs = nil
	log.Println("Unsubscribed from all subscriptions")
}

// Close 清理訂閱並關閉資源
func (s *Subscriber) Close() error {
	s.Unsubscribe()
	return nil
}