package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;"`
	CreatedAt  time.Time  `gorm:"index"`
	UserID     *uuid.UUID `gorm:"type:uuid;index"`
	User       *User      `gorm:"foreignKey:UserID"`
	BusinessID *uuid.UUID `gorm:"type:uuid;index"`

	Action   string         `gorm:"not null;index"` // POST, PUT, DELETE, etc.
	Resource string         `gorm:"not null;index"` // /api/v1/users
	Payload  datatypes.JSON `gorm:"type:jsonb"`     // Request Body (redacted)

	// Metrics for Billing
	DurationMs int64 `gorm:"not null"` // Execution time in ms
	ReqBytes   int   `gorm:"not null"` // Request Body Size
	RespBytes  int   `gorm:"not null"` // Response Body Size
	Status     int   `gorm:"not null"` // HTTP Status Code
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
