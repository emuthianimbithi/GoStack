package handlers

import (
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BillingHandler struct {
	service *services.BillingService
}

func NewBillingHandler(service *services.BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

// PaymentCallback handles webhooks from Stripe/Mpesa
func (h *BillingHandler) PaymentCallback(c *gin.Context) {
	provider := c.Param("provider") // stripe or mpesa

	// Mock parsing webhook payload
	// In reality, verify signature and parse JSON to get ref & status
	var payload struct {
		Ref    string `json:"ref"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		httpx.BadRequest(c, "Invalid Payload", err.Error())
		return
	}

	if err := h.service.HandleCallback(provider, payload.Ref, payload.Status); err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, gin.H{"received": true})
}

type SubscribeRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

func (h *BillingHandler) Subscribe(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	if !exists {
		httpx.Forbidden(c, "Context Missing")
		return
	}
	businessUUID := businessID.(*uuid.UUID)

	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request", err)
		return
	}

	planUUID, err := uuid.Parse(req.PlanID)
	if err != nil {
		httpx.BadRequest(c, "Invalid Plan ID", nil)
		return
	}

	checkoutURL, err := h.service.Subscribe(*businessUUID, planUUID)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, gin.H{"checkout_url": checkoutURL})
}

func (h *BillingHandler) Cancel(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	if !exists {
		httpx.Forbidden(c, "Context Missing")
		return
	}
	businessUUID := businessID.(*uuid.UUID)

	if err := h.service.CancelSubscription(*businessUUID); err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, gin.H{"message": "Subscription canceled"})
}
