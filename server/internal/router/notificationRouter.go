package router

import (
	"kycvault/internal/handlers"

	"github.com/gin-gonic/gin"
)

func NotificationRoutes(rg *gin.RouterGroup, h *handlers.NotificationHandler, auth gin.HandlerFunc) {
	notif := rg.Group("/notifications")
	notif.Use(auth)

	notif.GET("", h.Get)
	notif.PATCH("/:id/read", h.MarkRead)
}
