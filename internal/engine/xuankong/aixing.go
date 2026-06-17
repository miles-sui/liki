package xuankong

import (
	"liki/internal/engine/fengshui"
	"liki/internal/engine/tianwen"
)

// xuanKongStar holds the three stars (运星, 山星, 向星) for one palace.
type xuanKongStar struct {
	PalaceNum    int                  `json:"palace_num"`
	PeriodStar   fengshui.FlyingStar  `json:"period_star"`
	MountainStar fengshui.FlyingStar  `json:"mountain_star"`
	FacingStar   fengshui.FlyingStar  `json:"facing_star"`
}

// Chart is the complete 玄空飞星排盘 for a given坐向 and 运.
type Chart struct {
	Yun           SanYuanYun     `json:"yun"`
	SitMountain   int            `json:"sit_mountain"`  // 0-23,坐山 index
	FaceMountain  int            `json:"face_mountain"` // 0-23,朝向 index
	Palaces       [9]xuanKongStar `json:"palaces"`
	WangShan      bool           `json:"wang_shan"`
	WangXiang     bool           `json:"wang_xiang"`
	ShanXing      bool           `json:"shan_xing"`
	XiaShui       bool           `json:"xia_shui"`
	FanYin        bool           `json:"fan_yin"`
	FuYin         bool           `json:"fu_yin"`
	XingJiaHui    [9]xingJiaHui  `json:"xing_jia_hui"`
	ShouShanChuSha shouShanChuSha `json:"shou_shan_chu_sha"`
}

// ComputeChart computes the 玄空飞星盘 for a given坐向.
func ComputeChart(st tianwen.SolarTime, sitMountain, faceMountain int) Chart {
	return computeChart(sitMountain, faceMountain, st.Time().Year())
}

func computeChart(sitMountain, faceMountain int, year int) Chart {
	if sitMountain < 0 || sitMountain > 23 || faceMountain < 0 || faceMountain > 23 {
		return Chart{}
	}

	sit := fengshui.Mountains24Table[sitMountain]
	face := fengshui.Mountains24Table[faceMountain]

	yun := ComputeSanYuanYun(year)
	yunNum := yun.YunNumber

	// 1. Period star distribution.
	periodStars := flyStars(yunNum, true)

	// 2. Mountain star.
	sitPalace := mountainPalace(sitMountain)
	sitPeriodStar := periodStars[sitPalace-1]
	shanNum := sitPeriodStar.Number

	if ti := tiXingShanStar(sitMountain, shanNum); ti > 0 {
		shanNum = ti
	}
	shanForward := sit.YinYang == "阳"
	mountainStars := flyStars(shanNum, shanForward)

	// 3. Facing star.
	facePalace := mountainPalace(faceMountain)
	facePeriodStar := periodStars[facePalace-1]
	xiangNum := facePeriodStar.Number

	if ti := tiXingXiangStar(faceMountain, xiangNum); ti > 0 {
		xiangNum = ti
	}
	xiangForward := face.YinYang == "阳"
	facingStars := flyStars(xiangNum, xiangForward)

	// 4. Assemble the pan.
	pan := Chart{
		Yun:          yun,
		SitMountain:  sitMountain,
		FaceMountain: faceMountain,
	}

	for i := 0; i < 9; i++ {
		pan.Palaces[i] = xuanKongStar{
			PalaceNum:    i + 1,
			PeriodStar:   periodStars[i],
			MountainStar: mountainStars[i],
			FacingStar:   facingStars[i],
		}
	}

	// 5. Evaluate 旺山旺向/上山下水/反吟伏吟.
	pan.evaluate()

	// 6. 双星加会 + 收山出煞.
	pan.XingJiaHui = pan.computeXingJiaHui()
	pan.ShouShanChuSha = pan.computeShouShanChuSha()

	return pan
}

// flyStars distributes num stars following luoshu fly order.
func flyStars(centerNum int, forward bool) [9]fengshui.FlyingStar {
	var stars [9]fengshui.FlyingStar
	stars[4] = fengshui.StarByNumber(centerNum)

	for i, pn := range fengshui.LuoshuFlyOrder {
		var starNum int
		if forward {
			starNum = (centerNum + i + 1) % 9
		} else {
			starNum = (centerNum - i - 1 + 9) % 9
		}
		if starNum == 0 {
			starNum = 9
		}
		stars[pn-1] = fengshui.StarByNumber(starNum)
	}
	return stars
}

// mountainPalace returns which palace (1-9) a given 24-mountain index belongs to.
func mountainPalace(idx int) int {
	idx = idx % 24
	switch {
	case idx <= 1 || idx == 23:
		return 1 // 坎(子癸壬)
	case idx >= 2 && idx <= 4:
		return 8 // 艮(丑艮寅)
	case idx >= 5 && idx <= 7:
		return 3 // 震(甲卯乙)
	case idx >= 8 && idx <= 10:
		return 4 // 巽(辰巽巳)
	case idx >= 11 && idx <= 13:
		return 9 // 离(丙午丁)
	case idx >= 14 && idx <= 16:
		return 2 // 坤(未坤申)
	case idx >= 17 && idx <= 19:
		return 7 // 兑(庚酉辛)
	default: // 20-22
		return 6 // 乾(戌乾亥)
	}
}

func (p *Chart) evaluate() {
	sitPalace := mountainPalace(p.SitMountain)
	facePalace := mountainPalace(p.FaceMountain)
	yunNum := p.Yun.YunNumber

	sitMStar := p.Palaces[sitPalace-1].MountainStar.Number
	faceFStar := p.Palaces[facePalace-1].FacingStar.Number

	p.WangShan = sitMStar == yunNum
	p.WangXiang = faceFStar == yunNum

	sitPStar := p.Palaces[sitPalace-1].PeriodStar.Number
	facePStar := p.Palaces[facePalace-1].PeriodStar.Number
	p.ShanXing = sitPStar != yunNum && sitMStar != yunNum
	p.XiaShui = facePStar != yunNum && faceFStar != yunNum

	opposite := (sitPalace+4)%9 + 1
	p.FanYin = p.Palaces[opposite-1].PeriodStar.Number == yunNum

	for i := 0; i < 9; i++ {
		if p.Palaces[i].PeriodStar.Number == i+1 {
			p.FuYin = true
			break
		}
	}
}

// tiXingTable maps 24 mountain index → substitute star number (1-9).
var tiXingTable = [24]int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, //  子癸丑艮寅甲卯乙辰 (0-8)
	1, 2, 3, 4, 5, 6, 7, 8, 9, //  巽巳丙午丁未坤申庚 (9-17)
	1, 2, 3, 4, 5, 6, //         酉辛戌乾亥壬 (18-23)
}

func tiXingShanStar(sitIdx int, periodStarNum int) int {
	num := tiXingTable[sitIdx%24]
	sit := fengshui.Mountains24Table[sitIdx%24]
	if sit.YuanLong != "天元龙" {
		return num
	}
	return 0
}

func tiXingXiangStar(faceIdx int, periodStarNum int) int {
	return tiXingTable[faceIdx%24]
}

// -- 双星加会 (Double Star Combination) --------------------------------

type xingJiaHui struct {
	ShanNum    int    `json:"shan_num"`
	XiangNum   int    `json:"xiang_num"`
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Auspicious bool   `json:"auspicious"`
}

var xingJiaHuiTable = map[[2]int]xingJiaHui{
	{1, 4}: {1, 4, "一四同宫", "准发科名之显", true},
	{4, 1}: {4, 1, "四一同宫", "准发科名之显", true},
	{1, 6}: {1, 6, "一六共宗", "启八代之文章，官贵", true},
	{6, 1}: {6, 1, "六一联星", "文武双全，名利双收", true},
	{2, 7}: {2, 7, "二七同道", "火煞重重，定见火灾", false},
	{7, 2}: {7, 2, "七二同道", "火煞重重，定见火灾", false},
	{3, 8}: {3, 8, "三八为朋", "旺丁旺财，家业兴隆", true},
	{8, 3}: {8, 3, "八三为朋", "旺丁旺财，家业兴隆", true},
	{4, 9}: {4, 9, "四九为友", "富贵荣华，文笔生辉", true},
	{9, 4}: {9, 4, "九四为友", "富贵荣华，文笔生辉", true},
	{6, 8}: {6, 8, "六八同宫", "武科发迹，名利双收", true},
	{8, 6}: {8, 6, "八六同宫", "武科发迹，名利双收", true},
	{5, 5}: {5, 5, "五黄同宫", "大凶之象，损丁破财", false},
	{9, 1}: {9, 1, "九一同宫", "水火既济，文采风流", true},
	{1, 9}: {1, 9, "一九同宫", "水火既济，文采风流", true},
	{2, 5}: {2, 5, "二五交加", "损主重病，多灾多难", false},
	{5, 2}: {5, 2, "五二交加", "损主重病，多灾多难", false},
	{5, 9}: {5, 9, "九五交加", "紫黄毒药，犯之损人", false},
	{9, 5}: {9, 5, "五九交加", "紫黄毒药，犯之损人", false},
	{3, 7}: {3, 7, "三七穿心", "劫盗官非，破财损丁", false},
	{7, 3}: {7, 3, "七三穿心", "劫盗官非，破财损丁", false},
	{2, 3}: {2, 3, "二三斗牛", "官非口舌，夫妻反目", false},
	{3, 2}: {3, 2, "三二斗牛", "官非口舌，夫妻反目", false},
	{2, 9}: {2, 9, "二九同宫", "火土相生，旺丁旺财", true},
	{9, 2}: {9, 2, "九二同宫", "火土相生，旺丁旺财", true},
	{4, 7}: {4, 7, "四七同宫", "金木相战，刀伤之厄", false},
	{7, 4}: {7, 4, "七四同宫", "金木相战，刀伤之厄", false},
	{6, 9}: {6, 9, "六九同宫", "火照天门，丁财两败", false},
	{9, 6}: {9, 6, "九六同宫", "火照天门，丁财两败", false},
}

func (p *Chart) computeXingJiaHui() [9]xingJiaHui {
	var result [9]xingJiaHui
	for i, pal := range p.Palaces {
		key := [2]int{pal.MountainStar.Number, pal.FacingStar.Number}
		if entry, ok := xingJiaHuiTable[key]; ok {
			result[i] = entry
		} else {
			result[i] = xingJiaHui{
				ShanNum:    pal.MountainStar.Number,
				XiangNum:   pal.FacingStar.Number,
				Name:       "双星到向",
				Meaning:    "山向配合，需参合判断",
				Auspicious: pal.MountainStar.Auspicious && pal.FacingStar.Auspicious,
			}
		}
	}
	return result
}

// -- 收山出煞 (Mountain Containment & Sha Removal) --------------------

type shouShanChuSha struct {
	ZhengShen  int    `json:"zheng_shen"`
	LingShen   int    `json:"ling_shen"`
	ShouShanOK bool   `json:"shou_shan"`
	ChuShaOK   bool   `json:"chu_sha"`
	Assessment string `json:"assessment"`
}

func (p *Chart) computeShouShanChuSha() shouShanChuSha {
	yunNum := p.Yun.YunNumber
	zhengShen := yunNum
	lingShen := (yunNum + 4) % 9
	if lingShen == 0 {
		lingShen = 9
	}

	sitPalace := mountainPalace(p.SitMountain)
	facePalace := mountainPalace(p.FaceMountain)

	sitMStar := p.Palaces[sitPalace-1].MountainStar.Number
	faceFStar := p.Palaces[facePalace-1].FacingStar.Number

	shouShanOK := sitMStar == zhengShen
	chuShaOK := faceFStar == lingShen

	var assessment string
	if shouShanOK && chuShaOK {
		assessment = "收山出煞俱得，丁财两旺"
	} else if shouShanOK {
		assessment = "收山得宜，旺丁；出煞未得，财弱"
	} else if chuShaOK {
		assessment = "出煞得宜，旺财；收山未得，丁弱"
	} else {
		assessment = "收山出煞俱失，宜择时改向"
	}

	return shouShanChuSha{
		ZhengShen:  zhengShen,
		LingShen:   lingShen,
		ShouShanOK: shouShanOK,
		ChuShaOK:   chuShaOK,
		Assessment: assessment,
	}
}
