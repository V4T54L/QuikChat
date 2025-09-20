package usecase

import (
	"context"

	"chat-app/models"
	"chat-app/repository"

	"github.com/google/uuid"
)

type FriendUsecase interface {
	SendRequest(ctx context.Context, fromUserID uuid.UUID, toUsername string) error
	AcceptRequest(ctx context.Context, userID, requesterID uuid.UUID) error
	RejectRequest(ctx context.Context, userID, requesterID uuid.UUID) error
	Unfriend(ctx context.Context, userID, friendID uuid.UUID) error
	ListFriends(ctx context.Context, userID uuid.UUID) ([]*models.User, error)
	ListPendingRequests(ctx context.Context, userID uuid.UUID) ([]*models.User, error)
}

type friendUsecase struct {
	userRepo   repository.UserRepository
	friendRepo repository.FriendshipRepository
}

func NewFriendUsecase(userRepo repository.UserRepository, friendRepo repository.FriendshipRepository) FriendUsecase {
	return &friendUsecase{
		userRepo:   userRepo,
		friendRepo: friendRepo,
	}
}

func (u *friendUsecase) SendRequest(ctx context.Context, fromUserID uuid.UUID, toUsername string) error {
	toUser, err := u.userRepo.FindByUsername(ctx, toUsername)
	if err != nil {
		return err
	}

	if fromUserID == toUser.ID {
		return models.ErrCannotFriendSelf
	}

	// Check if a friendship or request already exists
	existing, err := u.friendRepo.Find(ctx, fromUserID, toUser.ID)
	if err != nil && err != models.ErrFriendRequestNotFound {
		return err
	}
	if existing != nil {
		if existing.Status == models.FriendshipStatusAccepted {
			return models.ErrAlreadyFriends
		}
		return models.ErrFriendRequestExists
	}

	friendship := &models.Friendship{
		UserID1: fromUserID,
		UserID2: toUser.ID,
		Status:  models.FriendshipStatusPending,
	}

	return u.friendRepo.Create(ctx, friendship)
}

func (u *friendUsecase) AcceptRequest(ctx context.Context, userID, requesterID uuid.UUID) error {
	friendship, err := u.friendRepo.Find(ctx, userID, requesterID)
	if err != nil {
		return err
	}

	if friendship.Status != models.FriendshipStatusPending {
		return models.ErrFriendRequestNotFound
	}

	if friendship.UserID1 != userID && friendship.UserID2 != userID {
		return models.ErrUnauthorized
	}

	return u.friendRepo.UpdateStatus(ctx, userID, requesterID, models.FriendshipStatusAccepted)
}

func (u *friendUsecase) RejectRequest(ctx context.Context, userID, requesterID uuid.UUID) error {
	friendship, err := u.friendRepo.Find(ctx, userID, requesterID)
	if err != nil {
		return err
	}

	if friendship.Status != models.FriendshipStatusPending {
		return models.ErrFriendRequestNotFound
	}

	if friendship.UserID1 != userID && friendship.UserID2 != userID {
		return models.ErrUnauthorized
	}

	return u.friendRepo.Delete(ctx, userID, requesterID)
}

func (u *friendUsecase) Unfriend(ctx context.Context, userID, friendID uuid.UUID) error {
	friendship, err := u.friendRepo.Find(ctx, userID, friendID)
	if err != nil {
		return err
	}

	if friendship.Status != models.FriendshipStatusAccepted {
		return models.ErrNotFriends
	}

	return u.friendRepo.Delete(ctx, userID, friendID)
}

func (u *friendUsecase) ListFriends(ctx context.Context, userID uuid.UUID) ([]*models.User, error) {
	return u.friendRepo.ListByUserID(ctx, userID, models.FriendshipStatusAccepted)
}

func (u *friendUsecase) ListPendingRequests(ctx context.Context, userID uuid.UUID) ([]*models.User, error) {
	return u.friendRepo.ListByUserID(ctx, userID, models.FriendshipStatusPending)
}

