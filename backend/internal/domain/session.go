package domain

import "time"

type Session struct {
	ID        string    `json:"id"` // This is the Refresh Token
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

