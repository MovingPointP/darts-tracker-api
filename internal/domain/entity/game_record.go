package entity

import (
	"errors"
	"time"
)

type GameType string

const (
	GameType01Game  GameType = "01game"
	GameTypeCricket GameType = "cricket"
	GameTypeCountUp GameType = "countup"
)

func (t GameType) Valid() bool {
	switch t {
	case GameType01Game, GameTypeCricket, GameTypeCountUp:
		return true
	default:
		return false
	}
}

// GameRecord は1ゲーム分の集計記録。投・ラウンド単位の詳細は持たない。
type GameRecord struct {
	ID       uint     `json:"id" gorm:"primaryKey"`
	UserID   string   `json:"user_id" gorm:"not null;type:uuid;index:idx_game_records_user_id;index:idx_game_records_user_game_date,composite:user_id,priority:1"`
	GameType GameType `json:"game_type" gorm:"not null;index:idx_game_records_user_game_date,composite:game_type,priority:2"`
	Value    float64  `json:"value" gorm:"not null"`
	Rating   *float64 `json:"rating"`
	Awards   map[string]int `json:"awards" gorm:"type:jsonb;serializer:json;default:'{}'"`
	// idx_game_records_user_game_dateは集計クエリ(AggregateRatingByDay)とフィルタクエリの高速化のために使用する
	PlayedAt  time.Time `json:"played_at" gorm:"not null;index:idx_game_records_user_game_date,composite:played_at,priority:3"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrGameRecordNotFound = errors.New("game record not found")
	ErrInvalidGameType    = errors.New("invalid game type")
	ErrValueOutOfRange    = errors.New("value out of range for this game type")
)
