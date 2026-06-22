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

// コンストラクタ
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

func (r *gormGameRecordRepository) FindAllByUserID(userID string, gameType *entity.GameType) ([]*entity.GameRecord, error) {
	var records []*entity.GameRecord

	query := r.db.Where("user_id = ?", userID)
	if gameType != nil {
		query = query.Where("game_type = ?", *gameType)
	}
	if err := query.Order("played_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *gormGameRecordRepository) Update(record *entity.GameRecord) error {
	return r.db.Save(record).Error
}

func (r *gormGameRecordRepository) Delete(id uint, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.GameRecord{}).Error
}
