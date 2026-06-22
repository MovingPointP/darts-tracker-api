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
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index;type:uuid"`
	GameType  GameType  `json:"game_type" gorm:"not null"`
	Value     float64   `json:"value" gorm:"not null"`
	Rating    *float64  `json:"rating"`
	PlayedAt  time.Time `json:"played_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var (
	ErrGameRecordNotFound = errors.New("game record not found")
	ErrInvalidGameType    = errors.New("invalid game type")
)
