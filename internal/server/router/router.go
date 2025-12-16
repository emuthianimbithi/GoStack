package router

import (
	"github.com/gin-gonic/gin"
	"github.com/emuthianimbithi/GoStack/internal/container"
	"github.com/emuthianimbithi/GoStack/internal/server/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouter(c *container.Container) *gin.Engine {
	r := gin.Default()

	// Global Middleware
	r.Use(otelgin.Middleware(c.Config.OTLP.ServiceName))
	// r.Use(middleware.RateLimitMiddleware(c.RDB, 100, time.Minute)) // Example global rate limit

	// Public Routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", c.AuthHandler.Login)
		authGroup.POST("/register", c.UserHandler.Register)
	}

	// Protected Routes
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(c.AuthService))
	api.Use(middleware.TenantMiddleware()) // Ensure Business Context
	api.Use(middleware.RBACMiddleware(c.PermService))
	{
		// Business Management (Master Admin)
		api.POST("/businesses", c.BusinessHandler.Create)
		api.GET("/businesses", c.BusinessHandler.List)

		// Menus
		api.GET("/menus", c.MenuHandler.GetMenu)

		// Roles & Permissions (Master Admin or Tenant Admin with permission)
		api.POST("/roles", c.RoleHandler.Create)
		api.GET("/roles", c.RoleHandler.List)
		api.POST("/roles/:id/permissions", c.RoleHandler.AssignPermission)

		// Example resource
		// api.GET("/users", c.UserHandler.List)

		// User Management
		api.GET("/users", c.UserHandler.List)
		api.DELETE("/users/:id", c.UserHandler.Delete)
	}

	return r
}
