package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// WebhookEndpoint represents a destination URL registered by a tenant to receive events.
type WebhookEndpoint struct {
	BaseModel
	BusinessID  uuid.UUID `gorm:"type:uuid;index;not null"`
	URL         string    `gorm:"not null"`
	Description string
	Secret      string         `gorm:"not null"`   // Used to sign the payload (HMAC)
	Events      datatypes.JSON `gorm:"type:jsonb"` // Array of event types: ["order.created", "user.signup"]
	Active      bool           `gorm:"default:true"`

	Business Business `gorm:"foreignKey:BusinessID"`
}

// WebhookDelivery tracks the history of events sent to endpoints.
type WebhookDelivery struct {
	BaseModel
	EndpointID uuid.UUID      `gorm:"type:uuid;index;not null"`
	EventID    string         `gorm:"index"` // Unique event ID
	EventType  string         `gorm:"index"` // e.g. "order.created"
	Payload    datatypes.JSON `gorm:"type:jsonb"`
	StatusCode int            // HTTP Status from the user's server
	Success    bool
	Attempt    int
	Error      string

	Endpoint WebhookEndpoint `gorm:"foreignKey:EndpointID"`
}
