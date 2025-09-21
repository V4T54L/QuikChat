package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
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
	ErrInvalidPassword    = errors.New("password is too short")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrSessionExpired     = errors.New("session has expired")
)

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
	// Validate username
	if err := validateUsername(input.Username); err != nil {
		return nil, err
	}
	// Validate password
	if len(input.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := util.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:           util.NewUUID(),
		Username:     strings.ToLower(input.Username),
		PasswordHash: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, input usecase.LoginInput) (*usecase.AuthTokens, error) {
	user, err := s.userRepo.GetByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !util.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Single device policy: remove all old sessions
	if err := s.sessionRepo.DeleteAllForUser(ctx, user.ID); err != nil {
		return nil, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*usecase.AuthTokens, error) {
	session, err := s.sessionRepo.GetByID(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	if session.ExpiresAt.Before(time.Now()) {
		// Clean up expired session
		_ = s.sessionRepo.Delete(ctx, session.ID)
		return nil, ErrSessionExpired
	}

	// Sliding window: delete old session and create a new one
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return nil, err
	}

	return s.createSession(ctx, session.UserID)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.sessionRepo.Delete(ctx, refreshToken)
}

func (s *authService) createSession(ctx context.Context, userID string) (*usecase.AuthTokens, error) {
	accessToken, err := util.GenerateAccessToken(userID, s.cfg.JWTSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        util.NewUUID(), // This is the refresh token
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
	}

	if err := s.sessionRepo.Store(ctx, session); err != nil {
		return nil, err
	}

	return &usecase.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: session.ID,
	}, nil
}

func validateUsername(username string) error {
	if len(username) < 4 || len(username) > 50 {
		return ErrInvalidUsername
	}
	// Allowed characters: lowercase letters (a-z), digits (0-9), and underscore (_)
	match, _ := regexp.MatchString("^[a-z0-9_]+$", username)
	if !match {
		return ErrInvalidUsername
	}
	return nil
}

