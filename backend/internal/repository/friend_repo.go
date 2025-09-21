package repository

import (
	"context"
	"errors"

	"chat-app/internal/domain"
)

var (
	ErrFriendRequestExists = errors.New("friend request already exists")
	ErrNotFound            = errors.New("resource not found")
)

type FriendRepository interface {
	CreateRequest(ctx context.Context, req *domain.FriendRequest) error
	GetRequestByID(ctx context.Context, id string) (*domain.FriendRequest, error)
	UpdateRequestStatus(ctx context.Context, id, status string) error
	AreFriends(ctx context.Context, userID1, userID2 string) (bool, error)
	AddFriendship(ctx context.Context, userID1, userID2 string) error
	RemoveFriendship(ctx context.Context, userID1, userID2 string) error
	GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendRequest, error)
	GetFriendsByUserID(ctx context.Context, userID string) ([]*domain.User, error)
	HasPendingRequest(ctx context.Context, userID1, userID2 string) (bool, error)
}

