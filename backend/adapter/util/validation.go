package util

import (
	"errors"
	"mime/multipart"
	"regexp"
	"strings"
)

var (
	usernameRegex    = regexp.MustCompile(`^[a-z0-9_]{4,50}$`)
	groupHandleRegex = regexp.MustCompile(`^[a-z0-9_]{4,50}$`)
)

func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("username must be 4-50 characters and contain only lowercase letters, digits, and underscores")
	}
	return nil
}

func ValidateGroupHandle(handle string) error {
	// Format: string prefix + #groupname
	parts := strings.Split(handle, "#")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("group handle must be in the format 'prefix#groupname'")
	}

	groupname := parts[1]
	if !groupHandleRegex.MatchString(groupname) {
		return errors.New("group name part of handle must be 4-50 characters and contain only lowercase letters, digits, and underscores")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func ValidateProfilePic(header *multipart.FileHeader) error {
	// Max size: 200 KB
	if header.Size > 200*1024 {
		return errors.New("profile picture size cannot exceed 200 KB")
	}

	// Allowed formats: png, jpg, jpeg, webp
	contentType := header.Header.Get("Content-Type")
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
		return nil
	default:
		return errors.New("invalid file format. Only png, jpg, jpeg, and webp are allowed")
	}
}

