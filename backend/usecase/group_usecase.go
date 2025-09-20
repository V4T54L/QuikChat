package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"time"

	"chat-app/backend/adapter/util"
	"chat-app/backend/models"
	"chat-app/backend/repository"

	"github.com/google/uuid"
)

type GroupUsecase interface {
	CreateGroup(ctx context.Context, ownerID uuid.UUID, handle, name string, photo multipart.File, photoHeader *multipart.FileHeader) (*models.Group, error)
	UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, name *string, photo multipart.File, photoHeader *multipart.FileHeader) (*models.Group, error)
	JoinGroup(ctx context.Context, userID uuid.UUID, groupHandle string) error
	LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error
	AddMember(ctx context.Context, adderID uuid.UUID, newMemberUsername string, groupID uuid.UUID) error
	RemoveMember(ctx context.Context, ownerID, memberID, groupID uuid.UUID) error
	TransferOwnership(ctx context.Context, currentOwnerID, newOwnerID, groupID uuid.UUID) error
	SearchGroups(ctx context.Context, query string) ([]*models.Group, error)
	GetGroupDetails(ctx context.Context, groupID uuid.UUID) (*models.Group, error)
	ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*models.User, error)
}

type groupUsecase struct {
	groupRepo    repository.GroupRepository
	userRepo     repository.UserRepository
	friendRepo   repository.FriendshipRepository
	fileRepo     repository.FileRepository
	eventUsecase EventUsecase
}

func NewGroupUsecase(groupRepo repository.GroupRepository, userRepo repository.UserRepository, friendRepo repository.FriendshipRepository, fileRepo repository.FileRepository, eventUsecase EventUsecase) GroupUsecase {
	return &groupUsecase{
		groupRepo:    groupRepo,
		userRepo:     userRepo,
		friendRepo:   friendRepo,
		fileRepo:     fileRepo,
		eventUsecase: eventUsecase,
	}
}

func (u *groupUsecase) CreateGroup(ctx context.Context, ownerID uuid.UUID, handle, name string, photo multipart.File, photoHeader *multipart.FileHeader) (*models.Group, error) {
	if err := util.ValidateGroupHandle(handle); err != nil {
		return nil, err
	}
	if name == "" || len(name) > 100 {
		return nil, fmt.Errorf("group name must be between 1 and 100 characters: %w", models.ErrBadRequest)
	}

	var photoURL string
	if photo != nil && photoHeader != nil {
		if err := util.ValidateProfilePic(photoHeader); err != nil {
			return nil, err
		}
		url, err := u.fileRepo.Save(photo, photoHeader)
		if err != nil {
			return nil, err
		}
		photoURL = url
	}

	group := &models.Group{
		ID:        uuid.New(),
		Handle:    handle,
		Name:      name,
		PhotoURL:  photoURL,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}

	if err := u.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (u *groupUsecase) UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, name *string, photo multipart.File, photoHeader *multipart.FileHeader) (*models.Group, error) {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.OwnerID != userID {
		return nil, models.ErrNotGroupOwner
	}

	if name != nil {
		if *name == "" || len(*name) > 100 {
			return nil, fmt.Errorf("group name must be between 1 and 100 characters: %w", models.ErrBadRequest)
		}
		group.Name = *name
	}

	if photo != nil && photoHeader != nil {
		if err := util.ValidateProfilePic(photoHeader); err != nil {
			return nil, err
		}
		url, err := u.fileRepo.Save(photo, photoHeader)
		if err != nil {
			return nil, err
		}
		group.PhotoURL = url
	}

	if err := u.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (u *groupUsecase) JoinGroup(ctx context.Context, userID uuid.UUID, groupHandle string) error {
	group, err := u.groupRepo.FindByHandle(ctx, groupHandle)
	if err != nil {
		return err
	}

	member := &models.GroupMember{
		GroupID:  group.ID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := u.groupRepo.AddMember(ctx, member); err != nil {
		return err
	}

	// Notify other group members
	go u.notifyGroupMembers(context.Background(), group.ID, userID, models.EventUserJoinedGroup, map[string]interface{}{
		"groupId":   group.ID,
		"groupName": group.Name,
		"userId":    userID,
	})

	return nil
}

func (u *groupUsecase) LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if _, err := u.groupRepo.FindMember(ctx, groupID, userID); err != nil {
		return models.ErrNotGroupMember
	}

	members, err := u.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return err
	}

	if err := u.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return err
	}

	// Notify remaining members
	go u.notifyGroupMembers(context.Background(), groupID, userID, models.EventUserLeftGroup, map[string]interface{}{
		"groupId":   groupID,
		"groupName": group.Name,
		"userId":    userID,
	})

	// Handle ownership transfer or group deletion
	if group.OwnerID == userID {
		if len(members) == 1 { // The owner was the last member
			return u.groupRepo.Delete(ctx, groupID)
		}

		oldestMember, err := u.groupRepo.GetOldestMember(ctx, groupID)
		if err != nil {
			// This case should be rare, but if it happens, delete the group
			u.groupRepo.Delete(ctx, groupID)
			return err
		}
		group.OwnerID = oldestMember.ID
		return u.groupRepo.Update(ctx, group)
	}

	return nil
}

func (u *groupUsecase) AddMember(ctx context.Context, adderID uuid.UUID, newMemberUsername string, groupID uuid.UUID) error {
	// Check if adder is a member
	if _, err := u.groupRepo.FindMember(ctx, groupID, adderID); err != nil {
		return models.ErrNotGroupMember
	}

	newMember, err := u.userRepo.FindByUsername(ctx, newMemberUsername)
	if err != nil {
		return err
	}

	// Check if they are friends
	fs, err := u.friendRepo.Find(ctx, adderID, newMember.ID)
	if err != nil || fs.Status != models.FriendshipStatusAccepted {
		return models.ErrNotFriends
	}

	member := &models.GroupMember{
		GroupID:  groupID,
		UserID:   newMember.ID,
		JoinedAt: time.Now(),
	}

	if err := u.groupRepo.AddMember(ctx, member); err != nil {
		return err
	}

	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err // Group should exist, but handle error just in case
	}

	// Notify the new member
	go u.notifyUser(context.Background(), newMember.ID, adderID, models.EventAddedToGroup, map[string]interface{}{
		"groupId":   groupID,
		"groupName": group.Name,
		"adderId":   adderID,
	})

	// Notify other group members
	go u.notifyGroupMembers(context.Background(), groupID, newMember.ID, models.EventUserJoinedGroup, map[string]interface{}{
		"groupId":   groupID,
		"groupName": group.Name,
		"userId":    newMember.ID,
		"adderId":   adderID,
	})

	return nil
}

func (u *groupUsecase) RemoveMember(ctx context.Context, ownerID, memberID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != ownerID {
		return models.ErrNotGroupOwner
	}

	if ownerID == memberID {
		return models.ErrCannotRemoveOwner
	}

	if err := u.groupRepo.RemoveMember(ctx, groupID, memberID); err != nil {
		return err
	}

	// Notify the removed member
	go u.notifyUser(context.Background(), memberID, ownerID, models.EventRemovedFromGroup, map[string]interface{}{
		"groupId":   groupID,
		"groupName": group.Name,
		"removerId": ownerID,
	})

	// Notify other group members
	go u.notifyGroupMembers(context.Background(), groupID, memberID, models.EventUserLeftGroup, map[string]interface{}{
		"groupId":   groupID,
		"groupName": group.Name,
		"userId":    memberID,
		"removerId": ownerID,
	})

	return nil
}

func (u *groupUsecase) TransferOwnership(ctx context.Context, currentOwnerID, newOwnerID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != currentOwnerID {
		return models.ErrNotGroupOwner
	}

	if _, err := u.groupRepo.FindMember(ctx, groupID, newOwnerID); err != nil {
		return models.ErrNotGroupMember
	}

	group.OwnerID = newOwnerID
	return u.groupRepo.Update(ctx, group)
}

func (u *groupUsecase) SearchGroups(ctx context.Context, query string) ([]*models.Group, error) {
	return u.groupRepo.FuzzySearchByHandle(ctx, query, 10)
}

func (u *groupUsecase) GetGroupDetails(ctx context.Context, groupID uuid.UUID) (*models.Group, error) {
	return u.groupRepo.FindByID(ctx, groupID)
}

func (u *groupUsecase) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*models.User, error) {
	return u.groupRepo.ListMembers(ctx, groupID)
}

func (u *groupUsecase) notifyGroupMembers(ctx context.Context, groupID, subjectUserID uuid.UUID, eventType models.EventType, payload map[string]interface{}) {
	members, err := u.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return
	}

	jsonPayload, _ := json.Marshal(payload)

	for _, member := range members {
		if member.ID == subjectUserID {
			continue
		}
		event := &models.Event{
			ID:          uuid.New(),
			Type:        eventType,
			Payload:     jsonPayload,
			RecipientID: member.ID,
			CreatedAt:   time.Now(),
		}
		u.eventUsecase.StoreEvent(ctx, event)
	}
}

func (u *groupUsecase) notifyUser(ctx context.Context, recipientID, senderID uuid.UUID, eventType models.EventType, payload map[string]interface{}) {
	jsonPayload, _ := json.Marshal(payload)
	event := &models.Event{
		ID:          uuid.New(),
		Type:        eventType,
		Payload:     jsonPayload,
		RecipientID: recipientID,
		SenderID:    &senderID,
		CreatedAt:   time.Now(),
	}
	u.eventUsecase.StoreEvent(ctx, event)
}

