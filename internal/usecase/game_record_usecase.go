package usecase

import (
	"fmt"
	"time"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
	"github.com/MovingPointP/darts-tracker-api/internal/domain/repository"
	"github.com/MovingPointP/darts-tracker-api/internal/usecase/rating"
)

type GameRecordUsecase interface {
	Create(userID string, gameType entity.GameType, value float64, playedAt time.Time) (*entity.GameRecord, error)
	Get(id uint, userID string) (*entity.GameRecord, error)
	GetAll(userID string, gameType *entity.GameType) ([]*entity.GameRecord, error)
	Update(id uint, userID string, value float64, playedAt time.Time) (*entity.GameRecord, error)
	Delete(id uint, userID string) error
}

type gameRecordUsecase struct {
	gameRecordRepo repository.GameRecordRepository
}

// コンストラクタ
func NewGameRecordUsecase(gameRecordRepo repository.GameRecordRepository) GameRecordUsecase {
	return &gameRecordUsecase{gameRecordRepo: gameRecordRepo}
}

func (u *gameRecordUsecase) Create(userID string, gameType entity.GameType, value float64, playedAt time.Time) (*entity.GameRecord, error) {
	if !gameType.Valid() {
		return nil, entity.ErrInvalidGameType
	}

	record := &entity.GameRecord{
		UserID:   userID,
		GameType: gameType,
		Value:    value,
		Rating:   calculateRating(gameType, value),
		PlayedAt: playedAt,
	}
	if err := u.gameRecordRepo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to create game record: %w", err)
	}
	return record, nil
}

func (u *gameRecordUsecase) Get(id uint, userID string) (*entity.GameRecord, error) {
	record, err := u.gameRecordRepo.FindByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game record: %w", err)
	}
	if record == nil {
		return nil, entity.ErrGameRecordNotFound
	}
	return record, nil
}

func (u *gameRecordUsecase) GetAll(userID string, gameType *entity.GameType) ([]*entity.GameRecord, error) {
	records, err := u.gameRecordRepo.FindAllByUserID(userID, gameType)
	if err != nil {
		return nil, fmt.Errorf("failed to get game records: %w", err)
	}
	return records, nil
}

func (u *gameRecordUsecase) Update(id uint, userID string, value float64, playedAt time.Time) (*entity.GameRecord, error) {
	// 記録の取得(所有権チェックを兼ねる)
	record, err := u.gameRecordRepo.FindByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game record: %w", err)
	}
	if record == nil {
		return nil, entity.ErrGameRecordNotFound
	}

	record.Value = value
	record.Rating = calculateRating(record.GameType, value)
	record.PlayedAt = playedAt

	if err := u.gameRecordRepo.Update(record); err != nil {
		return nil, fmt.Errorf("failed to update game record: %w", err)
	}
	return record, nil
}

func (u *gameRecordUsecase) Delete(id uint, userID string) error {
	// 記録の取得(所有権チェックを兼ねる)
	record, err := u.gameRecordRepo.FindByID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to get game record: %w", err)
	}
	if record == nil {
		return entity.ErrGameRecordNotFound
	}

	if err := u.gameRecordRepo.Delete(id, userID); err != nil {
		return fmt.Errorf("failed to delete game record: %w", err)
	}
	return nil
}

// calculateRating はゲーム種別に応じてレーティングを算出する。
// COUNTUPはレーティング算出対象外のためnilを返す。
func calculateRating(gameType entity.GameType, value float64) *float64 {
	var r float64
	switch gameType {
	case entity.GameType01Game:
		r = rating.CalculatePPRRating(value)
	case entity.GameTypeCricket:
		r = rating.CalculateMPRRating(value)
	default:
		return nil
	}
	return &r
}
