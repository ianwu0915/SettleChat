package chat

import (
	"log"
)

// What we do in Room: Fire a GoRoutine
// User can join or leave the room
// User can send Messgage
// Room will broadcast the message to every users in the room

type Room struct {
	ID        string
	Clients   map[string]*Client // userId -> Client
	Broadcast chan ChatMessage
}

func NewRoom(id string) *Room {
	return &Room{
		ID:        id,
		Clients:   make(map[string]*Client),
		Broadcast: make(chan ChatMessage),
	}
}

func (r *Room) AddClient(client *Client) {
	r.Clients[client.ID] = client
	log.Printf("[%s] %s joined", r.ID, client.Username)
}

func (r *Room) RemoveClient(client *Client) {
	if _, exisit := r.Clients[client.ID]; exisit {
		delete(r.Clients, client.ID)
		close(client.Send)
		log.Printf("[%s] %s left", r.ID, client.Username)
	}
}

func (r *Room) Run() {
	for {
		message := <-r.Broadcast
		for _, client := range r.Clients {

			// non-blocking
			select {
			case client.Send <- message:

			// If client is offline, or other problems
			default:
				close(client.Send)
				delete(r.Clients, client.ID)
			}
		}
	}
}
