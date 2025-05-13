package chat

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ianwu0915/SettleChat/internal/events"
	"github.com/ianwu0915/SettleChat/internal/messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	UnRegister chan *Client
	Store      *storage.PostgresStore
	Publisher  *messaging.Publisher
	Subscriber *messaging.Subscriber
	eventBus   events.EventBus
	mu         sync.Mutex
}

func NewHub(store *storage.PostgresStore, publisher *messaging.Publisher, subscriber *messaging.Subscriber) *Hub {
	hub := &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		UnRegister: make(chan *Client),
		Store:      store,
		Publisher:  publisher,
		Subscriber: subscriber,
		eventBus:   events.NewEventBus(),
	}

	hub.setupEventHandlers()
	return hub
}

func (h *Hub) setupEventHandlers() {
	h.eventBus.Subscribe(events.UserJoinedEvent, &userJoinedHandler{store: h.Store})
	h.eventBus.Subscribe(events.UserLeftEvent, &userLeftHandler{})
	h.eventBus.Subscribe(events.MessageSentEvent, &messageSentHandler{store: h.Store})
	h.eventBus.Subscribe(events.SystemMessageEvent, &systemMessageHandler{})
	h.eventBus.Subscribe(events.PresenceEvent, &presenceHandler{publisher: h.Publisher})
	h.eventBus.Subscribe(events.HistoryMessageEvent, &historyMessageHandler{hub: h})
}

func (h *Hub) getOrCreateRoom(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exist := h.Rooms[id]
	if !exist {
		room = NewRoom(id, h.Publisher, h.eventBus)
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

			// 獲取歷史消息並發布事件
			msgs, err := h.Store.GetRecentMessages(context.Background(), client.RoomID, 50)
			if err != nil {
				log.Printf("Error Retrieving Past Message from databases: %v", err)
				continue
			}

			event := events.NewEvent(events.HistoryMessageEvent, events.HistoryMessagePayload{
				RoomID:   client.RoomID,
				UserID:   client.ID,
				Username: client.Username,
				Messages: msgs,
			})

			if err := h.eventBus.Publish(event); err != nil {
				log.Printf("Failed to publish history message event: %v", err)
			}

		case client := <-h.UnRegister:
			h.mu.Lock()
			if room, ok := h.Rooms[client.RoomID]; ok {
				room.SaveRemoveClient(client)
			}
			h.mu.Unlock()
		}
	}
}

// 事件處理器實現
type userJoinedHandler struct {
	store *storage.PostgresStore
}

func (h *userJoinedHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.UserJoinedPayload)
	log.Printf("User %s joined room %s", payload.Username, payload.RoomID)
	return nil
}

type userLeftHandler struct{}

func (h *userLeftHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.UserLeftPayload)
	log.Printf("User %s left room %s", payload.Username, payload.RoomID)
	return nil
}

type messageSentHandler struct {
	store *storage.PostgresStore
}

func (h *messageSentHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.MessageSentPayload)
	ctx := context.Background()
	if err := h.store.SaveMessage(ctx, payload.Message); err != nil {
		log.Printf("Failed to save message to database: %v", err)
		return err
	}
	return nil
}

type systemMessageHandler struct{}

func (h *systemMessageHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.SystemMessagePayload)
	log.Printf("System message in room %s: %s", payload.RoomID, payload.Message)
	return nil
}

type presenceHandler struct {
	publisher *messaging.Publisher
}

func (h *presenceHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.PresencePayload)
	return h.publisher.PublishUserPresence(
		payload.RoomID,
		payload.UserID,
		payload.Username,
		payload.IsOnline,
	)
}

// findClient 在指定房間中查找客戶端
func (h *Hub) findClient(roomID, userID string) (*Client, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.Rooms[roomID]
	if !exists {
		return nil, false
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	client, exists := room.Clients[userID]
	return client, exists
}

type historyMessageHandler struct {
	hub *Hub
}

func (h *historyMessageHandler) Handle(event events.Event) error {
	payload := event.Payload().(events.HistoryMessagePayload)
	client, ok := h.hub.findClient(payload.RoomID, payload.UserID)
	if !ok {
		log.Printf("Client %s no longer in room %s", payload.Username, payload.RoomID)
		return nil
	}

	for i := len(payload.Messages) - 1; i >= 0; i-- {
		msg := payload.Messages[i]
		select {
		case client.Send <- msg:
			log.Printf("Successfully sent history message to %s", payload.Username)
		case <-time.After(2 * time.Second):
			log.Printf("⚠️ timed out sending history to %s", payload.Username)
			return nil
		default:
			log.Printf("Cannot send history to client %s, channel might be closed", payload.Username)
			return nil
		}
	}
	return nil
}


