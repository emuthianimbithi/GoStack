package models

import "github.com/google/uuid"

type User struct {
	BaseModel
	Email     string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"` // Hashed
	FirstName string `gorm:"not null"`
	LastName  string `gorm:"not null"`
	// Role      string `gorm:"default:'user'"` // Deprecated in favor of RoleID

	RoleID     *uuid.UUID  `gorm:"type:uuid;index"`
	AccessRole *AccessRole `gorm:"foreignKey:RoleID"`

	BusinessID *uuid.UUID `gorm:"type:uuid;index"`
	Business   *Business  `gorm:"foreignKey:BusinessID"`
}
