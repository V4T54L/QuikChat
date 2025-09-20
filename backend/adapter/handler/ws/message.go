package ws

import (
	"encoding/json"
	"github.com/google/uuid"
)

// Message represents a message sent over the WebSocket connection.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// InboundMessage represents a message received from a client.
type InboundMessage struct {
	Content     string    `json:"content"`
	RecipientID uuid.UUID `json:"recipientId"` // Can be a user ID or group ID
}

// OutboundMessage represents a message sent to a client.
type OutboundMessage struct {
	ID          uuid.UUID `json:"id"`
	Content     string    `json:"content"`
	SenderID    uuid.UUID `json:"senderId"`
	RecipientID uuid.UUID `json:"recipientId"` // Can be a user ID or group ID
	Timestamp   string    `json:"timestamp"`
}

