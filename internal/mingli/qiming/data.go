package qiming

import (
	"encoding/csv"
	_ "embed"
	"log"
	"strconv"
	"strings"

	"github.com/25types/25types/internal/ganzhi"
	"gopkg.in/yaml.v3"
)

//go:embed data/gsc_pinyin_with_tone.csv
var gscCSV []byte

//go:embed data/sancai_numbers.yaml
var sancaiNumbersYAML []byte

//go:embed data/sancai_configs.yaml
var sancaiConfigsYAML []byte

//go:embed data/zodiac.yaml
var zodiacYAML []byte

var CharDB []CharacterEntry
var CharByElement = make(map[ganzhi.Element][]CharacterEntry)
var CharByRune = make(map[rune]CharacterEntry)
var SanCaiNums map[int]sanCaiNum
var SanCaiCfg map[string]sanCaiCfgEntry
var ZodiacByBranch map[int]zodiacBranchEntry

type sanCaiNum struct {
	Element string
	Fortune string
	Desc    string
}

type sanCaiCfgEntry struct {
	Fortune string
	Desc    string
}

type zodiacBranchEntry struct {
	Animal    string
	Preferred []string
	Forbidden []string
}

var defaultData = struct {
	CharDB        []CharacterEntry
	CharByElement map[ganzhi.Element][]CharacterEntry
	CharByRune    map[rune]CharacterEntry
	SanCaiNums    map[int]sanCaiNum
	SanCaiCfg     map[string]sanCaiCfgEntry
	Zodiac        map[int]zodiacBranchEntry
}{}

var defaultEngine = &defaultData

func init() {
	if err := loadNaming(); err != nil {
		log.Printf("qiming: load qiming data: %v", err)
	}
	if err := loadZodiac(); err != nil {
		log.Printf("qiming: load zodiac: %v", err)
	}

	defaultData.CharDB = CharDB
	defaultData.CharByElement = CharByElement
	defaultData.CharByRune = CharByRune
	defaultData.SanCaiNums = SanCaiNums
	defaultData.SanCaiCfg = SanCaiCfg
	defaultData.Zodiac = ZodiacByBranch
}

// radicalToElement maps Kangxi radicals to five elements per docs/naming.md §3.7.
var radicalToElement = map[string]ganzhi.Element{
	// 木
	"木": ganzhi.ElemWood, "艹": ganzhi.ElemWood, "林": ganzhi.ElemWood,
	"竹": ganzhi.ElemWood, "禾": ganzhi.ElemWood, "米": ganzhi.ElemWood, "桑": ganzhi.ElemWood,
	"舟": ganzhi.ElemWood, "羽": ganzhi.ElemWood, "纟": ganzhi.ElemWood,
	"弓": ganzhi.ElemWood, "户": ganzhi.ElemWood, "门": ganzhi.ElemWood,
	"巾": ganzhi.ElemWood, "虍": ganzhi.ElemWood, "鹿": ganzhi.ElemWood,
	"生": ganzhi.ElemWood, "角": ganzhi.ElemWood, "弋": ganzhi.ElemWood,
	"龠": ganzhi.ElemWood, "乙": ganzhi.ElemWood, "麦": ganzhi.ElemWood,
	"谷": ganzhi.ElemWood, "青": ganzhi.ElemWood, "耒": ganzhi.ElemWood,
	"⺮": ganzhi.ElemWood, "衤": ganzhi.ElemWood, "衣": ganzhi.ElemWood,
	// 火
	"火": ganzhi.ElemFire, "日": ganzhi.ElemFire, "灬": ganzhi.ElemFire,
	"心": ganzhi.ElemFire, "忄": ganzhi.ElemFire, "目": ganzhi.ElemFire, "离": ganzhi.ElemFire,
	"丙": ganzhi.ElemFire, "丁": ganzhi.ElemFire, "马": ganzhi.ElemFire, "鸟": ganzhi.ElemFire,
	"礻": ganzhi.ElemFire, "饣": ganzhi.ElemFire, "见": ganzhi.ElemFire,
	"隹": ganzhi.ElemFire, "香": ganzhi.ElemFire, "舌": ganzhi.ElemFire,
	// 土
	"土": ganzhi.ElemEarth, "山": ganzhi.ElemEarth, "石": ganzhi.ElemEarth,
	"田": ganzhi.ElemEarth, "玉": ganzhi.ElemEarth, "王": ganzhi.ElemEarth,
	"瓦": ganzhi.ElemEarth, "阜": ganzhi.ElemEarth, "阝": ganzhi.ElemEarth,
	"艮": ganzhi.ElemEarth, "戊": ganzhi.ElemEarth, "己": ganzhi.ElemEarth,
	"犭": ganzhi.ElemEarth, "穴": ganzhi.ElemEarth, "广": ganzhi.ElemEarth,
	"虫": ganzhi.ElemEarth, "羊": ganzhi.ElemEarth, "牛": ganzhi.ElemEarth,
	"厂": ganzhi.ElemEarth, "皿": ganzhi.ElemEarth, "宀": ganzhi.ElemEarth,
	"龙": ganzhi.ElemEarth, "甘": ganzhi.ElemEarth, "黄": ganzhi.ElemEarth,
	"豸": ganzhi.ElemEarth, "士": ganzhi.ElemEarth, "缶": ganzhi.ElemEarth,
	// 金
	"金": ganzhi.ElemMetal, "钅": ganzhi.ElemMetal, "刀": ganzhi.ElemMetal,
	"刂": ganzhi.ElemMetal, "刃": ganzhi.ElemMetal, "戈": ganzhi.ElemMetal,
	"辛": ganzhi.ElemMetal, "庚": ganzhi.ElemMetal, "酉": ganzhi.ElemMetal,
	"口": ganzhi.ElemMetal, "囗": ganzhi.ElemMetal, "白": ganzhi.ElemMetal,
	"革": ganzhi.ElemMetal, "车": ganzhi.ElemMetal, "骨": ganzhi.ElemMetal,
	"立": ganzhi.ElemMetal, "言": ganzhi.ElemMetal, "讠": ganzhi.ElemMetal,
	"齿": ganzhi.ElemMetal, "矢": ganzhi.ElemMetal, "斤": ganzhi.ElemMetal,
	"矛": ganzhi.ElemMetal, "鼻": ganzhi.ElemMetal, "韦": ganzhi.ElemMetal,
	"殳": ganzhi.ElemMetal, "鼎": ganzhi.ElemMetal,
	// 水
	"水": ganzhi.ElemWater, "氵": ganzhi.ElemWater, "雨": ganzhi.ElemWater,
	"鱼": ganzhi.ElemWater, "风": ganzhi.ElemWater, "冫": ganzhi.ElemWater,
	"子": ganzhi.ElemWater, "壬": ganzhi.ElemWater, "癸": ganzhi.ElemWater, "亥": ganzhi.ElemWater,
	"女": ganzhi.ElemWater, "月": ganzhi.ElemWater, "贝": ganzhi.ElemWater,
	"鼠": ganzhi.ElemWater, "豕": ganzhi.ElemWater, "气": ganzhi.ElemWater,
	"血": ganzhi.ElemWater, "黑": ganzhi.ElemWater, "鬼": ganzhi.ElemWater,
}

// inferElementFromRadical returns the element implied by a Kangxi radical.
// Tries direct radical match first, then partial component match on the character itself.
func inferElementFromRadical(radical, char string) (ganzhi.Element, bool) {
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
		r := csv.NewReader(strings.NewReader(string(gscCSV)))
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
			elem := ElementFromChinese(rec[5])
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
			tone, _ := strconv.Atoi(rec[10])

			// Take the first reading when multiple pinyin are present ("zé,shì" → "ze").
			pinyin := rec[2]
			if idx := strings.IndexByte(pinyin, ','); idx >= 0 {
				pinyin = pinyin[:idx]
			}
			// Strip tone numbers and neutral-tone markers.
			pinyin = strings.TrimRight(pinyin, "0123456789·")

			ce := CharacterEntry{
				Char:        word,
				Element:     elem,
				Stroke:      stroke,
				Radical:     rec[3],
				Pinyin:      pinyin,
				Tone:        tone,
				Traditional: rec[6],
			}
			CharByElement[elem] = append(CharByElement[elem], ce)
			CharDB = append(CharDB, ce)
			for _, r := range word {
				CharByRune[r] = ce
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
		SanCaiNums = make(map[int]sanCaiNum)
		for n, v := range raw.Numbers {
			SanCaiNums[n] = sanCaiNum{
				Element: ElementYAMLToChinese(v.Element),
				Fortune: FortuneYAMLToChinese(v.Fortune),
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
		SanCaiCfg = make(map[string]sanCaiCfgEntry)
		for k, v := range raw.Configs {
			SanCaiCfg[k] = sanCaiCfgEntry{
				Fortune: FortuneYAMLToChinese(v.Fortune),
				Desc:    v.Desc,
			}
		}
	}

	return nil
}

var branchNameToNum = func() map[string]int {
	m := make(map[string]int, 12)
	for i := 1; i <= 12; i++ {
		m[ganzhi.BranchNames[i]] = i
	}
	return m
}()

func loadZodiac() error {
	var raw map[string]struct {
		Animal    string   `yaml:"animal"`
		Preferred []string `yaml:"preferred"`
		Forbidden []string `yaml:"forbidden"`
	}
	if err := yaml.Unmarshal(zodiacYAML, &raw); err != nil {
		return err
	}
	ZodiacByBranch = make(map[int]zodiacBranchEntry)
	for branchName, v := range raw {
		num, ok := branchNameToNum[branchName]
		if !ok {
			continue
		}
		ZodiacByBranch[num] = zodiacBranchEntry{
			Animal:    v.Animal,
			Preferred: v.Preferred,
			Forbidden: v.Forbidden,
		}
	}
	return nil
}
