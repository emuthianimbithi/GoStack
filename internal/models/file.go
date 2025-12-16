package models

import (
	"github.com/google/uuid"
)

type File struct {
	BaseModel
	BusinessID *uuid.UUID `gorm:"type:uuid;index"` // Optional, system files might vary
	UploaderID uuid.UUID  `gorm:"type:uuid;index"`

	Name     string // Original filename
	Key      string `gorm:"uniqueIndex"` // Storage Path/Key
	URL      string // Public URL
	MimeType string
	Size     int64
	Provider string // gcp, s3, local

	Business *Business `gorm:"foreignKey:BusinessID"`
}
