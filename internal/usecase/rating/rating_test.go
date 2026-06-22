package rating

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePPRRating(t *testing.T) {
	tests := []struct {
		name string
		ppr  float64
		want float64
	}{
		{"最下限", 0, 1.00},
		{"バンド1の中間", 20, 1.50},
		{"バンド境界(40)", 40, 2.00},
		{"バンド6の中間(62.5)", 62.5, 6.50},
		{"バンド13の中間(98.5)", 98.5, 13.50},
		{"最上位バンド下限(130)", 130, 18.00},
		{"最上位バンドを超える値(200)", 200, 18.00},
		{"負の値はクランプ", -10, 1.00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, CalculatePPRRating(tt.ppr), 0.001)
		})
	}
}

func TestCalculateMPRRating(t *testing.T) {
	tests := []struct {
		name string
		mpr  float64
		want float64
	}{
		{"最下限", 0, 1.00},
		{"バンド1の中間(0.65)", 0.65, 1.50},
		{"バンド5の中間(2.0)", 2.0, 5.50},
		{"最上位バンド下限(4.75)", 4.75, 18.00},
		{"最上位バンドを超える値(10)", 10, 18.00},
		{"負の値はクランプ", -1, 1.00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, CalculateMPRRating(tt.mpr), 0.001)
		})
	}
}
