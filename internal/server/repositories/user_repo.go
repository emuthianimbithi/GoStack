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

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.DB.WithContext(ctx).Delete(&models.User{}, id).Error
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
