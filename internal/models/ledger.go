package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type LedgerEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	BusinessID *uuid.UUID `gorm:"type:uuid;index"` // The tenant receiving money (if applicable) or paying
	Business   *Business  `gorm:"foreignKey:BusinessID"`

	// Transaction Details
	Amount      float64 `gorm:"not null"`                // e.g., 10.00
	Currency    string  `gorm:"size:3;default:'FJD'"`    // ISO code
	Provider    string  `gorm:"index"`                   // stripe, mpesa
	ProviderRef string  `gorm:"index"`                   // stripe_charge_id
	Status      string  `gorm:"index;default:'pending'"` // pending, paid, failed, refunded

	Description string
	Metadata    datatypes.JSON `gorm:"type:jsonb"`
}

func (LedgerEntry) TableName() string {
	return "billing_ledger"
}
