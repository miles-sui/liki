package liuyao

import (
	"math/rand"
	"testing"
	"time"

	"liki/internal/engine/ganzhi"
	"liki/internal/engine/tianwen"
)

// TestMonthWangShuai verifies 旺相休囚死 classification.
// Rules: 同我者旺, 月建生爻→相, 爻生月建→休, 爻克月建→囚, 月建克爻→死
func TestWangShuaiOf(t *testing.T) {
	tests := []struct {
		name     string
		lineZhi  ganzhi.Zhi
		monthZhi ganzhi.Zhi
		want     ganzhi.WangShuai
	}{
		// 月建寅(木)
		{"寅月同木→旺", ganzhi.ZhiYin, ganzhi.ZhiYin, ganzhi.WSWang},     // 寅木=寅木 → 旺
		{"寅月生火→相", ganzhi.ZhiSi, ganzhi.ZhiYin, ganzhi.WSXiang},     // 木生火,巳=火 → 相
		{"寅月水生→休", ganzhi.ZhiZi, ganzhi.ZhiYin, ganzhi.WSXiu},       // 水生木,子=水 → 休
		{"寅月金克→囚", ganzhi.ZhiShen, ganzhi.ZhiYin, ganzhi.WSQiu},     // 金克木,申=金 → 囚
		{"寅月克土→死", ganzhi.ZhiChen, ganzhi.ZhiYin, ganzhi.WSSi},      // 木克土,辰=土 → 死

		// 月建午(火)
		{"午月同火→旺", ganzhi.ZhiWu, ganzhi.ZhiWu, ganzhi.WSWang},
		{"午月生土→相", ganzhi.ZhiChen, ganzhi.ZhiWu, ganzhi.WSXiang},     // 火生土
		{"午月木生→休", ganzhi.ZhiMao, ganzhi.ZhiWu, ganzhi.WSXiu},       // 木生火
		{"午月水克→囚", ganzhi.ZhiZi, ganzhi.ZhiWu, ganzhi.WSQiu},        // 水克火
		{"午月克金→死", ganzhi.ZhiYou, ganzhi.ZhiWu, ganzhi.WSSi},         // 火克金

		// 月建申(金)
		{"申月同金→旺", ganzhi.ZhiShen, ganzhi.ZhiShen, ganzhi.WSWang},
		{"申月生水→相", ganzhi.ZhiZi, ganzhi.ZhiShen, ganzhi.WSXiang},     // 金生水
		{"申月土生→休", ganzhi.ZhiChen, ganzhi.ZhiShen, ganzhi.WSXiu},     // 土生金
		{"申月火克→囚", ganzhi.ZhiWu, ganzhi.ZhiShen, ganzhi.WSQiu},       // 火克金
		{"申月克木→死", ganzhi.ZhiYin, ganzhi.ZhiShen, ganzhi.WSSi},       // 金克木

		// 月建亥(水)
		{"亥月同水→旺", ganzhi.ZhiHai, ganzhi.ZhiHai, ganzhi.WSWang},
		{"亥月生木→相", ganzhi.ZhiYin, ganzhi.ZhiHai, ganzhi.WSXiang},     // 水生木
		{"亥月金生→休", ganzhi.ZhiShen, ganzhi.ZhiHai, ganzhi.WSXiu},      // 金生水
		{"亥月土克→囚", ganzhi.ZhiChen, ganzhi.ZhiHai, ganzhi.WSQiu},      // 土克水
		{"亥月克火→死", ganzhi.ZhiWu, ganzhi.ZhiHai, ganzhi.WSSi},         // 水克火
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(tt.lineZhi), tt.monthZhi)
			if got != tt.want {
				t.Errorf("WangShuaiOf(%s,%s) = %s, want %s",
					ganzhi.ZhiName(tt.lineZhi), ganzhi.ZhiName(tt.monthZhi),
					got.String(), tt.want.String())
			}
		})
	}
}

// TestChongZhi verifies 六冲 branch pairs.
func TestChongZhi(t *testing.T) {
	tests := []struct {
		z    ganzhi.Zhi
		want ganzhi.Zhi
	}{
		{ganzhi.ZhiZi, ganzhi.ZhiWu},   // 子午冲
		{ganzhi.ZhiWu, ganzhi.ZhiZi},   // 午子冲
		{ganzhi.ZhiChou, ganzhi.ZhiWei}, // 丑未冲
		{ganzhi.ZhiWei, ganzhi.ZhiChou}, // 未丑冲
		{ganzhi.ZhiYin, ganzhi.ZhiShen}, // 寅申冲
		{ganzhi.ZhiShen, ganzhi.ZhiYin}, // 申寅冲
		{ganzhi.ZhiMao, ganzhi.ZhiYou},  // 卯酉冲
		{ganzhi.ZhiYou, ganzhi.ZhiMao},  // 酉卯冲
		{ganzhi.ZhiChen, ganzhi.ZhiXu},  // 辰戌冲
		{ganzhi.ZhiXu, ganzhi.ZhiChen},  // 戌辰冲
		{ganzhi.ZhiSi, ganzhi.ZhiHai},   // 巳亥冲
		{ganzhi.ZhiHai, ganzhi.ZhiSi},   // 亥巳冲
	}

	for _, tt := range tests {
		t.Run(ganzhi.ZhiName(tt.z)+"冲"+ganzhi.ZhiName(tt.want), func(t *testing.T) {
			got := chongZhi(tt.z)
			if got != tt.want {
				t.Errorf("chongZhi(%s) = %s, want %s",
					ganzhi.ZhiName(tt.z), ganzhi.ZhiName(got), ganzhi.ZhiName(tt.want))
			}
		})
	}
}

// TestInvertDongYao verifies correct bit-flip for each line position.
func TestInvertDongYao(t *testing.T) {
	// 乾为天: all yang lines.
	qianYao := [6]YaoType{ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang}
	qianBin := yaosToBin(qianYao)            // 63
	qianGua := binaryToGuaTable[qianBin]      // guaTable index for 乾为天 = 0

	tests := []struct {
		name    string
		benGua  guaIndex
		dongYao []int
		wantBin int // expected binary encoding of result
	}{
		{"动初爻", qianGua, []int{1}, qianBin ^ 1},       // 63→天风姤
		{"动二爻", qianGua, []int{2}, qianBin ^ 2},       // 63→天山遁
		{"动上爻", qianGua, []int{6}, qianBin ^ 32},      // 63→泽天夬
		{"动初上", qianGua, []int{1, 6}, qianBin ^ 1 ^ 32},
		{"无动爻", qianGua, []int{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, hasBian := invertDongYao(tt.benGua, tt.dongYao)
			if len(tt.dongYao) == 0 {
				if hasBian {
					t.Error("expected hasBian=false for empty dongYao")
				}
				return
			}
			if !hasBian {
				t.Fatal("expected hasBian=true")
			}
			wantGua := binaryToGuaTable[tt.wantBin]
			if got != wantGua {
				t.Errorf("invertDongYao(...) = %d(%s), want %d(%s)",
					got, guaTable[got].Name, wantGua, guaTable[wantGua].Name)
			}
		})
	}
}

// TestYaosToGua verifies yaosToGua maps to correct guaTable entry.
func TestYaosToGua(t *testing.T) {
	tests := []struct {
		name string
		yaos [6]YaoType
	}{
		{"乾为天", [6]YaoType{ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang}},
		{"坤为地", [6]YaoType{ShaoYin, ShaoYin, ShaoYin, ShaoYin, ShaoYin, ShaoYin}},
		{"水火既济", [6]YaoType{ShaoYang, ShaoYin, ShaoYang, ShaoYin, ShaoYang, ShaoYin}},
		{"火水未济", [6]YaoType{ShaoYin, ShaoYang, ShaoYin, ShaoYang, ShaoYin, ShaoYang}},
		{"天风姤", [6]YaoType{ShaoYin, ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang}},
		{"风地观", [6]YaoType{ShaoYin, ShaoYin, ShaoYin, ShaoYin, ShaoYang, ShaoYang}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yaosToGua(tt.yaos)
			meta := guaTable[got]
			if meta.Name != tt.name {
				t.Errorf("yaosToGua(...) = %d (%s), want %s",
					got, meta.Name, tt.name)
			}
		})
	}
}

// TestDongYaoDetection verifies detecting changing lines.
func TestDongYaoDetection(t *testing.T) {
	yaos := [6]YaoType{ShaoYang, LaoYin, ShaoYang, LaoYang, ShaoYin, ShaoYin}
	dy := dongYao(yaos)
	if len(dy) != 2 {
		t.Fatalf("expected 2 dongYao, got %d: %v", len(dy), dy)
	}
	if dy[0] != 2 || dy[1] != 4 {
		t.Errorf("expected positions [2,4], got %v", dy)
	}
}

// TestComputeLiuQin verifies六亲 classification.
// Rule: 同我兄弟, 我生子孙, 生我父母, 我克妻财, 克我官鬼
func TestComputeLiuQin(t *testing.T) {
	// Palace = 乾宫(金), element = 金
	palaceElem := ganzhi.WxJin
	tests := []struct {
		name     string
		lineElem ganzhi.Wuxing
		want     LiuQin
	}{
		{"同金→兄弟", ganzhi.WxJin, QinXiongDi},
		{"金生水→子孙", ganzhi.WxShui, QinZiSun},
		{"土生金→父母", ganzhi.WxTu, QinFumu},
		{"金克木→妻财", ganzhi.WxMu, QinQiCai},
		{"火克金→官鬼", ganzhi.WxHuo, QinGuanGui},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeLiuQin(tt.lineElem, palaceElem)
			if got != tt.want {
				t.Errorf("computeLiuQin(%s,%s) = %s, want %s",
					tt.lineElem.String(), palaceElem.String(), got.String(), tt.want.String())
			}
		})
	}
}

// TestComputeChart_Smoke uses fixed yaos for deterministic verification.
func TestComputeChart_Smoke(t *testing.T) {
	// 2000-01-01 12:00:00 CST
	st := tianwen.SolarTime(time.Date(2000, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	// All ShaoYang → 乾为天, no changing lines.
	chart := ComputeChart(st, YongShiYao, [6]int{7, 7, 7, 7, 7, 7})

	if chart.Name != "乾为天" {
		t.Errorf("expected 乾为天, got %s", chart.Name)
	}
	if chart.Palace != "乾" {
		t.Errorf("expected palace 乾, got %s", chart.Palace)
	}
	if len(chart.DongYao) != 0 {
		t.Errorf("expected no dongYao, got %v", chart.DongYao)
	}
	// 世爻 should be at position 6 for 乾为天.
	if chart.YongShen.Position != 6 {
		t.Errorf("expected shiYao at 6, got %d", chart.YongShen.Position)
	}
	// All lines should have旺衰 set.
	for i := 0; i < 6; i++ {
		if chart.WangShuai[i].String() == "" {
			t.Errorf("line %d: empty wangshuai", i+1)
		}
	}
}

// =============================================================================
// LiuQin.String — 六亲名称
// =============================================================================

func TestLiuQin_String(t *testing.T) {
	tests := []struct {
		q    LiuQin
		want string
	}{
		{QinFumu, "父母"},
		{QinXiongDi, "兄弟"},
		{QinGuanGui, "官鬼"},
		{QinQiCai, "妻财"},
		{QinZiSun, "子孙"},
		{LiuQin(-1), "?"},
		{LiuQin(5), "?"},
		{LiuQin(100), "?"},
	}
	for _, tt := range tests {
		got := tt.q.String()
		if got != tt.want {
			t.Errorf("LiuQin(%d).String() = %s, want %s", int(tt.q), got, tt.want)
		}
	}
}

// =============================================================================
// LiuShou.String — 六兽名称
// =============================================================================

func TestLiuShou_String(t *testing.T) {
	tests := []struct {
		l    LiuShou
		want string
	}{
		{ShouQingLong, "青龙"},
		{ShouZhuQue, "朱雀"},
		{ShouGouChen, "勾陈"},
		{ShouTengShe, "螣蛇"},
		{ShouBaiHu, "白虎"},
		{ShouXuanWu, "玄武"},
		{LiuShou(-1), "?"},
		{LiuShou(6), "?"},
		{LiuShou(100), "?"},
	}
	for _, tt := range tests {
		got := tt.l.String()
		if got != tt.want {
			t.Errorf("LiuShou(%d).String() = %s, want %s", int(tt.l), got, tt.want)
		}
	}
}

// =============================================================================
// shakeCoins — 随机摇卦 (固定种子)
// =============================================================================

func TestShakeCoins_Deterministic(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	yaos := shakeCoins(rng)

	for i, y := range yaos {
		if y < 6 || y > 9 {
			t.Errorf("position %d: yao=%d, want 6-9", i+1, y)
		}
	}
	// Verify reproducibility.
	rng2 := rand.New(rand.NewSource(42))
	yaos2 := shakeCoins(rng2)
	if yaos != yaos2 {
		t.Error("shakeCoins not reproducible with same seed")
	}
}

// =============================================================================
// yongShenToLiuQin — 用神到六亲映射
// =============================================================================

func TestYongShenToLiuQin_All(t *testing.T) {
	tests := []struct {
		ys   YongShen
		want LiuQin
	}{
		{YongFumu, QinFumu},
		{YongXiongDi, QinXiongDi},
		{YongGuanGui, QinGuanGui},
		{YongQiCai, QinQiCai},
		{YongZiSun, QinZiSun},
		{YongShiYao, -1},
		{YongShen(100), -1},
	}
	for _, tt := range tests {
		got := yongShenToLiuQin(tt.ys)
		if got != tt.want {
			t.Errorf("yongShenToLiuQin(%s) = %s, want %s",
				tt.ys.String(), got.String(), tt.want.String())
		}
	}
}

// =============================================================================
// ordinal — 爻位名称
// =============================================================================

func TestOrdinal(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1, "初"}, {2, "二"}, {3, "三"}, {4, "四"}, {5, "五"}, {6, "上"},
		{0, "?"}, {7, "?"}, {-1, "?"}, {100, "?"},
	}
	for _, tt := range tests {
		got := ordinal(tt.n)
		if got != tt.want {
			t.Errorf("ordinal(%d) = %s, want %s", tt.n, got, tt.want)
		}
	}
}

// =============================================================================
// findShiYao — 找世爻
// =============================================================================

func TestFindShiYao_Found(t *testing.T) {
	p := &Chart{
		Lines: [6]Line{
			{Position: 1}, {Position: 2}, {Position: 3},
			{Position: 4, ShiYing: "世"}, {Position: 5}, {Position: 6},
		},
	}
	if got := p.findShiYao(); got != 4 {
		t.Errorf("findShiYao() = %d, want 4", got)
	}
}

func TestFindShiYao_NotFound(t *testing.T) {
	p := &Chart{
		Lines: [6]Line{
			{Position: 1}, {Position: 2}, {Position: 3},
			{Position: 4}, {Position: 5}, {Position: 6},
		},
	}
	if got := p.findShiYao(); got != 0 {
		t.Errorf("findShiYao() = %d, want 0", got)
	}
}

// =============================================================================
// findYongShen — 用神查找 (ben卦 + 变卦路径)
// =============================================================================

func TestFindYongShen_InBenGua(t *testing.T) {
	p := &Chart{
		Lines: [6]Line{
			{Position: 1, LiuQin: QinFumu},
			{Position: 2, LiuQin: QinQiCai},
			{Position: 3, LiuQin: QinGuanGui},
			{Position: 4, LiuQin: QinXiongDi},
			{Position: 5, LiuQin: QinZiSun},
			{Position: 6, LiuQin: QinFumu},
		},
	}
	if got, _ := p.findYongShen(YongQiCai); got != 2 {
		t.Errorf("findYongShen(YongQiCai) = %d, want 2", got)
	}
	// YongShiYao with no shiYao → findShiYao returns 0.
	if got, _ := p.findYongShen(YongShiYao); got != 0 {
		t.Errorf("findYongShen(YongShiYao) = %d, want 0", got)
	}
}

func TestFindYongShen_InBianGua(t *testing.T) {
	p := &Chart{
		Lines: [6]Line{
			{Position: 1, LiuQin: QinFumu},
			{Position: 2, LiuQin: QinFumu},
			{Position: 3, LiuQin: QinFumu},
			{Position: 4, LiuQin: QinFumu},
			{Position: 5, LiuQin: QinFumu},
			{Position: 6, LiuQin: QinFumu},
		},
		BianLines: [6]Line{
			{Position: 1, LiuQin: QinFumu},
			{Position: 2, LiuQin: QinFumu},
			{Position: 3, LiuQin: QinQiCai},
			{Position: 4, LiuQin: QinFumu},
			{Position: 5, LiuQin: QinFumu},
			{Position: 6, LiuQin: QinFumu},
		},
	}
	if got, isBian := p.findYongShen(YongQiCai); got != 3 || !isBian {
		t.Errorf("findYongShen(YongQiCai) in bianGua = (%d, %v), want (3, true)", got, isBian)
	}
}

func TestFindYongShen_NotFound(t *testing.T) {
	p := &Chart{
		Lines: [6]Line{
			{Position: 1, LiuQin: QinFumu},
			{Position: 2, LiuQin: QinFumu},
			{Position: 3, LiuQin: QinFumu},
			{Position: 4, LiuQin: QinFumu},
			{Position: 5, LiuQin: QinFumu},
			{Position: 6, LiuQin: QinFumu},
		},
	}
	if got, _ := p.findYongShen(YongQiCai); got != 0 {
		t.Errorf("findYongShen(YongQiCai) = %d, want 0", got)
	}
}

// =============================================================================
// findFuShen — 伏神查找
// =============================================================================

func TestFindFuShen_Found(t *testing.T) {
	// 天风姤 (乾宫, guaIndex=1, PalaceIdx=0)
	// 本宫卦 乾为天: 子水→子孙, 寅木→妻财, 辰土→父母, 午火→官鬼, 申金→兄弟, 戌土→父母
	// 搜索 父母 → 第一个匹配是 line 3 (辰)
	p := &Chart{BenGua: 1} // 天风姤, PalaceIdx=0
	fs := p.findFuShen(YongFumu)
	if fs == nil {
		t.Fatal("findFuShen returned nil")
	}
	if fs.Position != 3 {
		t.Errorf("Position = %d, want 3", fs.Position)
	}
	if fs.LiuQin != QinFumu {
		t.Errorf("LiuQin = %s, want 父母", fs.LiuQin.String())
	}
	if fs.Zhi != "辰" {
		t.Errorf("Zhi = %s, want 辰", fs.Zhi)
	}
}

func TestFindFuShen_DifferentTarget(t *testing.T) {
	// 天风姤 (乾宫), search for 官鬼 → base line 4 (午火, 火克金=官鬼)
	p := &Chart{BenGua: 1}
	fs := p.findFuShen(YongGuanGui)
	if fs == nil {
		t.Fatal("findFuShen returned nil")
	}
	if fs.Position != 4 {
		t.Errorf("Position = %d, want 4", fs.Position)
	}
	if fs.Zhi != "午" {
		t.Errorf("Zhi = %s, want 午", fs.Zhi)
	}
}

// =============================================================================
// computeYingQi — 应期推算 (完整路径)
// =============================================================================

func TestComputeYingQi_YongShenDongYao(t *testing.T) {
	// 乾为天, 三爻动(LaoYang=9) → 辰土父母为动爻.
	st := tianwen.SolarTime(time.Date(2000, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	chart := ComputeChart(st, YongFumu, [6]int{7, 7, 9, 7, 7, 7})

	yq := computeYingQi(&chart, YongFumu)
	if yq.YongShen != "父母" {
		t.Errorf("YongShen = %s, want 父母", yq.YongShen)
	}
	if yq.DongYaoPos != 3 {
		t.Errorf("DongYaoPos = %d, want 3 (三爻为动爻)", yq.DongYaoPos)
	}
	if yq.YingTime == "" {
		t.Error("expected YingTime not empty")
	}
	if yq.Assessment == "" {
		t.Error("expected Assessment not empty")
	}
}

func TestComputeYingQi_YongShenJingYao(t *testing.T) {
	// 用神在静爻上
	st := tianwen.SolarTime(time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	// All ShaoYang (=7, 静) → 乾为天, 世爻在6
	chart := ComputeChart(st, YongShiYao, [6]int{7, 7, 7, 7, 7, 7})

	yq := computeYingQi(&chart, YongShiYao)
	if yq.DongYaoPos != 0 {
		t.Errorf("expected DongYaoPos=0 for static yao, got %d", yq.DongYaoPos)
	}
	// 静爻待冲.
	if yq.YingTime != "" {
		t.Errorf("expected empty YingTime for static yao, got %s", yq.YingTime)
	}
}

func TestComputeYingQi_NotFound_WithFuShen(t *testing.T) {
	// 用神不上卦, 有伏神.
	// 构造一个全部爻位为父母的卦, 搜索妻财 → 乾宫本宫卦有妻财(寅)为伏神.
	p := &Chart{
		BenGua: 1, // 天风姤, PalaceIdx=0 (乾宫)
		Lines: [6]Line{
			{Position: 1, LiuQin: QinFumu, Zhi: ganzhi.ZhiZi},
			{Position: 2, LiuQin: QinFumu, Zhi: ganzhi.ZhiYin},
			{Position: 3, LiuQin: QinFumu, Zhi: ganzhi.ZhiChen},
			{Position: 4, LiuQin: QinFumu, Zhi: ganzhi.ZhiWu},
			{Position: 5, LiuQin: QinFumu, Zhi: ganzhi.ZhiShen},
			{Position: 6, LiuQin: QinFumu, Zhi: ganzhi.ZhiXu},
		},
		MonthZhi: ganzhi.ZhiYin,
		DayGan:   ganzhi.GanJia,
		DayZhi:   ganzhi.ZhiZi,
	}
	yq := computeYingQi(p, YongQiCai)
	if yq.YongShen != "妻财" {
		t.Errorf("YongShen = %s, want 妻财", yq.YongShen)
	}
	if yq.Assessment == "" {
		t.Error("expected Assessment not empty")
	}
}

// =============================================================================
// ComputeChart — 随机摇卦路径 (无 fixed yaos)
// =============================================================================

func TestComputeChart_RandomShake(t *testing.T) {
	st := tianwen.SolarTime(time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	chart := ComputeChart(st, YongQiCai, [6]int{})

	if chart.Name == "" {
		t.Error("Name is empty")
	}
	if chart.Palace == "" {
		t.Error("Palace is empty")
	}
	if len(chart.Lines) != 6 {
		t.Errorf("Lines len = %d, want 6", len(chart.Lines))
	}
	// 旺衰 and 日建关系 should all be set.
	for i := 0; i < 6; i++ {
		if chart.WangShuai[i].String() == "" {
			t.Errorf("line %d: empty wangshuai", i+1)
		}
		if chart.DayRelations[i].Relation == "" {
			t.Errorf("line %d: empty day relation", i+1)
		}
	}
	if chart.YongShen.Type != YongQiCai {
		t.Errorf("YongShen.Type = %s, want 妻财", chart.YongShen.Type.String())
	}
	if chart.YingQi.Assessment == "" {
		t.Error("YingQi.Assessment is empty")
	}
}

// =============================================================================
// ComputeChart — 变卦线类型验证
// =============================================================================

func TestComputeChart_BianLinesType(t *testing.T) {
	// 初爻 LaoYang (9) → 变出 ShaoYin (8), 二爻 LaoYin (6) → 变出 ShaoYang (7)
	st := tianwen.SolarTime(time.Date(2000, 3, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	chart := ComputeChart(st, YongFumu, [6]int{9, 6, 7, 7, 7, 7})

	if len(chart.DongYao) != 2 {
		t.Fatalf("expected 2 dongYao, got %d", len(chart.DongYao))
	}
	// 初爻 LaoYang → BianLine type should be ShaoYin.
	if chart.BianLines[0].Type != ShaoYin {
		t.Errorf("bian line 1: type=%d, want ShaoYin(8)", chart.BianLines[0].Type)
	}
	// 二爻 LaoYin → BianLine type should be ShaoYang.
	if chart.BianLines[1].Type != ShaoYang {
		t.Errorf("bian line 2: type=%d, want ShaoYang(7)", chart.BianLines[1].Type)
	}
	// 三爻 ShaoYang (未动) → 不变.
	if chart.BianLines[2].Type != ShaoYang {
		t.Errorf("bian line 3: type=%d, want ShaoYang(7)", chart.BianLines[2].Type)
	}
}

// =============================================================================
// YongShen.String — 用神名称
// =============================================================================

func TestYongShen_String(t *testing.T) {
	tests := []struct {
		y    YongShen
		want string
	}{
		{YongFumu, "父母"},
		{YongXiongDi, "兄弟"},
		{YongGuanGui, "官鬼"},
		{YongQiCai, "妻财"},
		{YongZiSun, "子孙"},
		{YongShiYao, "世爻"},
	}
	for _, tt := range tests {
		got := tt.y.String()
		if got != tt.want {
			t.Errorf("YongShen(%d).String() = %s, want %s", int(tt.y), got, tt.want)
		}
	}
}

// =============================================================================
// YaoType — IsYang / IsChanging
// =============================================================================

func TestYaoType_All(t *testing.T) {
	tests := []struct {
		yt         YaoType
		isYang     bool
		isChanging bool
	}{
		{LaoYin, false, true},
		{ShaoYang, true, false},
		{ShaoYin, false, false},
		{LaoYang, true, true},
	}
	for _, tt := range tests {
		if tt.yt.IsYang() != tt.isYang {
			t.Errorf("YaoType(%d).IsYang() = %v, want %v", tt.yt, tt.yt.IsYang(), tt.isYang)
		}
		if tt.yt.IsChanging() != tt.isChanging {
			t.Errorf("YaoType(%d).IsChanging() = %v, want %v", tt.yt, tt.yt.IsChanging(), tt.isChanging)
		}
	}
}

// =============================================================================
// WangShuai.String — 旺衰名称
// =============================================================================

func TestWangShuai_String(t *testing.T) {
	tests := []struct {
		ws   ganzhi.WangShuai
		want string
	}{
		{ganzhi.WSWang, "旺"}, {ganzhi.WSXiang, "相"}, {ganzhi.WSXiu, "休"}, {ganzhi.WSQiu, "囚"}, {ganzhi.WSSi, "死"},
	}
	for _, tt := range tests {
		if got := tt.ws.String(); got != tt.want {
			t.Errorf("WangShuai(%d).String() = %s, want %s", int(tt.ws), got, tt.want)
		}
	}
}

// TestComputeChart_ChangingLines checks变卦 with specific changing lines.
func TestComputeChart_ChangingLines(t *testing.T) {
	st := tianwen.SolarTime(time.Date(2000, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	// One changing line at position 2 (LaoYin=6).
	chart := ComputeChart(st, YongQiCai, [6]int{7, 6, 7, 7, 7, 7})

	if len(chart.DongYao) != 1 {
		t.Fatalf("expected 1 dongYao, got %d", len(chart.DongYao))
	}
	if chart.DongYao[0] != 2 {
		t.Errorf("dongYao position = %d, want 2", chart.DongYao[0])
	}
	// BenGua should be 乾为天 with one bit flipped.
	if chart.BenGua == chart.BianGua {
		t.Error("bianGua should differ from benGua")
	}
}

// TestDayGanShouOrder verifies六兽 order starts from correct 青龙 position.
func TestDayGanShouOrder(t *testing.T) {
	tests := []struct {
		dayGan ganzhi.Gan
		want0  LiuShou // expected first六兽
	}{
		{ganzhi.GanJia, ShouQingLong},
		{ganzhi.GanYi, ShouQingLong},
		{ganzhi.GanBing, ShouZhuQue},
		{ganzhi.GanDing, ShouZhuQue},
		{ganzhi.GanWu, ShouGouChen},
		{ganzhi.GanJi, ShouTengShe},
		{ganzhi.GanGeng, ShouBaiHu},
		{ganzhi.GanXin, ShouBaiHu},
		{ganzhi.GanRen, ShouXuanWu},
		{ganzhi.GanGui, ShouXuanWu},
	}

	for _, tt := range tests {
		t.Run(ganzhi.GanName(tt.dayGan), func(t *testing.T) {
			order := dayGanShouOrder(tt.dayGan)
			if order[0] != tt.want0 {
				t.Errorf("dayGanShouOrder(%s)[0] = %s, want %s",
					ganzhi.GanName(tt.dayGan), order[0].String(), tt.want0.String())
			}
		})
	}
}
