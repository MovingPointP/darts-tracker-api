package repository

import "github.com/MovingPointP/darts-tracker-api/internal/domain/entity"

type GameRecordRepository interface {
	Create(record *entity.GameRecord) error
	FindByID(id uint, userID string) (*entity.GameRecord, error)
	FindAllByUserID(userID string, gameType *entity.GameType) ([]*entity.GameRecord, error)
	Update(record *entity.GameRecord) error
	Delete(id uint, userID string) error
}
