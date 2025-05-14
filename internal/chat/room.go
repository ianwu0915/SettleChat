package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"
	"github.com/ianwu0915/SettleChat/internal/types"
)

// What we do in Room: Fire a GoRoutine
// User can join or leave the room
// User can send Messgage
// Room will broadcast the message to every users in the room

type Room struct {
	ID         string
	Clients    map[string]*Client
	Publisher  *messaging.NATSPublisher
	Subscriber *messaging.Subscriber
	EventBus   *messaging.EventBus
	Mu         sync.Mutex
}

func NewRoom(id string, publisher *messaging.NATSPublisher, subscriber *messaging.Subscriber, eventBus *messaging.EventBus) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[string]*Client),
		Publisher:  publisher,
		Subscriber: subscriber,
		EventBus:   eventBus,
	}
}

func (r *Room) AddClient(client *Client) {
	log.Printf("Adding client %s to room %s", client.ID, r.ID)

	r.Mu.Lock()
	r.Clients[client.ID] = client
	r.Mu.Unlock()

	// 1. 先訂閱歷史消息響應主題（這是必要的基礎設施操作，保留）
	if r.Subscriber != nil {
		historyResponseTopic := r.Subscriber.Topics.GetHistoryResponseTopic(r.ID, client.ID)
		log.Printf("Subscribing to history response topic: %s", historyResponseTopic)
		if err := r.Subscriber.SubscribeTopic(historyResponseTopic); err != nil {
			log.Printf("Failed to subscribe to history response topic for client %s: %v", client.ID, err)
		} else {
			log.Printf("Successfully subscribed to history response topic: %s", historyResponseTopic)
		}
	} else {
		log.Printf("Warning: Subscriber is nil for room %s", r.ID)
	}

	// 2. 發布客戶端連接事件 (使用 EventBus)
	if r.EventBus != nil {
		if err := r.EventBus.PublishConnectEvent(r.ID, client.ID, client.Username); err != nil {
			log.Printf("Failed to publish client connection event: %v", err)
		} else {
			log.Printf("新方法：Published client connection event for %s in room %s", client.Username, r.ID)
		}
		// } else {
		// 	// 兼容舊的方式，如果 EventBus 尚未設置
		// 	log.Println("舊方法")
		// 	connectionData, _ := json.Marshal(map[string]interface{}{
		// 		"room_id":    r.ID,
		// 		"user_id":    client.ID,
		// 		"username":   client.Username,
		// 		"timestamp":  time.Now(),
		// 		"event_type": "connect",
		// 	})
		// 	connectionTopic := r.Subscriber.Topics.GetConnectionTopic(r.ID)
		// 	if err := r.Publisher.Publish(connectionTopic, connectionData); err != nil {
		// 		log.Printf("Failed to publish client connection event: %v", err)
		// 	} else {
		// 		log.Printf("Published client connection event for %s in room %s", client.ID, r.ID)
		// 	}
	}

	// 3. 發布歷史消息請求事件
	if r.EventBus != nil {
		if err := r.EventBus.PublishHistoryRequestEvent(r.ID, client.ID, 50); err != nil {
			log.Printf("Failed to request history messages: %v", err)
		} else {
			log.Printf("新歷史消息方法：Successfully sent history request for client %s in room %s", client.ID, r.ID)
		}
		// } else {
		// 	// 兼容舊的方式
		// 	log.Println("舊歷史方法")
		// 	historyRequest := types.HistoryRequest{
		// 		RoomID: r.ID,
		// 		UserID: client.ID,
		// 		Limit:  50,
		// 	}
		// 	data, err := json.Marshal(historyRequest)
		// 	if err != nil {
		// 		log.Printf("Failed to marshal history request: %v", err)
		// 	} else {
		// 		requestTopic := r.Subscriber.Topics.GetHistoryRequestTopic(r.ID)
		// 		log.Printf("Sending history request to topic: %s", requestTopic)
		// 		if err := r.Publisher.Publish(requestTopic, data); err != nil {
		// 			log.Printf("Failed to request history messages: %v", err)
		// 		} else {
		// 			log.Printf("Successfully sent history request for client %s in room %s", client.ID, r.ID)
		// 		}
		// 	}
	}

	// 4. 發布用戶加入事件 (使用 EventBus)
	if r.EventBus != nil {
		if err := r.EventBus.PublishUserJoinedEvent(r.ID, client.ID, client.Username); err != nil {
			log.Printf("Failed to publish user joined event: %v", err)
		} else {
			log.Printf("Published user joined event for client %s in room %s", client.ID, r.ID)
		}
		// } else {
		// 	// 兼容舊的方式，如果 EventBus 尚未設置
		// 	userJoinedData := types.UserJoinedMessage{
		// 		RoomID:   r.ID,
		// 		UserID:   client.ID,
		// 		Username: client.Username,
		// 		JoinedAt: time.Now(),
		// 	}
		// 	joinedData, err := json.Marshal(userJoinedData)
		// 	if err != nil {
		// 		log.Printf("Failed to marshal user joined event: %v", err)
		// 	} else {
		// 		joinedTopic := r.Subscriber.Topics.GetUserJoinedTopic(r.ID)
		// 		if err := r.Publisher.Publish(joinedTopic, joinedData); err != nil {
		// 			log.Printf("Failed to publish user joined event: %v", err)
		// 		} else {
		// 			log.Printf("Published user joined event for client %s in room %s", client.ID, r.ID)
		// 		}
		// 	}
	}

	// 注意：不再直接發布在線狀態事件，這將由 UserJoinedHandler 處理
}

func (r *Room) SaveRemoveClient(client *Client) {
	r.Mu.Lock()
	if _, exist := r.Clients[client.ID]; exist {
		delete(r.Clients, client.ID)
		close(client.Send)

		// 1. 發布客戶端斷開連接事件
		if r.EventBus != nil {
			if err := r.EventBus.PublishDisconnectEvent(r.ID, client.ID, client.Username); err != nil {
				log.Printf("Failed to publish client disconnect event: %v", err)
			} else {
				log.Printf("Published client disconnect event for %s in room %s", client.ID, r.ID)
			}
		} else {
			// 兼容舊的方式
			log.Println("依然使用舊斷開連結方法")
			disconnectData, _ := json.Marshal(map[string]interface{}{
				"room_id":    r.ID,
				"user_id":    client.ID,
				"username":   client.Username,
				"timestamp":  time.Now(),
				"event_type": "disconnect",
			})
			disconnectTopic := r.Subscriber.Topics.GetConnectionTopic(r.ID)
			if err := r.Publisher.Publish(disconnectTopic, disconnectData); err != nil {
				log.Printf("Failed to publish client disconnect event: %v", err)
			} else {
				log.Printf("Published client disconnect event for %s in room %s", client.ID, r.ID)
			}
		}

		// 2. 發布用戶離開事件
		if r.EventBus != nil {
			if err := r.EventBus.PublishUserLeftEvent(r.ID, client.ID, client.Username); err != nil {
				log.Printf("Failed to publish user left event: %v", err)
			} else {
				log.Printf("Published user left event for client %s in room %s", client.ID, r.ID)
			}
		} else {
			// 兼容舊的方式
			log.Println("依然使用舊用戶離開方法")
			userLeftData := types.UserLeftMessage{
				RoomID:   r.ID,
				UserID:   client.ID,
				Username: client.Username,
				LeftAt:   time.Now(),
			}
			leftData, err := json.Marshal(userLeftData)
			if err != nil {
				log.Printf("Failed to marshal user left event: %v", err)
			} else {
				leftTopic := r.Subscriber.Topics.GetUserLeftTopic(r.ID)
				if err := r.Publisher.Publish(leftTopic, leftData); err != nil {
					log.Printf("Failed to publish user left event: %v", err)
				} else {
					log.Printf("Published user left event for client %s in room %s", client.ID, r.ID)
				}
			}
		}

		// 注意：不再直接發布離線狀態事件，這將由 UserLeftHandler 處理

		// 3. 取消訂閱歷史消息響應主題
		if r.Subscriber != nil {
			historyResponseTopic := r.Subscriber.Topics.GetHistoryResponseTopic(r.ID, client.ID)
			log.Printf("Unsubscribing from history response topic: %s", historyResponseTopic)
			if err := r.Subscriber.UnsubscribeTopic(historyResponseTopic); err != nil {
				log.Printf("Failed to unsubscribe from history response topic for client %s: %v", client.ID, err)
			} else {
				log.Printf("Successfully unsubscribed from history response topic: %s", historyResponseTopic)
			}
		}
	}
	r.Mu.Unlock()
}

func (r *Room) Run(subscriber *messaging.Subscriber) {
	// 訂閱房間的所有相關主題
	if err := subscriber.SubscribeToRoom(r.ID); err != nil {
		log.Printf("Failed to subscribe to room %s: %v", r.ID, err)
		return
	}

	log.Printf("Room %s is now active with subscriptions", r.ID)
}
