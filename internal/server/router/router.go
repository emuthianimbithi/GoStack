package router

import (
	"github.com/emuthianimbithi/GoStack/internal/container"
	"github.com/emuthianimbithi/GoStack/internal/server/middleware"
	"github.com/gin-gonic/gin"
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

	if c.Config.Modules.AuditEnabled && c.AuditService != nil {
		api.Use(middleware.AuditMiddleware(c.AuditService))
	}

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
		api.POST("/invites", c.UserHandler.Invite) // New Invite Route

		// Notifications
		api.GET("/notifications", c.NotificationHandler.List)
		api.POST("/notifications/:id/read", c.NotificationHandler.MarkRead)

		// Callbacks (Publicish? No, protected usually or separate group. Keeping in v1 for now but might need open endpoint for webhooks)
		// Webhooks usually need to be public or use signature verification.
		// For now, let's put them in a separate group if they shouldn't be auth protected, OR assume the provider has a key.
		// Let's create a separate group "callbacks" that is public but verified manually.
	}

	callbacks := r.Group("/api/v1/callbacks")
	// callbacks.Use(middleware.WebhookAuth()) // Verify Stripe Signature etc.
	{
		callbacks.POST("/:provider", c.BillingHandler.PaymentCallback)
	}

	// Billing Management
	billing := r.Group("/api/v1/billing")
	billing.Use(middleware.AuthMiddleware(c.AuthService))
	billing.Use(middleware.TenantMiddleware())
	billing.Use(middleware.RBACMiddleware(c.PermService))
	{
		billing.POST("/subscribe", c.BillingHandler.Subscribe)
		billing.POST("/cancel", c.BillingHandler.Cancel)
	}

	// Webhooks (Tenants managing their outbound hooks)
	webhooks := r.Group("/api/v1/webhooks")
	webhooks.Use(middleware.AuthMiddleware(c.AuthService))
	webhooks.Use(middleware.TenantMiddleware())
	webhooks.Use(middleware.RBACMiddleware(c.PermService))
	{
		webhooks.POST("", c.WebhookHandler.Create)
		webhooks.GET("", c.WebhookHandler.List)
	}

	// Files
	files := r.Group("/api/v1/files")
	files.Use(middleware.AuthMiddleware(c.AuthService))
	// Tenant middleware? Optional, files can be user-level or tenant-level. FileHandler handles both logic.
	// But let's assume authenticated user context is enough for now,
	// FileHandler manually extracts businessID from context if present (set by TenantMiddleware if used).
	// Let's rely on AuthMiddleware for UserID.
	{
		files.POST("/upload", c.FileHandler.Upload)
	}

	return r
}
