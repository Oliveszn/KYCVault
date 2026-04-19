package services

import (
	"context"
	"kycvault/internal/models"
	"kycvault/internal/repository"
	"kycvault/internal/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationService interface {
	Create(ctx context.Context, n *models.Notification)
	GetUserNotifications(ctx context.Context, userID uuid.UUID) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
}

type notificationService struct {
	repo   repository.NotifRepository
	hub    *websocket.Hub
	logger *zap.Logger
}

func NewNotificationService(
	repo repository.NotifRepository,
	hub *websocket.Hub,
	logger *zap.Logger,

) NotificationService {
	return &notificationService{
		repo:   repo,
		hub:    hub,
		logger: logger,
	}
}

func (s *notificationService) Create(ctx context.Context, n *models.Notification) {
	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Error("failed to create notification", zap.Error(err))
		return
	}

	// push real-time update
	s.hub.SendToUser(n.UserID, map[string]any{
		"type": "notification",
		"data": n,
	})
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID) ([]models.Notification, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}
