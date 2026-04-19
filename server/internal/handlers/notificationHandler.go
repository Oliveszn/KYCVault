package handlers

import (
	"kycvault/internal/middleware"
	"kycvault/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	svc    services.NotificationService
	logger *zap.Logger
}

func NewNotificationHandler(svc services.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		svc:    svc,
		logger: logger,
	}
}

func (h *NotificationHandler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	notifs, err := h.svc.GetUserNotifications(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch notifications")
		return
	}

	respond(c, http.StatusOK, "notifications retrieved", notifs)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	c.Status(http.StatusNoContent)
}
