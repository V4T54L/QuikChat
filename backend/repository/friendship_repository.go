package repository

import (
	"context"
	"chat-app/models"

	"github.com/google/uuid"
)

type FriendshipRepository interface {
	Create(ctx context.Context, friendship *models.Friendship) error
	UpdateStatus(ctx context.Context, userID1, userID2 uuid.UUID, status models.FriendshipStatus) error
	Delete(ctx context.Context, userID1, userID2 uuid.UUID) error
	Find(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Friendship, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, status models.FriendshipStatus) ([]*models.User, error)
}

