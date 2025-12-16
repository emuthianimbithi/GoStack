package models

import (
	"time"

	"github.com/google/uuid"
)

type UserInvite struct {
	BaseModel
	Email       string    `gorm:"not null"`
	BusinessID  uuid.UUID `gorm:"type:uuid;index;not null"`
	RoleID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Token       string    `gorm:"uniqueIndex;not null"`
	ExpiresAt   time.Time `gorm:"index;not null"`
	Status      string    `gorm:"default:'pending'"` // pending, accepted, expired
	InvitedByID uuid.UUID `gorm:"type:uuid"`

	Business  Business   `gorm:"foreignKey:BusinessID"`
	Role      AccessRole `gorm:"foreignKey:RoleID"`
	InvitedBy User       `gorm:"foreignKey:InvitedByID"`
}
