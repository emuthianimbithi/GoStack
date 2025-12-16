package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/emuthianimbithi/GoStack/internal/config"
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
)

type AuthService struct {
	config   config.AuthConfig
	userRepo *repositories.UserRepository
}

func NewAuthService(cfg config.AuthConfig, userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{
		config:   cfg,
		userRepo: userRepo,
	}
}

type Claims struct {
	UserID     string  `json:"user_id"`
	Role       string  `json:"role"`
	BusinessID *string `json:"bid,omitempty"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateTokenPair(user *models.User) (string, string, error) {
	// Access Token
	accessToken, err := s.generateToken(user, s.config.AccessDuration)
	if err != nil {
		return "", "", err
	}

	// Refresh Token (Longer lived)
	// In a real app, you might save this to DB to allow revocation
	refreshToken, err := s.generateToken(user, s.config.RefreshDuration*24*60) // Days to minutes
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) generateToken(user *models.User, durationMinutes int) (string, error) {
	roleName := ""
	if user.AccessRole != nil {
		roleName = user.AccessRole.Name
	}

	claims := Claims{
		UserID: user.ID.String(),
		Role:   roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(durationMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if user.BusinessID != nil {
		businessIDStr := user.BusinessID.String()
		claims.BusinessID = &businessIDStr
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
