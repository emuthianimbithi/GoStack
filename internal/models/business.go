package models

import (
	"github.com/google/uuid"
)

type Business struct {
	BaseModel
	Name string `gorm:"not null"`
	Slug string `gorm:"uniqueIndex;not null"`
}

type MenuItem struct {
	BaseModel
	Title        string `gorm:"not null"`
	Path         string `gorm:"not null"`
	Icon         string
	ParentID     *uuid.UUID `gorm:"type:uuid"`
	SortOrder    int        `gorm:"default:0"`
	FeatureGroup string     `gorm:"index"` // e.g. "billing", "users"
	RequiredRole string     // Optional: If set, user must have this role name (or "masteradmin")

	// Relationships
	Parent   *MenuItem  `gorm:"foreignKey:ParentID"`
	Children []MenuItem `gorm:"foreignKey:ParentID"`
}

type BusinessPermission struct {
	BaseModel
	BusinessID uuid.UUID `gorm:"type:uuid;index;not null"`
	Feature    string    `gorm:"index;not null"` // "billing", "users", etc. matches MenuItem.FeatureGroup or APIResource.Group

	Business Business `gorm:"foreignKey:BusinessID"`
}
