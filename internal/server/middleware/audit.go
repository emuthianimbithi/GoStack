package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func AuditMiddleware(auditService *services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read Request Body
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// Wrap Response Writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process Request
		c.Next()

		// Skip auditing GET requests unless configured otherwise (to save DB space)
		// But User asked for billing billing, so maybe we SHOULD audit GET if it costs money?
		// For now, let's stick to state-changing methods + configured ones.
		// Actually, simpler: Audit Everything that is an API call.
		if c.Request.Method == "OPTIONS" {
			return
		}

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		// Exact User/Business context
		userIDStr := c.GetString("userID")
		var userID *uuid.UUID
		if userIDStr != "" {
			if uid, err := uuid.Parse(userIDStr); err == nil {
				userID = &uid
			}
		}

		businessIDStr := c.GetString("businessID")
		var businessID *uuid.UUID
		if businessIDStr != "" {
			if bid, err := uuid.Parse(businessIDStr); err == nil {
				businessID = &bid
			}
		}

		// Don't log payload for GET to save space, or redacting passwords

		// Naive redaction:
		var payloadBytes []byte = reqBodyBytes
		if len(reqBodyBytes) > 0 {
			var bodyMap map[string]interface{}
			if err := json.Unmarshal(reqBodyBytes, &bodyMap); err == nil {
				redactSensitive(bodyMap)
				if redacted, err := json.Marshal(bodyMap); err == nil {
					payloadBytes = redacted
				}
			}
		}

		logEntry := &models.AuditLog{
			Action:     c.Request.Method,
			Resource:   c.Request.URL.Path,
			UserID:     userID,
			BusinessID: businessID,
			Payload:    datatypes.JSON(payloadBytes),
			DurationMs: duration,
			ReqBytes:   len(reqBodyBytes),
			RespBytes:  blw.body.Len(),
			Status:     status,
		}

		// Async log to not block response?
		// For billing accuracy, sync might be safer, but async is better for latency.
		// Go routine is fine.
		go auditService.Log(logEntry)
	}
}

func redactSensitive(data map[string]interface{}) {
	sensitiveKeys := []string{"password", "token", "secret", "refresh_token", "credit_card"}
	for k, v := range data {
		for _, sensitive := range sensitiveKeys {
			if k == sensitive {
				data[k] = "[REDACTED]"
			}
		}
		// Recursive redaction for nested maps
		if nested, ok := v.(map[string]interface{}); ok {
			redactSensitive(nested)
		}
	}
}
