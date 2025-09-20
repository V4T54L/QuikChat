package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	// Messaging
	EventMessageSent EventType = "message_sent"
	EventMessageAck  EventType = "message_ack"

	// Friend Management
	EventFriendRequestReceived EventType = "friend_request_received"
	EventFriendRequestAccepted EventType = "friend_request_accepted"
	EventFriendRequestRejected EventType = "friend_request_rejected"
	EventUnfriended            EventType = "unfriended"

	// Group Management
	EventAddedToGroup     EventType = "added_to_group"
	EventRemovedFromGroup EventType = "removed_from_group"
	EventUserJoinedGroup  EventType = "user_joined_group"
	EventUserLeftGroup    EventType = "user_left_group"
)

type Event struct {
	ID          uuid.UUID       `json:"id"`
	Type        EventType       `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	RecipientID uuid.UUID       `json:"-"`
	CreatedAt   time.Time       `json:"createdAt"`
	SenderID    *uuid.UUID      `json:"senderId,omitempty"` // Optional, for messages etc.
}

