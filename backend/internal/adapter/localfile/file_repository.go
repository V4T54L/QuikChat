package localfile

import (
	"chat-app/internal/repository"
	"chat-app/pkg/util"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type localFileRepository struct {
	uploadDir string
}

func NewLocalFileRepository(uploadDir string) (repository.FileRepository, error) {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &localFileRepository{uploadDir: uploadDir}, nil
}

func (r *localFileRepository) SaveProfilePicture(_ context.Context, fileData []byte, fileType string) (string, error) {
	var ext string
	switch fileType {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", fmt.Errorf("unsupported file type: %s", fileType)
	}

	filename := util.NewUUID() + ext
	filePath := filepath.Join(r.uploadDir, filename)

	err := os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}
