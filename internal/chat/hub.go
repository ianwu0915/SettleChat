package chat

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Store      *storage.PostgresStore
	Publisher *messaging.Publisher
	Subscriber *messaging.Subscriber
	mu         sync.Mutex
}

func NewHub(store *storage.PostgresStore, publisher *messaging.Publisher,subsbriber *messaging.Subscriber ) *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Store:      store,
		Publisher: publisher,
		Subscriber: subsbriber,
	}
}

func (h *Hub) getOrCreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exist := h.Rooms[id]
	if !exist {
		room = NewRoom(id, h.Publisher)
		h.Rooms[id] = room
		go room.Run(h.Subscriber)
	}

	return room
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			room := h.getOrCreateRoom(client.RoomID)
			room.AddClient(client)

			// Retrieve Historic message from the database
			msgs, err := client.Hub.Store.GetRecentMessages(context.Background(), client.RoomID, 50)
			if err != nil {
				log.Printf("Error Retrieving Past Message from databases: %v", err)
			}
			// Send historic Messgage Safely 
			for i := len(msgs) -1; i>=0; i-- {
				msg := msgs[i]
				// Check if the client exists in the room
				room.mu.Lock()
				_, exists := room.Clients[client.ID]
				room.mu.Unlock()

				if !exists {
					log.Printf("Client %s no longer in room, skipping history message", client.Username)
					break
				}

				select {
				case client.Send <- msg:
					log.Printf("Successfully send history message to %s", client.Username)
				case <-time.After(2 * time.Second):
					log.Printf("⚠️ timed out sending history to %s", client.Username)
				default:
					// 通道已滿或關閉，不進行處理
					log.Printf("Cannot send history to client %s, channel might be closed", client.Username)
				}
			}

		// Handle User Leave
		case client := <-h.UnRegister:
			// If the room exist
			h.mu.Lock()
			if room, ok := h.Rooms[client.RoomID]; ok {
				room.SaveRemoveClient(client)
				// if len(room.Clients) == 0 {
				// 	h.mu.Lock()
				// 	delete(h.Rooms, room.ID)
				// 	h.mu.Unlock()
				// }
			}
			h.mu.Unlock()
		}
	}
}


