package qiming

// SanCai holds the three-talent (三才) analysis.
type SanCai struct {
	Configuration string `json:"configuration"`
	Fortune       string `json:"fortune"`
	Description   string `json:"description"`
}

func fortuneYAMLToChinese(f string) string {
	switch f {
	case "ji":
		return "吉"
	case "da_ji":
		return "大吉"
	case "ban_ji":
		return "半吉"
	case "xiong":
		return "凶"
	}
	return f
}

// computeSanCai returns the three-talent analysis.
func computeSanCai(tianElem, renElem, diElem string) SanCai {
	key := tianElem + renElem + diElem
	if v, ok := sanCaiCfg[key]; ok {
		return SanCai{Configuration: key, Fortune: v.Fortune, Description: v.Desc}
	}
	return SanCai{Configuration: key, Fortune: "半吉", Description: "三才配置中等，无大吉亦无大凶。"}
}
