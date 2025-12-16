package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/emuthianimbithi/GoStack/internal/constants"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/httpx"
	"gorm.io/gorm"
)

type MenuHandler struct {
	db *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{db: db}
}

// GetMenu returns the menu structure for the current user/business
func (h *MenuHandler) GetMenu(c *gin.Context) {
	businessID, exists := c.Get("businessID")
	role, _ := c.Get("role")

	if !exists {
		// If master admin without impersonation could see everything, or nothing.
		// For now returning empty if no business context.
		// Or if master admin, show system menu?
		if role == "masteradmin" {
			// Return system menu (items with FeatureGroup="system" maybe?)
			// For simplicity, returning empty for now.
			httpx.JSON(c, http.StatusOK, []models.MenuItem{})
			return
		}
		httpx.Forbidden(c, "No Business Context Failed to load menu")
		return
	}

	// 1. Fetch all menu items
	var allItems []models.MenuItem
	// Order by SortOrder
	if err := h.db.Order("sort_order asc").Find(&allItems).Error; err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	// 2. Fetch Business Permissions (Enabled Features)
	var perms []models.BusinessPermission
	if err := h.db.Where("business_id = ?", businessID).Find(&perms).Error; err != nil {
		httpx.InternalError(c, err.Error())
		return
	}

	enabledFeatures := make(map[string]bool)
	for _, p := range perms {
		enabledFeatures[p.Feature] = true
	}

	// 3. Filter Items
	var allowedItems []models.MenuItem
	for _, item := range allItems {
		// If item belongs to a feature group, check if business has it enabled
		if item.FeatureGroup != "" {
			if !enabledFeatures[item.FeatureGroup] {
				continue
			}
		}
		// Check Role Requirement
		if item.RequiredRole != "" {
			// User must have this role. "masteradmin" bypasses everything?
			// The user said "admins can only see their business roles only".
			// Assuming "masteradmin" sees everything or just system stuff.
			// Let's implement strict check:
			if role != constants.RoleMasterAdmin.String() && role != item.RequiredRole {
				continue
			}
		}

		allowedItems = append(allowedItems, item)
	}

	// 4. Build Hierarchy (naive implementation)
	tree := buildMenuTree(allowedItems, nil)
	httpx.JSON(c, http.StatusOK, tree)
}

func buildMenuTree(items []models.MenuItem, parentID *string) []models.MenuItem {
	var nodes []models.MenuItem
	// This is O(N^2) naive, fine for small menus. Optimize if needed.
	for _, i := range items {
		// Check parent match
		if (parentID == nil && i.ParentID == nil) ||
			(parentID != nil && i.ParentID != nil && i.ParentID.String() == *parentID) {

			children := buildMenuTree(items, strPtr(i.ID.String()))
			i.Children = children
			nodes = append(nodes, i)
		}
	}
	return nodes
}

func strPtr(s string) *string {
	return &s
}
