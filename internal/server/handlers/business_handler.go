package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"gorm.io/gorm"
)

type BusinessHandler struct {
	db *gorm.DB
}

func NewBusinessHandler(db *gorm.DB) *BusinessHandler {
	return &BusinessHandler{db: db}
}

func (h *BusinessHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	business := models.Business{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := h.db.Create(&business).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			httpx.BadRequest(c, "Business slug already exists", nil)
			return
		}
		httpx.InternalError(c, err.Error())
		return
	}

	httpx.JSON(c, http.StatusCreated, business)
}

func (h *BusinessHandler) List(c *gin.Context) {
	var businesses []models.Business
	meta, err := httpx.Paginate(c, h.db.Model(&models.Business{}), &businesses)
	if err != nil {
		httpx.InternalError(c, err.Error())
		return
	}
	httpx.JSON(c, http.StatusOK, httpx.WrapPaged(businesses, meta))
}
