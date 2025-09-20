package repository

import "mime/multipart"

type FileRepository interface {
	Save(file multipart.File, header *multipart.FileHeader) (string, error)
}

