package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	RefreshToken uuid.UUID `json:"refreshToken"`
	UserID       uuid.UUID `json:"userId"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

