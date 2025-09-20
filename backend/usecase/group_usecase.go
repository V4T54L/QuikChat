package usecase

import (
	"context"
	"fmt"
	"mime/multipart"

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
	groupRepo  repository.GroupRepository
	userRepo   repository.UserRepository
	friendRepo repository.FriendshipRepository
	fileRepo   repository.FileRepository
}

func NewGroupUsecase(groupRepo repository.GroupRepository, userRepo repository.UserRepository, friendRepo repository.FriendshipRepository, fileRepo repository.FileRepository) GroupUsecase {
	return &groupUsecase{
		groupRepo:  groupRepo,
		userRepo:   userRepo,
		friendRepo: friendRepo,
		fileRepo:   fileRepo,
	}
}

func (u *groupUsecase) CreateGroup(ctx context.Context, ownerID uuid.UUID, handle, name string, photo multipart.File, photoHeader *multipart.FileHeader) (*models.Group, error) {
	if err := util.ValidateGroupHandle(handle); err != nil {
		return nil, err
	}

	var photoURL string
	var err error
	if photo != nil && photoHeader != nil {
		if err := util.ValidateProfilePic(photoHeader); err != nil {
			return nil, err
		}
		photoURL, err = u.fileRepo.Save(photo, photoHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to save group photo: %w", err)
		}
	}

	group := &models.Group{
		ID:       uuid.New(),
		Handle:   handle,
		Name:     name,
		PhotoURL: photoURL,
		OwnerID:  ownerID,
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

	if photo != nil && photoHeader != nil {
		if err := util.ValidateProfilePic(photoHeader); err != nil {
			return nil, err
		}
		photoURL, err := u.fileRepo.Save(photo, photoHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to save group photo: %w", err)
		}
		group.PhotoURL = photoURL
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
		GroupID: group.ID,
		UserID:  userID,
	}

	return u.groupRepo.AddMember(ctx, member)
}

func (u *groupUsecase) LeaveGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if err := u.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return err
	}

	// If the owner leaves, transfer ownership
	if group.OwnerID == userID {
		oldestMember, err := u.groupRepo.GetOldestMember(ctx, groupID)
		if err != nil {
			// If no members are left, the group can be deleted or left ownerless.
			// Let's assume it gets deleted.
			if err == models.ErrGroupNotFound { // No members left
				return u.groupRepo.Delete(ctx, groupID)
			}
			return fmt.Errorf("failed to find new owner: %w", err)
		}
		group.OwnerID = oldestMember.ID
		if err := u.groupRepo.Update(ctx, group); err != nil {
			return fmt.Errorf("failed to transfer ownership: %w", err)
		}
	}

	return nil
}

func (u *groupUsecase) AddMember(ctx context.Context, adderID uuid.UUID, newMemberUsername string, groupID uuid.UUID) error {
	// Check if adder is a member of the group
	if _, err := u.groupRepo.FindMember(ctx, groupID, adderID); err != nil {
		return models.ErrNotGroupMember
	}

	// Find the user to be added
	newMember, err := u.userRepo.FindByUsername(ctx, newMemberUsername)
	if err != nil {
		return err
	}

	// Check if adder and new member are friends
	friendship, err := u.friendRepo.Find(ctx, adderID, newMember.ID)
	if err != nil || friendship.Status != models.FriendshipStatusAccepted {
		return models.ErrNotFriends
	}

	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  newMember.ID,
	}

	return u.groupRepo.AddMember(ctx, member)
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

	return u.groupRepo.RemoveMember(ctx, groupID, memberID)
}

func (u *groupUsecase) TransferOwnership(ctx context.Context, currentOwnerID, newOwnerID, groupID uuid.UUID) error {
	group, err := u.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != currentOwnerID {
		return models.ErrNotGroupOwner
	}

	// Check if new owner is a member
	if _, err := u.groupRepo.FindMember(ctx, groupID, newOwnerID); err != nil {
		return err
	}

	group.OwnerID = newOwnerID
	return u.groupRepo.Update(ctx, group)
}

func (u *groupUsecase) SearchGroups(ctx context.Context, query string) ([]*models.Group, error) {
	return u.groupRepo.FuzzySearchByHandle(ctx, query, 10) // Limit to 10 results
}

func (u *groupUsecase) GetGroupDetails(ctx context.Context, groupID uuid.UUID) (*models.Group, error) {
	return u.groupRepo.FindByID(ctx, groupID)
}

func (u *groupUsecase) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*models.User, error) {
	return u.groupRepo.ListMembers(ctx, groupID)
}

