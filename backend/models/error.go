package models

import "errors"

var (
	// User & Auth
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInternalServer     = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")

	// Friendship
	ErrFriendRequestExists   = errors.New("friend request already exists")
	ErrAlreadyFriends        = errors.New("users are already friends")
	ErrNotFriends            = errors.New("users are not friends")
	ErrFriendRequestNotFound = errors.New("friend request not found")
	ErrCannotFriendSelf      = errors.New("cannot send friend request to yourself")

	// Group
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupHandleTaken   = errors.New("group handle is already taken")
	ErrNotGroupOwner      = errors.New("user is not the group owner")
	ErrNotGroupMember     = errors.New("user is not a group member")
	ErrAlreadyGroupMember = errors.New("user is already a group member")
	ErrCannotRemoveOwner  = errors.New("cannot remove the group owner")
)
