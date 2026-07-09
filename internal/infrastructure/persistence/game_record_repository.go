package persistence

import (
	"errors"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
	"github.com/MovingPointP/darts-tracker-api/internal/domain/repository"
	"gorm.io/gorm"
)

type gormGameRecordRepository struct {
	db *gorm.DB
}

func NewGameRecordRepository(db *gorm.DB) repository.GameRecordRepository {
	return &gormGameRecordRepository{db: db}
}

func (r *gormGameRecordRepository) Create(record *entity.GameRecord) error {
	return r.db.Create(record).Error
}

func (r *gormGameRecordRepository) FindByID(id uint, userID string) (*entity.GameRecord, error) {
	var record entity.GameRecord
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

func (r *gormGameRecordRepository) FindWithFilter(userID string, filter repository.RecordsFilter) (*repository.PagedRecords, error) {
	query := r.db.Model(&entity.GameRecord{}).Where("user_id = ?", userID)
	if filter.GameType != nil {
		query = query.Where("game_type = ?", *filter.GameType)
	}
	if filter.From != nil {
		query = query.Where("played_at >= ?", *filter.From)
	}
	if filter.To != nil {
		// 終了日は当日23:59:59まで含める
		query = query.Where("played_at < ?", filter.To.AddDate(0, 0, 1))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	records := []*entity.GameRecord{}
	if err := query.Order("played_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &repository.PagedRecords{
		Records: records,
		Total:   total,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}, nil
}

func (r *gormGameRecordRepository) AggregateRatingByDay(userID string, gameType entity.GameType) ([]*repository.DailyRating, error) {
	results := []*repository.DailyRating{}
	err := r.db.Model(&entity.GameRecord{}).
		Select("TO_CHAR(played_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date, ROUND(AVG(rating)::numeric, 2) AS rating").
		Where("user_id = ? AND game_type = ? AND rating IS NOT NULL", userID, gameType).
		Group("TO_CHAR(played_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (r *gormGameRecordRepository) GetSummary(userID string, gameType entity.GameType) (*repository.GameSummary, error) {
	type aggregateRow struct {
		TotalGames int64    `gorm:"column:total_games"`
		BestValue  *float64 `gorm:"column:best_value"`
		BestRating *float64 `gorm:"column:best_rating"`
	}
	var agg aggregateRow
	if err := r.db.Model(&entity.GameRecord{}).
		Where("user_id = ? AND game_type = ?", userID, gameType).
		Select("COUNT(*) AS total_games, MAX(value) AS best_value, MAX(rating) AS best_rating").
		Row().Scan(&agg.TotalGames, &agg.BestValue, &agg.BestRating); err != nil {
		return nil, err
	}

	type awardRow struct {
		Key   string `gorm:"column:key"`
		Total int    `gorm:"column:total"`
	}
	var rows []awardRow
	if err := r.db.Raw(`
		SELECT a.key, SUM(a.value::int) AS total
		FROM game_records g
		CROSS JOIN LATERAL jsonb_each_text(g.awards) AS a(key, value)
		WHERE g.user_id = ? AND g.game_type = ?
		GROUP BY a.key
		ORDER BY total DESC
	`, userID, gameType).Scan(&rows).Error; err != nil {
		return nil, err
	}

	awards := map[string]int{}
	for _, row := range rows {
		awards[row.Key] = row.Total
	}

	return &repository.GameSummary{
		TotalGames: agg.TotalGames,
		BestValue:  agg.BestValue,
		BestRating: agg.BestRating,
		Awards:     awards,
	}, nil
}

func (r *gormGameRecordRepository) Update(record *entity.GameRecord) error {
	return r.db.Save(record).Error
}

func (r *gormGameRecordRepository) Delete(id uint, userID string) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.GameRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrGameRecordNotFound
	}
	return nil
}
