package handlers

import (
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	service *services.UploadService
}

func NewFileHandler(service *services.UploadService) *FileHandler {
	return &FileHandler{service: service}
}

func (h *FileHandler) Upload(c *gin.Context) {
	// Parse Business
	businessID, exists := c.Get("businessID")
	var businessUUID *uuid.UUID
	if exists {
		bID := businessID.(*uuid.UUID)
		businessUUID = bID
	}

	// Parse User
	userID := c.MustGet("userID").(string)
	userUUID, _ := uuid.Parse(userID)

	// Get File
	file, err := c.FormFile("file")
	if err != nil {
		httpx.BadRequest(c, "No file provided", err.Error())
		return
	}

	// Upload
	record, err := h.service.UploadFile(file, userUUID, businessUUID)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.Ok(c, record)
}
