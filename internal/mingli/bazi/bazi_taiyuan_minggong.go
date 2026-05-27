package bazi

// TaiYuanMingGong holds the 胎元, 命宫 and 身宫 pillars for a chart.
type TaiYuanMingGong struct {
	TaiYuan  Pillar `json:"tai_yuan"`
	MingGong Pillar `json:"ming_gong"`
	ShenGong Pillar `json:"shen_gong"`
}

// ComputeTaiYuanMingGong computes the 三垣 (three palaces): 胎元, 命宫, 身宫.
//
// 胎元: month stem + 1, month branch + 3 (wrapping around 60-cycle).
// 命宫: count from 子(1) backward to birth month → result, then from result forward
//
//	to birth hour (子时=1) → result is 命宫 branch. 命宫 stem = 五虎遁 formula:
//	(year stem × 2 + mingGong branch index) % 10.
//
// 身宫: count from 子(1) forward to birth month → result, then from result backward
//
//	to birth hour → result is 身宫 branch. 身宫 stem = same 五虎遁 formula.
func ComputeTaiYuanMingGong(monthPillar Pillar, yearStem Stem, birthMonth, birthHour int) TaiYuanMingGong {
	// 胎元: month stem+1 (mod 10), month branch+3 (mod 12)
	tyStem := int(monthPillar.Stem) + 1
	if tyStem > 10 {
		tyStem -= 10
	}
	tyBranch := int(monthPillar.Branch) + 3
	if tyBranch > 12 {
		tyBranch -= 12
	}
	taiYuan := Pillar{Stem: Stem(tyStem), Branch: Branch(tyBranch)}

	// 命宫: 以子为正月逆数至生月
	// 正月=子(1), 二月=亥(12), 三月=戌(11), ...
	// month branch = (1 - (month - 1) + 12) % 12 → if 0 → 12
	monthOnZi := (1 - (birthMonth - 1) + 12) % 12
	if monthOnZi == 0 {
		monthOnZi = 12
	}

	// 从结果顺数至生时（子时=1, 丑时=2, ...）
	hourBranch := HourBranchFromSolarTime(float64(birthHour * 60))
	// Forward: monthOnZi → count up to hourBranch
	mgBranch := (monthOnZi + int(hourBranch) - 1) % 12
	if mgBranch == 0 {
		mgBranch = 12
	}

	// 命宫天干：五虎遁 formula. Convert branch number (子=1) to month index (寅=1).
	mgMonthIdx := ((mgBranch - 3 + 12) % 12) + 1
	mgStem := (int(yearStem)*2 + mgMonthIdx) % 10
	if mgStem == 0 {
		mgStem = 10
	}

	mingGong := Pillar{Stem: Stem(mgStem), Branch: Branch(mgBranch)}

	// 身宫: 以子为正月顺数至生月
	// 正月=子(1), 二月=丑(2), ..., 十二月=亥(12) — direct mapping.
	shenStart := birthMonth

	// 从所得宫位逆数至生时（子时=1, 丑时=2, ...）
	// Backward: shenStart → count down to hourBranch
	sgBranch := (shenStart - int(hourBranch) + 1 + 12) % 12
	if sgBranch == 0 {
		sgBranch = 12
	}

	// 身宫天干：五虎遁 formula. Convert branch number (子=1) to month index (寅=1).
	sgMonthIdx := ((sgBranch - 3 + 12) % 12) + 1
	sgStem := (int(yearStem)*2 + sgMonthIdx) % 10
	if sgStem == 0 {
		sgStem = 10
	}

	shenGong := Pillar{Stem: Stem(sgStem), Branch: Branch(sgBranch)}

	return TaiYuanMingGong{TaiYuan: taiYuan, MingGong: mingGong, ShenGong: shenGong}
}
