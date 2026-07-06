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

	// アワード名のリネーム対応
	conn.Exec(`UPDATE game_records SET awards = (awards - 'THREE IN THE BLACK') || jsonb_build_object('3 IN THE BLACK', awards->'THREE IN THE BLACK') WHERE awards ? 'THREE IN THE BLACK'`)
	conn.Exec(`UPDATE game_records SET awards = (awards - 'THREE IN A BED') || jsonb_build_object('3 IN A BED', awards->'THREE IN A BED') WHERE awards ? 'THREE IN A BED'`)

	// awards カラムを JSON配列→JSON オブジェクトに変換（string[]→map[string]int への型変更対応）
	// ["ONE BULL", "TON 80"] → {"ONE BULL": 1, "TON 80": 1}、[] → {}
	conn.Exec(`
		UPDATE game_records
		SET awards = (
			SELECT COALESCE(jsonb_object_agg(award, cnt), '{}')
			FROM (
				SELECT award, count(*)::int AS cnt
				FROM jsonb_array_elements_text(awards) AS award
				GROUP BY award
			) sub
		)
		WHERE jsonb_typeof(awards) = 'array'
	`)

	return conn, nil
}
