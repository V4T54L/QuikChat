package http

import (
	"chat-app/backend/usecase"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

type UserHandler struct {
	userUsecase usecase.UserUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

type FriendHandler struct {
	friendUsecase usecase.FriendUsecase
}

func NewFriendHandler(friendUsecase usecase.FriendUsecase) *FriendHandler {
	return &FriendHandler{friendUsecase: friendUsecase}
}

type GroupHandler struct {
	groupUsecase usecase.GroupUsecase
}

func NewGroupHandler(groupUsecase usecase.GroupUsecase) *GroupHandler {
	return &GroupHandler{groupUsecase: groupUsecase}
}
