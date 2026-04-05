package repository

import (
	"context"
	"fmt"
	"kycvault/internal/models"

	"gorm.io/gorm"
)

type AuditRepository interface {
	CreateAuditEvent(ctx context.Context, event *models.AuditEvent) error
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) CreateAuditEvent(ctx context.Context, event *models.AuditEvent) error {
	result := r.db.WithContext(ctx).Create(event)
	if result.Error != nil {
		return fmt.Errorf("repository: create audit event: %w", result.Error)
	}
	return nil
}
