package util

import "github.com/google/uuid"

func NewUUID() string {
	return uuid.New().String()
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
