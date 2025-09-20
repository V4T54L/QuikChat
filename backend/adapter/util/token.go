package util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken() (uuid.UUID, time.Time, error)
	GetRefreshTokenExp() time.Duration
}

type tokenGenerator struct {
	jwtSecret       string
	accessTokenExp  time.Duration
	refreshTokenExp time.Duration
}

func NewTokenGenerator(secret string, accessExp, refreshExp time.Duration) TokenGenerator {
	return &tokenGenerator{
		jwtSecret:       secret,
		accessTokenExp:  accessExp,
		refreshTokenExp: refreshExp,
	}
}

func (t *tokenGenerator) GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(t.accessTokenExp).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.jwtSecret))
}

func (t *tokenGenerator) GenerateRefreshToken() (uuid.UUID, time.Time, error) {
	refreshToken := uuid.New()
	expiresAt := time.Now().Add(t.refreshTokenExp)
	return refreshToken, expiresAt, nil
}

func (t *tokenGenerator) GetRefreshTokenExp() time.Duration {
	return t.refreshTokenExp
}

