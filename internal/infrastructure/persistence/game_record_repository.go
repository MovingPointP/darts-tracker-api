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

	var records []*entity.GameRecord
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
	var results []*repository.DailyRating
	err := r.db.Model(&entity.GameRecord{}).
		Select("TO_CHAR(played_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date, ROUND(AVG(rating)::numeric, 2) AS rating").
		Where("user_id = ? AND game_type = ? AND rating IS NOT NULL", userID, gameType).
		Group("TO_CHAR(played_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}

func (r *gormGameRecordRepository) Update(record *entity.GameRecord) error {
	return r.db.Save(record).Error
}

func (r *gormGameRecordRepository) Delete(id uint, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.GameRecord{}).Error
}
