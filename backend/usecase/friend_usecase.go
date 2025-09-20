package usecase

import (
	"context"
	"encoding/json"
	"time"

	"chat-app/backend/models"
	"chat-app/backend/repository"

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
	userRepo     repository.UserRepository
	friendRepo   repository.FriendshipRepository
	eventUsecase EventUsecase
}

func NewFriendUsecase(userRepo repository.UserRepository, friendRepo repository.FriendshipRepository, eventUsecase EventUsecase) FriendUsecase {
	return &friendUsecase{
		userRepo:     userRepo,
		friendRepo:   friendRepo,
		eventUsecase: eventUsecase,
	}
}

func (u *friendUsecase) SendRequest(ctx context.Context, fromUserID uuid.UUID, toUsername string) error {
	if fromUserID.String() == toUsername {
		return models.ErrCannotFriendSelf
	}

	toUser, err := u.userRepo.FindByUsername(ctx, toUsername)
	if err != nil {
		return err
	}

	if fromUserID == toUser.ID {
		return models.ErrCannotFriendSelf
	}

	// Check if a friendship or request already exists
	_, err = u.friendRepo.Find(ctx, fromUserID, toUser.ID)
	if err == nil {
		return models.ErrFriendRequestExists
	} else if err != models.ErrFriendRequestNotFound {
		return err
	}

	// Create friendship request
	friendship := &models.Friendship{
		UserID1:   fromUserID,
		UserID2:   toUser.ID,
		Status:    models.FriendshipStatusPending,
		CreatedAt: time.Now(),
	}
	if err := u.friendRepo.Create(ctx, friendship); err != nil {
		return err
	}

	// Create and store event for the recipient
	fromUser, err := u.userRepo.FindByID(ctx, fromUserID)
	if err != nil {
		return err // Should not happen if fromUserID is from a valid token
	}

	payload, _ := json.Marshal(fromUser)
	event := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventFriendRequestReceived,
		Payload:     payload,
		RecipientID: toUser.ID,
		SenderID:    &fromUserID,
		CreatedAt:   time.Now().UTC(),
	}

	return u.eventUsecase.StoreEvent(ctx, event)
}

func (u *friendUsecase) AcceptRequest(ctx context.Context, userID, requesterID uuid.UUID) error {
	// Ensure a pending request exists from requesterID to userID
	fs, err := u.friendRepo.Find(ctx, requesterID, userID)
	if err != nil {
		return err
	}
	if fs.Status != models.FriendshipStatusPending {
		return models.ErrFriendRequestNotFound
	}

	if err := u.friendRepo.UpdateStatus(ctx, requesterID, userID, models.FriendshipStatusAccepted); err != nil {
		return err
	}

	// Create and store event for the original requester
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(user)
	event := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventFriendRequestAccepted,
		Payload:     payload,
		RecipientID: requesterID,
		SenderID:    &userID,
		CreatedAt:   time.Now().UTC(),
	}

	return u.eventUsecase.StoreEvent(ctx, event)
}

func (u *friendUsecase) RejectRequest(ctx context.Context, userID, requesterID uuid.UUID) error {
	// Ensure a pending request exists from requesterID to userID
	fs, err := u.friendRepo.Find(ctx, requesterID, userID)
	if err != nil {
		return err
	}
	if fs.Status != models.FriendshipStatusPending {
		return models.ErrFriendRequestNotFound
	}

	if err := u.friendRepo.Delete(ctx, requesterID, userID); err != nil {
		return err
	}

	// Create and store event for the original requester
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"username": user.Username})
	event := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventFriendRequestRejected,
		Payload:     payload,
		RecipientID: requesterID,
		SenderID:    &userID,
		CreatedAt:   time.Now().UTC(),
	}

	return u.eventUsecase.StoreEvent(ctx, event)
}

func (u *friendUsecase) Unfriend(ctx context.Context, userID, friendID uuid.UUID) error {
	// Ensure they are friends
	fs, err := u.friendRepo.Find(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if fs.Status != models.FriendshipStatusAccepted {
		return models.ErrNotFriends
	}

	if err := u.friendRepo.Delete(ctx, userID, friendID); err != nil {
		return err
	}

	// Create and store event for the unfriended user
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"username": user.Username})
	event := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventUnfriended,
		Payload:     payload,
		RecipientID: friendID,
		SenderID:    &userID,
		CreatedAt:   time.Now().UTC(),
	}

	return u.eventUsecase.StoreEvent(ctx, event)
}

func (u *friendUsecase) ListFriends(ctx context.Context, userID uuid.UUID) ([]*models.User, error) {
	return u.friendRepo.ListByUserID(ctx, userID, models.FriendshipStatusAccepted)
}

func (u *friendUsecase) ListPendingRequests(ctx context.Context, userID uuid.UUID) ([]*models.User, error) {
	return u.friendRepo.ListByUserID(ctx, userID, models.FriendshipStatusPending)
}
