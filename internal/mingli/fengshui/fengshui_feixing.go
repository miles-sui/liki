package fengshui

import "github.com/25types/25types/internal/ganzhi"

// -- 紫白飞星 (Purple-White Flying Stars) -----------------------------------------

// FlyingStar holds a single star with its attributes.
type FlyingStar struct {
	Number     int            `json:"number"`
	Color      string         `json:"color"`
	Name       string         `json:"name"`
	Element    ganzhi.Element `json:"element"`
	Auspicious bool           `json:"auspicious"`
}

var starTable = [10]FlyingStar{
	{},
	{1, "白", "一白贪狼", ganzhi.ElemWater, true},
	{2, "黑", "二黑巨门", ganzhi.ElemEarth, false},
	{3, "碧", "三碧禄存", ganzhi.ElemWood, false},
	{4, "绿", "四绿文曲", ganzhi.ElemWood, true},
	{5, "黄", "五黄廉贞", ganzhi.ElemEarth, false},
	{6, "白", "六白武曲", ganzhi.ElemMetal, true},
	{7, "赤", "七赤破军", ganzhi.ElemMetal, false},
	{8, "白", "八白左辅", ganzhi.ElemEarth, true},
	{9, "紫", "九紫右弼", ganzhi.ElemFire, true},
}

func StarByNumber(n int) FlyingStar {
	if n >= 1 && n <= 9 {
		return starTable[n]
	}
	return FlyingStar{}
}

// Palace is one of the nine palaces in the 洛书 (Luoshu) grid.
type Palace struct {
	Number    int            `json:"number"`
	Name      string         `json:"name"`
	Direction string         `json:"direction"`
	Element   ganzhi.Element `json:"element"`
}

var palaceTable = [10]Palace{
	{},
	{1, "坎", "北", ganzhi.ElemWater},
	{2, "坤", "西南", ganzhi.ElemEarth},
	{3, "震", "东", ganzhi.ElemWood},
	{4, "巽", "东南", ganzhi.ElemWood},
	{5, "中", "中", ganzhi.ElemEarth},
	{6, "乾", "西北", ganzhi.ElemMetal},
	{7, "兑", "西", ganzhi.ElemMetal},
	{8, "艮", "东北", ganzhi.ElemEarth},
	{9, "离", "南", ganzhi.ElemFire},
}

func PalaceByNumber(n int) Palace {
	if n >= 1 && n <= 9 {
		return palaceTable[n]
	}
	return Palace{}
}

var luoshuFlyOrder = [8]int{6, 7, 8, 9, 1, 2, 3, 4}

type YearStarResult struct {
	Year       int           `json:"year"`
	CenterStar FlyingStar    `json:"center_star"`
	Palaces    [9]PalaceStar `json:"palaces"`
}

type PalaceStar struct {
	Palace `json:",inline"`
	Star   FlyingStar `json:"star"`
}

// ComputeYearStars computes the annual purple-white flying star distribution for a given year.
//
// Formula: 下元甲子(1984)七赤入中, each year the center star decreases by 1.
func ComputeYearStars(year int) YearStarResult {
	var jiaZiYear int
	var jiaZiStar int
	switch {
	case year >= 1984:
		jiaZiYear, jiaZiStar = 1984, 7
	case year >= 1924:
		jiaZiYear, jiaZiStar = 1924, 4
	case year >= 1864:
		jiaZiYear, jiaZiStar = 1864, 1
	default:
		cycle := (1864 - year) / 60
		if (1864-year)%60 != 0 {
			cycle++
		}
		jiaZiYear = 1864 - cycle*60
		phase := (jiaZiYear / 60) % 3
		jiaZiStar = []int{1, 4, 7}[phase%3]
	}

	diff := year - jiaZiYear
	centerNum := (jiaZiStar - diff%9 + 9) % 9
	if centerNum == 0 {
		centerNum = 9
	}

	centerStar := StarByNumber(centerNum)

	var palaces [9]PalaceStar
	palaces[4] = PalaceStar{Palace: PalaceByNumber(5), Star: centerStar}

	for i, pn := range luoshuFlyOrder {
		starNum := (centerNum + i + 1) % 9
		if starNum == 0 {
			starNum = 9
		}
		palaces[pn-1] = PalaceStar{Palace: PalaceByNumber(pn), Star: StarByNumber(starNum)}
	}

	return YearStarResult{Year: year, CenterStar: centerStar, Palaces: palaces}
}

// DirectionStar returns the annual star for a given palace number (1-9).
func (r YearStarResult) DirectionStar(palaceNum int) PalaceStar {
	if palaceNum >= 1 && palaceNum <= 9 {
		return r.Palaces[palaceNum-1]
	}
	return PalaceStar{}
}

// -- 八宅 四吉四凶 ------------------------------------------

// EightMansionDirs returns the 八宅 four auspicious and four inauspicious directions
// for a given minggua (大游年 formula). Returns palace numbers 1-9.
func EightMansionDirs(guaNum int) (auspicious [4]int, inauspicious [4]int) {
	patterns := map[int]struct {
		shengQi, tianYi, yanNian, fuWei int
		huoHai, wuGui, liuSha, jueMing  int
	}{
		1: {6, 8, 9, 1, 2, 3, 7, 4},
		2: {9, 1, 6, 2, 8, 7, 4, 3},
		3: {8, 6, 1, 3, 7, 2, 9, 4},
		4: {1, 9, 8, 4, 5, 6, 3, 2},
		6: {4, 3, 2, 6, 8, 7, 9, 1},
		7: {3, 4, 6, 7, 8, 9, 1, 2},
		8: {2, 7, 3, 8, 9, 1, 5, 6},
		9: {8, 3, 4, 9, 6, 5, 2, 1},
	}

	p, ok := patterns[guaNum]
	if !ok {
		return
	}
	auspicious = [4]int{p.shengQi, p.tianYi, p.yanNian, p.fuWei}
	inauspicious = [4]int{p.huoHai, p.wuGui, p.liuSha, p.jueMing}
	return
}

func palaceDirs() [10]string {
	return [10]string{"", "北", "西南", "东", "东南", "中", "西北", "西", "东北", "南"}
}

// BaZhaiDirections holds the 八宅 four-auspicious-four-inauspicious directions by name.
type BaZhaiDirections struct {
	ShengQi []string `json:"sheng_qi"`
	TianYi  []string `json:"tian_yi"`
	YanNian []string `json:"yan_nian"`
	FuWei   []string `json:"fu_wei"`
	HuoHai  []string `json:"huo_hai"`
	WuGui   []string `json:"wu_gui"`
	LiuSha  []string `json:"liu_sha"`
	JueMing []string `json:"jue_ming"`
}

// BaZhaiDirectionsForGua returns the eight mansion directions as named strings.
func BaZhaiDirectionsForGua(guaNum int) BaZhaiDirections {
	aus, inaus := EightMansionDirs(guaNum)
	dirs := palaceDirs()
	return BaZhaiDirections{
		ShengQi: []string{dirs[aus[0]]},
		TianYi:  []string{dirs[aus[1]]},
		YanNian: []string{dirs[aus[2]]},
		FuWei:   []string{dirs[aus[3]]},
		HuoHai:  []string{dirs[inaus[0]]},
		WuGui:   []string{dirs[inaus[1]]},
		LiuSha:  []string{dirs[inaus[2]]},
		JueMing: []string{dirs[inaus[3]]},
	}
}
