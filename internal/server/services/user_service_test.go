package services_test

import (
	"context"
	"testing"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	// Use a unique name for each test's DB to ensure isolation
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(
		&models.User{},
		&models.Business{},
		&models.AccessRole{},
		&models.UserInvite{},
	)
	return db
}

func TestRegisterUser_Success(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)
	authSvc := services.NewAuthService(config.AuthConfig{}, repo)
	userSvc := services.NewUserService(repo, authSvc, nil)

	req := services.RegisterRequest{
		BusinessName: "Acme Corp",
		Email:        "admin@acme.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
	}

	err := userSvc.RegisterUser(context.Background(), req)
	assert.NoError(t, err)

	var business models.Business
	err = db.Where("name = ?", "Acme Corp").First(&business).Error
	assert.NoError(t, err)
	assert.Equal(t, "Acme Corp", business.Slug)

	var user models.User
	err = db.Preload("AccessRole").Where("email = ?", "admin@acme.com").First(&user).Error
	require.NoError(t, err)

	require.NotNil(t, user.AccessRole)
	assert.Equal(t, "admin", user.AccessRole.Name)
	require.NotNil(t, user.BusinessID)
	assert.Equal(t, business.ID, *user.BusinessID)
}

func TestRegisterUser_Duplicate(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)
	authSvc := services.NewAuthService(config.AuthConfig{}, repo)
	userSvc := services.NewUserService(repo, authSvc, nil)

	req := services.RegisterRequest{
		BusinessName: "Acme Corp",
		Email:        "admin@acme.com",
		Password:     "password123",
		FirstName:    "John",
		LastName:     "Doe",
	}

	_ = userSvc.RegisterUser(context.Background(), req)

	err := userSvc.RegisterUser(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
}

func TestListUsers_Scoped(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)
	userSvc := services.NewUserService(repo, nil, nil)

	// Setup: 2 Businesses
	biz1 := models.Business{Name: "Biz1", Slug: "biz1"}
	biz2 := models.Business{Name: "Biz2", Slug: "biz2"}
	db.Create(&biz1)
	db.Create(&biz2)

	// Setup: 2 Users
	user1 := models.User{Email: "u1@biz1.com", Password: "x", FirstName: "A", LastName: "B", BusinessID: &biz1.ID}
	user2 := models.User{Email: "u2@biz2.com", Password: "x", FirstName: "C", LastName: "D", BusinessID: &biz2.ID}
	db.Create(&user1)
	db.Create(&user2)

	// 1. List for Biz1 -> Should get user1 only
	users, err := userSvc.ListUsers(context.Background(), &biz1.ID)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, user1.ID, users[0].ID)

	// 2. List Global (nil) -> Should get both
	usersGlobal, err := userSvc.ListUsers(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, usersGlobal, 2)
}

func TestDeleteUser_Scoped(t *testing.T) {
	db := setupTestDB()
	repo := repositories.NewUserRepository(db)
	userSvc := services.NewUserService(repo, nil, nil)

	biz1 := models.Business{Name: "Biz1", Slug: "biz1"}
	biz2 := models.Business{Name: "Biz2", Slug: "biz2"}
	db.Create(&biz1)
	db.Create(&biz2)

	user1 := models.User{Email: "u1@biz1.com", Password: "x", FirstName: "A", LastName: "B", BusinessID: &biz1.ID}
	db.Create(&user1)

	// 1. Valid Delete (Same Business)
	err := userSvc.DeleteUser(context.Background(), user1.ID, &biz1.ID)
	assert.NoError(t, err)

	// Verify Deleted
	var count int64
	db.Model(&models.User{}).Where("id = ?", user1.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	// Recreation for next test
	user2 := models.User{Email: "u2@biz1.com", Password: "x", FirstName: "C", LastName: "D", BusinessID: &biz1.ID}
	db.Create(&user2)

	// 2. Invalid Delete (Different Business/Tenant Isolation)
	err = userSvc.DeleteUser(context.Background(), user2.ID, &biz2.ID)
	assert.Error(t, err) // Should fail
	assert.Contains(t, err.Error(), "user not found or access denied")
}
