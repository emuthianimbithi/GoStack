package services

import (
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
	// fcmClient *messaging.Client // Future: Firebase integration
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Send creates a notification in the DB. In the future, it will also push to FCM.
func (s *NotificationService) Send(userID uuid.UUID, title, message, notifType string) error {
	n := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		IsRead:  false,
	}
	return s.db.Create(&n).Error
}

func (s *NotificationService) ListByUser(userID uuid.UUID, limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&notifs).Error
	return notifs, err
}

func (s *NotificationService) MarkAsRead(id uuid.UUID, userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (s *NotificationService) MarkAllRead(userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
