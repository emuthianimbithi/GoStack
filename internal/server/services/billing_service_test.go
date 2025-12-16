package services_test

import (
	"testing"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBillingTestDB() *gorm.DB {
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(
		&models.LedgerEntry{},
		&models.Plan{},
		&models.Subscription{},
		&models.Business{},
	)
	return db
}

func TestRecordTransaction(t *testing.T) {
	db := setupBillingTestDB()
	cfg := config.BillingConfig{}
	svc := services.NewBillingService(db, cfg)

	bizID := uuid.New()
	entry, err := svc.RecordTransaction(&bizID, 100.0, "mpesa", "ws_123", "Test Payment")

	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, 100.0, entry.Amount)
	assert.Equal(t, "pending", entry.Status)

	// Verify DB persistence
	var saved models.LedgerEntry
	err = db.First(&saved, entry.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "ws_123", saved.ProviderRef)
}

func TestHandleCallback(t *testing.T) {
	db := setupBillingTestDB()
	cfg := config.BillingConfig{}
	svc := services.NewBillingService(db, cfg)

	// Create pending transaction
	bizID := uuid.New()
	_, err := svc.RecordTransaction(&bizID, 50.0, "mpesa", "ref_999", "Pending")
	require.NoError(t, err)

	// Update Status
	err = svc.HandleCallback("mpesa", "ref_999", "completed")
	assert.NoError(t, err)

	// Verify
	var saved models.LedgerEntry
	err = db.Where("provider_ref = ?", "ref_999").First(&saved).Error
	assert.NoError(t, err)
	assert.Equal(t, "completed", saved.Status)
}

func TestHandleCallback_NotFound(t *testing.T) {
	db := setupBillingTestDB()
	cfg := config.BillingConfig{}
	svc := services.NewBillingService(db, cfg)

	err := svc.HandleCallback("mpesa", "non_existent", "completed")
	assert.Error(t, err)
	assert.Equal(t, "transaction not found", err.Error())
}

func TestSubscribe_Mpesa_FreePlan(t *testing.T) {
	db := setupBillingTestDB()
	cfg := config.BillingConfig{MpesaConsumerKey: "key"}
	svc := services.NewBillingService(db, cfg)

	plan := models.Plan{
		Name:  "Free Tier",
		Price: 0,
	}
	db.Create(&plan)

	err := svc.SubscribeMpesa(uuid.New(), plan.ID, "254700000000")
	assert.Error(t, err)
	assert.Equal(t, "cannot process free plan via mpesa", err.Error())
}
