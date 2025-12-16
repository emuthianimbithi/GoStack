package services

import (
	"github.com/emuthianimbithi/GoStack/internal/models"
	"github.com/emuthianimbithi/GoStack/internal/server/repositories"
)

type AuditService struct {
	repo *repositories.AuditRepository
}

func NewAuditService(repo *repositories.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(entry *models.AuditLog) error {
	// Here we could also stream to a dedicated log service or queue
	return s.repo.Create(entry)
}
