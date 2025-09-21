package domain

import "time"

type FriendRequest struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Status     string    `json:"status"` // e.g., "pending", "accepted", "rejected"
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Fields for joining with users table
	Sender   *User `json:"sender,omitempty"`
	Receiver *User `json:"receiver,omitempty"`
}

const (
	FriendRequestStatusPending  = "pending"
	FriendRequestStatusAccepted = "accepted"
	FriendRequestStatusRejected = "rejected"
)

