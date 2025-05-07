package chat

import (
	"log"
	"golang.org/x/text/message"
)

// What we do in Room: Fire a GoRoutine
// User can join or leave the room
// User can send Messgage
// Room will broadcast the message to every users in the room

type Room struct {
	ID string
	Clients map[string]*Client // userId -> Client
	Broadcast chan ChatMessage 
	Register chan *Client
	Unregister chan *Client 
}

func NewRoom(id string) *Room {
	return &Room{
		ID: id,
		Clients: make(map[string]*Client),
		Broadcast: make(chan ChatMessage),
		Register: make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (r *Room) Run () {
	for {
		select {
		// Handle User Join 
		case client := <- r.Register:
			r.Clients[client.ID] = client
			log.Printf("[%s] %s joined", client.ID, client.Username)
		// Handle User Leave
		case client := <-r.Unregister:
			delete(r.Clients, client.ID)

			// Prevent client from receivign message
			close(client.Send)
			log.Printf("[%s] %s left", client.ID, client.Username)
		
		// Handle Message broadcast
		case message := <-r.Broadcast:
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
}