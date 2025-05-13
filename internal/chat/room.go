package chat

import (
	"fmt"
	"log"
	"sync"

	"github.com/ianwu0915/SettleChat/internal/events"
	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// What we do in Room: Fire a GoRoutine
// User can join or leave the room
// User can send Messgage
// Room will broadcast the message to every users in the room

type Room struct {
	ID        string
	Clients   map[string]*Client // userId -> Client
	Publisher *messaging.Publisher
	eventBus  events.EventBus
	// Broadcast chan storage.ChatMessage
	mu sync.Mutex
}

func NewRoom(id string, publisher *messaging.Publisher, eventBus events.EventBus) *Room {
	return &Room{
		ID:        id,
		Clients:   make(map[string]*Client),
		// Broadcast: make(chan storage.ChatMessage),
		Publisher: publisher,
		eventBus:  eventBus,
	}
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	r.Clients[client.ID] = client
	r.mu.Unlock()

	// 發布用戶加入事件
	event := events.NewEvent(events.UserJoinedEvent, events.UserJoinedPayload{
		RoomID:   r.ID,
		UserID:   client.ID,
		Username: client.Username,
	})
	
	if err := r.eventBus.Publish(event); err != nil {
		log.Printf("Failed to publish user joined event: %v", err)
	}

	// 發布在線狀態事件
	presenceEvent := events.NewEvent(events.PresenceEvent, events.PresencePayload{
		RoomID:   r.ID,
		UserID:   client.ID,
		Username: client.Username,
		IsOnline: true,
	})

	if err := r.eventBus.Publish(presenceEvent); err != nil {
		log.Printf("Failed to publish presence event: %v", err)
	}
}

func (r *Room) SaveRemoveClient(client *Client) {
	r.mu.Lock()
	if _, exist := r.Clients[client.ID]; exist {
		delete(r.Clients, client.ID)
		close(client.Send)
		
		// 發布用戶離開事件
		event := events.NewEvent(events.UserLeftEvent, events.UserLeftPayload{
			RoomID:   r.ID,
			UserID:   client.ID,
			Username: client.Username,
		})
		
		if err := r.eventBus.Publish(event); err != nil {
			log.Printf("Failed to publish user left event: %v", err)
		}

		// 發布離線狀態事件
		presenceEvent := events.NewEvent(events.PresenceEvent, events.PresencePayload{
			RoomID:   r.ID,
			UserID:   client.ID,
			Username: client.Username,
			IsOnline: false,
		})

		if err := r.eventBus.Publish(presenceEvent); err != nil {
			log.Printf("Failed to publish presence event: %v", err)
		}
	}
	r.mu.Unlock()
}

func (r *Room) Run(subscriber *messaging.Subscriber) {
	err := subscriber.SubscribeToRoom(r.ID, func(msg storage.ChatMessage) error {
		r.mu.Lock()
		defer r.mu.Unlock()

		event := events.NewEvent(events.MessageSentEvent, events.MessageSentPayload{
			Message: msg,
		})
		
		if err := r.eventBus.Publish(event); err != nil {
			log.Printf("Failed to publish message sent event: %v", err)
		}

		for clientID, client := range r.Clients {
			select {
			case client.Send <- msg:
			default:
				log.Printf("Failed to send message to client %s, removing from room %s", clientID, r.ID)
				close(client.Send)
				delete(r.Clients, clientID)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Failed to subscribe to room %s: %v", r.ID, err)
		return
	}

	err = subscriber.SubscribeToPresence(r.ID, func(roomID, userID, username string, isOnline bool) error {
		status := "joined"
		if !isOnline {
			status = "left"
		}
		
		systemMsg := fmt.Sprintf("%s %s the room", username, status)
		event := events.NewEvent(events.SystemMessageEvent, events.SystemMessagePayload{
			RoomID:  r.ID,
			Message: systemMsg,
		})
		
		if err := r.eventBus.Publish(event); err != nil {
			log.Printf("Failed to publish system message event: %v", err)
		}

		return nil
	})
	
	if err != nil {
		log.Printf("Failed to subscribe to presence for room %s: %v", r.ID, err)
	}
	
	log.Printf("Room %s is now active with subscriptions", r.ID)


	// for {
	// 	message := <-r.Broadcast
	// 	for _, client := range r.Clients {

	// 		// non-blocking
	// 		select {
	// 		case client.Send <- message:

	// 		// If client is offline, or other problems
	// 		default:
	// 			close(client.Send)
	// 			delete(r.Clients, client.ID)
	// 		}
	// 	}
	// }
}
