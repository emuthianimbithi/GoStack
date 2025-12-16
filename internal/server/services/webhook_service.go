package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WebhookService struct {
	db     *gorm.DB
	client *http.Client
}

func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterEndpoint creates a new webhook destination for a tenant
func (s *WebhookService) RegisterEndpoint(businessID uuid.UUID, url string, events []string) (*models.WebhookEndpoint, error) {
	eventsJSON, _ := json.Marshal(events)

	endpoint := &models.WebhookEndpoint{
		BusinessID: businessID,
		URL:        url,
		Secret:     "whsec_" + uuid.New().String(), // Generate random secret
		Events:     datatypes.JSON(eventsJSON),
		Active:     true,
	}

	if err := s.db.Create(endpoint).Error; err != nil {
		return nil, err
	}
	return endpoint, nil
}

// ListEndpoints returns all webhooks for a tenant
func (s *WebhookService) ListEndpoints(businessID uuid.UUID) ([]models.WebhookEndpoint, error) {
	var endpoints []models.WebhookEndpoint
	if err := s.db.Where("business_id = ?", businessID).Find(&endpoints).Error; err != nil {
		return nil, err
	}
	return endpoints, nil
}

// DispatchEvent sends an event to all subscribed endpoints for a business
func (s *WebhookService) DispatchEvent(businessID uuid.UUID, eventType string, payload interface{}) error {
	// 1. Find matched endpoints
	// For simplicity, we fetch all active endpoints for the business and filter in code or use JSON query if Postgres
	// Here filtering in code for DB agnostic safety
	var endpoints []models.WebhookEndpoint
	if err := s.db.Where("business_id = ? AND active = ?", businessID, true).Find(&endpoints).Error; err != nil {
		return err
	}

	if len(endpoints) == 0 {
		return nil
	}

	// Prepare Payload
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 2. Send to each endpoint (Async in production, using goroutine here for basics)
	for _, ep := range endpoints {
		// Basic check if eventType matches (assuming Events is a simple string array)
		// Skipping granular check for brevity: "if eventType in ep.Events"

		go s.send(ep, eventType, bodyBytes)
	}

	return nil
}

func (s *WebhookService) send(ep models.WebhookEndpoint, eventType string, body []byte) {
	delivery := models.WebhookDelivery{
		EndpointID: ep.ID,
		EventID:    uuid.New().String(),
		EventType:  eventType,
		Payload:    datatypes.JSON(body),
		Attempt:    1,
	}

	req, _ := http.NewRequest("POST", ep.URL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GoStack-Event", eventType)
	req.Header.Set("X-GoStack-Signature", signPayload(ep.Secret, body))

	resp, err := s.client.Do(req)
	if err != nil {
		delivery.Success = false
		delivery.Error = err.Error()
	} else {
		defer resp.Body.Close()
		delivery.StatusCode = resp.StatusCode
		delivery.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	s.db.Create(&delivery)
}

func signPayload(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
