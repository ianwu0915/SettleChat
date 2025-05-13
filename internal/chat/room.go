package chat

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/ianwu0915/SettleChat/internal/messaging"
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
	Mu         sync.Mutex
}

func NewRoom(id string, publisher *messaging.NATSPublisher, subscriber *messaging.Subscriber) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[string]*Client),
		Publisher:  publisher,
		Subscriber: subscriber,
	}
}

func (r *Room) AddClient(client *Client) {
	log.Printf("Adding client %s to room %s", client.ID, r.ID)

	r.Mu.Lock()
	r.Clients[client.ID] = client
	r.Mu.Unlock()

	// 1. 先訂閱歷史消息響應主題
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

	// 2. 然後發送歷史消息請求
	historyRequest := types.HistoryRequest{
		RoomID: r.ID,
		UserID: client.ID,
		Limit:  50,
	}
	data, err := json.Marshal(historyRequest)
	if err != nil {
		log.Printf("Failed to marshal history request: %v", err)
	} else {
		requestTopic := r.Subscriber.Topics.GetHistoryRequestTopic(r.ID)
		log.Printf("Sending history request to topic: %s", requestTopic)
		if err := r.Publisher.Publish(requestTopic, data); err != nil {
			log.Printf("Failed to request history messages: %v", err)
		} else {
			log.Printf("Successfully sent history request for client %s in room %s", client.ID, r.ID)
		}
	}

	// 3. 發布用戶加入事件
	if err := r.Publisher.PublishUserJoined(r.ID, client.ID, client.Username); err != nil {
		log.Printf("Failed to publish user joined event: %v", err)
	} else {
		log.Printf("Published user joined event for client %s in room %s", client.ID, r.ID)
	}

	// 4. 發布在線狀態事件
	if err := r.Publisher.PublishUserPresence(r.ID, client.ID, client.Username, true); err != nil {
		log.Printf("Failed to publish presence event: %v", err)
	} else {
		log.Printf("Published presence event for client %s in room %s", client.ID, r.ID)
	}
}

func (r *Room) SaveRemoveClient(client *Client) {
	r.Mu.Lock()
	if _, exist := r.Clients[client.ID]; exist {
		delete(r.Clients, client.ID)
		close(client.Send)
		
		// 發布用戶離開事件
		if err := r.Publisher.PublishUserLeft(r.ID, client.ID, client.Username); err != nil {
			log.Printf("Failed to publish user left event: %v", err)
		}

		// 發布離線狀態事件
		if err := r.Publisher.PublishUserPresence(r.ID, client.ID, client.Username, false); err != nil {
			log.Printf("Failed to publish presence event: %v", err)
		}

		// 取消訂閱歷史消息響應主題
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
