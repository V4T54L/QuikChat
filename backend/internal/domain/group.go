package domain

import "time"

const (
	GroupRoleOwner  = "owner"
	GroupRoleMember = "member"
)

type Group struct {
	ID            string    `json:"id"`
	Handle        string    `json:"handle"`
	Name          string    `json:"name"`
	OwnerID       string    `json:"owner_id"`
	ProfilePicURL string    `json:"profile_pic_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type GroupMember struct {
	GroupID   string    `json:"group_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	User      *User     `json:"user,omitempty"` // For joining
}

