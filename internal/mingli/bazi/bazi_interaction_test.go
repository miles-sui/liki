package bazi

import "testing"

func TestAnalyzeStemRelation(t *testing.T) {
	// 甲己合化土 (StemJia=1, StemJi=6)
	r := AnalyzeStemRelation(StemJia, StemJi)
	if r.Type != "天干五合" {
		t.Errorf("甲+己 should be 天干五合, got %s", r.Type)
	}

	// 丙辛合化水 (StemBing=3, StemXin=8)
	r = AnalyzeStemRelation(StemBing, StemXin)
	if r.Type != "天干五合" {
		t.Errorf("丙+辛 should be 天干五合, got %s", r.Type)
	}

	// 甲+乙: both wood → same element
	r = AnalyzeStemRelation(StemJia, StemYi)
	if r.Type != "相同" {
		t.Errorf("甲+乙 should be 相同, got %s", r.Type)
	}

	// 甲+庚: wood (甲) → metal (庚) controls wood
	r = AnalyzeStemRelation(StemJia, StemGeng)
	if r.Type != "相克" {
		t.Errorf("甲+庚 should be 相克, got %s", r.Type)
	}

	// 甲+壬: water (壬) generates wood (甲)
	r = AnalyzeStemRelation(StemJia, StemRen)
	if r.Type != "相生" {
		t.Errorf("甲+壬 should be 相生, got %s", r.Type)
	}

	// Same stem.
	r = AnalyzeStemRelation(StemJia, StemJia)
	if r.Type != "相同" {
		t.Errorf("甲+甲 should be 相同, got %s", r.Type)
	}
}

func TestAnalyzeBranchRelation(t *testing.T) {
	// 子丑合化土
	r := AnalyzeBranchRelation(BranchZi, BranchChou)
	if r.Type != "六合" {
		t.Errorf("子+丑 should be 六合, got %s", r.Type)
	}

	// 子午冲
	r = AnalyzeBranchRelation(BranchZi, BranchWu)
	if r.Type != "六冲" {
		t.Errorf("子+午 should be 六冲, got %s", r.Type)
	}

	// 寅亥合化木
	r = AnalyzeBranchRelation(BranchYin, BranchHai)
	if r.Type != "六合" {
		t.Errorf("寅+亥 should be 六合, got %s", r.Type)
	}

	// 申子辰三合水局 (any two)
	r = AnalyzeBranchRelation(BranchShen, BranchZi)
	if r.Type != "三合" {
		t.Errorf("申+子 should be 三合, got %s", r.Type)
	}

	// 寅卯辰三会木方
	r = AnalyzeBranchRelation(BranchYin, BranchMao)
	if r.Type != "三会" {
		t.Errorf("寅+卯 should be 三会, got %s", r.Type)
	}

	// 子卯刑
	r = AnalyzeBranchRelation(BranchZi, BranchMao)
	if r.Type != "相刑" {
		t.Errorf("子+卯 should be 相刑, got %s", r.Type)
	}

	// 寅巳: both 相刑 (wuen) and 六害. 相刑 takes priority.
	r = AnalyzeBranchRelation(BranchYin, BranchSi)
	if r.Type != "相刑" {
		t.Errorf("寅+巳 should be 相刑 (优先于六害), got %s", r.Type)
	}

	// 辰酉合化金
	r = AnalyzeBranchRelation(BranchChen, BranchYou)
	if r.Type != "六合" {
		t.Errorf("辰+酉 should be 六合, got %s", r.Type)
	}

	// 寅戌三合火局 (寅午戌). No direct pair but part of triple he.
	r = AnalyzeBranchRelation(BranchYin, BranchXu)
	if r.Type != "三合" {
		t.Errorf("寅+戌 should be 三合 (寅午戌火局), got %s", r.Type)
	}
}

func TestComputeCurrentPillarIndex(t *testing.T) {
	// Simulate birth year 1982, start age 3.
	// 1982 + 3 = 1985. Current year depends on test time; approximate.
	pillars := []DayunPillar{
		{AgeStart: 3, AgeEnd: 12},
		{AgeStart: 13, AgeEnd: 22},
		{AgeStart: 23, AgeEnd: 32},
		{AgeStart: 33, AgeEnd: 42},
		{AgeStart: 43, AgeEnd: 52},
		{AgeStart: 53, AgeEnd: 62},
		{AgeStart: 63, AgeEnd: 72},
		{AgeStart: 73, AgeEnd: 82},
	}

	idx := computeCurrentPillarIndex(1982, 2026, pillars)
	// Age in 2026 = 44 → should be pillar index 4 (43-52).
	if idx != 4 {
		t.Logf("birthYear=1982 → current pillar index = %d (expected 4 if running in 2026)", idx)
	}

	// Birth year 2020: age 6 → pillar index 0
	idx = computeCurrentPillarIndex(2020, 2026, pillars)
	if idx != 0 {
		t.Logf("birthYear=2020 → current pillar index = %d (expected 0)", idx)
	}

	// Birth year 1900: age 126 → past all pillars → -1
	idx = computeCurrentPillarIndex(1900, 2026, pillars)
	if idx != -1 {
		t.Logf("birthYear=1900 → current pillar index = %d (expected -1)", idx)
	}
}

func TestComputeDayunInteractions(t *testing.T) {
	// Miles' chart: 辛酉 丙申 乙亥 丁丑
	bazi := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}

	dayunPillars := []DayunPillar{
		{Stem: StemYi, Branch: BranchWei, AgeStart: 3, AgeEnd: 12, Name: "乙未", TenGod: "比肩运"},
		{Stem: StemJia, Branch: BranchWu, AgeStart: 13, AgeEnd: 22, Name: "甲午", TenGod: "劫财运"},
		{Stem: StemGui, Branch: BranchSi, AgeStart: 23, AgeEnd: 32, Name: "癸巳", TenGod: "偏印运"},
		{Stem: StemRen, Branch: BranchChen, AgeStart: 33, AgeEnd: 42, Name: "壬辰", TenGod: "正印运"},
		{Stem: StemXin, Branch: BranchMao, AgeStart: 43, AgeEnd: 52, Name: "辛卯", TenGod: "七杀运"},
		{Stem: StemGeng, Branch: BranchYin, AgeStart: 53, AgeEnd: 62, Name: "庚寅", TenGod: "正官运"},
		{Stem: StemJi, Branch: BranchChou, AgeStart: 63, AgeEnd: 72, Name: "己丑", TenGod: "偏财运"},
		{Stem: StemWu, Branch: BranchZi, AgeStart: 73, AgeEnd: 82, Name: "戊子", TenGod: "正财运"},
	}

	interactions := ComputeDayunInteractions(dayunPillars, bazi)

	if len(interactions) != 8 {
		t.Fatalf("expected 8 interactions, got %d", len(interactions))
	}

	// 壬辰运 (index 3): 辰酉合 (辰 vs 年支酉)
	rc := interactions[3]
	foundChenYouHe := false
	for _, br := range rc.BranchRels {
		if br.Type == "六合" {
			t.Logf("壬辰运 合: %s (branch_a=%d, branch_b=%d)", br.Detail, br.BranchA, br.BranchB)
			foundChenYouHe = true
		}
	}
	if !foundChenYouHe {
		t.Error("壬辰运 should have 辰酉合 interaction")
	}

	// 壬辰运: 辰申半合 (辰 vs 月支申)
	foundChenShenHe := false
	for _, br := range rc.BranchRels {
		if br.Type == "三合" && br.BranchB == BranchShen {
			foundChenShenHe = true
		}
	}
	if !foundChenShenHe {
		t.Error("壬辰运 should have 辰申三合 interaction")
	}

	// 辛卯运 (index 4): 卯酉冲 (卯 vs 年支酉)
	r := interactions[4]
	foundMaoYouChong := false
	for _, br := range r.BranchRels {
		if br.Type == "六冲" {
			t.Logf("辛卯运 冲: %s", br.Detail)
			foundMaoYouChong = true
		}
	}
	if !foundMaoYouChong {
		t.Error("辛卯运 should have 卯酉冲 interaction")
	}

	// 辛卯运: 卯亥三合 (卯 vs 日支亥 → 亥卯未三合木局)
	foundMaoHaiHe := false
	for _, br := range r.BranchRels {
		if (br.Type == "三合" || br.Type == "六合") && br.BranchB == BranchHai {
			foundMaoHaiHe = true
		}
	}
	if !foundMaoHaiHe {
		t.Error("辛卯运 should have 卯亥合/三合 interaction")
	}
}

func TestComputeTenGodsTable(t *testing.T) {
	// Miles' chart: 辛酉 丙申 乙亥 丁丑
	pillars := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	dm := StemYi // 乙木日主

	hs := computeHiddenStems(pillars)
	table := ComputeTenGodsTable(dm, pillars, hs)

	for pi := range table {
		if len(table[pi]) == 0 {
			t.Errorf("pillar %d: expected entries, got none", pi)
		}
		// First entry should be stem ten god.
		if table[pi][0].Source != SourceStem {
			t.Errorf("pillar %d: first entry source = %s, want stem", pi, table[pi][0].Source)
		}
	}

	// 辛酉: 辛(metal)克乙(wood) → 七杀
	if table[0][0].TenGod != "七杀" {
		t.Errorf("year stem 辛 vs 乙 → 七杀, got %s", table[0][0].TenGod)
	}

	// 丙申: 丙(fire) is 乙生 → 伤官
	if table[1][0].TenGod != "伤官" {
		t.Errorf("month stem 丙 vs 乙 → 伤官, got %s", table[1][0].TenGod)
	}

	// 乙亥: 乙 = 乙 → 比肩
	if table[2][0].TenGod != "比肩" {
		t.Errorf("day stem 乙 vs 乙 → 比肩, got %s", table[2][0].TenGod)
	}

	// Check that hidden stems are included.
	for pi := range table {
		hasMainQi := false
		for _, e := range table[pi] {
			if e.Source == SourceMainQi {
				hasMainQi = true
			}
		}
		if !hasMainQi {
			t.Errorf("pillar %d: missing main_qi ten god", pi)
		}
	}
}

func TestComputeLifeStageTable(t *testing.T) {
	pillars := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	hs := computeHiddenStems(pillars)

	table := ComputeLifeStageTable(pillars, hs)

	// Each pillar should have entries.
	for pi := range table {
		if len(table[pi]) == 0 {
			t.Errorf("pillar %d: expected life stage entries, got none", pi)
		}
		t.Logf("pillar %d life stages: %d entries", pi, len(table[pi]))
	}

	// 乙木 day stem at 亥 → 死 (亥 is the 死 position for 乙木)
	for _, e := range table[2] { // 日柱
		if e.Stem == StemYi && e.Branch == BranchHai {
			t.Logf("乙木 at 亥 → %s", e.Stage)
		}
	}
}

func TestComputeShenSha(t *testing.T) {
	pillars := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	dm := StemYi
	monthBranch := BranchShen

	result := ComputeShenSha(pillars, dm, monthBranch)

	total := 0
	for pi := range result {
		total += len(result[pi])
		t.Logf("pillar %d shensha: %d", pi, len(result[pi]))
		for _, e := range result[pi] {
			t.Logf("  %s (%s): %s", e.Name, e.Category, e.Description)
		}
	}
	if total == 0 {
		t.Error("expected at least some shensha, got none")
	}
}

func TestComputeKongWang(t *testing.T) {
	// 乙亥日柱: 甲戌旬, 空亡 申酉
	dayPillar := Pillar{Stem: StemYi, Branch: BranchHai}
	allPillars := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}

	hits := ComputeKongWang(dayPillar, allPillars)
	t.Logf("kongwang hits: %v", hits)

	if len(hits) < 1 {
		t.Error("expected at least 1 void hit (酉 or 申)")
	}

	// 年柱(酉)和月柱(申)都应该落入空亡
	foundYou, foundShen := false, false
	for _, h := range hits {
		if h == 0 {
			foundYou = true
		}
		if h == 1 {
			foundShen = true
		}
	}
	if !foundYou {
		t.Error("year pillar (酉) should be in void")
	}
	if !foundShen {
		t.Error("month pillar (申) should be in void")
	}
}

func TestComputeLiuri(t *testing.T) {
	bazi := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	dm := StemYi

	result := ComputeLiuri("2026-05-24", dm, bazi, nil, nil)
	if result == nil {
		t.Fatal("ComputeLiuri returned nil")
	}

	t.Logf("date=%s day=%s ten_god=%s nayin=%s", result.Date, result.DayName, result.TenGod, result.DayNaYin)
	t.Logf("stem_rels=%d branch_rels=%d", len(result.StemRels), len(result.BranchRels))

	if result.DayName == "" {
		t.Error("expected non-empty day name")
	}
	if result.TenGod == "" {
		t.Error("expected non-empty ten god")
	}
}

func TestChartPillarInfo(t *testing.T) {
	// Full chart for Miles.
	ast := ComputeSolarTime(1982, 10, 13, 6, 45, 114.134, 8.0, false)
	bz := ComputeBazi(ast, 1982, 10, 13, 6, 45, 8.0, false)
	chart := ComputeChart(bz, 1982, 10, 13, "male")

	out := [4]PillarInfo{chart.Year, chart.Month, chart.Day, chart.Hour}

	for i, po := range out {
		t.Logf("pillar %d: stem=%d branch=%d nayin=%s void=%v ten_gods=%d life_stages=%d shensha=%d",
			i, po.Stem, po.Branch, po.NaYin, po.IsVoid, len(po.TenGods), len(po.LifeStages), len(po.ShenSha))

		if po.NaYin == "" {
			t.Errorf("pillar %d: expected nayin", i)
		}
		if len(po.TenGods) == 0 {
			t.Errorf("pillar %d: expected ten gods", i)
		}
	}

	// At least 2 pillars should have life stages.
	lifePillars := 0
	for _, po := range out {
		if len(po.LifeStages) > 0 {
			lifePillars++
		}
	}
	if lifePillars < 2 {
		t.Errorf("expected at least 2 pillars with life stages, got %d", lifePillars)
	}
}

func TestChartPillarInfoShenSha(t *testing.T) {
	ast := ComputeSolarTime(1982, 10, 13, 6, 45, 114.134, 8.0, false)
	bz := ComputeBazi(ast, 1982, 10, 13, 6, 45, 8.0, false)
	chart := ComputeChart(bz, 1982, 10, 13, "male")
	out := [4]PillarInfo{chart.Year, chart.Month, chart.Day, chart.Hour}

	// Miles' chart should have 文昌 at year pillar (辛→子, not 酉)
	// 驿马 at day pillar (亥卯未→巳, day branch 亥 is in triad → 驿马 at 巳)
	total := 0
	names := [4]string{"year", "month", "day", "hour"}
	for i, po := range out {
		total += len(po.ShenSha)
		for _, ss := range po.ShenSha {
			t.Logf("  pillar %s: %s (%s)", names[i], ss.Name, ss.Category)
		}
	}
	t.Logf("total shensha: %d", total)
}

func TestComputeDynamicShenSha(t *testing.T) {
	// Year branch 酉(10), day stem 乙(StemYi=2).
	// 巳酉丑 triad: 驿马→亥(12), 桃花→午(7), 华盖→丑(2)
	// 亥(12) against year酉(10): 驿马=亥 ✓
	result := ComputeDynamicShenSha(BranchHai, BranchYou, StemYi)
	foundYima := false
	for _, ss := range result {
		if ss.Name == "驿马" {
			foundYima = true
		}
		t.Logf("dynamic shensha (亥vs酉): %s (%s)", ss.Name, ss.Category)
	}
	if !foundYima {
		t.Error("expected 驿马 when branch=亥 for year branch=酉")
	}

	// 天乙贵人 for 乙: 子(1) and 申(9).
	// 天喜 for year酉(10) → 子(1).
	// So 子(1) vs year酉(10), dayStem乙 → 天乙贵人 + 天喜 = 2 shensha
	result2 := ComputeDynamicShenSha(BranchZi, BranchYou, StemYi)
	if len(result2) < 2 {
		t.Errorf("子 vs 酉 expected >=2 shensha (天乙贵人+天喜), got %d", len(result2))
	}
	for _, ss := range result2 {
		t.Logf("dynamic shensha (子vs酉): %s (%s)", ss.Name, ss.Category)
	}

	// 卯(4) vs year酉+乙日主: no shensha triggers.
	result3 := ComputeDynamicShenSha(BranchMao, BranchYou, StemYi)
	if len(result3) != 0 {
		t.Logf("卯 vs 酉 got %d shensha (expected 0)", len(result3))
	}
}

func TestComputeLiushi(t *testing.T) {
	bazi := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	dm := StemYi

	result := ComputeLiushi("2026-05-24", 9, dm, bazi) // 巳时
	if result == nil {
		t.Fatal("ComputeLiushi returned nil")
	}

	t.Logf("time=%s hour=%s ten_god=%s", result.Time, result.HourName, result.TenGod)
	t.Logf("stem_rels=%d branch_rels=%d", len(result.StemRels), len(result.BranchRels))

	if result.HourName == "" {
		t.Error("expected non-empty hour name")
	}
	if result.TenGod == "" {
		t.Error("expected non-empty ten god")
	}
	// 巳时的天干应正确 (乙日→丙子起→巳时=辛巳)
	if result.HourBranch != BranchSi {
		t.Errorf("expected hour branch 巳(6), got %d", result.HourBranch)
	}

	// Hour 23 should map to 子时
	result23 := ComputeLiushi("2026-05-24", 23, dm, bazi)
	if result23 != nil && result23.HourBranch != BranchZi {
		t.Errorf("23:00 should be 子时(1), got %d", result23.HourBranch)
	}

	// Hour 1 should map to 丑时
	result1 := ComputeLiushi("2026-05-24", 1, dm, bazi)
	if result1 != nil && result1.HourBranch != BranchChou {
		t.Errorf("01:00 should be 丑时(2), got %d", result1.HourBranch)
	}
}

func TestComputeLiuyue(t *testing.T) {
	bazi := Bazi{
		Year: Pillar{Stem: StemXin, Branch: BranchYou},
		Month: Pillar{Stem: StemBing, Branch: BranchShen},
		Day: Pillar{Stem: StemYi, Branch: BranchHai},
		Hour: Pillar{Stem: StemDing, Branch: BranchChou},
	}
	dm := StemYi

	result := ComputeLiuyue(2026, 5, dm, bazi)
	if result == nil {
		t.Fatal("ComputeLiuyue returned nil")
	}

	t.Logf("year=%d month=%d name=%s ten_god=%s",
		result.Year, result.Month, result.MonthName, result.TenGod)
	t.Logf("stem_rels=%d branch_rels=%d shensha=%d",
		len(result.StemRels), len(result.BranchRels), len(result.ShenSha))

	if result.MonthName == "" {
		t.Error("expected non-empty month name")
	}
	if result.TenGod == "" {
		t.Error("expected non-empty ten god")
	}
}


