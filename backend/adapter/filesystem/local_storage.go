package filesystem

import (
	"chat-app/backend/repository"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type localStorage struct {
	storageDir string
	routePath  string
}

func NewLocalStorage(storageDir, routePath string) repository.FileRepository {
	return &localStorage{
		storageDir: storageDir,
		routePath:  routePath,
	}
}

func (l *localStorage) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	randomFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(l.storageDir, randomFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	fileURL := filepath.Join(l.routePath, randomFilename)
	return fileURL, nil
}

