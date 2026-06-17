package liuyao

import "liki/internal/engine/ganzhi"

// WangShuai classifies 旺衰 level.
type WangShuai int

const (
	WSWang  WangShuai = iota // 旺
	WSXiang                   // 相
	WSXiu                     // 休
	WSQiu                     // 囚
	WSSi                      // 死
)

var wangShuaiNames = [5]string{"旺", "相", "休", "囚", "死"}

func (w WangShuai) String() string { return wangShuaiNames[w] }

func monthWangShuai(lineZhi ganzhi.Zhi, monthZhi ganzhi.Zhi) WangShuai {
	le := int(ganzhi.ZhiWuxing(lineZhi))
	me := int(ganzhi.ZhiWuxing(monthZhi))

	if le == (me%5)+1 { // month generates line → 旺
		return WSWang
	}
	if me == (le%5)+1 { // line generates month → 休
		return WSXiu
	}
	if le == me { // same → 相
		return WSXiang
	}
	if le == ((me+1)%5)+1 { // line overcomes month → 囚
		return WSQiu
	}
	return WSSi // month overcomes line → 死
}

// DayRelation describes 日建与爻的关系.
type DayRelation struct {
	Relation string `json:"relation"` // 生/扶/克/冲/合
	Strength string `json:"strength"` // 旺/平/衰
}

func dayInteraction(lineZhi ganzhi.Zhi, dayZhi ganzhi.Zhi) DayRelation {
	li, di := int(lineZhi), int(dayZhi)
	le, de := int(ganzhi.ZhiWuxing(lineZhi)), int(ganzhi.ZhiWuxing(dayZhi))

	rel := DayRelation{}

	// 冲: 地支六冲 (子午, 丑未, 寅申, 卯酉, 辰戌, 巳亥)
	if (li+6)%12 == di%12 || (di+6)%12 == li%12 {
		rel.Relation = "冲"
		rel.Strength = "衰"
		return rel
	}

	// 合: 地支六合 (子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未)
	hePairs := [][2]int{{1,2}, {3,12}, {4,11}, {5,10}, {6,9}, {7,8}}
	for _, p := range hePairs {
		if (li == p[0] && di == p[1]) || (li == p[1] && di == p[0]) {
			rel.Relation = "合"
			rel.Strength = "旺"
			return rel
		}
	}

	// 生: day generates line
	if sheng(de, le) {
		rel.Relation = "生"
		rel.Strength = "旺"
		return rel
	}

	// 扶: same element
	if le == de {
		rel.Relation = "扶"
		rel.Strength = "旺"
		return rel
	}

	// 克: day overcomes line
	if ke(de, le) {
		rel.Relation = "克"
		rel.Strength = "衰"
		return rel
	}

	rel.Relation = "平"
	rel.Strength = "平"
	return rel
}

// YingQi holds 应期 prediction.
type YingQi struct {
	YongShen    string `json:"yong_shen"`
	DongYaoPos  int    `json:"dong_yao_pos"` // 动爻位置
	YingTime    string `json:"ying_time"`    // 应期描述
	Assessment  string `json:"assessment"`   // 综合判断
}

func computeYingQi(p *Chart, typ YongShen) YingQi {

	yq := YingQi{
		YongShen: typ.String(),
	}

	yongPos := p.findYongShen(typ)
	if yongPos == 0 {
		// 用神不上卦，找伏神.
		fs := p.findFuShen(typ)
		if fs != nil {
			yq.Assessment = typ.String() + "不上卦，伏于" + ganzhi.ZhiName(p.Lines[fs.Position-1].Zhi) + "之下，待" + fs.Zhi + "年月冲出为应"
			return yq
		}
		yq.Assessment = typ.String() + "不上卦，问事不吉"
		return yq
	}

	// Check if the用神 line is a动爻.
	yao := p.Lines[yongPos-1]
	if yao.Type.IsChanging() {
		yq.DongYaoPos = yongPos
		yq.YingTime = "动爻临值之时（" + ganzhi.ZhiName(yao.Zhi) + "年月）为应"
	}

	// Month旺衰.
	ws := monthWangShuai(yao.Zhi, p.MonthZhi)

	// Day interaction.
	di := dayInteraction(yao.Zhi, p.DayZhi)

	yq.Assessment = typ.String() + "在" + ordinal(yongPos) + "爻" +
		"，月建" + ws.String() + "，日建" + di.Relation + "(" + di.Strength + ")"
	if yq.DongYaoPos > 0 {
		yq.Assessment += "，" + yq.YingTime
	} else {
		yq.Assessment += "，静爻待冲。冲" + ganzhi.ZhiName(chongZhi(yao.Zhi)) + "之时为应"
	}

	return yq
}

func ordinal(n int) string {
	names := [7]string{"", "初", "二", "三", "四", "五", "上"}
	if n >= 1 && n <= 6 { return names[n] }
	return "?"
}

func chongZhi(z ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.Zhi((int(z) + 6) % 12)
}

