package usecase

import (
	"chat-app/internal/domain"
	"context"
	"mime/multipart"
)

type UpdateUserInput struct {
	Username *string
	Password *string
}

type UserUsecase interface {
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateUserInput) (*domain.User, error)
	UpdateProfilePicture(ctx context.Context, userID string, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}

