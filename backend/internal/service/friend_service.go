package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"chat-app/internal/usecase"
	"chat-app/pkg/util"
)

var (
	ErrFriendRequestYourself    = errors.New("cannot send friend request to yourself")
	ErrAlreadyFriends           = errors.New("users are already friends")
	ErrFriendRequestExists      = errors.New("a pending friend request already exists")
	ErrFriendRequestInvalid     = errors.New("invalid friend request")
	ErrFriendRequestNotReceiver = errors.New("only the receiver can accept or reject a friend request")
)

type friendService struct {
	friendRepo repository.FriendRepository
	userRepo   repository.UserRepository
	eventUcase usecase.EventUsecase
}

func NewFriendService(friendRepo repository.FriendRepository, userRepo repository.UserRepository, eventUsecase usecase.EventUsecase) usecase.FriendUsecase {
	return &friendService{
		friendRepo: friendRepo,
		userRepo:   userRepo,
		eventUcase: eventUsecase,
	}
}

func (s *friendService) SendFriendRequest(ctx context.Context, senderID, receiverUsername string) (*domain.FriendRequest, error) {
	sender, err := s.userRepo.GetByID(ctx, senderID)
	if err != nil || sender == nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	receiver, err := s.userRepo.GetByUsername(ctx, receiverUsername)
	if err != nil {
		return nil, fmt.Errorf("error fetching receiver: %w", err)
	}
	if receiver == nil {
		return nil, ErrUserNotFound
	}

	if senderID == receiver.ID {
		return nil, ErrFriendRequestYourself
	}

	areFriends, err := s.friendRepo.AreFriends(ctx, senderID, receiver.ID)
	if err != nil {
		return nil, err
	}
	if areFriends {
		return nil, ErrAlreadyFriends
	}

	hasPending, err := s.friendRepo.HasPendingRequest(ctx, senderID, receiver.ID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrFriendRequestExists
	}

	req := &domain.FriendRequest{
		ID:         util.NewUUID(),
		SenderID:   senderID,
		ReceiverID: receiver.ID,
		Status:     domain.FriendRequestStatusPending,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.friendRepo.CreateRequest(ctx, req); err != nil {
		if errors.Is(err, repository.ErrFriendRequestExists) {
			return nil, ErrFriendRequestExists
		}
		return nil, err
	}

	// Create event for receiver
	eventPayload := map[string]interface{}{
		"request_id": req.ID,
		"sender": map[string]string{
			"id":              sender.ID,
			"username":        sender.Username,
			"profile_pic_url": sender.ProfilePicURL,
		},
	}
	s.eventUcase.CreateAndBufferEvent(ctx, "friend_request_received", eventPayload, receiver.ID)

	req.Sender = sender
	req.Receiver = receiver
	return req, nil
}

func (s *friendService) AcceptFriendRequest(ctx context.Context, userID, requestID string) error {
	req, err := s.friendRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrFriendRequestInvalid
		}
		return err
	}

	if req.ReceiverID != userID {
		return ErrFriendRequestNotReceiver
	}

	if req.Status != domain.FriendRequestStatusPending {
		return ErrFriendRequestInvalid
	}

	if err := s.friendRepo.AddFriendship(ctx, req.SenderID, req.ReceiverID); err != nil {
		return err
	}

	if err := s.friendRepo.UpdateRequestStatus(ctx, requestID, domain.FriendRequestStatusAccepted); err != nil {
		// Attempt to rollback friendship, but don't fail the whole operation if this fails
		_ = s.friendRepo.RemoveFriendship(ctx, req.SenderID, req.ReceiverID)
		return err
	}

	// Create event for sender
	receiver, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error fetching receiver user for event: %v\n", err)
	} else {
		eventPayload := map[string]interface{}{
			"request_id": requestID,
			"receiver": map[string]string{
				"id":              receiver.ID,
				"username":        receiver.Username,
				"profile_pic_url": receiver.ProfilePicURL,
			},
		}
		s.eventUcase.CreateAndBufferEvent(ctx, "friend_request_accepted", eventPayload, req.SenderID)
	}

	return nil
}

func (s *friendService) RejectFriendRequest(ctx context.Context, userID, requestID string) error {
	req, err := s.friendRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrFriendRequestInvalid
		}
		return err
	}

	if req.ReceiverID != userID && req.SenderID != userID {
		return ErrFriendRequestNotReceiver // Or a more generic "not authorized"
	}

	if req.Status != domain.FriendRequestStatusPending {
		return ErrFriendRequestInvalid
	}

	if err := s.friendRepo.UpdateRequestStatus(ctx, requestID, domain.FriendRequestStatusRejected); err != nil {
		return err
	}

	// Create event for the other user
	var otherUserID string
	if userID == req.SenderID {
		otherUserID = req.ReceiverID
	} else {
		otherUserID = req.SenderID
	}

	rejecter, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		fmt.Printf("Error fetching rejecting user for event: %v\n", err)
	} else {
		eventPayload := map[string]interface{}{
			"request_id": requestID,
			"user": map[string]string{
				"id":       rejecter.ID,
				"username": rejecter.Username,
			},
		}
		s.eventUcase.CreateAndBufferEvent(ctx, "friend_request_rejected", eventPayload, otherUserID)
	}

	return nil
}

func (s *friendService) Unfriend(ctx context.Context, userID, friendID string) error {
	areFriends, err := s.friendRepo.AreFriends(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if !areFriends {
		return nil // Idempotent
	}

	if err := s.friendRepo.RemoveFriendship(ctx, userID, friendID); err != nil {
		return err
	}

	// Create event for the unfriended user
	unfriender, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		fmt.Printf("Error fetching unfriender user for event: %v\n", err)
	} else {
		eventPayload := map[string]interface{}{
			"user": map[string]string{
				"id":       unfriender.ID,
				"username": unfriender.Username,
			},
		}
		s.eventUcase.CreateAndBufferEvent(ctx, "unfriended", eventPayload, friendID)
	}

	return nil
}

func (s *friendService) GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendRequest, error) {
	return s.friendRepo.GetPendingRequests(ctx, userID)
}

func (s *friendService) ListFriends(ctx context.Context, userID string) ([]*domain.User, error) {
	return s.friendRepo.GetFriendsByUserID(ctx, userID)
}

