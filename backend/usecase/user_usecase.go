package usecase

import (
	"chat-app/backend/adapter/util"
	"chat-app/backend/models"
	"chat-app/backend/repository"
	"context"
	"errors"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
)

type UserUsecase interface {
	Register(ctx context.Context, username, password string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, username, password *string, profilePic multipart.File, profilePicHeader *multipart.FileHeader) (*models.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
	fileRepo repository.FileRepository
}

func NewUserUsecase(userRepo repository.UserRepository, fileRepo repository.FileRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		fileRepo: fileRepo,
	}
}

func (u *userUsecase) Register(ctx context.Context, username, password string) (*models.User, error) {
	if err := util.ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := util.ValidatePassword(password); err != nil {
		return nil, err
	}

	existingUser, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, models.ErrUsernameTaken
	}

	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     strings.ToLower(username),
		PasswordHash: hashedPassword,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *userUsecase) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return u.userRepo.FindByUsername(ctx, username)
}

func (u *userUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, username, password *string, profilePic multipart.File, profilePicHeader *multipart.FileHeader) (*models.User, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if username != nil {
		if err := util.ValidateUsername(*username); err != nil {
			return nil, err
		}
		// Check if new username is taken by another user
		existingUser, err := u.userRepo.FindByUsername(ctx, *username)
		if err != nil && !errors.Is(err, models.ErrUserNotFound) {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, models.ErrUsernameTaken
		}
		user.Username = strings.ToLower(*username)
	}

	if password != nil {
		if err := util.ValidatePassword(*password); err != nil {
			return nil, err
		}
		hashedPassword, err := util.HashPassword(*password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if profilePic != nil {
		if err := util.ValidateProfilePic(profilePicHeader); err != nil {
			return nil, err
		}
		picURL, err := u.fileRepo.Save(profilePic, profilePicHeader)
		if err != nil {
			return nil, err
		}
		user.ProfilePicURL = picURL
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

