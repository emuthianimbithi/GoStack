package handlers

import (
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler(service *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.Unauthorized(c, "Invalid User ID")
		return
	}

	notifs, err := h.service.ListByUser(userID, 50)
	if err != nil {
		httpx.InternalError(c, "Failed to fetch notifications")
		return
	}
	httpx.Ok(c, notifs)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr) // Middleware ensures validity usually

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.BadRequest(c, "Invalid Notification ID", nil)
		return
	}

	if err := h.service.MarkAsRead(id, userID); err != nil {
		httpx.InternalError(c, "Failed to mark as read")
		return
	}
	httpx.NoContent(c)
}
