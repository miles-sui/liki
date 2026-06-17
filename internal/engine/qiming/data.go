package qiming

import (
	"liki/internal/engine/ganzhi"
	"bytes"
	_ "embed"
	"encoding/csv"
	"log"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed data/gsc_pinyin_with_tone.csv
var gscCSV []byte

//go:embed data/sancai_numbers.yaml
var sancaiNumbersYAML []byte

//go:embed data/sancai_configs.yaml
var sancaiConfigsYAML []byte

var charByElement = make(map[Wuxing][]Character)
var charByRune = make(map[rune]Character)
var sanCaiNums map[int]sanCaiNum
var sanCaiCfg map[string]sanCaiCfgEntry

type sanCaiNum struct {
	Element string
	Fortune string
	Desc    string
}

type sanCaiCfgEntry struct {
	Fortune string
	Desc    string
}

func init() {
	if err := loadNaming(); err != nil {
		log.Fatalf("qiming: load qiming data: %v", err)
	}
}

// radicalToElement maps Kangxi radicals to five elements per docs/naming.md §3.7.
var radicalToElement = map[string]Wuxing{
	// 木
	"木": ganzhi.WxMu, "艹": ganzhi.WxMu, "林": ganzhi.WxMu,
	"竹": ganzhi.WxMu, "禾": ganzhi.WxMu, "米": ganzhi.WxMu, "桑": ganzhi.WxMu,
	"舟": ganzhi.WxMu, "羽": ganzhi.WxMu, "纟": ganzhi.WxMu,
	"弓": ganzhi.WxMu, "户": ganzhi.WxMu, "门": ganzhi.WxMu,
	"巾": ganzhi.WxMu, "虍": ganzhi.WxMu, "鹿": ganzhi.WxMu,
	"生": ganzhi.WxMu, "角": ganzhi.WxMu, "弋": ganzhi.WxMu,
	"龠": ganzhi.WxMu, "乙": ganzhi.WxMu, "麦": ganzhi.WxMu,
	"谷": ganzhi.WxMu, "青": ganzhi.WxMu, "耒": ganzhi.WxMu,
	"⺮": ganzhi.WxMu, "衤": ganzhi.WxMu, "衣": ganzhi.WxMu,
	// 火
	"火": ganzhi.WxHuo, "日": ganzhi.WxHuo, "灬": ganzhi.WxHuo,
	"心": ganzhi.WxHuo, "忄": ganzhi.WxHuo, "目": ganzhi.WxHuo, "离": ganzhi.WxHuo,
	"丙": ganzhi.WxHuo, "丁": ganzhi.WxHuo, "马": ganzhi.WxHuo, "鸟": ganzhi.WxHuo,
	"礻": ganzhi.WxHuo, "饣": ganzhi.WxHuo, "见": ganzhi.WxHuo,
	"隹": ganzhi.WxHuo, "香": ganzhi.WxHuo, "舌": ganzhi.WxHuo,
	// 土
	"土": ganzhi.WxTu, "山": ganzhi.WxTu, "石": ganzhi.WxTu,
	"田": ganzhi.WxTu, "玉": ganzhi.WxTu, "王": ganzhi.WxTu,
	"瓦": ganzhi.WxTu, "阜": ganzhi.WxTu, "阝": ganzhi.WxTu,
	"艮": ganzhi.WxTu, "戊": ganzhi.WxTu, "己": ganzhi.WxTu,
	"犭": ganzhi.WxTu, "穴": ganzhi.WxTu, "广": ganzhi.WxTu,
	"虫": ganzhi.WxTu, "羊": ganzhi.WxTu, "牛": ganzhi.WxTu,
	"厂": ganzhi.WxTu, "皿": ganzhi.WxTu, "宀": ganzhi.WxTu,
	"龙": ganzhi.WxTu, "甘": ganzhi.WxTu, "黄": ganzhi.WxTu,
	"豸": ganzhi.WxTu, "士": ganzhi.WxTu, "缶": ganzhi.WxTu,
	// 金
	"金": ganzhi.WxJin, "钅": ganzhi.WxJin, "刀": ganzhi.WxJin,
	"刂": ganzhi.WxJin, "刃": ganzhi.WxJin, "戈": ganzhi.WxJin,
	"辛": ganzhi.WxJin, "庚": ganzhi.WxJin, "酉": ganzhi.WxJin,
	"口": ganzhi.WxJin, "囗": ganzhi.WxJin, "白": ganzhi.WxJin,
	"革": ganzhi.WxJin, "车": ganzhi.WxJin, "骨": ganzhi.WxJin,
	"立": ganzhi.WxJin, "言": ganzhi.WxJin, "讠": ganzhi.WxJin,
	"齿": ganzhi.WxJin, "矢": ganzhi.WxJin, "斤": ganzhi.WxJin,
	"矛": ganzhi.WxJin, "鼻": ganzhi.WxJin, "韦": ganzhi.WxJin,
	"殳": ganzhi.WxJin, "鼎": ganzhi.WxJin,
	// 水
	"水": ganzhi.WxShui, "氵": ganzhi.WxShui, "雨": ganzhi.WxShui,
	"鱼": ganzhi.WxShui, "风": ganzhi.WxShui, "冫": ganzhi.WxShui,
	"子": ganzhi.WxShui, "壬": ganzhi.WxShui, "癸": ganzhi.WxShui, "亥": ganzhi.WxShui,
	"女": ganzhi.WxShui, "月": ganzhi.WxShui, "贝": ganzhi.WxShui,
	"鼠": ganzhi.WxShui, "豕": ganzhi.WxShui, "气": ganzhi.WxShui,
	"血": ganzhi.WxShui, "黑": ganzhi.WxShui, "鬼": ganzhi.WxShui,
}

// inferElementFromRadical returns the element implied by a Kangxi radical.
// Tries direct radical match first, then partial component match on the character itself.
func inferElementFromRadical(radical, char string) (Wuxing, bool) {
	if e, ok := radicalToElement[radical]; ok {
		return e, true
	}
	for _, r := range char {
		if e, ok := radicalToElement[string(r)]; ok {
			return e, true
		}
	}
	return 0, false
}

func loadNaming() error {
	{
		r := csv.NewReader(bytes.NewReader(gscCSV))
		records, err := r.ReadAll()
		if err != nil {
			return err
		}
		for i, rec := range records {
			if i == 0 {
				continue // skip header
			}
			if len(rec) < 11 {
				continue
			}
			word := rec[1]
			if word == "" {
				continue
			}
			elem := wuxingFromChinese(rec[5])
			if elem == 0 {
				var ok bool
				elem, ok = inferElementFromRadical(rec[3], word)
				if !ok {
					continue
				}
			}
			stroke, err := strconv.Atoi(rec[4])
			if err != nil {
				continue
			}
			tone, err := strconv.Atoi(rec[10])
			if err != nil {
				tone = 0
			}

			// Take the first reading when multiple pinyin are present ("zé,shì" → "ze").
			pinyin := rec[2]
			if idx := strings.IndexByte(pinyin, ','); idx >= 0 {
				pinyin = pinyin[:idx]
			}
			// Strip tone numbers and neutral-tone markers.
			pinyin = strings.TrimRight(pinyin, "0123456789·")

			ce := Character{
				Char:        word,
				Element:     elem,
				Stroke:      stroke,
				Radical:     rec[3],
				Pinyin:      pinyin,
				Tone:        tone,
				Traditional: rec[6],
			}
			charByElement[elem] = append(charByElement[elem], ce)
			for _, r := range word {
				charByRune[r] = ce
			}
		}
	}

	{
		var raw struct {
			Numbers map[int]struct {
				Element string `yaml:"element"`
				Fortune string `yaml:"fortune"`
				Desc    string `yaml:"desc"`
			} `yaml:"numbers"`
		}
		if err := yaml.Unmarshal(sancaiNumbersYAML, &raw); err != nil {
			return err
		}
		sanCaiNums = make(map[int]sanCaiNum)
		for n, v := range raw.Numbers {
			sanCaiNums[n] = sanCaiNum{
				Element: elementYAMLToChinese(v.Element),
				Fortune: fortuneYAMLToChinese(v.Fortune),
				Desc:    v.Desc,
			}
		}
	}

	{
		var raw struct {
			Configs map[string]struct {
				Fortune string `yaml:"fortune"`
				Desc    string `yaml:"desc"`
			} `yaml:"configs"`
		}
		if err := yaml.Unmarshal(sancaiConfigsYAML, &raw); err != nil {
			return err
		}
		sanCaiCfg = make(map[string]sanCaiCfgEntry)
		for k, v := range raw.Configs {
			sanCaiCfg[k] = sanCaiCfgEntry{
				Fortune: fortuneYAMLToChinese(v.Fortune),
				Desc:    v.Desc,
			}
		}
	}

	return nil
}

