package usecase

import (
	"context"

	"chat-app/internal/domain"
)

type FriendUsecase interface {
	SendFriendRequest(ctx context.Context, senderID, receiverUsername string) (*domain.FriendRequest, error)
	AcceptFriendRequest(ctx context.Context, userID, requestID string) error
	RejectFriendRequest(ctx context.Context, userID, requestID string) error
	Unfriend(ctx context.Context, userID, friendID string) error
	GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendRequest, error)
	ListFriends(ctx context.Context, userID string) ([]*domain.User, error)
}

