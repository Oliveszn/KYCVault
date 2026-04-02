package router

import (
	"kycvault/internal/handlers"

	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	AuthHandler    *handlers.AuthHandler
	AuthMiddleware gin.HandlerFunc
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")

	// Auth routes
	authGroup := api.Group("/auth")
	RegisterRoutes(authGroup, deps.AuthHandler, deps.AuthMiddleware)

	return r
}
