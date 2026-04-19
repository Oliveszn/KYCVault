package repository

import (
	"context"
	"errors"

	"kycvault/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrNotificationNotFound = errors.New("We couldn't find the notification, try again later")
)

type NotifRepository interface {
	Create(ctx context.Context, notif *models.Notification) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, notifID uuid.UUID, userID uuid.UUID) error
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

func (r *notifRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	var notifs []models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifs).Error

	return notifs, err
}

func (r *notifRepository) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotificationNotFound
	}
	return nil
}
