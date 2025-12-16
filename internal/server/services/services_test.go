package services_test

import (
	"log"
	"testing"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGenerateTokenPair(t *testing.T) {
	// Setup not strictly needed for logic test but helpful if we used Repo
	cfg := config.AuthConfig{
		JWTSecret:       "test-secret",
		AccessDuration:  15, // minutes
		RefreshDuration: 7,  // days
	}
	svc := services.NewAuthService(cfg, nil)

	uid := uuid.New()
	bid := uuid.New()
	user := &models.User{
		BaseModel: models.BaseModel{ID: uid},
		Email:     "test@example.com",
		RoleID:    nil, // AccessRole is what matters
		AccessRole: &models.AccessRole{
			Name: "admin",
		},
		BusinessID: &bid,
	}

	access, refresh, err := svc.GenerateTokenPair(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	// Validate Access Token
	claims, err := svc.ValidateToken(access)
	assert.NoError(t, err)
	assert.Equal(t, uid.String(), claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, bid.String(), *claims.BusinessID)
}

func TestValidateToken_Invalid(t *testing.T) {
	cfg := config.AuthConfig{JWTSecret: "test-secret"}
	svc := services.NewAuthService(cfg, nil)

	_, err := svc.ValidateToken("invalid.token.string")
	assert.Error(t, err)
}

// Notification Tests

func setupNotifTestDB() *gorm.DB {
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&models.Notification{}, &models.User{})
	return db
}

func TestNotificationFlow(t *testing.T) {
	db := setupNotifTestDB()
	svc := services.NewNotificationService(db)

	userID := uuid.New()

	// 1. Send Notification
	err := svc.Send(userID, "Welcome", "Hello World", "info")
	assert.NoError(t, err)

	// 2. List Notifications
	notifs, err := svc.ListByUser(userID, 10)
	assert.NoError(t, err)
	assert.Len(t, notifs, 1)
	assert.Equal(t, "Welcome", notifs[0].Title)
	assert.False(t, notifs[0].IsRead)
	notifID := notifs[0].ID

	// 3. Mark as Read
	err = svc.MarkAsRead(notifID, userID)
	assert.NoError(t, err)

	// Validate Read Status
	var n models.Notification
	db.First(&n, notifID)
	assert.True(t, n.IsRead)

	// 4. Mark All Read
	// Create another unread one
	svc.Send(userID, "Another", "msg", "warning")

	err = svc.MarkAllRead(userID)
	assert.NoError(t, err)

	count := int64(0)
	db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count)
	assert.Equal(t, int64(0), count)
}
