package usecase

import (
	"chat-app/backend/adapter/util"
	"chat-app/backend/models"
	"chat-app/backend/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
	Refresh(ctx context.Context, refreshToken string) (newAccessToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

type authUsecase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	tokenGen    util.TokenGenerator
}

func NewAuthUsecase(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, tokenGen util.TokenGenerator) AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokenGen:    tokenGen,
	}
}

func (a *authUsecase) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := a.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return "", "", models.ErrInvalidCredentials
		}
		return "", "", err
	}

	if !util.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", models.ErrInvalidCredentials
	}

	// Single device policy: remove old sessions
	if err := a.sessionRepo.DeleteByUserID(ctx, user.ID); err != nil {
		// Log error but continue, as this is not critical for login
		// log.Printf("failed to delete old sessions for user %s: %v", user.ID, err)
	}

	accessToken, err := a.tokenGen.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", err
	}

	refreshToken, expiresAt, err := a.tokenGen.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	session := &models.Session{
		RefreshToken: refreshToken,
		UserID:       user.ID,
		ExpiresAt:    expiresAt,
	}

	if err := a.sessionRepo.Create(ctx, session); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.String(), nil
}

func (a *authUsecase) Refresh(ctx context.Context, refreshTokenStr string) (string, error) {
	refreshToken, err := uuid.Parse(refreshTokenStr)
	if err != nil {
		return "", models.ErrInvalidToken
	}

	session, err := a.sessionRepo.Find(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	if session.ExpiresAt.Before(time.Now()) {
		// Clean up expired session
		_ = a.sessionRepo.Delete(ctx, refreshToken)
		return "", models.ErrSessionNotFound
	}

	// Sliding window: extend session expiry
	session.ExpiresAt = time.Now().Add(a.tokenGen.GetRefreshTokenExp())
	if err := a.sessionRepo.Create(ctx, session); err != nil { // Create will UPSERT
		return "", err
	}

	newAccessToken, err := a.tokenGen.GenerateAccessToken(session.UserID)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

func (a *authUsecase) Logout(ctx context.Context, refreshTokenStr string) error {
	refreshToken, err := uuid.Parse(refreshTokenStr)
	if err != nil {
		return models.ErrInvalidToken
	}
	return a.sessionRepo.Delete(ctx, refreshToken)
}

