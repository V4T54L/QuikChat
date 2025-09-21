package websocket

import (
	"chat-app/internal/usecase"
	"context"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	eventUsecase usecase.EventUsecase
}

func NewHub(eventUsecase usecase.EventUsecase) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		eventUsecase: eventUsecase,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
			log.Printf("Client connected: %s. Total clients: %d", client.userID, len(h.clients))
			// TODO: Fetch and send undelivered events
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("Client disconnected: %s. Total clients: %d", client.userID, len(h.clients))
			}
		case message := <-h.broadcast:
			// This is a simple broadcast. A real app would target specific users.
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
		}
	}
}

// BroadcastEvent sends an event to a specific user if they are online,
// otherwise it buffers the event.
func (h *Hub) BroadcastEvent(ctx context.Context, eventType string, payload interface{}, recipientID string) {
	if client, ok := h.clients[recipientID]; ok {
		// User is online, send directly
		client.SendEvent(map[string]interface{}{"type": eventType, "payload": payload})
	} else {
		// User is offline, buffer the event
		err := h.eventUsecase.CreateAndBufferEvent(ctx, eventType, payload, recipientID)
		if err != nil {
			log.Printf("Error buffering event for offline user %s: %v", recipientID, err)
		}
	}
}

