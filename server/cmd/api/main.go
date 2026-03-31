package main

import (
	"context"
	"fmt"
	"kycvault/internal/config"
	"kycvault/internal/database"
	"kycvault/internal/logger"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"os/signal"
)

func main() {
	cfg, err := config.LoadConfig()

	if err != nil {
		panic("failed to load config")
	}

	logger.InitLogger(cfg.ENV)
	// Ensure logger syncs before program exits
	defer zap.L().Sync()

	err = database.InitDatabase(&cfg)

	if err != nil {
		panic("failed to connect to database")
	}

	err = database.Migrate()

	if err != nil {
		fmt.Printf("Migration error: %v\n", err)
		panic("failed to migrate database")
	}

	router := mux.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: corsMiddleware.Handler(router),
	}

	defer database.CloseDB()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}
	}()

	fmt.Println("Server started on port " + cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Server shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited properly")
}
