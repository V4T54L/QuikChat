package models

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameTaken     = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrSessionNotFound   = errors.New("session not found or expired")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInternalServer    = errors.New("internal server error")
	ErrBadRequest        = errors.New("bad request")
)

