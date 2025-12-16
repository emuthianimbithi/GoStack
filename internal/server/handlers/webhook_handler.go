package handlers

import (
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	service *services.WebhookService
}

func NewWebhookHandler(service *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

type CreateWebhookRequest struct {
	URL    string   `json:"url" binding:"required,url"`
	Events []string `json:"events" binding:"required"`
}

func (h *WebhookHandler) Create(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	if !exists {
		httpx.Forbidden(c, "Context Missing")
		return
	}
	businessUUID := businessID.(*uuid.UUID)

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request", err)
		return
	}

	ep, err := h.service.RegisterEndpoint(*businessUUID, req.URL, req.Events)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, ep)
}

func (h *WebhookHandler) List(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	if !exists {
		httpx.Forbidden(c, "Context Missing")
		return
	}
	businessUUID := businessID.(*uuid.UUID)

	endpoints, err := h.service.ListEndpoints(*businessUUID)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, endpoints)
}
