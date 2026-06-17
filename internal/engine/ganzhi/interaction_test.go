package ganzhi

import "testing"

// ── IsGanHe ──

func TestIsGanHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Gan
	}{
		{"甲己合", GanJia, GanJi},
		{"乙庚合", GanYi, GanGeng},
		{"丙辛合", GanBing, GanXin},
		{"丁壬合", GanDing, GanRen},
		{"戊癸合", GanWu, GanGui},
	}
	for _, tc := range tests {
		if !IsGanHe(tc.a, tc.b) {
			t.Errorf("IsGanHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsGanHe(tc.b, tc.a) {
			t.Errorf("IsGanHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsGanHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Gan
	}{
		{"甲乙", GanJia, GanYi},
		{"甲丙", GanJia, GanBing},
		{"己庚", GanJi, GanGeng},
		{"乙丙", GanYi, GanBing},
	}
	for _, tc := range tests {
		if IsGanHe(tc.a, tc.b) {
			t.Errorf("IsGanHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsGanHe_SameStem(t *testing.T) {
	for _, g := range []Gan{GanJia, GanYi, GanBing, GanDing, GanWu} {
		if IsGanHe(g, g) {
			t.Errorf("IsGanHe(%d,%d)=true, same stem should be false", g, g)
		}
	}
}

// ── IsZhiHe ──

func TestIsZhiHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合", ZhiZi, ZhiChou},
		{"寅亥合", ZhiYin, ZhiHai},
		{"卯戌合", ZhiMao, ZhiXu},
		{"辰酉合", ZhiChen, ZhiYou},
		{"巳申合", ZhiSi, ZhiShen},
		{"午未合", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if !IsZhiHe(tc.a, tc.b) {
			t.Errorf("IsZhiHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsZhiHe(tc.b, tc.a) {
			t.Errorf("IsZhiHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsZhiHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子寅", ZhiZi, ZhiYin},
		{"丑卯", ZhiChou, ZhiMao},
		{"子午冲不是合", ZhiZi, ZhiWu},
	}
	for _, tc := range tests {
		if IsZhiHe(tc.a, tc.b) {
			t.Errorf("IsZhiHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsTripleHe ──

func TestIsTripleHe_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"申子（水局）", ZhiShen, ZhiZi},
		{"子辰（水局）", ZhiZi, ZhiChen},
		{"亥卯（木局）", ZhiHai, ZhiMao},
		{"卯未（木局）", ZhiMao, ZhiWei},
		{"寅午（火局）", ZhiYin, ZhiWu},
		{"午戌（火局）", ZhiWu, ZhiXu},
		{"巳酉（金局）", ZhiSi, ZhiYou},
		{"酉丑（金局）", ZhiYou, ZhiChou},
	}
	for _, tc := range tests {
		if !IsTripleHe(tc.a, tc.b) {
			t.Errorf("IsTripleHe(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsTripleHe(tc.b, tc.a) {
			t.Errorf("IsTripleHe(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsTripleHe_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"申寅（水火不同局）", ZhiShen, ZhiYin},
		{"子午（水土不同局）", ZhiZi, ZhiWu},
		{"亥寅（木木但不同三合）", ZhiHai, ZhiYin},
	}
	for _, tc := range tests {
		if IsTripleHe(tc.a, tc.b) {
			t.Errorf("IsTripleHe(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsTripleHui ──

func TestIsTripleHui_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"寅卯（东方木）", ZhiYin, ZhiMao},
		{"卯辰（东方木）", ZhiMao, ZhiChen},
		{"巳午（南方火）", ZhiSi, ZhiWu},
		{"午未（南方火）", ZhiWu, ZhiWei},
		{"申酉（西方金）", ZhiShen, ZhiYou},
		{"酉戌（西方金）", ZhiYou, ZhiXu},
		{"亥子（北方水）", ZhiHai, ZhiZi},
		{"子丑（北方水）", ZhiZi, ZhiChou},
	}
	for _, tc := range tests {
		if !IsTripleHui(tc.a, tc.b) {
			t.Errorf("IsTripleHui(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsTripleHui(tc.b, tc.a) {
			t.Errorf("IsTripleHui(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsTripleHui_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"寅巳（不同方）", ZhiYin, ZhiSi},
		{"子午（不同方）", ZhiZi, ZhiWu},
		{"申寅（不同方）", ZhiShen, ZhiYin},
	}
	for _, tc := range tests {
		if IsTripleHui(tc.a, tc.b) {
			t.Errorf("IsTripleHui(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsLiuChong ──

func TestIsLiuChong_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子午冲", ZhiZi, ZhiWu},
		{"丑未冲", ZhiChou, ZhiWei},
		{"寅申冲", ZhiYin, ZhiShen},
		{"卯酉冲", ZhiMao, ZhiYou},
		{"辰戌冲", ZhiChen, ZhiXu},
		{"巳亥冲", ZhiSi, ZhiHai},
	}
	for _, tc := range tests {
		if !IsLiuChong(tc.a, tc.b) {
			t.Errorf("IsLiuChong(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsLiuChong(tc.b, tc.a) {
			t.Errorf("IsLiuChong(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsLiuChong_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合不是冲", ZhiZi, ZhiChou},
		{"寅亥合不是冲", ZhiYin, ZhiHai},
		{"午未合不是冲", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if IsLiuChong(tc.a, tc.b) {
			t.Errorf("IsLiuChong(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsLiuChong_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiYin, ZhiWu} {
		if IsLiuChong(z, z) {
			t.Errorf("IsLiuChong(%d,%d)=true, same branch should be false", z, z)
		}
	}
}

// ── IsXing ──

func TestIsXing_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		// 无礼之刑
		{"子卯无礼之刑", ZhiZi, ZhiMao},
		// 无恩之刑
		{"寅巳无恩之刑", ZhiYin, ZhiSi},
		{"巳申无恩之刑", ZhiSi, ZhiShen},
		{"寅申无恩之刑", ZhiYin, ZhiShen},
		// 恃势之刑
		{"丑未恃势之刑", ZhiChou, ZhiWei},
		{"未戌恃势之刑", ZhiWei, ZhiXu},
		{"丑戌恃势之刑", ZhiChou, ZhiXu},
	}
	for _, tc := range tests {
		if !IsXing(tc.a, tc.b) {
			t.Errorf("IsXing(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsXing(tc.b, tc.a) {
			t.Errorf("IsXing(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsXing_SelfXing(t *testing.T) {
	// 自刑：辰午酉亥
	for _, z := range []Zhi{ZhiChen, ZhiWu, ZhiYou, ZhiHai} {
		if !IsXing(z, z) {
			t.Errorf("IsXing(%d,%d)=false, self-xing should be true", z, z)
		}
	}
	// Non-self-xing branches should not be self-xing
	for _, z := range []Zhi{ZhiZi, ZhiChou, ZhiYin, ZhiMao, ZhiSi, ZhiWei, ZhiShen, ZhiXu} {
		if IsXing(z, z) {
			t.Errorf("IsXing(%d,%d)=true, should not be self-xing", z, z)
		}
	}
}

func TestIsXing_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑无刑", ZhiZi, ZhiChou},
		{"寅卯无刑", ZhiYin, ZhiMao},
		{"子丑合不是刑", ZhiZi, ZhiChou},
		{"午未合不是刑", ZhiWu, ZhiWei},
	}
	for _, tc := range tests {
		if IsXing(tc.a, tc.b) {
			t.Errorf("IsXing(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

// ── IsHai ──

func TestIsHai_Positive(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子未害", ZhiZi, ZhiWei},
		{"丑午害", ZhiChou, ZhiWu},
		{"寅巳害", ZhiYin, ZhiSi},
		{"卯辰害", ZhiMao, ZhiChen},
		{"申亥害", ZhiShen, ZhiHai},
		{"酉戌害", ZhiYou, ZhiXu},
	}
	for _, tc := range tests {
		if !IsHai(tc.a, tc.b) {
			t.Errorf("IsHai(%d,%d)=false, want true (%s)", tc.a, tc.b, tc.name)
		}
		if !IsHai(tc.b, tc.a) {
			t.Errorf("IsHai(%d,%d)=false, want true (reversed %s)", tc.b, tc.a, tc.name)
		}
	}
}

func TestIsHai_Negative(t *testing.T) {
	tests := []struct {
		name string
		a, b Zhi
	}{
		{"子丑合不是害", ZhiZi, ZhiChou},
		{"子午冲不是害", ZhiZi, ZhiWu},
		{"寅亥合不是害", ZhiYin, ZhiHai},
	}
	for _, tc := range tests {
		if IsHai(tc.a, tc.b) {
			t.Errorf("IsHai(%d,%d)=true, want false (%s)", tc.a, tc.b, tc.name)
		}
	}
}

func TestIsHai_SameBranch(t *testing.T) {
	for _, z := range []Zhi{ZhiZi, ZhiWu, ZhiYin} {
		if IsHai(z, z) {
			t.Errorf("IsHai(%d,%d)=true, same branch should be false", z, z)
		}
	}
}

// ── inBranchList ──

func TestInBranchList(t *testing.T) {
	branches := []int{1, 3, 5}
	tests := []struct {
		name string
		z    Zhi
		want bool
	}{
		{"子在内", ZhiZi, true},
		{"寅在内", ZhiYin, true},
		{"辰在内", ZhiChen, true},
		{"丑不在内", ZhiChou, false},
		{"卯不在内", ZhiMao, false},
	}
	for _, tc := range tests {
		got := inBranchList(branches, tc.z)
		if got != tc.want {
			t.Errorf("inBranchList(%v,%d)=%v, want %v (%s)", branches, tc.z, got, tc.want, tc.name)
		}
	}
}

func TestInBranchList_Empty(t *testing.T) {
	if inBranchList(nil, ZhiZi) {
		t.Error("inBranchList(nil, 子)=true, want false")
	}
	if inBranchList([]int{}, ZhiZi) {
		t.Error("inBranchList([], 子)=true, want false")
	}
}
