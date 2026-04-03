package router

import (
	"kycvault/internal/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	AuthHandler    *handlers.AuthHandler
	AuthMiddleware gin.HandlerFunc
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	r := gin.Default()

	//CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api/v1")

	// Auth routes
	authGroup := api.Group("/auth")
	RegisterRoutes(authGroup, deps.AuthHandler, deps.AuthMiddleware)

	return r
}
