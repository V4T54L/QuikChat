package models

import (
	"time"

	"github.com/google/uuid"
)

type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
)

type Friendship struct {
	UserID1   uuid.UUID        `json:"userId1"`
	UserID2   uuid.UUID        `json:"userId2"`
	Status    FriendshipStatus `json:"status"`
	CreatedAt time.Time        `json:"createdAt"`
}

