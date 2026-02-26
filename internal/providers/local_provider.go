package providers

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type LocalUploadProvider struct {
	basePath string
	log      zerolog.Logger
}

func NewLocalUploadProvider(basePath string, logger zerolog.Logger) *LocalUploadProvider {
	return &LocalUploadProvider{basePath: basePath}
}

func (p *LocalUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {

	fullPath := filepath.Join(p.basePath, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
		return "", err
	}

	// Open source
	src, err := file.Open()
	if err != nil {
		return "", err
	}

	// create destination
	dst, err := os.Create(fullPath) // #nosec G304
	if err != nil {
		return "", err
	}

	defer func() {
		if err := src.Close(); err != nil {
			p.log.Printf("failed to close src: %v", err)
		}
		if err := dst.Close(); err != nil {
			p.log.Printf("failed to close dst: %v", err)
		}
	}()

	// read from source to destination
	if _, err := dst.ReadFrom(src); err != nil {
		return "", err
	}

	return fmt.Sprintf("/uploads/%s", path), nil

}

func (p *LocalUploadProvider) DeleteFile(path string) error {
	fullPath := filepath.Join(p.basePath, path)
	return os.Remove(fullPath)
}
