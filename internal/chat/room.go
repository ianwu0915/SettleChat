package chat

import (
	"log"
	"sync"
	"fmt"
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
	// Broadcast chan storage.ChatMessage
	mu sync.Mutex
}

func NewRoom(id string, publisher *messaging.Publisher) *Room {
	return &Room{
		ID:        id,
		Clients:   make(map[string]*Client),
		// Broadcast: make(chan storage.ChatMessage),
		Publisher: publisher,
	}
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	r.Clients[client.ID] = client
	r.mu.Unlock()
	log.Printf("[%s] %s joined", r.ID, client.Username)

	if err := r.Publisher.PublishUserPresence(r.ID, client.ID, client.Username, true); err != nil {
		log.Printf("Failed to publish presence for user %s joining room %s: %v", 
			client.Username, r.ID, err)
	}

	log.Printf("[%s] %s joined", r.ID, client.Username)
}

func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	if _, exisit := r.Clients[client.ID]; exisit {
		delete(r.Clients, client.ID)
		close(client.Send)
		log.Printf("[%s] %s left", r.ID, client.Username)
	}
	r.mu.Unlock()
	if err := r.Publisher.PublishUserPresence(r.ID, client.ID, client.Username, false); err != nil {
		log.Printf("Failed to publish presence for user %s leaving room %s: %v", 
			client.Username, r.ID, err)
	}

	log.Printf("[%s] %s left", r.ID, client.Username)

}

func (r *Room) Run(subscriber *messaging.Subscriber) {

	err := subscriber.SubscribeToRoom(r.ID, func(msg storage.ChatMessage) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		// Send all message received from subscription to all client in the room
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

	// Subscribe to Onlice Presence status change
	err = subscriber.SubscribeToPresence(r.ID, func(roomID, userID, username string, isOnline bool) error {
		status := "joined"
		if !isOnline {
			status = "left"
		}
		
		// 發送系統消息通知
		systemMsg := fmt.Sprintf("%s %s the room", username, status)
		return r.Publisher.PublishSystemMessage(r.ID, systemMsg)
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
