package repository

import (
	"context"

	"kycvault/internal/models"

	"gorm.io/gorm"
)

type NotifRepository interface {
	Create(ctx context.Context, notif *models.Notification) error
}

type notifRepository struct {
	db *gorm.DB
}

func NewNotifRepository(db *gorm.DB) NotifRepository {
	return &notifRepository{db: db}
}

func (r *notifRepository) Create(ctx context.Context, notif *models.Notification) error {
	return r.db.WithContext(ctx).Create(notif).Error
}
