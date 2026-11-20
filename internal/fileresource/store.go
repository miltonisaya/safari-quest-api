package fileresource

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const maxUploadBytes = 20 << 20

// ErrFileType is returned when the upload extension is not an image, document, or video.
var ErrFileType = errors.New("Upload an image, document, or video")

// ErrFileTooLarge is returned when the upload exceeds 20 MB.
var ErrFileTooLarge = errors.New("File is larger than 20 MB")

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".pdf": true, ".txt": true, ".doc": true, ".docx": true,
	".mp4": true, ".webm": true, ".mov": true,
}

// Store saves the upload under dir and returns the relative path.
func Store(header *multipart.FileHeader, dir string) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExt[ext] {
		return "", ErrFileType
	}
	if header.Size > maxUploadBytes {
		return "", ErrFileTooLarge
	}
	src, err := header.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := uuid.New().String() + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(src, maxUploadBytes+1))
	if err != nil {
		return "", err
	}
	if written > maxUploadBytes {
		_ = os.Remove(filepath.Join(dir, name))
		return "", ErrFileTooLarge
	}
	return filepath.ToSlash(filepath.Join(dir, name)), nil
}
