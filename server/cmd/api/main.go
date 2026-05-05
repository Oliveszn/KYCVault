package main

import (
	"context"
	"fmt"
	"kycvault/internal/config"
	"kycvault/internal/database"
	"kycvault/internal/handlers"
	"kycvault/internal/infra/facepp"
	"kycvault/internal/infra/storage"
	"kycvault/internal/logger"
	"kycvault/internal/middleware"
	"kycvault/internal/repository"
	"kycvault/internal/router"
	"kycvault/internal/services"
	"kycvault/internal/utils"
	"kycvault/internal/websocket"
	"kycvault/internal/worker"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"os/signal"
)

func main() {
	//CONFIG
	cfg, err := config.LoadConfig()

	if err != nil {
		panic("failed to load config")
	}

	//LOGGER
	logger.InitLogger(cfg.ENV)
	// Ensure logger syncs before program exits
	defer zap.L().Sync()

	//CONNECT DB
	err = database.InitDatabase(&cfg)

	if err != nil {
		panic("failed to connect to database")
	}

	//MIGRATE DB
	err = database.Migrate()

	if err != nil {
		fmt.Printf("Migration error: %v\n", err)
		panic("failed to migrate database")
	}

	// CREATE INDEXES
	err = database.CreateIndexes()
	if err != nil {
		panic(fmt.Errorf("failed to create indexes: %w", err))
	}

	defer database.CloseDB()

	jwtUtil, err := utils.NewJWTUtil(utils.JWTConfig{
		AccessSecret:    cfg.JWT_ACCESS_SECRET,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          cfg.JWT_ISSUER,
	})
	if err != nil {
		zap.L().Error("invalid jwt config", zap.String("error", err.Error()))
		os.Exit(1)
	}

	isProd := cfg.ENV == "production"
	cookieCfg := utils.CookieConfig{
		Domain:   cfg.COOKIE_DOMAIN,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode, //change to strict mode in prod
	}

	wsHub := websocket.NewHub()
	wsHandler := handlers.NewWSHandler(wsHub, zap.L(), cfg.CORSAllowedOrigins)

	notifRepo := repository.NewNotifRepository(database.GetDB())

	notifSvc := services.NewNotificationService(notifRepo, wsHub, zap.L())

	notifHandler := handlers.NewNotificationHandler(notifSvc, zap.L())

	authRepo := repository.NewAuthRepository(database.GetDB())
	authSvc := services.NewAuthService(authRepo, jwtUtil, zap.L())
	authHandler := handlers.NewAuthHandler(authSvc, jwtUtil, cookieCfg, zap.L())
	authMiddleware := middleware.Authenticate(jwtUtil, zap.L())

	auditRepo := repository.NewAuditRepository(database.GetDB())
	auditSvc := services.NewAuditService(auditRepo, zap.L())
	kycRepo := repository.NewKYCRepository(database.GetDB())
	kycSvc := services.NewKYCService(kycRepo, auditSvc, zap.L(), notifSvc)
	kycHandler := handlers.NewKYCHandler(kycSvc, zap.L())

	awsCfg := aws.Config{
		Region:      cfg.AWSRegion,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AWS_ACCESS_KEY, cfg.AWS_SECRET_ACCESS_KEY, ""),
	}

	awsClient := s3.NewFromConfig(awsCfg)
	storageClient := storage.NewS3Client(awsClient, cfg.S3Bucket)

	docRepo := repository.NewDocumentRepository(database.GetDB())
	docSvc := services.NewDocumentService(
		docRepo, kycRepo, kycSvc,
		storageClient, cfg.S3Bucket,
		auditSvc, zap.L(),
	)
	docHandler := handlers.NewDocumentHandler(docSvc, zap.L())

	faceppClient := facepp.NewClient(
		cfg.FACE_API_KEY,
		cfg.FACE_API_SECRET,
	)
	faceRepo := repository.NewFaceRepository(database.GetDB())
	faceSvc := services.NewFaceService(
		faceRepo,
		docRepo,
		kycSvc,
		faceppClient,
		storageClient,
		cfg.S3Bucket,
		auditSvc,
		zap.L(),
	)
	faceHandler := handlers.NewFaceHandler(faceSvc, zap.L())

	// Router
	r := router.NewRouter(router.RouterDependencies{
		AuthHandler:         authHandler,
		KycHandler:          kycHandler,
		DocHandler:          docHandler,
		FaceHandler:         faceHandler,
		WSHandler:           wsHandler,
		NotificationHandler: notifHandler,
		AuthMiddleware:      authMiddleware,
		CORSOrigins:         cfg.CORSAllowedOrigins,
	})

	// server := &http.Server{
	// 	Addr:    ":" + cfg.ServerPort,
	// 	Handler: r,
	// }
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.ServerPort // fallback for local dev
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())

	tokenWorker := worker.NewTokenCleanupWorker(
		authRepo,
		zap.L(),
		15*time.Minute,
	)

	tokenWorker.Start(ctx)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}
	}()

	fmt.Println("Server started on port " + port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Server shutting down...")

	cancel()

	// graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited properly")
}
