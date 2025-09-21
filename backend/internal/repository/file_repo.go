package repository

import (
	"context"
)

type FileRepository interface {
	SaveProfilePicture(ctx context.Context, fileData []byte, fileType string) (filename string, err error)
}

