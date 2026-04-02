package router

import (
	"kycvault/internal/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler, authMiddleware gin.HandlerFunc) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)

	// All routes below require a valid access token.
	protected := rg.Group("")
	protected.Use(authMiddleware)
	{
		protected.POST("/logout", h.Logout)
		protected.POST("/logout-all", h.LogoutAll)
		protected.GET("/me", h.Me)
	}
}
