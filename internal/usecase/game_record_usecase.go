package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/MovingPointP/darts-tracker-api/internal/domain/entity"
	"github.com/MovingPointP/darts-tracker-api/internal/domain/repository"
	"github.com/MovingPointP/darts-tracker-api/internal/usecase/rating"
)

type GameRecordUsecase interface {
	Create(userID string, gameType entity.GameType, value float64, playedAt time.Time, awards map[string]int) (*entity.GameRecord, error)
	Get(id uint, userID string) (*entity.GameRecord, error)
	GetWithFilter(userID string, filter repository.RecordsFilter) (*repository.PagedRecords, error)
	GetDailyRatings(userID string, gameType entity.GameType) ([]*repository.DailyRating, error)
	Update(id uint, userID string, value float64, playedAt time.Time, awards map[string]int) (*entity.GameRecord, error)
	Delete(id uint, userID string) error
}

type gameRecordUsecase struct {
	gameRecordRepo repository.GameRecordRepository
}

func NewGameRecordUsecase(gameRecordRepo repository.GameRecordRepository) GameRecordUsecase {
	return &gameRecordUsecase{gameRecordRepo: gameRecordRepo}
}

func (u *gameRecordUsecase) Create(userID string, gameType entity.GameType, value float64, playedAt time.Time, awards map[string]int) (*entity.GameRecord, error) {
	if !gameType.Valid() {
		return nil, entity.ErrInvalidGameType
	}
	if value > maxValueForGameType(gameType) {
		return nil, entity.ErrValueOutOfRange
	}

	record := &entity.GameRecord{
		UserID:   userID,
		GameType: gameType,
		Value:    value,
		Rating:   calculateRating(gameType, value),
		Awards:   awards,
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

func (u *gameRecordUsecase) GetWithFilter(userID string, filter repository.RecordsFilter) (*repository.PagedRecords, error) {
	result, err := u.gameRecordRepo.FindWithFilter(userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get game records: %w", err)
	}
	return result, nil
}

func (u *gameRecordUsecase) GetDailyRatings(userID string, gameType entity.GameType) ([]*repository.DailyRating, error) {
	if !gameType.Valid() || gameType == entity.GameTypeCountUp {
		return nil, entity.ErrInvalidGameType
	}
	ratings, err := u.gameRecordRepo.AggregateRatingByDay(userID, gameType)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate ratings: %w", err)
	}
	return ratings, nil
}

func (u *gameRecordUsecase) Update(id uint, userID string, value float64, playedAt time.Time, awards map[string]int) (*entity.GameRecord, error) {
	record, err := u.gameRecordRepo.FindByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game record: %w", err)
	}
	if record == nil {
		return nil, entity.ErrGameRecordNotFound
	}
	if value > maxValueForGameType(record.GameType) {
		return nil, entity.ErrValueOutOfRange
	}

	record.Value = value
	record.Rating = calculateRating(record.GameType, value)
	record.Awards = awards
	record.PlayedAt = playedAt

	if err := u.gameRecordRepo.Update(record); err != nil {
		return nil, fmt.Errorf("failed to update game record: %w", err)
	}
	return record, nil
}

func (u *gameRecordUsecase) Delete(id uint, userID string) error {
	if err := u.gameRecordRepo.Delete(id, userID); err != nil {
		if errors.Is(err, entity.ErrGameRecordNotFound) {
			return entity.ErrGameRecordNotFound
		}
		return fmt.Errorf("failed to delete game record: %w", err)
	}
	return nil
}

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

// maxValueForGameType はゲーム種別ごとの理論上の最大値。
// 01Game: 1ラウンド(3投)の最大点(トリプル20×3) = 180
// クリケット: 1ラウンド(3投)の最大マーク数(トリプル=3マーク×3投) = 9
// COUNTUP: DARTSLIVE標準の8ラウンド(24投)でのトリプル20連続 = 1440
func maxValueForGameType(gameType entity.GameType) float64 {
	switch gameType {
	case entity.GameType01Game:
		return 180
	case entity.GameTypeCricket:
		return 9
	case entity.GameTypeCountUp:
		return 1440
	default:
		return 0
	}
}
