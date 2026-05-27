package bazi

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// -- Stem / Branch lookups --

func TestStemElement(t *testing.T) {
	cases := map[Stem]Element{
		StemJia: ElemWood, StemYi: ElemWood,
		StemBing: ElemFire, StemDing: ElemFire,
		StemWu: ElemEarth, StemJi: ElemEarth,
		StemGeng: ElemMetal, StemXin: ElemMetal,
		StemRen: ElemWater, StemGui: ElemWater,
	}
	for s, want := range cases {
		if got := StemElement(s); got != want {
			t.Errorf("StemElement(%d)=%d, want %d", s, got, want)
		}
	}
}

func TestStemYinYang(t *testing.T) {
	yangStems := []Stem{StemJia, StemBing, StemWu, StemGeng, StemRen}
	yinStems := []Stem{StemYi, StemDing, StemJi, StemXin, StemGui}
	for _, s := range yangStems {
		if got := StemYinYang(s); got != Yang {
			t.Errorf("StemYinYang(%d)=Yin, want Yang", s)
		}
	}
	for _, s := range yinStems {
		if got := StemYinYang(s); got != Yin {
			t.Errorf("StemYinYang(%d)=Yang, want Yin", s)
		}
	}
}

func TestBranchElement(t *testing.T) {
	// 寅卯=木, 巳午=火, 辰戌丑未=土, 申酉=金, 亥子=水
	cases := map[Branch]Element{
		BranchYin: ElemWood, BranchMao: ElemWood,
		BranchSi: ElemFire, BranchWu: ElemFire,
		BranchChen: ElemEarth, BranchXu: ElemEarth, BranchChou: ElemEarth, BranchWei: ElemEarth,
		BranchShen: ElemMetal, BranchYou: ElemMetal,
		BranchHai: ElemWater, BranchZi: ElemWater,
	}
	for b, want := range cases {
		if got := BranchElement(b); got != want {
			t.Errorf("BranchElement(%d)=%d, want %d", b, got, want)
		}
	}
}

// -- sixtyCycleName / sixtyCycleName --

func TestSixtyCycleName(t *testing.T) {
	// 甲子=0, 乙丑=1, ..., 癸亥=59
	if got := sixtyCycleName(StemJia, BranchZi); got != 0 {
		t.Errorf("甲子 index=%d, want 0", got)
	}
	if got := sixtyCycleName(StemGui, BranchHai); got != 59 {
		t.Errorf("癸亥 index=%d, want 59", got)
	}
	if got := sixtyCycleName(StemJia, BranchXu); got != 10 {
		t.Errorf("甲戌 index=%d, want 10 (day pillar baseline)", got)
	}
}

func TestSixtyCycleName_Consistency(t *testing.T) {
	// sixtyCycleName should go from 0 to 59 and wrap.
	for s := StemJia; s <= StemGui; s++ {
		for b := BranchZi; b <= BranchHai; b++ {
			idx := sixtyCycleName(s, b)
			if idx < 0 || idx > 59 {
				t.Errorf("sixtyCycleName(%d,%d)=%d out of range", s, b, idx)
			}
		}
	}
}

// -- JulianDay --

func TestJulianDay_Known(t *testing.T) {
	// J2000.0 = 2000-01-01 12:00 TT ≈ JD 2451545
	if got := JulianDay(2000, 1, 1); got != 2451545 {
		t.Errorf("JulianDay(2000,1,1)=%d, want 2451545", got)
	}
}

func TestJulianDay_1900Baseline(t *testing.T) {
	// 1900-01-01 = JD 2415021 (the anchor for day pillar computation).
	if got := JulianDay(1900, 1, 1); got != 2415021 {
		t.Errorf("JulianDay(1900,1,1)=%d, want 2415021", got)
	}
}

func TestJulianDay_CrossCentury(t *testing.T) {
	// 1900-03-01: month > 2, no year adjustment.
	jd1 := JulianDay(1900, 2, 28)
	jd2 := JulianDay(1900, 3, 1)
	if jd2-jd1 != 1 {
		t.Errorf("Feb 28→Mar 1 diff=%d, want 1 (1900 not leap)", jd2-jd1)
	}

	// 2000 is a leap year (divisible by 400).
	jd3 := JulianDay(2000, 2, 28)
	jd4 := JulianDay(2000, 3, 1)
	if jd4-jd3 != 2 {
		t.Errorf("Feb 28→Mar 1 diff=%d, want 2 (2000 leap)", jd4-jd3)
	}
}

// -- DayOfYear --

func TestDayOfYear(t *testing.T) {
	if got := DayOfYear(2024, 1, 1); got != 1 {
		t.Errorf("Jan 1 = %d, want 1", got)
	}
	if got := DayOfYear(2024, 12, 31); got != 366 {
		t.Errorf("Dec 31 2024 = %d, want 366 (leap)", got)
	}
	if got := DayOfYear(2023, 12, 31); got != 365 {
		t.Errorf("Dec 31 2023 = %d, want 365", got)
	}
}

// -- LiChunDay --

func TestLiChunDay_Range(t *testing.T) {
	for year := 1950; year <= 2100; year++ {
		m, d := LiChunDay(year)
		if m != 2 {
			t.Errorf("LiChunDay(%d) month=%d, want 2", year, m)
		}
		if d < 3 || d > 5 {
			t.Errorf("LiChunDay(%d) day=%d, want 3-5", year, d)
		}
	}
}

// -- YearPillar --

func TestYearPillar_Known(t *testing.T) {
	// 2024 = 甲辰 (Stem=1, Branch=5).
	yp := YearPillar(2024, 3, 1) // after 立春.
	if yp.Stem != 1 || yp.Branch != 5 {
		t.Errorf("YearPillar(2024) = %d-%d, want 甲辰=1-5", yp.Stem, yp.Branch)
	}
	// 2023 = 癸卯 (Stem=10, Branch=4).
	yp2 := YearPillar(2023, 3, 1)
	if yp2.Stem != 10 || yp2.Branch != 4 {
		t.Errorf("YearPillar(2023) = %d-%d, want 癸卯=10-4", yp2.Stem, yp2.Branch)
	}
}

func TestYearPillar_BeforeLiChun(t *testing.T) {
	// Jan 1 is always before 立春 — year should be (year-1).
	// 2024-01-01: year-1=2023 → 癸卯 (10-4).
	yp := YearPillar(2024, 1, 1)
	if yp.Stem != 10 || yp.Branch != 4 {
		t.Errorf("YearPillar(2024,1,1) = %d-%d, want 癸卯=10-4", yp.Stem, yp.Branch)
	}
}

// -- DayPillar --

func TestDayPillar_Baseline(t *testing.T) {
	// 1900-01-01 = 甲戌 (Stem=1, Branch=11).
	dp := DayPillar(1900, 1, 1)
	if dp.Stem != 1 || dp.Branch != 11 {
		t.Errorf("DayPillar(1900,1,1) = %d-%d, want 甲戌=1-11", dp.Stem, dp.Branch)
	}
}

func TestDayPillar_Known(t *testing.T) {
	// Cross-checked against known dates.
	// 1984-02-04 = 戊辰 (5-5), verified in smoke test.
	dp := DayPillar(1984, 2, 4)
	if dp.Stem != 5 || dp.Branch != 5 {
		t.Errorf("DayPillar(1984,2,4) = %d-%d, want 戊辰=5-5", dp.Stem, dp.Branch)
	}
	// 2024-02-04 = 戊戌 (5-11).
	dp2 := DayPillar(2024, 2, 4)
	if dp2.Stem != 5 || dp2.Branch != 11 {
		t.Errorf("DayPillar(2024,2,4) = %d-%d, want 戊戌=5-11", dp2.Stem, dp2.Branch)
	}
}

func TestDayPillar_Consecutive(t *testing.T) {
	dp1 := DayPillar(2024, 2, 4)
	dp2 := DayPillar(2024, 2, 5)
	idx1 := sixtyCycleName(dp1.Stem, dp1.Branch)
	idx2 := sixtyCycleName(dp2.Stem, dp2.Branch)
	if (idx2-idx1+60)%60 != 1 {
		t.Errorf("consecutive days not consecutive in 60-cycle: %d → %d", idx1, idx2)
	}
}

// -- HourBranchFromSolarTime --

func TestHourBranchFromSolarTime(t *testing.T) {
	// 子时: 23:00-01:00 = [1380,1440) ∪ [0,60).
	if got := HourBranchFromSolarTime(0); got != BranchZi {
		t.Errorf("HourBranch(0)=%d, want 子=1", got)
	}
	if got := HourBranchFromSolarTime(30); got != BranchZi {
		t.Errorf("HourBranch(30)=%d, want 子=1", got)
	}
	if got := HourBranchFromSolarTime(1380); got != BranchZi {
		t.Errorf("HourBranch(1380)=%d, want 子=1", got)
	}
	// 丑时: 01:00-03:00 = [60,180).
	if got := HourBranchFromSolarTime(60); got != BranchChou {
		t.Errorf("HourBranch(60)=%d, want 丑=2", got)
	}
	// 午时: 11:00-13:00 = [660,780).
	if got := HourBranchFromSolarTime(720); got != BranchWu {
		t.Errorf("HourBranch(720)=%d, want 午=7", got)
	}
	// 酉时: 17:00-19:00 = [1020,1140).
	if got := HourBranchFromSolarTime(1080); got != BranchYou {
		t.Errorf("HourBranch(1080)=%d, want 酉=10", got)
	}
}

// -- IsDST (China 1986-1991) --

func TestIsDST(t *testing.T) {
	if !IsDST(1986, 6, 1) {
		t.Error("1986-06-01 should be DST")
	}
	if IsDST(1986, 1, 1) {
		t.Error("1986-01-01 should NOT be DST")
	}
	if IsDST(1985, 6, 1) {
		t.Error("1985 should NOT be DST")
	}
	if IsDST(1992, 6, 1) {
		t.Error("1992 should NOT be DST")
	}
}

// -- ComputeSolarTime --

func TestComputeSolarTime_BeijingNoon(t *testing.T) {
	// Beijing 120°E, timezone UTC+8.
	ast := ComputeSolarTime(2024, 6, 21, 12, 0, 120.0, 8.0, false)
	// Should be close to 720 (12:00) ± 15 minutes (EoT range).
	if ast < 705 || ast > 735 {
		t.Errorf("Beijing noon solar time = %.1f, want ~720", ast)
	}
}

func TestComputeSolarTime_Longitude(t *testing.T) {
	// Urumqi E87.6, timezone UTC+8 — large westward offset.
	ast := ComputeSolarTime(2024, 6, 21, 12, 0, 87.6, 8.0, false)
	// Longitude offset = 4*(87.6-120) = -129.6 min. Solar time ≈ 720-129.6 ± EoT.
	if ast > 620 {
		t.Errorf("Urumqi solar time = %.1f, want < 620 (westward)", ast)
	}
}

func TestComputeSolarTime_DST(t *testing.T) {
	withDST := ComputeSolarTime(1988, 7, 15, 12, 0, 120.0, 120.0, true)
	withoutDST := ComputeSolarTime(1988, 7, 15, 12, 0, 120.0, 120.0, false)
	if withDST >= withoutDST {
		t.Errorf("DST should be ~60 min earlier: dst=%.1f, no-dst=%.1f", withDST, withoutDST)
	}
}

// -- HiddenStems --

func TestHiddenStemsForBranch(t *testing.T) {
	// 子藏癸 — one hidden stem.
	hs := HiddenStemsForBranch(BranchZi)
	if hs.Main == nil || *hs.Main != 10 {
		t.Errorf("子 main=%v, want 癸(10)", hs.Main)
	}
	if hs.Mid != nil || hs.Minor != nil {
		t.Error("子 should only have main stem")
	}

	// 辰藏戊乙癸 — three hidden stems.
	hs2 := HiddenStemsForBranch(BranchChen)
	if hs2.Main == nil || *hs2.Main != 5 {
		t.Errorf("辰 main=%v, want 戊(5)", hs2.Main)
	}
	if hs2.Mid == nil || hs2.Minor == nil {
		t.Error("辰 should have mid and minor stems")
	}
}

// -- TenGod type --

func TestTenGodType_SameElement(t *testing.T) {
	// 甲(木,yang) vs 甲(木,yang) = 比肩.
	if got := TenGodType(ElemWood, Yang, ElemWood, Yang); got != TenGodBiJian {
		t.Errorf("甲-甲 = %d, want 比肩=%d", got, TenGodBiJian)
	}
	// 甲(木,yang) vs 乙(木,yin) = 劫财.
	if got := TenGodType(ElemWood, Yang, ElemWood, Yin); got != TenGodJieCai {
		t.Errorf("甲-乙 = %d, want 劫财=%d", got, TenGodJieCai)
	}
}

func TestTenGodType_Generates(t *testing.T) {
	// 甲(木,yang) 生 丙(火,yang) = 食神.
	if got := TenGodType(ElemWood, Yang, ElemFire, Yang); got != TenGodShiShen {
		t.Errorf("甲-丙 = %d, want 食神=%d", got, TenGodShiShen)
	}
	// 甲(木,yang) 生 丁(火,yin) = 伤官.
	if got := TenGodType(ElemWood, Yang, ElemFire, Yin); got != TenGodShangGuan {
		t.Errorf("甲-丁 = %d, want 伤官=%d", got, TenGodShangGuan)
	}
}

func TestTenGodType_GeneratedBy(t *testing.T) {
	// 甲(木) 被 壬(水,yang) 生 = 偏印.
	if got := TenGodType(ElemWood, Yang, ElemWater, Yang); got != TenGodPianYin {
		t.Errorf("被壬生 = %d, want 偏印=%d", got, TenGodPianYin)
	}
	// 甲(木) 被 癸(水,yin) 生 = 正印.
	if got := TenGodType(ElemWood, Yang, ElemWater, Yin); got != TenGodZhengYin {
		t.Errorf("被癸生 = %d, want 正印=%d", got, TenGodZhengYin)
	}
}

func TestTenGodType_Controls(t *testing.T) {
	// 甲(木) 克 戊(土,yang) = 偏财.
	if got := TenGodType(ElemWood, Yang, ElemEarth, Yang); got != TenGodPianCai {
		t.Errorf("甲-戊 = %d, want 偏财=%d", got, TenGodPianCai)
	}
	// 甲(木) 克 己(土,yin) = 正财.
	if got := TenGodType(ElemWood, Yang, ElemEarth, Yin); got != TenGodZhengCai {
		t.Errorf("甲-己 = %d, want 正财=%d", got, TenGodZhengCai)
	}
}

// -- Large fortune (大运) --

func TestFortunePillars(t *testing.T) {
	// Start from 丙寅月 (index of 丙寅 in 60-cycle is 2).
	pillars := fortunePillars(Pillar{Stem: StemBing, Branch: BranchYin}, true, 8)
	if len(pillars) != 8 {
		t.Fatalf("got %d pillars, want 8", len(pillars))
	}
	// First forward pillar should be the next in 60-cycle: 丁卯 (idx 3).
	firstIdx := sixtyCycleName(pillars[0].Stem, pillars[0].Branch)
	if firstIdx != 3 {
		t.Errorf("first forward pillar idx=%d, want 3 (丁卯)", firstIdx)
	}
	if pillars[0].Stem != StemDing || pillars[0].Branch != BranchMao {
		t.Errorf("first forward pillar=%d-%d, want 丁卯=4-4", pillars[0].Stem, pillars[0].Branch)
	}
}

func TestFortuneStartAge(t *testing.T) {
	// Male, Yang stem (甲) → forward, next 节 after birth.
	age := fortuneStartAge(1984, 2, 5, true)
	if age < 0 || age > 40 {
		t.Errorf("start age = %d, want 0-40", age)
	}
	// Female, Yang stem (甲) → backward, previous 节 before birth.
	age2 := fortuneStartAge(1984, 2, 5, false)
	if age2 < 0 || age2 > 40 {
		t.Errorf("start age (backward) = %d, want 0-40", age2)
	}
}

// -- Na Yin --

func TestNaYin_Known(t *testing.T) {
	// 甲子=海中金, 丙寅=炉中火.
	idx0 := sixtyCycleName(StemJia, BranchZi)
	if name, ok := defaultEngine.NayinTable[idx0]; ok {
		if name != "海中金" {
			t.Errorf("甲子纳音=%s, want 海中金", name)
		}
	} else {
		t.Error("甲子 not in nayin table")
	}

	idx1 := sixtyCycleName(StemBing, BranchYin)
	if name, ok := defaultEngine.NayinTable[idx1]; ok {
		if name != "炉中火" {
			t.Errorf("丙寅纳音=%s, want 炉中火", name)
		}
	}
}

// -- Life stages --

func TestLifeStages_AllStems(t *testing.T) {
	for s := StemJia; s <= StemGui; s++ {
		stages, ok := defaultEngine.LifeStagesTable[int(s)]
		if !ok {
			t.Errorf("no life stages for stem %d", s)
			continue
		}
		if len(stages) != 12 {
			t.Errorf("stem %d has %d stages, want 12", s, len(stages))
		}
	}
}

// -- Element count --

func TestElementCount(t *testing.T) {
	bz := Bazi{
		Year:  Pillar{Stem: StemJia, Branch: BranchYin},   // 甲寅: 木+木
		Month: Pillar{Stem: StemBing, Branch: BranchWu},   // 丙午: 火+火
		Day:   Pillar{Stem: StemWu, Branch: BranchChen},   // 戊辰: 土+土
		Hour:  Pillar{Stem: StemGeng, Branch: BranchShen}, // 庚申: 金+金
	}
	hs := [4]HiddenStemsOut{
		{Main: StemJia},   // 寅藏甲
		{Main: StemBing},  // 午藏丙
		{Main: StemWu},    // 辰藏戊
		{Main: StemGeng},  // 申藏庚
	}
	count := computeElementCount(bz, hs)
	if count[ElemWood] != 3 {
		t.Errorf("Wood count=%d, want 3 (甲寅+甲藏)", count[ElemWood])
	}
	if count[ElemFire] != 3 {
		t.Errorf("Fire count=%d, want 3", count[ElemFire])
	}
	if count[ElemEarth] != 3 {
		t.Errorf("Earth count=%d, want 3", count[ElemEarth])
	}
	if count[ElemMetal] != 3 {
		t.Errorf("Metal count=%d, want 3", count[ElemMetal])
	}
}

// -- Match (合八字) --

func TestIsStemHe(t *testing.T) {
	// 甲-己 = 合.
	if !IsStemHe(StemJia, StemJi) {
		t.Error("甲-己 should be stem-he")
	}
	if !IsStemHe(StemJi, StemJia) {
		t.Error("己-甲 should be stem-he (symmetric)")
	}
	// 甲-乙 = NOT 合.
	if IsStemHe(StemJia, StemYi) {
		t.Error("甲-乙 should NOT be stem-he")
	}
}

func TestIsBranchHe(t *testing.T) {
	// 子-丑 = 合.
	if !IsBranchHe(BranchZi, BranchChou) {
		t.Error("子-丑 should be branch-he")
	}
	// 子-寅 = NOT 合.
	if IsBranchHe(BranchZi, BranchYin) {
		t.Error("子-寅 should NOT be branch-he")
	}
}

func TestIsLiuChong(t *testing.T) {
	// 子-午 = 冲.
	if !IsLiuChong(BranchZi, BranchWu) {
		t.Error("子-午 should be 冲")
	}
	// 子-丑 = NOT 冲.
	if IsLiuChong(BranchZi, BranchChou) {
		t.Error("子-丑 should NOT be 冲")
	}
}

func TestControls_Relation(t *testing.T) {
	if !Ke(ElemWood, ElemEarth) {
		t.Error("Wood should control Earth")
	}
	if !Ke(ElemFire, ElemMetal) {
		t.Error("Fire should control Metal")
	}
	if Ke(ElemWood, ElemFire) {
		t.Error("Wood should NOT control Fire (sheng)")
	}
}

func TestGenerates_Relation(t *testing.T) {
	if !Sheng(ElemWood, ElemFire) {
		t.Error("Wood should generate Fire")
	}
	if !Sheng(ElemWater, ElemWood) {
		t.Error("Water should generate Wood")
	}
}

// -- ComputeChart integration --

func TestComputeChart_TimezoneHours(t *testing.T) {
	// Verify timezone is interpreted as hours, not degrees.
	// Longitude 127.17°E (Harbin), timezone 8.0 (UTC+8).
	// If buggy (tz=8 treated as degrees): lonOffset = 4*(127.17-8) = 476.68 min ≈ +8h → solar time ~17:00
	// Correct (tz=8 converted to 120° internally): lonOffset = 4*(127.17-120) = 28.68 min → solar time ~9:28
	ast := ComputeSolarTime(1985, 6, 7, 9, 0, 127.17, 8.0, false)
	bz := ComputeBazi(ast, 1985, 6, 7, 9, 0, 8.0, false)
	chart := ComputeChart(bz, 1985, 6, 7, "male")

	// Solar time should be near 9:28 (568 min), not near 17:00 (1020 min).
	ast = chart.SolarTime
	if ast < 540 || ast > 600 {
		t.Errorf("solar time = %.1f min, want 540–600 (≈9:00–10:00). Bug would give ~1020 (17:00)", ast)
	}

	// Hour pillar should be 巳 (9:00–11:00), not 申 (15:00–17:00).
	if chart.Hour.Branch != BranchSi {
		t.Errorf("hour branch = %d, want 巳(6)", chart.Hour.Branch)
	}
}

func TestComputeChart_Integration(t *testing.T) {
	// 1984-02-05 12:00 Beijing: 甲子 丙寅 己巳 庚午.
	ast := ComputeSolarTime(1984, 2, 5, 12, 0, 120.0, 8.0, false)
	bz := ComputeBazi(ast, 1984, 2, 5, 12, 0, 8.0, false)
	chart := ComputeChart(bz, 1984, 2, 5, "male")

	if chart.Year.Stem != StemJia || chart.Year.Branch != BranchZi {
		t.Errorf("year=%d-%d, want 甲子=1-1", chart.Year.Stem, chart.Year.Branch)
	}
	if chart.Month.Stem != StemBing || chart.Month.Branch != BranchYin {
		t.Errorf("month=%d-%d, want 丙寅=3-3", chart.Month.Stem, chart.Month.Branch)
	}
	if chart.Day.Stem != 6 || chart.Day.Branch != 6 {
		t.Errorf("day=%d-%d, want 己巳=6-6", chart.Day.Stem, chart.Day.Branch)
	}

	// Check extensions.
	if len(chart.NaYinArray()) != 4 {
		t.Errorf("got %d nayin, want 4", len(chart.NaYinArray()))
	}
	if len(chart.TenGodsArray()) != 4 {
		t.Errorf("got %d ten god pairs, want 4", len(chart.TenGodsArray()))
	}
	if chart.ElementCount == nil {
		t.Error("element count should not be nil")
	}
	if chart.Dayun.StartAge < 0 {
		t.Errorf("dayun start age = %d, want >= 0", chart.Dayun.StartAge)
	}
	if chart.SolarTime < 0 || chart.SolarTime >= 1440 {
		t.Errorf("solar time = %.1f, want [0,1440)", chart.SolarTime)
	}
}

func TestComputeChart_2024LiChun(t *testing.T) {
	// 2024-02-04 18:00 Beijing (after 立春 at 16:12): 甲辰 丙寅 戊戌 辛酉.
	ast := ComputeSolarTime(2024, 2, 4, 18, 0, 120.0, 8.0, false)
	bz := ComputeBazi(ast, 2024, 2, 4, 18, 0, 8.0, false)
	chart := ComputeChart(bz, 2024, 2, 4, "male")

	if chart.Month.Stem != StemBing || chart.Month.Branch != BranchYin {
		t.Errorf("month=%d-%d, want 丙寅=3-3 (after 立春)", chart.Month.Stem, chart.Month.Branch)
	}
	if chart.Day.Stem != 5 || chart.Day.Branch != 11 {
		t.Errorf("day=%d-%d, want 戊戌=5-11", chart.Day.Stem, chart.Day.Branch)
	}
}

// -- SearchCities --

func TestSearchCities_EnglishPrefix(t *testing.T) {

	cities := SearchCities("Bei")
	if len(cities) == 0 {
		t.Fatal("SearchCities(Bei) returned no results")
	}
	found := false
	for _, c := range cities {
		if c.NameZh == "北京" || c.Name == "Beijing" {
			found = true
			if c.NameZh != "北京" {
				t.Errorf("Beijing name_zh = %q, want 北京", c.NameZh)
			}
		}
	}
	if !found {
		t.Error("Beijing not found in SearchCities(Bei)")
	}
}

func TestSearchCities_ChinesePrefix(t *testing.T) {

	cities := SearchCities("北")
	if len(cities) == 0 {
		t.Fatal("SearchCities(北) returned no results")
	}
	found := false
	for _, c := range cities {
		if c.NameZh == "北京" {
			found = true
		}
	}
	if !found {
		t.Error("北京 not found in SearchCities(北)")
	}
}

func TestSearchCities_ShortQuery(t *testing.T) {
	if cities := SearchCities(""); cities != nil {
		t.Errorf("SearchCities('') = %v, want nil", cities)
	}
}

func TestSearchCities_MaxResults(t *testing.T) {

	cities := SearchCities("a")
	if len(cities) > 20 {
		t.Errorf("SearchCities returned %d results, want <= 20", len(cities))
	}
}
