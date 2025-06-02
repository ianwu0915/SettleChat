package chat

import (
	"log"
	"sync"

	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/messaging/nats"
	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
)

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Store      *storage.PostgresStore
	Publisher  *nats.NATSPublisher
	Subscriber *nats.Subscriber
	Topics     types.TopicFormatter
	EventBus   *messaging.EventBus
	mu         sync.Mutex
}

func NewHub(store *storage.PostgresStore, publisher *nats.NATSPublisher, subscriber *nats.Subscriber, topics types.TopicFormatter, eventbus *messaging.EventBus) *Hub {
	hub := &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Store:      store,
		Publisher:  publisher,
		Subscriber: subscriber,
		Topics:     topics,
		EventBus:   eventbus,
	}

	return hub
}

func (h *Hub) getOrCreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exist := h.Rooms[id]
	if !exist {
		log.Printf("Creating new room: %s", id)
		room = NewRoom(id, h.Publisher, h.Subscriber, h.EventBus)
		h.Rooms[id] = room
		go room.Run(h.Subscriber)
		log.Printf("Room %s created and started", id)
	} else {
		log.Printf("Found existing room: %s", id)
	}

	return room
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			room := h.getOrCreateRoom(client.RoomID)
			room.AddClient(client)

		case client := <-h.UnRegister:
			h.mu.Lock()
			if room, ok := h.Rooms[client.RoomID]; ok {
				room.SaveRemoveClient(client)
			}
			h.mu.Unlock()
		}
	}
}

// FindClient 在指定房間中查找客戶端
func (h *Hub) FindClient(roomID, userID string) (*Client, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.Rooms[roomID]
	if !exists {
		return nil, false
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	client, exists := room.Clients[userID]
	return client, exists
}

// Close gracefully shuts down the hub and all client connections
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, room := range h.Rooms {
		room.Mu.Lock()
		for _, client := range room.Clients {
			close(client.Send)
			client.Conn.Close()
		}
		room.Mu.Unlock()
	}
	h.Rooms = make(map[string]*Room)
}

// GetRoom 獲取指定ID的房間
func (h *Hub) GetRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.Rooms[id]
}
