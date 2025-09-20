package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	ProfilePicURL string    `json:"profilePicUrl"`
	CreatedAt    time.Time `json:"createdAt"`
}

