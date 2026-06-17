// Package ganzhi provides the core stem-branch (干支) type system shared
// by all Chinese metaphysics packages.
//
// Public API:
//
// # Primitive types
//
//	Gan (天干)       — GanJia .. GanGui
//	Zhi (地支)       — ZhiZi .. ZhiHai
//	Wuxing (五行)    — WuxingMu, WuxingHuo, WuxingTu, WuxingJin, WuxingShui
//	YinYang (阴阳)   — Yin, Yang
//	Gender           — Male, Female
//	Zhu              — 干支柱 (Gan + Zhi)
//	Bazi             — 四柱 (Nian, Yue, Ri, Shi)
//
// # Lookup tables
//
//	GanNames, ZhiNames               — 干支中文名
//	GanHes, ZhiHes                   — 天干五合 / 地支六合
//	TripleHeList, TripleHuiList      — 三合 / 三会
//	ChongPairs, HaiPairs             — 六冲 / 六害
//	XingGroups                       — 相刑
//	HiddenStemsTable                 — 藏干表
//	LifeStagesTable, StageNamesZH    — 十二长生
//	NayinTable                       — 纳音
//	HourRanges                       — 时辰范围
//
// # Conversion
//
//	GanWuxing, GanYinYang, GanName
//	ZhiWuxing, ZhiName, ZodiacLabel, ZhiHourRangeLabel, ZhiSeasonLabel
//	WuxingFromChinese, WuxingFromString
//
// # Stem-branch queries
//
//	IsGanHe, IsZhiHe, IsTripleHe, IsTripleHui
//	IsLiuChong, IsXing, IsHai, IsAnHe, IsPo
//	SixtyCycleName
//
// # Five-element computation
//
//	Sheng, Ke
//
// # Ten gods (十神)
//
//	TenGodType, TenGodName, TenGodFromGan
//
// # Hidden stems / RenYuan / NaYin
//
//	HiddenStemsForBranch, HiddenStems
//	RenYuanPhasesForBranch, RenYuanPhase
//	NaYinLabel
package ganzhi
