package services_test

import (
	"context"
	"testing"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.User{}, &models.Business{})
	return db
}

func TestRegisterUser_Success(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)

	// Mock Auth Service (not really needed for Register unless it calls it?)
	// UserService.NewUserService takes AuthService but RegisterUser doesn't use it.
	// We can pass nil for now or a dummy.
	authSvc := services.NewAuthService(config.AuthConfig{}, repo)
	userSvc := services.NewUserService(repo, authSvc)

	req := services.RegisterRequest{
		BusinessName: "Acme Corp",
		Email:        "admin@acme.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
	}

	err := userSvc.RegisterUser(context.Background(), req)
	assert.NoError(t, err)

	// Verify Business Created
	var business models.Business
	err = db.Where("name = ?", "Acme Corp").First(&business).Error
	assert.NoError(t, err)
	assert.Equal(t, "Acme Corp", business.Slug)

	// Verify User Created and Linked
	var user models.User
	// Preload AccessRole to check the name
	err = db.Preload("AccessRole").Where("email = ?", "admin@acme.com").First(&user).Error
	require.NoError(t, err)

	// Check AccessRole relation
	require.NotNil(t, user.AccessRole)
	assert.Equal(t, "admin", user.AccessRole.Name)

	require.NotNil(t, user.BusinessID)
	assert.Equal(t, business.ID, *user.BusinessID)
}

func TestRegisterUser_Duplicate(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)
	authSvc := services.NewAuthService(config.AuthConfig{}, repo)
	userSvc := services.NewUserService(repo, authSvc)

	req := services.RegisterRequest{
		BusinessName: "Acme Corp",
		Email:        "admin@acme.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
	}

	// First registration
	_ = userSvc.RegisterUser(context.Background(), req)

	// Second registration (Same email)
	err := userSvc.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
}
