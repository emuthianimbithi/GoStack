package repositories

import (
	"context"

	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID, businessID *uuid.UUID) error {
	query := r.DB.WithContext(ctx).Where("id = ?", id)
	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	result := query.Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, businessID *uuid.UUID) ([]models.User, error) {
	var users []models.User
	query := r.DB.WithContext(ctx).Preload("AccessRole")

	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	err := query.Find(&users).Error
	return users, err
}

func (r *UserRepository) CreateInvite(ctx context.Context, invite *models.UserInvite) error {
	return r.DB.WithContext(ctx).Create(invite).Error
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("AccessRole").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("AccessRole").First(&user, id).Error
	return &user, err
}
