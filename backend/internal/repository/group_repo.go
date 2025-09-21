package repository

import (
	"chat-app/internal/domain"
	"context"
)

var (
	ErrGroupHandleExists = NewRepositoryError("group handle already exists")
	ErrGroupMemberExists = NewRepositoryError("user is already a member of this group")
)

type GroupRepository interface {
	Create(ctx context.Context, group *domain.Group) error
	CreateMember(ctx context.Context, member *domain.GroupMember) error
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	GetByHandle(ctx context.Context, handle string) (*domain.Group, error)
	Update(ctx context.Context, group *domain.Group) error
	FindMember(ctx context.Context, groupID, userID string) (*domain.GroupMember, error)
	RemoveMember(ctx context.Context, groupID, userID string) error
	GetMembersWithUserDetails(ctx context.Context, groupID string) ([]*domain.GroupMember, error)
	GetGroupsByUserID(ctx context.Context, userID string) ([]*domain.Group, error)
	SearchByHandle(ctx context.Context, query string) ([]*domain.Group, error)
	GetOldestMember(ctx context.Context, groupID, excludeUserID string) (*domain.GroupMember, error)
}

