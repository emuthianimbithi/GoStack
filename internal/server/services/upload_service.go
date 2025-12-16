package services

import (
	"mime/multipart"
	"path/filepath"

	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadService struct {
	db *gorm.DB
	// gcpClient *storage.Client
}

func NewUploadService(db *gorm.DB) *UploadService {
	return &UploadService{db: db}
}

// UploadFile handles the upload logic (interface for S3/GCP/Local)
func (s *UploadService) UploadFile(file *multipart.FileHeader, uploaderID uuid.UUID, businessID *uuid.UUID) (*models.File, error) {
	// 1. Validate
	// 2. Upload to Provider (Stubbed Local for now)
	// In prod:
	// bucket := s.gcpClient.Bucket("my-bucket")
	// obj := bucket.Object(key).NewWriter(ctx)

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	key := "uploads/" + filename
	// Mock URL
	url := "https://storage.googleapis.com/my-bucket/" + key

	// 3. Save to DB
	f := &models.File{
		BusinessID: businessID,
		UploaderID: uploaderID,
		Name:       file.Filename,
		Key:        key,
		URL:        url,
		MimeType:   file.Header.Get("Content-Type"),
		Size:       file.Size,
		Provider:   "gcp-stub",
	}

	if err := s.db.Create(f).Error; err != nil {
		return nil, err
	}

	return f, nil
}
