package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/emuthianimbithi/GoStack/internal/permissions"
)

func RBACMiddleware(permService *permissions.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No role found in context"})
			return
		}

		method := c.Request.Method
		path := c.FullPath() // Use FullPath to get the registered route path (e.g., /api/v1/users/:id)

		if !permService.IsAllowed(c.Request.Context(), role.(string), method, path) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.Next()
	}
}
