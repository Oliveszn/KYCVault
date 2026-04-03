package database

import (
	"errors"
	"fmt"
	"kycvault/internal/config"
	"kycvault/internal/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

var AllModels = []interface{}{
	&models.User{},
	&models.RefreshToken{},
}

func InitDatabase(cfg *config.Config) error {
	var err error

	//gorm thowrs many logs in console so we add tis to silence it
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      false,
		},
	)

	// Connect using Neon DB URI
	// DB, err = gorm.Open(postgres.Open(cfg.DB_URI), &gorm.Config{})
	DB, err = gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  cfg.DB_URI,
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{
			Logger: newLogger,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// Connection pooling
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)

	fmt.Println("Database connected successfully")
	return nil
}

func Migrate() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	for _, model := range AllModels {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	fmt.Println("Database migrated successfully")
	return nil
}

func CloseDB() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func GetDB() *gorm.DB {
	return DB
}
