package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	BaseModel
	Name        string `gorm:"not null"` // Free, Pro, Enterprise
	Description string
	Price       float64 `gorm:"not null"`
	Currency    string  `gorm:"default:'USD'"`
	Interval    string  `gorm:"default:'month'"` // month, year

	// External Provider IDs
	StripePriceID string
	MpesaPlanID   string

	Features []PlanFeature `gorm:"foreignKey:PlanID"`
}

type PlanFeature struct {
	BaseModel
	PlanID       uuid.UUID `gorm:"type:uuid;index;not null"`
	FeatureKey   string    `gorm:"index"` // max_users, audit_logs
	FeatureValue string    // "5", "true", "unlimited"
}

type Subscription struct {
	BaseModel
	BusinessID uuid.UUID `gorm:"type:uuid;index;not null"`
	PlanID     uuid.UUID `gorm:"type:uuid;index;not null"`

	Status string `gorm:"index;default:'active'"` // active, past_due, canceled, incomplete

	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool

	Provider    string `gorm:"default:'stripe'"`
	ProviderRef string `gorm:"index"` // sub_123456

	Business Business `gorm:"foreignKey:BusinessID"`
	Plan     Plan     `gorm:"foreignKey:PlanID"`
}
