package handlers

import (
	"net/http"

	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	if err := h.userService.RegisterUser(c.Request.Context(), req); err != nil {
		httpx.BadRequest(c, "Failed to register user", err.Error()) // Or conflict if exists
		return
	}

	httpx.JSON(c, http.StatusCreated, gin.H{"message": "user registered successfully"})
}

func (h *UserHandler) List(c *gin.Context) {
	var businessIDPtr *uuid.UUID
	if bid, exists := c.Get("businessID"); exists {
		if idStr, ok := bid.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				businessIDPtr = &parsed
			}
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), businessIDPtr)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}
	httpx.JSON(c, http.StatusOK, users)
}

type InviteRequest struct {
	Email  string `json:"email" binding:"required,email"`
	RoleID string `json:"role_id" binding:"required"`
}

func (h *UserHandler) Invite(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	userIDStr := c.GetString("userID")
	if !exists || userIDStr == "" {
		httpx.Forbidden(c, "Context Missing")
		return
	}

	businessUUID := businessID.(*uuid.UUID)
	inviterUUID, _ := uuid.Parse(userIDStr)

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request", err)
		return
	}

	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		httpx.BadRequest(c, "Invalid Role ID", nil)
		return
	}

	err = h.userService.InviteUser(c.Request.Context(), req.Email, *businessUUID, roleUUID, inviterUUID)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, gin.H{"message": "Invitation sent successfully"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpx.BadRequest(c, "Invalid User ID", nil)
		return
	}

	var businessIDPtr *uuid.UUID
	if bid, exists := c.Get("businessID"); exists {
		if idStr, ok := bid.(string); ok {
			if parsed, err := uuid.Parse(idStr); err == nil {
				businessIDPtr = &parsed
			}
		}
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id, businessIDPtr); err != nil {
		httpx.InternalError(c, err.Error()) // Or NotFound/Forbidden based on error
		return
	}
	httpx.JSON(c, http.StatusOK, gin.H{"message": "user deleted"})
}
