package domain

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	RecipientID string          `json:"recipient_id"`
	IsDelivered bool            `json:"is_delivered"`
	CreatedAt   time.Time       `json:"created_at"`
}

