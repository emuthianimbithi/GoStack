package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/emuthianimbithi/GoStack/internal/constants"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo        *repositories.UserRepository
	authService *AuthService
}

func NewUserService(repo *repositories.UserRepository, authService *AuthService) *UserService {
	return &UserService{
		repo:        repo,
		authService: authService,
	}
}

type RegisterRequest struct {
	BusinessName string `json:"business_name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
}

func (s *UserService) RegisterUser(ctx context.Context, req RegisterRequest) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Check if user exists
		var count int64
		if err := tx.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("user already exists")
		}

		// 2. Create Business
		business := models.Business{
			Name: req.BusinessName,
			Slug: req.BusinessName, // Simple slug for now, better slugify needed
		}
		if err := tx.Create(&business).Error; err != nil {
			return err
		}

		// 3. Hash Password
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// 4. Create Admin Role for this Business
		adminRole := models.AccessRole{
			Name:        constants.RoleAdmin.String(),
			Description: "Business Administrator",
			BusinessID:  &business.ID,
		}
		if err := tx.Create(&adminRole).Error; err != nil {
			return err
		}

		// 5. Create User (Assigned to Admin Role)
		user := &models.User{
			Email:      req.Email,
			Password:   string(hashedBytes),
			FirstName:  req.FirstName,
			LastName:   req.LastName,
			RoleID:     &adminRole.ID,
			BusinessID: &business.ID,
		}

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Find User
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// 3. Generate Tokens
	return s.authService.GenerateTokenPair(user)
}

func (s *UserService) ListUsers(ctx context.Context, businessID *uuid.UUID) ([]models.User, error) {
	var users []models.User
	query := s.repo.DB.WithContext(ctx).Preload("AccessRole")

	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	err := query.Find(&users).Error
	return users, err
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID, businessID *uuid.UUID) error {
	// Ensure user belongs to the business (if scoped)
	query := s.repo.DB.WithContext(ctx).Where("id = ?", id)
	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	result := query.Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found or access denied")
	}
	return nil
}
