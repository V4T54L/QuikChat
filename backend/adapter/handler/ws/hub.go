package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"chat-app/models"
	"chat-app/usecase"
	"github.com/google/uuid"
)

// ClientMessage is a message from a client to the hub.
type ClientMessage struct {
	client  *Client
	message []byte
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[uuid.UUID]*Client
	// Inbound messages from the clients.
	broadcast chan *ClientMessage
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
	// Event usecase
	eventUsecase usecase.EventUsecase
	groupUsecase usecase.GroupUsecase
}

func NewHub(eventUsecase usecase.EventUsecase, groupUsecase usecase.GroupUsecase) *Hub {
	return &Hub{
		broadcast:    make(chan *ClientMessage),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		clients:      make(map[uuid.UUID]*Client),
		eventUsecase: eventUsecase,
		groupUsecase: groupUsecase,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userID] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
		case clientMessage := <-h.broadcast:
			h.handleMessage(clientMessage.client, clientMessage.message)
		}
	}
}

func (h *Hub) handleMessage(sender *Client, rawMessage []byte) {
	var msg Message
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("error unmarshalling message: %v", err)
		return
	}

	switch msg.Type {
	case string(models.EventMessageSent):
		var inbound InboundMessage
		if err := json.Unmarshal(msg.Payload, &inbound); err != nil {
			log.Printf("error unmarshalling inbound message payload: %v", err)
			return
		}
		h.processAndRelayMessage(sender.userID, inbound)
	}
}

func (h *Hub) processAndRelayMessage(senderID uuid.UUID, inbound InboundMessage) {
	// Create an event for the message
	outboundPayload := OutboundMessage{
		ID:          uuid.New(),
		Content:     inbound.Content,
		SenderID:    senderID,
		RecipientID: inbound.RecipientID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	payloadBytes, _ := json.Marshal(outboundPayload)

	// Check if recipient is a group or a user
	members, err := h.groupUsecase.ListGroupMembers(context.Background(), inbound.RecipientID)
	isGroup := err == nil && len(members) > 0

	var recipients []uuid.UUID
	if isGroup {
		for _, member := range members {
			if member.ID != senderID { // Don't send to self
				recipients = append(recipients, member.ID)
			}
		}
	} else {
		recipients = append(recipients, inbound.RecipientID)
	}

	// Store and send event to all recipients
	for _, recipientID := range recipients {
		event := &models.Event{
			ID:          uuid.New(),
			Type:        models.EventMessageSent,
			Payload:     payloadBytes,
			RecipientID: recipientID,
			SenderID:    &senderID,
			CreatedAt:   time.Now().UTC(),
		}
		if err := h.eventUsecase.StoreEvent(context.Background(), event); err != nil {
			log.Printf("failed to store event: %v", err)
			continue
		}
		h.DeliverEvent(event)
	}

	// Send acknowledgment back to sender
	ackEvent := &models.Event{
		ID:          outboundPayload.ID, // Use message ID for ack
		Type:        models.EventMessageAck,
		Payload:     payloadBytes,
		RecipientID: senderID,
		CreatedAt:   time.Now().UTC(),
	}
	h.DeliverEvent(ackEvent)
}

// DeliverEvent sends a single event to a connected client if they are online.
func (h *Hub) DeliverEvent(event *models.Event) {
	if client, ok := h.clients[event.RecipientID]; ok {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("error marshalling event for delivery: %v", err)
			return
		}
		select {
		case client.send <- data:
		default:
			// Client's send buffer is full, assume they are lagging and disconnect.
			close(client.send)
			delete(h.clients, client.userID)
		}
	}
}

