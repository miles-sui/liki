package ganzhi

// Stem-and-branch relation lookup tables. These are standard traditional
// Chinese metaphysics pairings that never change, so they are defined as
// compile-time constants rather than loaded from config.

var (
	// 天干五合 pairs: (stem_a, stem_b) → 化五行 element
	stemHePairs = [][3]int{
		{1, 6, 3},  // 甲+己 → 土
		{2, 7, 4},  // 乙+庚 → 金
		{3, 8, 5},  // 丙+辛 → 水
		{4, 9, 1},  // 丁+壬 → 木
		{5, 10, 2}, // 戊+癸 → 火
	}

	// 地支六合 pairs: (branch_a, branch_b)
	branchHePairs = [][2]int{
		{1, 2},  // 子+丑
		{3, 12}, // 寅+亥
		{4, 11}, // 卯+戌
		{5, 10}, // 辰+酉
		{6, 9},  // 巳+申
		{7, 8},  // 午+未
	}

	// 地支三合 groups: three branches → element
	tripleHeGroups = [][4]int{
		{9, 1, 5, 5},   // 申子辰 → 水
		{12, 4, 8, 1},  // 亥卯未 → 木
		{3, 7, 11, 2},  // 寅午戌 → 火
		{6, 10, 2, 4},  // 巳酉丑 → 金
	}

	// 地支三会 groups: three consecutive branches → element
	tripleHuiGroups = [][4]int{
		{3, 4, 5, 1},    // 寅卯辰 → 木
		{6, 7, 8, 2},    // 巳午未 → 火
		{9, 10, 11, 4},  // 申酉戌 → 金
		{12, 1, 2, 5},   // 亥子丑 → 水
	}

	// 六冲 pairs
	chongPairs = [][2]int{
		{1, 7},  // 子午冲
		{2, 8},  // 丑未冲
		{3, 9},  // 寅申冲
		{4, 10}, // 卯酉冲
		{5, 11}, // 辰戌冲
		{6, 12}, // 巳亥冲
	}

	// 相刑 groups
	xingGroups = []struct {
		branches []int
	}{
		{[]int{1, 4}},             // 无礼之刑: 子卯
		{[]int{3, 6, 9}},          // 无恩之刑: 寅巳申
		{[]int{2, 8, 11}},         // 恃势之刑: 丑未戌
		{[]int{5, 7, 10, 12}},     // 自刑: 辰午酉亥
	}

	// 六害 pairs
	haiPairs = [][2]int{
		{1, 8},   // 子未害
		{2, 7},   // 丑午害
		{3, 6},   // 寅巳害
		{4, 5},   // 卯辰害
		{9, 12},  // 申亥害
		{10, 11}, // 酉戌害
	}
)

func containsInt(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// IsStemHe returns true if the two stems form a 天干五合 pair.
func IsStemHe(a, b Stem) bool {
	ai, bi := int(a), int(b)
	for _, p := range stemHePairs {
		if (ai == p[0] && bi == p[1]) || (ai == p[1] && bi == p[0]) {
			return true
		}
	}
	return false
}

// IsBranchHe returns true if the two branches form a 地支六合 pair.
func IsBranchHe(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, p := range branchHePairs {
		if (ai == p[0] && bi == p[1]) || (ai == p[1] && bi == p[0]) {
			return true
		}
	}
	return false
}

// IsTripleHe returns true if the two branches belong to the same 三合 group.
func IsTripleHe(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, g := range tripleHeGroups {
		if (ai == g[0] || ai == g[1] || ai == g[2]) && (bi == g[0] || bi == g[1] || bi == g[2]) {
			return true
		}
	}
	return false
}

// IsTripleHui returns true if the two branches belong to the same 三会 group.
func IsTripleHui(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, g := range tripleHuiGroups {
		if (ai == g[0] || ai == g[1] || ai == g[2]) && (bi == g[0] || bi == g[1] || bi == g[2]) {
			return true
		}
	}
	return false
}

// IsLiuChong returns true if the two branches form a 六冲 pair.
func IsLiuChong(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, p := range chongPairs {
		if (ai == p[0] && bi == p[1]) || (ai == p[1] && bi == p[0]) {
			return true
		}
	}
	return false
}

// IsXing returns true if the two branches belong to the same 相刑 group.
func IsXing(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, g := range xingGroups {
		if containsInt(g.branches, ai) && containsInt(g.branches, bi) {
			return true
		}
	}
	return false
}

// IsHai returns true if the two branches form a 六害 pair.
func IsHai(a, b Branch) bool {
	ai, bi := int(a), int(b)
	for _, p := range haiPairs {
		if (ai == p[0] && bi == p[1]) || (ai == p[1] && bi == p[0]) {
			return true
		}
	}
	return false
}
