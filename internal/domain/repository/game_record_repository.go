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

// Period は集計対象の期間フィルタ(任意)。FromもToもnilなら全期間。
// Toはその日の終わりまで含める運用(永続化層で+1日する)。
type Period struct {
	From *time.Time
	To   *time.Time
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

// GameSummary は種目ごとの集計結果。
type GameSummary struct {
	TotalGames int64          `json:"total_games"`
	BestValue  *float64       `json:"best_value"`
	AvgValue   *float64       `json:"avg_value"`
	BestRating *float64       `json:"best_rating"`
	AvgRating  *float64       `json:"avg_rating"`
	Awards     map[string]int `json:"awards"`
}

type GameRecordRepository interface {
	Create(record *entity.GameRecord) error
	FindByID(id uint, userID string) (*entity.GameRecord, error)
	FindWithFilter(userID string, filter RecordsFilter) (*PagedRecords, error)
	AggregateRatingByDay(userID string, gameType entity.GameType, period Period) ([]*DailyRating, error)
	GetSummary(userID string, gameType entity.GameType, period Period) (*GameSummary, error)
	Update(record *entity.GameRecord) error
	Delete(id uint, userID string) error
}
