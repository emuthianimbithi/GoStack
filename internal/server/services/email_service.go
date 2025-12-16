package services

import (
	"log"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gorm.io/gorm"
)

type EmailService struct {
	cfg config.EmailConfig
	db  *gorm.DB
}

func NewEmailService(cfg config.EmailConfig, db *gorm.DB) *EmailService {
	return &EmailService{cfg: cfg, db: db}
}

func (s *EmailService) Send(toEmail, toName, subject, htmlContent string, businessID *uuid.UUID) error {
	apiKey := s.cfg.SendGridKey
	fromEmail := s.cfg.FromEmail
	fromName := s.cfg.FromName

	// 1. Check if Business has Custom Settings
	if businessID != nil {
		var settings models.BusinessSettings
		if err := s.db.Where("business_id = ?", businessID).First(&settings).Error; err == nil {
			if settings.SendGridKey != "" {
				apiKey = settings.SendGridKey
			}
			if settings.SenderEmail != "" {
				fromEmail = settings.SenderEmail
			}
			if settings.SenderName != "" {
				fromName = settings.SenderName
			}
		}
	}

	if apiKey == "" {
		log.Println("[EmailService] No API Key found. Skipping email:", subject)
		return nil
	}

	from := mail.NewEmail(fromName, fromEmail)
	to := mail.NewEmail(toName, toEmail)
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(apiKey)

	// In production, handle retries
	response, err := client.Send(message)
	if err != nil {
		log.Printf("[EmailService] Error sending email: %v", err)
		return err
	}

	if response.StatusCode >= 400 {
		log.Printf("[EmailService] SendGrid failed: %d %s", response.StatusCode, response.Body)
	}

	return nil
}
