package database

import (
	"errors"
	"fmt"
	"kycvault/internal/config"
	"kycvault/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var AllModels = []interface{}{
	&models.User{},
}

func InitDatabase(cfg *config.Config) error {
	var err error

	// Connect using Neon DB URI
	DB, err = gorm.Open(postgres.Open(cfg.DB_URI), &gorm.Config{})
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
