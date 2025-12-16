package permissions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/emuthianimbithi/GoStack/internal/constants"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewService(db *gorm.DB, rdb *redis.Client) *Service {
	return &Service{db: db, rdb: rdb}
}

// IsAllowed checks if a role has access to a resource (cached)
func (s *Service) IsAllowed(ctx context.Context, role, method, path string) bool {
	key := fmt.Sprintf("rbac:%s:%s:%s", role, method, path)

	// 1. Check Redis
	if val, err := s.rdb.Get(ctx, key).Result(); err == nil {
		return val == "1"
	}

	// 2. DB Fallback
	allowed := s.queryDB(role, method, path)

	// 3. Cache it (TTL 1 hour)
	val := "0"
	if allowed {
		val = "1"
	}
	s.rdb.Set(ctx, key, val, time.Hour)

	return allowed
}

// InvalidateCache clears the cache for a specific role and resource
func (s *Service) InvalidateCache(ctx context.Context, role, method, path string) error {
	key := fmt.Sprintf("rbac:%s:%s:%s", role, method, path)
	return s.rdb.Del(ctx, key).Err()
}

// InvalidateAll clears all RBAC related keys
func (s *Service) InvalidateAll(ctx context.Context) error {
	// Scan for keys rbac:*
	var cursor uint64
	var keys []string
	var err error

	// WARNING: In production with millions of keys, SCAN should be used carefully with a cursor loop.
	// For simplicity in this demo, we assume relatively few RBAC keys or accept the overhead.
	for {
		keys, cursor, err = s.rdb.Scan(ctx, cursor, "rbac:*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (s *Service) queryDB(roleName, method, path string) bool {
	var count int64
	err := s.db.Model(&models.RolePermission{}).
		Joins("JOIN access_roles ON access_roles.id = role_permissions.role_id").
		Joins("JOIN api_resources ON api_resources.id = role_permissions.resource_id").
		Where("access_roles.name = ? AND api_resources.method = ? AND api_resources.path = ? AND role_permissions.allowed = ?", roleName, method, path, true).
		Count(&count).Error

	return err == nil && count > 0
}

// CreateRole creates a new role (System or Tenant scoped)
func (s *Service) CreateRole(ctx context.Context, name, description string, businessID *uuid.UUID) (*models.AccessRole, error) {
	role := &models.AccessRole{
		Name:        name,
		Description: description,
		BusinessID:  businessID,
	}
	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// CreateResource creates a new API resource
func (s *Service) CreateResource(ctx context.Context, method, path, group, desc string) (*models.APIResource, error) {
	res := &models.APIResource{
		Method:      method,
		Path:        path,
		Group:       group,
		Description: desc,
	}
	if err := s.db.Create(res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// AssignPermission assigns a permission to a role and invalidates cache
func (s *Service) AssignPermission(ctx context.Context, roleID, resourceID uuid.UUID, allowed bool) error {
	// 1. Create/Update RolePermission
	// Upsert usually preferable
	perm := models.RolePermission{
		RoleID:     roleID,
		ResourceID: resourceID,
		Allowed:    allowed,
	}

	err := s.db.Where(models.RolePermission{RoleID: roleID, ResourceID: resourceID}).
		Assign(models.RolePermission{Allowed: allowed}).
		FirstOrCreate(&perm).Error

	if err != nil {
		return err
	}

	// 2. Invalidate Cache
	// We need the role name and resource method/path to invalidate specific key.
	// Fetching them is one extra query, or we just InvalidateAll for simplicity in this MVP?
	// Let's InvalidateAll for now as permission changes are rare.
	return s.InvalidateAll(ctx)
}

// ListRoles returns roles visible to the context (System roles + Tenant roles)
func (s *Service) ListRoles(ctx context.Context, businessID *string) ([]models.AccessRole, error) {
	var roles []models.AccessRole
	query := s.db.Model(&models.AccessRole{})

	if businessID != nil {
		// Show Tenant Roles AND System Roles?
		// User said "admins can only see their business roles only".
		// If they need to see masteradmin (to know they aren't it?), maybe not.
		// Let's strict scope: Only business roles.
		// BUT wait, 'admin' role created during register is a Business Role now.
		query = query.Where("business_id = ?", *businessID)
	} else {
		// If no businessID (System Context), show System Roles?
		// Or show all? Let's show Global Roles (BusinessID IS NULL)
		query = query.Where("business_id IS NULL")
	}

	err := query.Find(&roles).Error
	return roles, err
}

// EnsureSystemRoles seeds default system roles
func (s *Service) EnsureSystemRoles() error {
	ctx := context.Background()
	// Only System Level Roles here.
	// The "admin" role for a business is created dynamically during registration.
	// "masteradmin" is the only global one.
	roles := []struct {
		Name string
		Desc string
	}{
		{constants.RoleMasterAdmin.String(), "System Administrator"},
	}

	for _, r := range roles {
		// System roles have nil BusinessID
		if _, err := s.CreateRole(ctx, r.Name, r.Desc, nil); err != nil {
			// If already exists (duplicate key), ignore. In real app check specific error.
			// assuming gorm error wrapping or check logic
			continue
		}
	}
	return nil
}
