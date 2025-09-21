package usecase

import (
	"chat-app/internal/domain"
	"context"
	"mime/multipart"
)

type CreateGroupInput struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
}

type UpdateGroupInput struct {
	Name *string `json:"name"`
}

type AddGroupMemberInput struct {
	UserID string `json:"user_id"`
}

type TransferOwnershipInput struct {
	NewOwnerID string `json:"new_owner_id"`
}

type GroupDetails struct {
	*domain.Group
	Members []*domain.GroupMember `json:"members"`
}

type GroupUsecase interface {
	CreateGroup(ctx context.Context, userID string, input CreateGroupInput, file multipart.File, fileHeader *multipart.FileHeader) (*domain.Group, error)
	GetGroupDetails(ctx context.Context, groupID string) (*GroupDetails, error)
	UpdateGroup(ctx context.Context, userID, groupID string, input UpdateGroupInput, file multipart.File, fileHeader *multipart.FileHeader) (*domain.Group, error)
	SearchGroups(ctx context.Context, query string) ([]*domain.Group, error)
	JoinGroup(ctx context.Context, userID, groupHandle string) (*domain.Group, error)
	LeaveGroup(ctx context.Context, userID, groupID string) error
	AddMember(ctx context.Context, currentUserID, groupID, friendID string) error
	RemoveMember(ctx context.Context, ownerID, groupID, memberID string) error
	TransferOwnership(ctx context.Context, currentOwnerID, groupID, newOwnerID string) error
	ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error)
}

