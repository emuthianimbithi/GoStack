package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	UserID uuid.UUID `gorm:"type:uuid;index;not null"`
	User   User      `gorm:"foreignKey:UserID"`

	Title    string         `gorm:"not null"`
	Message  string         `gorm:"not null"` // or Body
	Type     string         `gorm:"index"`    // info, warning, error, invite
	IsRead   bool           `gorm:"default:false;index"`
	Metadata datatypes.JSON `gorm:"type:jsonb"` // e.g., link to resource
}

func (Notification) TableName() string {
	return "notifications"
}
