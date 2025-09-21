package repository

import (
	"context"
	"chat-app/internal/domain"
)

type SessionRepository interface {
	Store(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}

