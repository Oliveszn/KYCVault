package router

import (
	"kycvault/internal/handlers"

	"github.com/gin-gonic/gin"
)

func FaceRoutes(r *gin.RouterGroup, h handlers.FaceHandler, auth gin.HandlerFunc) {
	face := r.Group("/kyc/sessions/:id/face")
	face.Use(auth)
	{
		// POST /kyc/sessions/:id/face — submit selfie, triggers Smile job
		face.POST("", h.StartVerification)

		// GET  /kyc/sessions/:id/face — poll for result (React wizard calls this)
		face.GET("", h.GetVerification)
	}

}
