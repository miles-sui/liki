package qiming

// SanCai holds the three-talent (三才) analysis.
type SanCai struct {
	Configuration string `json:"configuration"`
	Fortune       string `json:"fortune"`
	Description   string `json:"description"`
}

func FortuneYAMLToChinese(f string) string {
	switch f {
	case "ji":
		return "吉"
	case "da_ji":
		return "吉"
	case "ban_ji":
		return "半吉"
	case "xiong":
		return "凶"
	}
	return f
}

// ComputeSanCai returns the three-talent analysis.
func ComputeSanCai(tianElem, renElem, diElem string) SanCai {
	e := defaultEngine
	key := tianElem + renElem + diElem
	if e != nil {
		if v, ok := e.SanCaiCfg[key]; ok {
			return SanCai{Configuration: key, Fortune: v.Fortune, Description: v.Desc}
		}
	}
	return SanCai{Configuration: key, Fortune: "半吉", Description: "三才配置中等，无大吉亦无大凶。"}
}
