package models

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID        uuid.UUID `json:"id"`
	Handle    string    `json:"handle"`
	Name      string    `json:"name"`
	PhotoURL  string    `json:"photoUrl"`
	OwnerID   uuid.UUID `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

type GroupMember struct {
	GroupID  uuid.UUID `json:"groupId"`
	UserID   uuid.UUID `json:"userId"`
	JoinedAt time.Time `json:"joinedAt"`
}

