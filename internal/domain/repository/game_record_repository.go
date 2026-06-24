package repository

import (
	"time"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
)

// RecordsFilter は記録一覧取得時のフィルタ・ページネーション条件。
type RecordsFilter struct {
	GameType *entity.GameType
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// DailyRating は日別の平均レーティング。
type DailyRating struct {
	Date   string  `json:"date"`
	Rating float64 `json:"rating"`
}

// PagedRecords はページネーション付き記録一覧。
type PagedRecords struct {
	Records []*entity.GameRecord `json:"records"`
	Total   int64                `json:"total"`
	Limit   int                  `json:"limit"`
	Offset  int                  `json:"offset"`
}

type GameRecordRepository interface {
	Create(record *entity.GameRecord) error
	FindByID(id uint, userID string) (*entity.GameRecord, error)
	FindWithFilter(userID string, filter RecordsFilter) (*PagedRecords, error)
	AggregateRatingByDay(userID string, gameType entity.GameType) ([]*DailyRating, error)
	Update(record *entity.GameRecord) error
	Delete(id uint, userID string) error
}
