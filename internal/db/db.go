package db

import (
	"log"
	"time"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.DBConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Connection Pool settings
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// OpenTelemetry Instrumentation
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		log.Printf("failed to use otelgorm plugin: %v", err)
	}

	return db
}

func AutoMigrate(db *gorm.DB) error {
	// Register all models here
	return db.AutoMigrate(
		&models.User{},
		&models.Business{},
		&models.AccessRole{},
		&models.APIResource{},
		&models.RolePermission{},
		&models.MenuItem{},
		&models.BusinessPermission{},
		&models.BusinessSettings{},
		&models.AuditLog{},
		&models.Notification{},
		&models.UserInvite{},
		&models.LedgerEntry{},
		&models.Plan{},
		&models.PlanFeature{},
		&models.Subscription{},
		&models.WebhookEndpoint{},
		&models.WebhookDelivery{},
		&models.File{},
	)
}
