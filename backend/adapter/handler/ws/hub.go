package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"chat-app/backend/models"
	"chat-app/backend/usecase"
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
	mu           sync.RWMutex
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
			h.mu.Lock()
			h.clients[client.userID] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
		case clientMessage := <-h.broadcast:
			h.handleMessage(clientMessage.client, clientMessage.message)
		}
	}
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) handleMessage(sender *Client, rawMessage []byte) {
	var msg Message
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("error unmarshalling message: %v", err)
		return
	}

	switch msg.Type {
	case "message_sent":
		var inbound InboundMessage
		if err := json.Unmarshal(msg.Payload, &inbound); err != nil {
			log.Printf("error unmarshalling inbound message payload: %v", err)
			return
		}

		content := strings.TrimSpace(inbound.Content)
		if len(content) == 0 || len(content) > 200 {
			log.Printf("invalid message length from user %s", sender.userID)
			return
		}
		inbound.Content = content

		h.processAndRelayMessage(sender.userID, inbound)
	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

func (h *Hub) processAndRelayMessage(senderID uuid.UUID, inbound InboundMessage) {
	ctx := context.Background()
	outboundPayload := OutboundMessage{
		ID:          uuid.New(),
		Content:     inbound.Content,
		SenderID:    senderID,
		RecipientID: inbound.RecipientID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	payloadBytes, _ := json.Marshal(outboundPayload)

	var recipients []uuid.UUID

	// Check if recipient is a group or a user
	group, err := h.groupUsecase.GetGroupDetails(ctx, inbound.RecipientID)
	if err == nil && group != nil {
		members, err := h.groupUsecase.ListGroupMembers(ctx, group.ID)
		if err != nil {
			log.Printf("error listing group members: %v", err)
			return
		}
		for _, member := range members {
			recipients = append(recipients, member.ID)
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
		if err := h.eventUsecase.StoreEvent(ctx, event); err != nil {
			log.Printf("failed to store event: %v", err)
			continue
		}
		h.DeliverEvent(event)
	}

	// Send acknowledgment back to sender
	ackPayload, err := json.Marshal(map[string]string{"messageId": outboundPayload.ID.String()})
	if err != nil {
		log.Printf("error marshalling ack payload: %v", err)
		return
	}

	ackEvent := &models.Event{
		ID:          uuid.New(), // New ID for the ack event itself
		Type:        models.EventMessageAck,
		Payload:     ackPayload,
		RecipientID: senderID,
		CreatedAt:   time.Now().UTC(),
	}
	h.DeliverEvent(ackEvent)
}

// DeliverEvent sends a single event to a connected client if they are online.
func (h *Hub) DeliverEvent(event *models.Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshalling event for delivery: %v", err)
		return
	}

	h.mu.RLock()
	client, ok := h.clients[event.RecipientID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.send <- payload:
		default:
			// Client's send buffer is full, assume it's lagging and disconnect.
			log.Printf("client %s lagging, disconnecting", client.userID)
			h.unregister <- client
		}
	}
}

