package db

import (
	"fmt"
	"os"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 接続
	conn, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// マイグレーション
	if err := conn.AutoMigrate(&entity.GameRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return conn, nil
}
