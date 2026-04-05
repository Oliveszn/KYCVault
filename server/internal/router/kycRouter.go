package router

import (
	"kycvault/internal/handlers"
	"kycvault/internal/middleware"

	"github.com/gin-gonic/gin"
)

func KycRoutes(r *gin.RouterGroup, h *handlers.KYCHandler, authMiddleware gin.HandlerFunc) {
	//user roles requires auth
	user := r.Group("/kyc")
	user.Use(authMiddleware)
	{
		user.POST("/sessions", h.InitiateSession)
		user.GET("/sessions/active", h.GetActiveSession)
		user.GET("/sessions/history", h.GetSessionHistory)
		user.GET("/sessions/:id", h.GetSession)
	}

	//admin requires auth and admin role
	admin := r.Group("/admin/kyc")
	admin.Use(
		authMiddleware,
		middleware.RequireRole("admin"),
	)
	{
		admin.GET("/sessions", h.GetSessionQueue)
		admin.GET("/sessions/counts", h.GetStatusCounts)
		admin.GET("/sessions/:id", h.GetSessionAdmin)
		admin.POST("/sessions/:id/approve", h.ApproveSession)
		admin.POST("/sessions/:id/reject", h.RejectSession)
	}
}
