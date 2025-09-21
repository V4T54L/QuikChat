package usecase

import (
	"context"
	"chat-app/internal/domain"
)

type SignUpInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthUsecase interface {
	SignUp(ctx context.Context, input SignUpInput) (*domain.User, error)
	Login(ctx context.Context, input LoginInput) (*AuthTokens, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

