package router

import (
	"kycvault/internal/handlers"
	"kycvault/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires document routes onto an existing router group.
// All routes sit under /kyc/sessions/:id/documents so they're naturally
// scoped to a session and the session ID is always in the path.
func docRoutes(r *gin.RouterGroup, h *handlers.DocumentHandler, authMiddleware gin.HandlerFunc) {
	//user roles requires auth
	session := r.Group("/kyc/sessions/:id/documents")
	session.Use(authMiddleware)
	{
		// POST   /kyc/sessions/:id/documents  — upload front or back
		session.POST("", h.UploadDocument)

		// GET    /kyc/sessions/:id/documents  — list documents for a session
		session.GET("", h.ListDocuments)
	}

	// Admin generate a presigned URL to view any document image
	admin := r.Group("/admin/documents")
	admin.Use(
		authMiddleware,
		middleware.RequireRole("admin"),
	)
	{
		admin.GET("/:doc_id/url", h.GetPresignedURL)
	}
}
