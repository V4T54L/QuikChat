package util

import (
	"errors"
	"mime/multipart"
	"regexp"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{4,50}$`)

func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("invalid username format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password is too short")
	}
	return nil
}

func ValidateProfilePic(header *multipart.FileHeader) error {
	// Max size: 200 KB
	if header.Size > 200*1024 {
		return errors.New("file size exceeds 200KB")
	}

	// Allowed formats: png, jpg, jpeg, webp
	allowedTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return errors.New("invalid file type")
	}

	return nil
}
```
