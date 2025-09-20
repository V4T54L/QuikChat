package repository

import (
	"chat-app/backend/models"
	"context"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	Find(ctx context.Context, refreshToken uuid.UUID) (*models.Session, error)
	Delete(ctx context.Context, refreshToken uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

