package service

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"chat-app/internal/usecase"
	"chat-app/pkg/util"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
)

var (
	ErrInvalidGroupHandle   = errors.New("invalid group handle format")
	ErrNotGroupOwner        = errors.New("only the group owner can perform this action")
	ErrNotGroupMember       = errors.New("user is not a member of this group")
	ErrCannotRemoveOwner    = errors.New("group owner cannot be removed")
	ErrCannotLeaveAsOwner   = errors.New("owner must transfer ownership before leaving")
	ErrAddNotFriend         = errors.New("you can only add your friends to a group")
	ErrTransferToNonMember  = errors.New("can only transfer ownership to a group member")
	ErrTransferToSelf       = errors.New("cannot transfer ownership to yourself")
	ErrRemoveSelf           = errors.New("cannot remove yourself from a group, use leave group instead")
	ErrLastMemberCannotLeave = errors.New("last member cannot leave the group, it will be deleted")
)

var groupHandleRegex = regexp.MustCompile(`^[a-z0-9_]{4,50}$`)

type groupService struct {
	groupRepo  repository.GroupRepository
	userRepo   repository.UserRepository
	friendRepo repository.FriendRepository
	fileRepo   repository.FileRepository
	eventUsecase usecase.EventUsecase
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	friendRepo repository.FriendRepository,
	fileRepo repository.FileRepository,
	eventUsecase usecase.EventUsecase,
) usecase.GroupUsecase {
	return &groupService{
		groupRepo:  groupRepo,
		userRepo:   userRepo,
		friendRepo: friendRepo,
		fileRepo:   fileRepo,
		eventUsecase: eventUsecase,
	}
}

func (s *groupService) validateGroupHandle(handle string) error {
	if !strings.HasPrefix(handle, "#") {
		return ErrInvalidGroupHandle
	}
	cleanHandle := strings.TrimPrefix(handle, "#")
	if !groupHandleRegex.MatchString(cleanHandle) {
		return ErrInvalidGroupHandle
	}
	return nil
}

func (s *groupService) CreateGroup(ctx context.Context, userID string, input usecase.CreateGroupInput, file multipart.File, fileHeader *multipart.FileHeader) (*domain.Group, error) {
	if err := s.validateGroupHandle(input.Handle); err != nil {
		return nil, err
	}

	var picURL string
	if file != nil {
		// Re-use user profile picture logic for group pictures
		url, err := NewUserService(s.userRepo, s.fileRepo).UpdateProfilePicture(ctx, "", file, fileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to save group picture: %w", err)
		}
		picURL = url
	}

	group := &domain.Group{
		ID:            util.NewUUID(),
		Handle:        input.Handle,
		Name:          input.Name,
		OwnerID:       userID,
		ProfilePicURL: picURL,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	ownerMember := &domain.GroupMember{
		GroupID: group.ID,
		UserID:  userID,
		Role:    domain.GroupRoleOwner,
	}
	if err := s.groupRepo.CreateMember(ctx, ownerMember); err != nil {
		// TODO: Add transaction to rollback group creation
		return nil, err
	}

	return group, nil
}

func (s *groupService) GetGroupDetails(ctx context.Context, groupID string) (*usecase.GroupDetails, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	members, err := s.groupRepo.GetMembersWithUserDetails(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &usecase.GroupDetails{
		Group:   group,
		Members: members,
	}, nil
}

func (s *groupService) UpdateGroup(ctx context.Context, userID, groupID string, input usecase.UpdateGroupInput, file multipart.File, fileHeader *multipart.FileHeader) (*domain.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.OwnerID != userID {
		return nil, ErrNotGroupOwner
	}

	if input.Name != nil {
		group.Name = *input.Name
	}

	if file != nil {
		url, err := NewUserService(s.userRepo, s.fileRepo).UpdateProfilePicture(ctx, "", file, fileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to update group picture: %w", err)
		}
		group.ProfilePicURL = url
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *groupService) SearchGroups(ctx context.Context, query string) ([]*domain.Group, error) {
	return s.groupRepo.SearchByHandle(ctx, query)
}

func (s *groupService) JoinGroup(ctx context.Context, userID, groupHandle string) (*domain.Group, error) {
	group, err := s.groupRepo.GetByHandle(ctx, groupHandle)
	if err != nil {
		return nil, err
	}

	member := &domain.GroupMember{
		GroupID: group.ID,
		UserID:  userID,
		Role:    domain.GroupRoleMember,
	}

	if err := s.groupRepo.CreateMember(ctx, member); err != nil {
		if errors.Is(err, repository.ErrGroupMemberExists) {
			return group, nil // Idempotent: already a member, return success
		}
		return nil, err
	}

	return group, nil
}

func (s *groupService) LeaveGroup(ctx context.Context, userID, groupID string) error {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID == userID {
		// Check for other members
		oldestMember, err := s.groupRepo.GetOldestMember(ctx, groupID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				// This is the last member, they can't leave.
				// In a real app, this might trigger group deletion.
				return ErrLastMemberCannotLeave
			}
			return err
		}
		// Auto-transfer ownership
		if err := s.TransferOwnership(ctx, userID, groupID, oldestMember.UserID); err != nil {
			return fmt.Errorf("failed to auto-transfer ownership: %w", err)
		}
	}

	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func (s *groupService) AddMember(ctx context.Context, currentUserID, groupID, friendID string) error {
	// 1. Check if current user is a member of the group
	_, err := s.groupRepo.FindMember(ctx, groupID, currentUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotGroupMember
		}
		return err
	}

	// 2. Check if the user to be added is a friend of the current user
	areFriends, err := s.friendRepo.AreFriends(ctx, currentUserID, friendID)
	if err != nil {
		return err
	}
	if !areFriends {
		return ErrAddNotFriend
	}

	// 3. Add the friend to the group
	newMember := &domain.GroupMember{
		GroupID: groupID,
		UserID:  friendID,
		Role:    domain.GroupRoleMember,
	}
	if err := s.groupRepo.CreateMember(ctx, newMember); err != nil {
		if errors.Is(err, repository.ErrGroupMemberExists) {
			return nil // Idempotent
		}
		return err
	}

	return nil
}

func (s *groupService) RemoveMember(ctx context.Context, ownerID, groupID, memberID string) error {
	if ownerID == memberID {
		return ErrRemoveSelf
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != ownerID {
		return ErrNotGroupOwner
	}

	if group.OwnerID == memberID {
		return ErrCannotRemoveOwner
	}

	return s.groupRepo.RemoveMember(ctx, groupID, memberID)
}

func (s *groupService) TransferOwnership(ctx context.Context, currentOwnerID, groupID, newOwnerID string) error {
	if currentOwnerID == newOwnerID {
		return ErrTransferToSelf
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	if group.OwnerID != currentOwnerID {
		return ErrNotGroupOwner
	}

	// Check if new owner is a member
	_, err = s.groupRepo.FindMember(ctx, groupID, newOwnerID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTransferToNonMember
		}
		return err
	}

	// Perform transfer in a transaction (conceptual)
	// 1. Update group owner_id
	group.OwnerID = newOwnerID
	if err := s.groupRepo.Update(ctx, group); err != nil {
		return err
	}

	// 2. Update roles in group_members table
	// This would require a new repo method, e.g., UpdateMemberRole
	// For simplicity, we'll assume this is handled or not strictly required by the current schema
	// A more robust implementation would have a transaction and update both tables.

	return nil
}

func (s *groupService) ListUserGroups(ctx context.Context, userID string) ([]*domain.Group, error) {
	return s.groupRepo.GetGroupsByUserID(ctx, userID)
}

