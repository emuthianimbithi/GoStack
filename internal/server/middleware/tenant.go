package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")

		// 1. Handle Master Admin Switching
		if role == "masteradmin" {
			impersonateID := c.GetHeader("X-Impersonate-Business-ID")
			if impersonateID != "" {
				// Master Admin is impersonating a business
				c.Set("businessID", impersonateID)
			}
			// If no header, Master Admin is context-free (system wide) or operating on their own context
			// We don't force a businessID in this case
			c.Next()
			return
		}

		// 2. Normal Users MUST have a Business ID
		_, exists := c.Get("businessID")
		if !exists {
			httpx.Forbidden(c, "Context missing Business ID")
			c.Abort()
			return
		}

		// 3. (Optional) Check Key Business Permissions/Features if we had a service injected here
		// For now, we rely on the businessID being present to scope DB queries.

		c.Next()
	}
}
