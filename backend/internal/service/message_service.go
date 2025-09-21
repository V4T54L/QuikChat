package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"chat-app/backend/internal/domain"
	"chat-app/backend/internal/repository"
	"chat-app/backend/internal/usecase"
	"chat-app/backend/pkg/util"
)

var (
	ErrMessageTooLong      = errors.New("message content exceeds 200 characters")
	ErrInvalidConversation = errors.New("invalid conversation or user not a member")
)

type messageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	groupRepo   repository.GroupRepository
	eventUsecase  usecase.EventUsecase
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
	groupRepo repository.GroupRepository,
	eventUsecase usecase.EventUsecase,
) usecase.MessageUsecase {
	return &messageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		eventUsecase:  eventUsecase,
	}
}

func (s *messageService) SendMessage(ctx context.Context, senderID string, input usecase.SendMessageInput) (*domain.Message, error) {
	if utf8.RuneCountInString(input.Content) > 200 {
		return nil, ErrMessageTooLong
	}

	recipients, err := s.getConversationRecipients(ctx, senderID, input.ConversationID)
	if err != nil {
		return nil, err
	}

	sender, err := s.userRepo.GetByID(ctx, senderID)
	if err != nil || sender == nil {
		return nil, ErrUserNotFound
	}

	message := &domain.Message{
		ID:             util.NewUUID(),
		ConversationID: input.ConversationID,
		SenderID:       senderID,
		Content:        input.Content,
		CreatedAt:      time.Now().UTC(),
		Sender: &domain.User{
			ID:            sender.ID,
			Username:      sender.Username,
			ProfilePicURL: sender.ProfilePicURL,
		},
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	for _, recipientID := range recipients {
		s.eventUsecase.CreateAndBufferEvent(ctx, "new_message", message, recipientID)
	}

	return message, nil
}

func (s *messageService) GetMessageHistory(ctx context.Context, userID, conversationID string, before time.Time, limit int) ([]*domain.Message, error) {
	isMember, err := s.isUserConversationMember(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrInvalidConversation
	}

	return s.messageRepo.GetByConversationID(ctx, conversationID, before, limit)
}

func (s *messageService) getConversationRecipients(ctx context.Context, senderID, conversationID string) ([]string, error) {
	// Check if it's a group conversation
	if _, err := util.ParseUUID(conversationID); err == nil {
		group, err := s.groupRepo.GetByID(ctx, conversationID)
		if err != nil {
			return nil, ErrInvalidConversation
		}
		members, err := s.groupRepo.GetMembersWithUserDetails(ctx, group.ID)
		if err != nil {
			return nil, err
		}

		var recipientIDs []string
		isSenderMember := false
		for _, member := range members {
			if member.UserID == senderID {
				isSenderMember = true
			} else {
				recipientIDs = append(recipientIDs, member.UserID)
			}
		}

		if !isSenderMember {
			return nil, ErrInvalidConversation
		}
		return recipientIDs, nil
	}

	// Assume it's a P2P conversation
	ids := strings.Split(conversationID, ":")
	if len(ids) != 2 {
		return nil, ErrInvalidConversation
	}
	user1ID, user2ID := ids[0], ids[1]

	if senderID != user1ID && senderID != user2ID {
		return nil, ErrInvalidConversation
	}

	var recipientID string
	if senderID == user1ID {
		recipientID = user2ID
	} else {
		recipientID = user1ID
	}

	return []string{recipientID}, nil
}

func (s *messageService) isUserConversationMember(ctx context.Context, userID, conversationID string) (bool, error) {
	// Check if it's a group conversation
	if _, err := util.ParseUUID(conversationID); err == nil {
		member, err := s.groupRepo.FindMember(ctx, conversationID, userID)
		if err != nil && err != repository.ErrNotFound {
			return false, err
		}
		return member != nil, nil
	}

	// Assume it's a P2P conversation
	ids := strings.Split(conversationID, ":")
	if len(ids) != 2 {
		return false, nil // Invalid format, so not a member
	}
	return userID == ids[0] || userID == ids[1], nil
}

