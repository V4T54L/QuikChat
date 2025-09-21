package service

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"chat-app/internal/usecase"
	"chat-app/pkg/util"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"regexp"
)

var (
	ErrProfileUpdateFailed = errors.New("profile update failed")
	ErrInvalidFileType     = errors.New("invalid file type for profile picture, allowed: png, jpg, jpeg, webp")
	ErrFileSizeExceeded    = errors.New("file size exceeds the 200KB limit")
)

type userService struct {
	userRepo repository.UserRepository
	fileRepo repository.FileRepository
}

func NewUserService(userRepo repository.UserRepository, fileRepo repository.FileRepository) usecase.UserUsecase {
	return &userService{
		userRepo: userRepo,
		fileRepo: fileRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, input usecase.UpdateUserInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if input.Username != nil {
		if err := validateUsername(*input.Username); err != nil {
			return nil, err
		}
		existingUser, err := s.userRepo.GetByUsername(ctx, *input.Username)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, ErrUserExists
		}
		user.Username = *input.Username
	}

	if input.Password != nil {
		if len(*input.Password) < 8 {
			return nil, ErrInvalidPassword
		}
		hashedPassword, err := util.HashPassword(*input.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, ErrProfileUpdateFailed
	}

	return user, nil
}

func (s *userService) UpdateProfilePicture(ctx context.Context, userID string, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > 200*1024 { // 200 KB
		return "", ErrFileSizeExceeded
	}

	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := []string{"image/png", "image/jpeg", "image/jpg", "image/webp"}
	isValidType := false
	for _, t := range allowedTypes {
		if contentType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return "", ErrInvalidFileType
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	filename, err := s.fileRepo.SaveProfilePicture(ctx, fileBytes, contentType)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	user.ProfilePicURL = "/uploads/" + filename
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", ErrProfileUpdateFailed
	}

	return user.ProfilePicURL, nil
}

func validateUsername(username string) error {
	if len(username) < 4 || len(username) > 50 {
		return ErrInvalidUsername
	}
	match, _ := regexp.MatchString("^[a-z0-9_]+$")
	if !match {
		return ErrInvalidUsername
	}
	return nil
}

