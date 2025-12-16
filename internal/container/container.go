package container

import (
	"log"
	"os"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/db"
	"github.com/emuthianimbithi/GoStack/internal/permissions"
	"github.com/emuthianimbithi/GoStack/internal/server/handlers"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	// Core
	DB     *gorm.DB
	RDB    *redis.Client
	Config config.AppConfig

	// Repositories (Internal usage mostly, but exposed if needed)
	UserRepository  *repositories.UserRepository
	AuditRepository *repositories.AuditRepository

	// Services
	AuthService         *services.AuthService
	UserService         *services.UserService
	PermService         *permissions.Service
	AuditService        *services.AuditService
	EmailService        *services.EmailService
	NotificationService *services.NotificationService
	BillingService      *services.BillingService
	WebhookService      *services.WebhookService
	UploadService       *services.UploadService

	// Handlers
	AuthHandler         *handlers.AuthHandler
	UserHandler         *handlers.UserHandler
	BusinessHandler     *handlers.BusinessHandler
	MenuHandler         *handlers.MenuHandler
	RoleHandler         *handlers.RoleHandler
	NotificationHandler *handlers.NotificationHandler
	BillingHandler      *handlers.BillingHandler
	WebhookHandler      *handlers.WebhookHandler
	FileHandler         *handlers.FileHandler
}

// New creates and initializes a new Container
func New(cfg config.AppConfig) *Container {
	c := &Container{Config: cfg}

	c.initInfrastructure()
	c.initRepositories()
	c.initServices()
	c.initHandlers()

	return c
}

func (c *Container) initInfrastructure() {
	// Database
	c.DB = db.Connect(c.Config.DB)

	// Redis
	// In a real scenario, move REDIS_URL to config package
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}
	c.RDB = redis.NewClient(opt)
}

func (c *Container) initRepositories() {
	c.UserRepository = repositories.NewUserRepository(c.DB)
	c.AuditRepository = repositories.NewAuditRepository(c.DB)
}

func (c *Container) initServices() {
	c.PermService = permissions.NewService(c.DB, c.RDB)
	c.AuthService = services.NewAuthService(c.Config.Auth, c.UserRepository)

	// Ensure EmailService is init BEFORE UserService
	c.EmailService = services.NewEmailService(c.Config.Email, c.DB)

	c.UserService = services.NewUserService(c.UserRepository, c.AuthService, c.EmailService)

	if c.Config.Modules.AuditEnabled {
		c.AuditService = services.NewAuditService(c.AuditRepository)
	}

	c.NotificationService = services.NewNotificationService(c.DB)
	c.BillingService = services.NewBillingService(c.DB, c.Config.Billing)
	c.WebhookService = services.NewWebhookService(c.DB)
	c.UploadService = services.NewUploadService(c.DB)
}

func (c *Container) initHandlers() {
	c.AuthHandler = handlers.NewAuthHandler(c.UserService)
	// c.UserHandler = handlers.NewUserHandler(c.UserService) // Already below? No, let's keep order clean
	c.UserHandler = handlers.NewUserHandler(c.UserService)
	c.BusinessHandler = handlers.NewBusinessHandler(c.DB)
	c.MenuHandler = handlers.NewMenuHandler(c.DB)
	c.RoleHandler = handlers.NewRoleHandler(c.PermService)
	c.NotificationHandler = handlers.NewNotificationHandler(c.NotificationService)
	c.BillingHandler = handlers.NewBillingHandler(c.BillingService)
	c.WebhookHandler = handlers.NewWebhookHandler(c.WebhookService)
	c.FileHandler = handlers.NewFileHandler(c.UploadService)
}
