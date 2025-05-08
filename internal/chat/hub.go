package chat

import (
	"sync"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Store      storage.PostgresStore
	mu         sync.Mutex
}

func NewHub(store storage.PostgresStore) *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Store: store,
	}
}

func (h *Hub) getOrCreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exist := h.Rooms[id]
	if !exist {
		room = NewRoom(id)
		h.Rooms[id] = room
		go room.Run()
	}

	return room
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			room := h.getOrCreateRoom(client.RoomID)
			room.AddClient(client)

		// Handle User Leave
		case client := <-h.UnRegister:
			// If the room exist
			if room, ok := h.Rooms[client.RoomID]; ok {
				room.RemoveClient(client)
				if len(room.Clients) == 0 {
					h.mu.Lock()
					delete(h.Rooms, room.ID)
					h.mu.Unlock()
				}
			}
		}
	}
}
