package models

import (
	"github.com/google/uuid"
)

type AccessRole struct {
	BaseModel
	Name        string `gorm:"index;not null"` // e.g., "admin", "viewer"
	Description string
	BusinessID  *uuid.UUID `gorm:"type:uuid;index"` // Null for system roles (masteradmin), set for tenant roles
	Business    *Business  `gorm:"foreignKey:BusinessID"`
}

type APIResource struct {
	BaseModel
	Method      string `gorm:"index;not null"` // GET, POST, etc.
	Path        string `gorm:"index;not null"` // /api/v1/users
	Group       string `gorm:"index"`          // e.g. "User Management"
	Description string
}

type RolePermission struct {
	BaseModel
	RoleID     uuid.UUID   `gorm:"type:uuid;index;not null"`
	ResourceID uuid.UUID   `gorm:"type:uuid;index;not null"`
	Role       AccessRole  `gorm:"foreignKey:RoleID"`
	Resource   APIResource `gorm:"foreignKey:ResourceID"`
	Allowed    bool        `gorm:"default:true"`
}
