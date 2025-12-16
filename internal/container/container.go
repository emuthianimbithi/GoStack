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
	UserRepository *repositories.UserRepository

	// Services
	AuthService *services.AuthService
	UserService *services.UserService
	PermService *permissions.Service

	// Handlers
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	BusinessHandler *handlers.BusinessHandler
	MenuHandler     *handlers.MenuHandler
	RoleHandler     *handlers.RoleHandler
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
}

func (c *Container) initServices() {
	c.PermService = permissions.NewService(c.DB, c.RDB)
	c.AuthService = services.NewAuthService(c.Config.Auth, c.UserRepository)
	c.UserService = services.NewUserService(c.UserRepository, c.AuthService)
}

func (c *Container) initHandlers() {
	c.UserHandler = handlers.NewUserHandler(c.UserService)
	c.AuthHandler = handlers.NewAuthHandler(c.UserService)
	c.BusinessHandler = handlers.NewBusinessHandler(c.DB)
	c.MenuHandler = handlers.NewMenuHandler(c.DB)
	c.RoleHandler = handlers.NewRoleHandler(c.PermService)
}
