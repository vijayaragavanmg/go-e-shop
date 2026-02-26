package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/vijayaragavanmg/learning-go-shop/internal/interfaces"
)

var _ UploadServiceInterface = (*UploadService)(nil)

type UploadService struct {
	provider interfaces.UploadProvider
}

func NewUploadService(provider interfaces.UploadProvider) *UploadService {
	return &UploadService{provider: provider}
}

func (s *UploadService) UploadProductImage(productID uint, file *multipart.FileHeader) (string, error) {

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isValidImageExt(ext) {
		return "", fmt.Errorf("invalid file type: %s", ext)
	}

	newFileName := uuid.New().String() + ext
	path := fmt.Sprintf("products/%d/%s", productID, newFileName)

	return s.provider.UploadFile(file, path)
}

func isValidImageExt(ext string) bool {
	validExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	return slices.Contains(validExts, ext)
}
