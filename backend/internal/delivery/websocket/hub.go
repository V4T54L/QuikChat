package websocket

import (
	"context"
	"encoding/json"
	"log"

	"chat-app/backend/internal/usecase"
)

// HubEvent represents an event received from a client via WebSocket.
type HubEvent struct {
	Client  *Client
	Type    string
	Payload json.RawMessage
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan *HubEvent // Changed to HubEvent

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	eventUsecase   usecase.EventUsecase
	messageUsecase usecase.MessageUsecase // Added messageUsecase
}

func NewHub(eventUsecase usecase.EventUsecase, messageUsecase usecase.MessageUsecase) *Hub { // Added messageUsecase parameter
	return &Hub{
		broadcast:      make(chan *HubEvent), // Changed to HubEvent
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[string]*Client),
		eventUsecase:   eventUsecase,
		messageUsecase: messageUsecase, // Initialized messageUsecase
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
			log.Printf("Client connected: %s. Total clients: %d", client.userID, len(h.clients)) // Kept original log
			// TODO: Fetch and send undelivered events
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				log.Printf("Client disconnected: %s. Total clients: %d", client.userID, len(h.clients)) // Kept original log
			}
		case event := <-h.broadcast: // Changed to HubEvent
			h.handleIncomingEvent(event) // Delegated to new handler
		}
	}
}

// handleIncomingEvent processes events received from WebSocket clients.
func (h *Hub) handleIncomingEvent(event *HubEvent) {
	ctx := context.Background()
	switch event.Type {
	case "send_message":
		var input usecase.SendMessageInput
		if err := json.Unmarshal(event.Payload, &input); err != nil {
			log.Printf("Error unmarshalling send_message payload: %v", err)
			// Optionally send an error event back to the client
			return
		}

		if _, err := h.messageUsecase.SendMessage(ctx, event.Client.userID, input); err != nil {
			log.Printf("Error from SendMessage usecase: %v", err)
			// Optionally send an error event back to the client
		}
	case "auth":
		// This can be used for initial auth or re-auth
		log.Printf("Client %s authenticated via WebSocket", event.Client.userID)
	default:
		log.Printf("Unknown incoming event type: %s", event.Type)
	}
}

// BroadcastEvent sends an event to a specific user if they are online,
// otherwise it buffers the event.
func (h *Hub) BroadcastEvent(ctx context.Context, eventType string, payload interface{}, recipientID string) {
	eventPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling event payload: %v", err)
		return
	}

	event := map[string]interface{}{
		"type":    eventType,
		"payload": json.RawMessage(eventPayload),
	}

	if client, ok := h.clients[recipientID]; ok {
		client.SendEvent(event)
	} else {
		if err := h.eventUsecase.CreateAndBufferEvent(ctx, eventType, payload, recipientID); err != nil {
			log.Printf("Error buffering event for offline user %s: %v", recipientID, err)
		}
	}
}
