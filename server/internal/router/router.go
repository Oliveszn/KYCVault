package router

import (
	"kycvault/internal/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	AuthHandler         *handlers.AuthHandler
	KycHandler          *handlers.KYCHandler
	DocHandler          *handlers.DocumentHandler
	FaceHandler         *handlers.FaceHandler
	WSHandler           *handlers.WSHandler
	NotificationHandler *handlers.NotificationHandler
	AuthMiddleware      gin.HandlerFunc
	CORSOrigins         string
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	r := gin.Default()

	//CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{deps.CORSOrigins},
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

	KycRoutes(api, deps.KycHandler, deps.AuthMiddleware)

	docRoutes(api, deps.DocHandler, deps.AuthMiddleware)

	FaceRoutes(api, *deps.FaceHandler, deps.AuthMiddleware)

	NotificationRoutes(api, deps.NotificationHandler, deps.AuthMiddleware)

	//websockets
	r.GET("/ws", deps.AuthMiddleware, deps.WSHandler.Connect)
	return r
}
