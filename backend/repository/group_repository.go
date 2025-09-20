package repository

import (
	"chat-app/backend/models"
	"context"

	"github.com/google/uuid"
)

type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	Update(ctx context.Context, group *models.Group) error
	Delete(ctx context.Context, groupID uuid.UUID) error
	FindByID(ctx context.Context, groupID uuid.UUID) (*models.Group, error)
	FindByHandle(ctx context.Context, handle string) (*models.Group, error)
	FuzzySearchByHandle(ctx context.Context, query string, limit int) ([]*models.Group, error)

	AddMember(ctx context.Context, member *models.GroupMember) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	FindMember(ctx context.Context, groupID, userID uuid.UUID) (*models.GroupMember, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]*models.User, error)
	GetOldestMember(ctx context.Context, groupID uuid.UUID) (*models.User, error)
}
