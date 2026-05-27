package fengshui

import (
	"github.com/25types/25types/internal/ganzhi"
	"github.com/25types/25types/internal/mingli/bazi"
)

// 合参 (Hecan / Combined Reference) — parallel presentation of multiple Feng Shui
// systems alongside BaZi yong/xi shen. No scoring or synthesis — each system
// speaks for itself.

// HeCanResult bundles fate trigram, BaZhai directions, annual flying stars,
// pillar bagua, and BaZi yong/xi shen (pass-through).
type HeCanResult struct {
	MingGua     MingGuaResult      `json:"ming_gua"`
	BaZhaiDirs  BaZhaiDirections   `json:"ba_zhai_dirs"`
	YearStars   YearStarResult     `json:"year_stars"`
	YongShen    bazi.YongShenResult `json:"yong_shen"`
	PillarBagua [4]Trigram         `json:"pillar_bagua"`
}

// ComputeHeCan assembles a combined Feng Shui reference. It computes the
// fate trigram internally from birth year and gender, derives pillar bagua
// from the four pillars, and passes through the yong-shen analysis.
func ComputeHeCan(
	birthYear int,
	gender ganzhi.Gender,
	bz ganzhi.Bazi,
	yongShen bazi.YongShenResult,
	year int,
) HeCanResult {
	mingGua := ComputeMingGua(gender, birthYear)

	return HeCanResult{
		MingGua:    mingGua,
		BaZhaiDirs: BaZhaiDirectionsForGua(mingGua.GuaNumber),
		YearStars:  ComputeYearStars(year),
		YongShen:   yongShen,
		PillarBagua: [4]Trigram{
			PillarNaJia(bz.Year),
			PillarNaJia(bz.Month),
			PillarNaJia(bz.Day),
			PillarNaJia(bz.Hour),
		},
	}
}
