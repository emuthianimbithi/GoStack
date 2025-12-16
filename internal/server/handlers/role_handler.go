package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/emuthianimbithi/GoStack/internal/permissions"
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
)

type RoleHandler struct {
	permService *permissions.Service
}

func NewRoleHandler(ps *permissions.Service) *RoleHandler {
	return &RoleHandler{permService: ps}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request", err.Error())
		return
	}

	// Extract BusinessID from context (set by TenantMiddleware)
	var businessIDPtr *uuid.UUID
	if bid, exists := c.Get("businessID"); exists {
		if idStr, ok := bid.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				businessIDPtr = &parsed
			}
		}
	}

	role, err := h.permService.CreateRole(c.Request.Context(), req.Name, req.Description, businessIDPtr)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}
	httpx.JSON(c, http.StatusCreated, role)
}

func (h *RoleHandler) List(c *gin.Context) {
	var businessIDPtr *string
	if bid, exists := c.Get("businessID"); exists {
		if idStr, ok := bid.(string); ok {
			businessIDPtr = &idStr
		}
	}

	roles, err := h.permService.ListRoles(c.Request.Context(), businessIDPtr)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}
	httpx.JSON(c, http.StatusOK, roles)
}

func (h *RoleHandler) AssignPermission(c *gin.Context) {
	roleIDStr := c.Param("id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		httpx.BadRequest(c, "Invalid Role ID", nil)
		return
	}

	var req struct {
		ResourceID uuid.UUID `json:"resource_id" binding:"required"`
		Allowed    bool      `json:"allowed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request", err.Error())
		return
	}

	if err := h.permService.AssignPermission(c.Request.Context(), roleID, req.ResourceID, req.Allowed); err != nil {
		httpx.InternalError(c, err.Error())
		return
	}
	httpx.JSON(c, http.StatusOK, gin.H{"status": "assigned"})
}
