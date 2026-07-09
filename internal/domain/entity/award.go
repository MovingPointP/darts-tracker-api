package entity

import (
	"errors"
	"fmt"
)

var validAwards = map[string]struct{}{
	"ONE BULL":       {},
	"LOW TON":        {},
	"HIGH TON":       {},
	"TON 80":         {},
	"HAT TRICK":      {},
	"3 IN THE BLACK": {},
	"3 IN A BED":     {},
	"WHITE HORSE":    {},
	"5 MARK":         {},
	"9 MARK":         {},
}

var (
	ErrInvalidAwardName  = errors.New("invalid award name")
	ErrInvalidAwardCount = errors.New("award count must be positive")
)

// ValidateAwards はアワード名がホワイトリストに含まれ、かつ回数が1以上であることを検証する。
func ValidateAwards(awards map[string]int) error {
	for name, count := range awards {
		if _, ok := validAwards[name]; !ok {
			return fmt.Errorf("%w: %s", ErrInvalidAwardName, name)
		}
		if count <= 0 {
			return fmt.Errorf("%w: %s", ErrInvalidAwardCount, name)
		}
	}
	return nil
}
