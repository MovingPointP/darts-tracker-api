// Package rating はDARTSLIVE風の非公式換算表(dartsmap.com)を参考にした
// レーティング算出ロジックを提供する。01Game(PPR)とクリケット(MPR)それぞれ
// 独立に算出し、統合は行わない。公式の正確な計算式は非公開のため、あくまで
// 参考値として扱う。
package rating

import "math"

type band struct {
	Min, Max float64 // 半開区間 [Min, Max)
	Rating   float64 // このバンド下限でのレーティング
}

// pprBands は01Game(1ラウンド平均点 = PPR)からレーティングへの換算バンド。
var pprBands = []band{
	{0, 40, 1},
	{40, 45, 2},
	{45, 50, 3},
	{50, 55, 4},
	{55, 60, 5},
	{60, 65, 6},
	{65, 70, 7},
	{70, 75, 8},
	{75, 80, 9},
	{80, 85, 10},
	{85, 90, 11},
	{90, 95, 12},
	{95, 102, 13},
	{102, 109, 14},
	{109, 116, 15},
	{116, 123, 16},
	{123, 130, 17},
	{130, math.Inf(1), 18},
}

// mprBands はクリケット(1ラウンド平均マーク数 = MPR)からレーティングへの換算バンド。
var mprBands = []band{
	{0, 1.3, 1},
	{1.3, 1.5, 2},
	{1.5, 1.7, 3},
	{1.7, 1.9, 4},
	{1.9, 2.1, 5},
	{2.1, 2.3, 6},
	{2.3, 2.5, 7},
	{2.5, 2.7, 8},
	{2.7, 2.9, 9},
	{2.9, 3.1, 10},
	{3.1, 3.3, 11},
	{3.3, 3.5, 12},
	{3.5, 3.75, 13},
	{3.75, 4.0, 14},
	{4.0, 4.25, 15},
	{4.25, 4.5, 16},
	{4.5, 4.75, 17},
	{4.75, math.Inf(1), 18},
}

// CalculatePPRRating は01Gameの1ラウンド平均点(PPR)からレーティングを算出する。
func CalculatePPRRating(ppr float64) float64 {
	return interpolate(ppr, pprBands)
}

// CalculateMPRRating はクリケットの1ラウンド平均マーク数(MPR)からレーティングを算出する。
func CalculateMPRRating(mpr float64) float64 {
	return interpolate(mpr, mprBands)
}

// interpolate はバンド内で線形補間し、小数2桁に丸めたレーティングを返す。
// 値がバンド下限を下回る場合は最下位バンドの下限に、上限バンドを超える場合は
// 上限バンドのレーティング(18.00)にクランプする。
func interpolate(value float64, bands []band) float64 {
	if value < bands[0].Min {
		value = bands[0].Min
	}
	for _, b := range bands {
		if value >= b.Min && value < b.Max {
			r := b.Rating + (value-b.Min)/(b.Max-b.Min)
			return round2(r)
		}
	}
	last := bands[len(bands)-1]
	return round2(last.Rating)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
