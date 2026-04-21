package router

import (
	"kycvault/internal/handlers"
	"kycvault/internal/middleware"

	"github.com/gin-gonic/gin"
)

func FaceRoutes(r *gin.RouterGroup, h handlers.FaceHandler, auth gin.HandlerFunc) {
	face := r.Group("/kyc/sessions/:id/face")
	face.Use(auth)
	{
		// POST /kyc/sessions/:id/face — submit selfie
		face.POST("", h.StartVerification)

		// GET  /kyc/sessions/:id/face poll for result (React wizard calls this)
		face.GET("", h.GetVerification)
	}

	admin := r.Group("/admin/face")
	admin.Use(
		auth,
		middleware.RequireRole("admin"),
	)
	{
		admin.GET("/:id/face", h.GetVerificationAdmin)
		admin.GET("/:id/selfie-url", h.GetSelfieURL)
		admin.POST("/:verificationId/review", h.ReviewVerification)
	}

}
