package domain

import "time"

type User struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	PasswordHash  string    `json:"-"` // Do not expose password hash
	ProfilePicURL string    `json:"profile_pic_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

