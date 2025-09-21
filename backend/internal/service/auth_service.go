package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"chat-app/internal/usecase"
	"chat-app/pkg/config"
	"chat-app/pkg/util"
)

var (
	ErrUserExists         = errors.New("user with this username already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidUsername    = errors.New("username format is invalid")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters long")
	ErrSessionNotFound    = errors.New("session not found or already logged out")
	ErrSessionExpired     = errors.New("session has expired")
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{4,50}$`)

type authService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	cfg         *config.Config
}

func NewAuthService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, cfg *config.Config) usecase.AuthUsecase {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (s *authService) SignUp(ctx context.Context, input usecase.SignUpInput) (*domain.User, error) {
	if err := validateUsername(input.Username); err != nil {
		return nil, err
	}
	if len(input.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	existingUser, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := util.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:           util.NewUUID(),
		Username:     input.Username, // Removed strings.ToLower as regex implies lowercase
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, input usecase.LoginInput) (*usecase.AuthTokens, error) {
	user, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !util.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Single device policy: Invalidate all previous sessions for this user
	if err := s.sessionRepo.DeleteAllForUser(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("failed to invalidate previous sessions: %w", err)
	}

	return s.createSession(ctx, user.ID)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*usecase.AuthTokens, error) {
	session, err := s.sessionRepo.GetByID(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		_ = s.sessionRepo.Delete(ctx, session.ID)
		return nil, ErrSessionExpired
	}

	// Invalidate the old refresh token
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, fmt.Errorf("failed to delete old session: %w", err)
	}

	return s.createSession(ctx, session.UserID)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	err := s.sessionRepo.Delete(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *authService) createSession(ctx context.Context, userID string) (*usecase.AuthTokens, error) {
	accessToken, err := util.GenerateAccessToken(userID, s.cfg.JWTSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	session := &domain.Session{
		ID:        util.NewUUID(), // This is the refresh token
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	if err := s.sessionRepo.Store(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &usecase.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: session.ID,
	}, nil
}

func validateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}
