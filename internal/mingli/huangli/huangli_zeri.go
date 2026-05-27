package huangli

// -- 择日体系：喜神/财神/福神/彭祖百忌 -----------------------------------------------
// These belong to the huangli day-selection (择日) domain, NOT bazi fortune-telling.
// They are pure stem→value lookups — no analysis, no scoring.

// -- 喜神方位 (joy god direction by day stem) ---------------------------------------

var xiShenDirection = [11]string{
	"", "东北", "西北", "西南", "正南", "东南", // 甲己→东北, 乙庚→西北, 丙辛→西南, 丁壬→南, 戊癸→东南
	"东北", "西北", "西南", "正南", "东南",
}

// XiShenDirection returns the 喜神方位 for a given day stem.
func XiShenDirection(stem Stem) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return xiShenDirection[stem]
	}
	return ""
}

// -- 财神方位 (wealth god direction by day stem) ------------------------------------

var caiShenDirection = [11]string{
	"", "东北", "东北", "正西", "正西", "正北", // 甲, 乙, 丙, 丁, 戊
	"正北", "正东", "正东", "正南", "正南", // 己, 庚, 辛, 壬, 癸
}

// CaiShenDirection returns the 财神方位 for a given day stem.
func CaiShenDirection(stem Stem) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return caiShenDirection[stem]
	}
	return ""
}

// -- 福神方位 (blessing god direction by day stem) ----------------------------------

var fuShenDirection = [11]string{
	"", "东南", "东南", "西北", "正东", "正南", // 甲, 乙, 丙, 丁, 戊
	"正南", "西南", "西南", "西北", "正西", // 己, 庚, 辛, 壬, 癸
}

// FuShenDirection returns the 福神方位 for a given day stem.
func FuShenDirection(stem Stem) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return fuShenDirection[stem]
	}
	return ""
}

// -- 彭祖百忌 (Peng Zu daily taboos by stem and branch) -----------------------------

var pengZuStemTaboo = [11]string{
	"", "甲不开仓财物耗散", "乙不栽植千株不长",
	"丙不修灶必见灾殃", "丁不剃头头必生疮",
	"戊不受田田主不祥", "己不破券二比并亡",
	"庚不经络织机虚张", "辛不合酱主人不尝",
	"壬不汲水更难提防", "癸不词讼理弱敌强",
}

var pengZuBranchTaboo = [13]string{
	"", "子不问卜自惹祸殃", "丑不冠带主不还乡",
	"寅不祭祀神鬼不尝", "卯不穿井水泉不香",
	"辰不哭泣必主重丧", "巳不远行财物伏藏",
	"午不苫盖屋主更张", "未不服药毒气入肠",
	"申不安床鬼祟入房", "酉不会客醉坐颠狂",
	"戌不吃犬作怪上床", "亥不嫁娶不利新郎",
}

// PengZuStemTaboo returns the Peng Zu taboo for a given day stem.
func PengZuStemTaboo(stem Stem) string {
	if int(stem) >= 1 && int(stem) <= 10 {
		return pengZuStemTaboo[stem]
	}
	return ""
}

// PengZuBranchTaboo returns the Peng Zu taboo for a given day branch.
func PengZuBranchTaboo(branch Branch) string {
	if int(branch) >= 1 && int(branch) <= 12 {
		return pengZuBranchTaboo[branch]
	}
	return ""
}

// -- 汇总 ---------------------------------------------------------------------------

// DayStemDirections holds all three direction lookups for a day stem.
type DayStemDirections struct {
	DayStem  Stem   `json:"day_stem"`
	XiShen   string `json:"xi_shen"`   // 喜神方位
	CaiShen  string `json:"cai_shen"`  // 财神方位
	FuShen   string `json:"fu_shen"`   // 福神方位
}

// DayTaboos holds the Peng Zu taboos for a day (stem + branch).
type DayTaboos struct {
	DayStem     Stem   `json:"day_stem"`
	DayBranch   Branch `json:"day_branch"`
	StemTaboo   string `json:"stem_taboo"`
	BranchTaboo string `json:"branch_taboo"`
}

// ComputeDayDirections returns the three direction lookups for a given day stem.
func ComputeDayDirections(stem Stem) DayStemDirections {
	return DayStemDirections{
		DayStem:  stem,
		XiShen:   XiShenDirection(stem),
		CaiShen:  CaiShenDirection(stem),
		FuShen:   FuShenDirection(stem),
	}
}

// ComputeDayTaboos returns the Peng Zu taboos for a given day pillar.
func ComputeDayTaboos(dayPillar Pillar) DayTaboos {
	return DayTaboos{
		DayStem:     dayPillar.Stem,
		DayBranch:   dayPillar.Branch,
		StemTaboo:   PengZuStemTaboo(dayPillar.Stem),
		BranchTaboo: PengZuBranchTaboo(dayPillar.Branch),
	}
}

// -- 黄道黑道十二神 (Yellow/Black Path 12 Day Stars) --------------------------------
// Determined by month branch (青龙 start) + day branch offset.
// 黄道 = auspicious (6 stars), 黑道 = inauspicious (6 stars).

// HuangDaoStar holds one of the 12 yellow/black path stars.
type HuangDaoStar struct {
	Index    int    `json:"index"`    // 0-11
	Name     string `json:"name"`     // e.g. "青龙"
	Path     string `json:"path"`     // "黄道" or "黑道"
	Sequence int    `json:"sequence"` // position in the 12-star cycle (0=青龙)
}

var huangDaoStars = [12]HuangDaoStar{
	{0, "青龙", "黄道", 0},
	{1, "明堂", "黄道", 1},
	{2, "天刑", "黑道", 2},
	{3, "朱雀", "黑道", 3},
	{4, "金匮", "黄道", 4},
	{5, "天德", "黄道", 5},
	{6, "白虎", "黑道", 6},
	{7, "玉堂", "黄道", 7},
	{8, "天牢", "黑道", 8},
	{9, "玄武", "黑道", 9},
	{10, "司命", "黄道", 10},
	{11, "勾陈", "黑道", 11},
}

// qingLongStart maps month branch to the branch where 青龙 starts.
var qingLongStart = map[int]int{
	3: 1, 9: 1,  // 寅申→子
	4: 3, 10: 3, // 卯酉→寅
	5: 5, 11: 5, // 辰戌→辰
	6: 7, 12: 7, // 巳亥→午
	7: 9, 1: 9,  // 午子→申
	8: 11, 2: 11, // 未丑→戌
}

// HuangDaoForDay returns the yellow/black path star for a given month branch and day branch.
func HuangDaoForDay(monthBranch, dayBranch Branch) HuangDaoStar {
	start, ok := qingLongStart[int(monthBranch)]
	if !ok {
		return HuangDaoStar{}
	}
	offset := (int(dayBranch) - start + 12) % 12
	return huangDaoStars[offset]
}

// AllHuangDaoStars returns the full 12-star table.
func AllHuangDaoStars() [12]HuangDaoStar {
	return huangDaoStars
}
