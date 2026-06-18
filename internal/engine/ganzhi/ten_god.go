package ganzhi

import (
	"encoding/json"
	"fmt"
)

// TenGod classifies the ten-god relationship (十神).
type TenGod int

const (
	TenGodBiJian    TenGod = iota // 比肩
	TenGodJieCai                  // 劫财
	TenGodShiShen                 // 食神
	TenGodShangGuan               // 伤官
	TenGodPianCai                 // 偏财
	TenGodZhengCai                // 正财
	TenGodQiSha                   // 七杀
	TenGodZhengGuan               // 正官
	TenGodPianYin                 // 偏印
	TenGodZhengYin                // 正印
)

var tenGodNamesZH = [10]string{
	"比肩", "劫财", "食神", "伤官", "偏财",
	"正财", "七杀", "正官", "偏印", "正印",
}

func (tg TenGod) String() string {
	if tg >= 0 && int(tg) < len(tenGodNamesZH) {
		return tenGodNamesZH[tg]
	}
	return ""
}

func (tg TenGod) MarshalJSON() ([]byte, error) {
	return json.Marshal(tg.String())
}

func (tg *TenGod) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		parsed, err := ParseTenGod(name)
		if err != nil {
			return &json.UnmarshalTypeError{Value: "string", Type: nil, Field: name}
		}
		*tg = parsed
		return nil
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*tg = TenGod(i)
		return nil
	}
	return &json.UnmarshalTypeError{Value: string(data), Type: nil}
}

// ParseTenGod converts a Chinese ten-god name to a TenGod value.
func ParseTenGod(s string) (TenGod, error) {
	for i, name := range tenGodNamesZH {
		if name == s {
			return TenGod(i), nil
		}
	}
	return -1, fmt.Errorf("unknown tengod: %q", s)
}

// TenGodName returns the Chinese name for a ten god type.
func TenGodName(tg TenGod) string { return tg.String() }

// TenGodFromGan returns the TenGod for another stem relative to the day master.
func TenGodFromGan(dayMaster, other Gan) TenGod {
	dmElem := GanWuxing(dayMaster)
	otherElem := GanWuxing(other)
	dmYY := GanYinYang(dayMaster)
	otherYY := GanYinYang(other)
	return TenGodType(dmElem, dmYY, otherElem, otherYY)
}

// TenGodType classifies the ten-god relationship between day master and another stem.
func TenGodType(dmElem Wuxing, dmYY YinYang, otherElem Wuxing, otherYY YinYang) TenGod {
	switch {
	case dmElem == otherElem:
		if dmYY == otherYY {
			return TenGodBiJian
		}
		return TenGodJieCai
	case Sheng(dmElem, otherElem):
		if dmYY == otherYY {
			return TenGodShiShen
		}
		return TenGodShangGuan
	case Sheng(otherElem, dmElem):
		if dmYY == otherYY {
			return TenGodPianYin
		}
		return TenGodZhengYin
	case Ke(dmElem, otherElem):
		if dmYY == otherYY {
			return TenGodPianCai
		}
		return TenGodZhengCai
	default:
		if dmYY == otherYY {
			return TenGodQiSha
		}
		return TenGodZhengGuan
	}
}
