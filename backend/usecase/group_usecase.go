package usecase

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"time"

	"chat-app/adapter/util"
	"chat-app/models"
	"chat-app/repository"

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

	var photoURL string
	if photo != nil {
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
		group.Name = *name
	}

	if photo != nil {
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

	// Generate events
	user, _ := u.userRepo.FindByID(ctx, userID)

	// Event for other group members
	eventPayload, _ := json.Marshal(map[string]string{
		"groupId":   group.ID.String(),
		"groupName": group.Name,
		"userId":    user.ID.String(),
		"userName":  user.Username,
	})

	members, _ := u.groupRepo.ListMembers(ctx, group.ID)
	for _, m := range members {
		if m.ID != userID {
			event := &models.Event{
				ID:          uuid.New(),
				Type:        models.EventUserJoinedGroup,
				Payload:     eventPayload,
				RecipientID: m.ID,
				SenderID:    &userID,
				CreatedAt:   time.Now().UTC(),
			}
			u.eventUsecase.StoreEvent(ctx, event)
		}
	}
	return nil
}

func (u *groupUsecase) LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	members, err := u.groupRepo.ListMembers(ctx, groupID)
	if err != nil {
		return err
	}

	if err := u.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return err
	}

	// Generate event for remaining members
	eventPayload, _ := json.Marshal(map[string]string{
		"groupId":   group.ID.String(),
		"groupName": group.Name,
		"userId":    user.ID.String(),
		"userName":  user.Username,
	})

	for _, m := range members {
		if m.ID != userID {
			event := &models.Event{
				ID:          uuid.New(),
				Type:        models.EventUserLeftGroup,
				Payload:     eventPayload,
				RecipientID: m.ID,
				SenderID:    &userID,
				CreatedAt:   time.Now().UTC(),
			}
			u.eventUsecase.StoreEvent(ctx, event)
		}
	}

	// Handle ownership transfer or group deletion
	if group.OwnerID == userID {
		if len(members)-1 == 0 {
			// Last member left, delete group
			return u.groupRepo.Delete(ctx, groupID)
		}
		// Transfer ownership to the oldest member
		oldestMember, err := u.groupRepo.GetOldestMember(ctx, groupID)
		if err != nil {
			return err
		}
		group.OwnerID = oldestMember.ID
		return u.groupRepo.Update(ctx, group)
	}

	return nil
}

func (u *groupUsecase) AddMember(ctx context.Context, adderID uuid.UUID, newMemberUsername string, groupID uuid.UUID) error {
	// Check if adder is a member
	_, err := u.groupRepo.FindMember(ctx, groupID, adderID)
	if err != nil {
		return models.ErrNotGroupMember
	}

	newMember, err := u.userRepo.FindByUsername(ctx, newMemberUsername)
	if err != nil {
		return err
	}

	// Check if adder and new member are friends
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

	// Generate events
	adder, _ := u.userRepo.FindByID(ctx, adderID)
	group, _ := u.groupRepo.FindByID(ctx, groupID)

	// Event for the new member
	addedPayload, _ := json.Marshal(map[string]string{
		"groupId":   group.ID.String(),
		"groupName": group.Name,
		"adderName": adder.Username,
	})
	addedEvent := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventAddedToGroup,
		Payload:     addedPayload,
		RecipientID: newMember.ID,
		SenderID:    &adderID,
		CreatedAt:   time.Now().UTC(),
	}
	u.eventUsecase.StoreEvent(ctx, addedEvent)

	// Event for other group members
	members, _ := u.groupRepo.ListMembers(ctx, groupID)
	memberNotifPayload, _ := json.Marshal(map[string]string{
		"groupId":       group.ID.String(),
		"groupName":     group.Name,
		"adderName":     adder.Username,
		"newMemberId":   newMember.ID.String(),
		"newMemberName": newMember.Username,
	})
	for _, m := range members {
		if m.ID != newMember.ID && m.ID != adder.ID {
			event := &models.Event{
				ID:          uuid.New(),
				Type:        models.EventUserJoinedGroup, // Re-using this for simplicity
				Payload:     memberNotifPayload,
				RecipientID: m.ID,
				SenderID:    &adderID,
				CreatedAt:   time.Now().UTC(),
			}
			u.eventUsecase.StoreEvent(ctx, event)
		}
	}
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

	// Generate events
	owner, _ := u.userRepo.FindByID(ctx, ownerID)
	removedMember, _ := u.userRepo.FindByID(ctx, memberID)

	// Event for the removed member
	removedPayload, _ := json.Marshal(map[string]string{
		"groupId":     group.ID.String(),
		"groupName":   group.Name,
		"removerName": owner.Username,
	})
	removedEvent := &models.Event{
		ID:          uuid.New(),
		Type:        models.EventRemovedFromGroup,
		Payload:     removedPayload,
		RecipientID: removedMember.ID,
		SenderID:    &ownerID,
		CreatedAt:   time.Now().UTC(),
	}
	u.eventUsecase.StoreEvent(ctx, removedEvent)

	// Event for other group members
	members, _ := u.groupRepo.ListMembers(ctx, groupID)
	memberNotifPayload, _ := json.Marshal(map[string]string{
		"groupId":           group.ID.String(),
		"groupName":         group.Name,
		"removerName":       owner.Username,
		"removedMemberId":   removedMember.ID.String(),
		"removedMemberName": removedMember.Username,
	})
	for _, m := range members {
		event := &models.Event{
			ID:          uuid.New(),
			Type:        models.EventUserLeftGroup, // Re-using this for simplicity
			Payload:     memberNotifPayload,
			RecipientID: m.ID,
			SenderID:    &ownerID,
			CreatedAt:   time.Now().UTC(),
		}
		u.eventUsecase.StoreEvent(ctx, event)
	}
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

	_, err = u.groupRepo.FindMember(ctx, groupID, newOwnerID)
	if err != nil {
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

