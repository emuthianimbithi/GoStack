package main

import (
	"context"
	"log"
	"os"

	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/db"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/permissions"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
	"github.com/emuthianimbithi/GoStack/internal/server/services"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	gormDB := db.Connect(cfg.DB)

	// Redis optional for seeding if not used in user creation, but perm service might need it
	opt, _ := redis.ParseURL("redis://localhost:6379/0") // simplified
	rdb := redis.NewClient(opt)

	// 1. Seed System Roles (RBAC)
	permSvc := permissions.NewService(gormDB, rdb)
	if err := permSvc.EnsureSystemRoles(); err != nil {
		log.Fatalf("failed to seed roles: %v", err)
	}

	// 2. Seed Initial Admin
	email := os.Getenv("SEED_ADMIN_EMAIL")
	pass := os.Getenv("SEED_ADMIN_PASSWORD")
	if email != "" && pass != "" {
		seedAdmin(gormDB, email, pass)
	} else {
		log.Println("Skipping admin seed: SEED_ADMIN_EMAIL or SEED_ADMIN_PASSWORD not set")
	}
}

func seedAdmin(db *gorm.DB, email, password string) {
	var count int64
	db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		log.Println("Admin already exists")
		return
	}

	// Reuse UserService logic or duplicate for simplicity
	// Here we duplicate to avoid complex dependency injection in script
	repo := repositories.NewUserRepository(db)
	// We need auth service just for struct? No, hash password manually or use service helper if public

	// Ideally use service:
	// authSvc := services.NewAuthService(config.AuthConfig{}, repo)
	// userSvc := services.NewUserService(repo, authSvc)
	// But RegisterUser does simple hash.

	log.Printf("Seeding admin %s...", email)

	// Create struct manually to avoid full service dependency chain if desired,
	// but using service ensures consistency.

	dummyAuth := services.NewAuthService(config.AuthConfig{}, repo)
	userSvc := services.NewUserService(repo, dummyAuth, nil)

	err := userSvc.RegisterUser(context.Background(), services.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: "Admin",
		LastName:  "User",
	})

	if err != nil {
		log.Printf("Failed to seed admin: %v", err)
	} else {
		log.Println("Admin seeded successfully")
	}
}
