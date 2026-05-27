# Naming（起名）— 实现规格

> 八字→喜用神→候选字→组合→三才五格+音韵+字义并排→Top 20。命理引擎驱动，无主观评分。综合五行补益、生肖喜忌、三才五格三大流派。

---

## 1. 聚合边界

```
BirthInfo ──→ MingliChart ──→ 喜用神 + 生肖
                                │
                                ▼
                          候选字池 (五行筛选)
                                │
                         生肖过滤 (忌字剔除)
                                │
                         组合生成 (姓 + 字 × 字)
                                │
                     ┌──────────┼──────────┐
                     ▼          ▼          ▼
                   三才五格    音韵       字义
                     │          │          │
                     └──────────┼──────────┘
                                ▼
                          多维并排 → Top 20
```

| 概念 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| Deficiency | 值对象 | 实时计算 | 从 element_count 派生，判定喜用神方向 |
| ZodiacHint | 值对象 | 查表 | 年柱地支 → 宜用/忌用部首 |
| CharacterEntry | 数据 | 静态加载 | CSV 全量加载 7734 字，含康熙笔画、部首、拼音、声调。五行缺失时由部首法推断 |
| WuGe | 值对象 | 实时计算 | 天格/人格/地格/外格/总格，笔画+五行+吉凶 |
| SanCai | 值对象 | 实时计算 | 天人地三才五行生克配置 |
| NameCandidate | 值对象 | 实时计算 | 一个名字组合 + WuGe + SanCai + 音韵 + 字义标记 |
| NamingResult | 值对象 | 实时计算 | 喜用神 + 生肖提示 + Top 20 NameCandidate |

**不建表理由（引擎部分）**：所有计算从 MingliChart 和静态字库确定性派生，遵循 DPD-013。汉字库启动时由 `//go:embed` 从 CSV 加载到内存。起名历史记录由 Reports 系统管理。

---

## 2. 领域类型

### 2.1 输入

```go
type NamingRequest struct {
    Surname   string    `json:"surname"`    // 姓氏，单字
    BirthInfo BirthInfo `json:"birth_info"` // 与 BaZi 共用
}
```

### 2.2 喜用神 + 生肖

```go
type NamingAnalysis struct {
    Surname    string     `json:"surname"`     // 姓氏
    YongShen   string     `json:"yong_shen"`   // 用神元素（中文，如 "水"）
    XiShen     []string   `json:"xi_shen"`     // 喜神元素（可多选）
    ZodiacHint ZodiacHint `json:"zodiac_hint"` // 生肖提示
}

type ZodiacHint struct {
    Animal         string   `json:"animal"`           // 生肖（年柱地支）
    PreferredStems []string `json:"preferred_radicals"` // 宜用部首
    ForbiddenStems []string `json:"forbidden_radicals"` // 忌用部首
}
```

### 2.3 汉字条目

```go
type CharacterEntry struct {
    Char        string  `json:"char"`        // 汉字
    Element     Element `json:"element"`     // 五行（部首法判定）
    Stroke      int     `json:"stroke"`      // 简体笔画
    Radical     string  `json:"radical"`     // 部首（如 "氵"）
    Pinyin      string  `json:"pinyin"`      // 拼音（无声调）
    Tone        int     `json:"tone"`        // 声调 1-4，轻声=0
    Traditional string  `json:"traditional,omitempty"` // 繁体
}
```

**没有 `score` 字段。** 字库排序按笔画数（少优先）+ 拼音，不做主观评分。

### 2.4 三才五格

```go
type WuGe struct {
    TianGe GeResult `json:"tian_ge"`  // 天格 — 姓氏+1
    RenGe  GeResult `json:"ren_ge"`   // 人格 — 姓氏+名首字
    DiGe   GeResult `json:"di_ge"`    // 地格 — 名首字+名尾字
    WaiGe  GeResult `json:"wai_ge"`   // 外格 — 总格-人格+1
    ZongGe GeResult `json:"zong_ge"`  // 总格 — 全名笔画和
}

type GeResult struct {
    Stroke      int    `json:"stroke"`     // 笔画数
    Element     string `json:"element"`    // 1-81 数理五行（中文）
    Fortune     string `json:"fortune"`    // 吉/凶/半吉
    Description string `json:"description"` // 数理解释
}

type SanCai struct {
    Configuration string `json:"configuration"` // 三才五行序列，如 "金土土"
    Fortune       string `json:"fortune"`       // 吉/凶/半吉
    Description   string `json:"description"`   // 天人地生克解释
}
```

**五格笔画算法**（康熙字典笔画）：
- 天格 = 姓氏笔画 + 1（单姓）
- 人格 = 姓氏笔画 + 名字首字笔画
- 地格 = 名字首字笔画 + 名字尾字笔画（单名则 + 1）
- 外格 = 总格 - 人格 + 1
- 总格 = 姓氏笔画 + 所有名字字笔画

**三才配置**：天格/人格/地格三者的五行（1-81 数理五行）之间的生克关系。

```
木生火、火生土、土生金、金生水、水生木
木克土、火克金、土克水、金克木、水克火
```

### 2.5 名字候选

```go
type NameCandidate struct {
    Name        string           `json:"name"`        // 全名
    Characters  []CharacterEntry `json:"characters"`  // 名字部分的字
    WuGe        WuGe             `json:"wu_ge"`       // 五格
    SanCai      SanCai           `json:"san_cai"`     // 三才配置
    Phonetic    PhoneticMark     `json:"phonetic"`    // 音韵标记（非评分）
    Highlights  []string         `json:"highlights"`  // 属性亮点，如 ["三才全吉", "平仄交替", "五行补益到位"]
}

type PhoneticMark struct {
    Tones       string `json:"tones"`        // 声调序列，如 "3-4-2"（上-去-阳平）
    IsPingZe    bool   `json:"is_ping_ze"`   // 是否平仄交替
    HasHomophone bool  `json:"has_homophone"` // 是否有谐音歧义
    HomophoneNote string `json:"homophone_note"` // 谐音提示，如 "近音'没辙'"
}
```

`Highlights` 是**属性标记**，不是评分。如："三才全吉""五行补益到位""平仄交替""寓意优美"。用户看到的是名字的属性说明，而非数字分数。

### 2.6 完整结果

```go
type NamingResult struct {
    Analysis   NamingAnalysis  `json:"analysis"`
    Candidates []NameCandidate `json:"candidates"` // Top 20，按三才五格排序
}
```

---

## 3. 行为规则

### 3.1 流水线总览

```
① 五行定候选池           ② 生肖过滤          ③ 组合生成
───────────────── → ───────────────── → ─────────────────
按喜用神取字库       剔除生肖忌字        姓 + 候选字 × 候选字
(50-100字/元素)       (部首黑名单)        (max ~2500组合)
                                              │
                                              ▼
④ 多维并排                               ⑤ Top 20 输出
───────────────── ← ───────────────── ← ─────────────────
三才五格 │ 音韵 │ 字义                   含属性标记，不含分数
```

**五行和生肖是门槛**——不通过即淘汰。三才五格、音韵、字义是并排维度——没有权重，每个维度独立标记。

### 3.2 第一步：五行定候选池

从 `mingge` 分析获取喜用神元素，取对应元素的字库为候选池。

```
喜用神 = 水 → 取水部字库（~80字）
喜用神 = 金 → 取金部字库（~60字）
```

若喜用神有多个（如金和水），两个字库合并去重。

### 3.3 第二步：生肖过滤

年柱地支 → 生肖 → 忌用部首表。从候选池中剔除含忌用部首的字。

```
年柱地支 = 子（鼠）→ 忌 "午" "马" 部首 → 从候选池剔除
年柱地支 = 午（马）→ 忌 "子" 部首     → 从候选池剔除
```

生肖宜忌表（地支 → 宜用部首 / 忌用部首），存储于 `internal/mingli/qiming/data/zodiac.yaml`。

### 3.4 第三步：组合生成

姓氏 + 候选字 × 候选字 → 所有可能的二字名组合。

```
姓 "李" + 候选池 [沐, 泽, 涵, ...] × [沐, 泽, 涵, ...]
→ 李沐沐, 李沐泽, 李沐涵, 李泽沐, ...
```

组合数 = N × N（N 约 40-80），最大约 6400 个。每个组合计算三才五格、音韵、字义。

### 3.5 第四步：多维并排

每个组合计算三个维度：

**三才五格**（排序主维度）：
1. 三才配置吉凶 — 全吉 > 二吉一半吉 > 一吉 > 全凶
2. 五格吉数 — 吉格多者优先

**音韵**：
- 平仄交替：仄平仄 或 平仄平 > 平平平 或 仄仄仄
- 谐音检查：匹配常见词语/负面词库

**字义**：
- 名字组合后是否有不佳联想（如"李坏"）
- 字义是否协调（不互相矛盾）

### 3.6 第五步：排序输出

先按三才配置好坏分档，档内按五格吉数排，同档同分按音韵平仄交替排。

取 Top 20 返回。每个返回 `highlights` 属性标记——不是分数，是"三才全吉""五行补益到位""平仄交替"等用户看得懂的特征说明。

### 3.7 汉字五行判定（部首法）

根据部首/偏旁确定五行。判定顺序：

1. **CSV wuxing 列优先** — 直接读取 ben-hua CSV 中的五行标注
2. **部首精确匹配** — 查 radical→element 映射表（~105 部首，见 `data.go:radicalToElement`）
3. **偏旁推断** — 字中含明确五行偏旁（如含"氵"→水），遍历字符 rune 查映射表
4. **兜底** — 无法判定则跳过该字不入库

部首映射表（完整 105 部首，`internal/mingli/qiming/data.go`）：

| 五行 | 部首 |
|------|------|
| 木 | 木 艹 林 竹 禾 米 桑 舟 羽 纟 弓 户 门 巾 虍 鹿 生 角 弋 龠 乙 麦 谷 青 耒 ⺮ 衤 衣 |
| 火 | 火 日 灬 心 忄 目 离 丙 丁 马 鸟 礻 饣 见 隹 香 舌 |
| 土 | 土 山 石 田 玉 王 瓦 阜 阝 艮 戊 己 犭 穴 广 虫 羊 牛 厂 皿 宀 龙 甘 黄 豸 士 缶 |
| 金 | 金 钅 刀 刂 刃 戈 辛 庚 酉 口 囗 白 革 车 骨 立 言 讠 齿 矢 斤 矛 鼻 韦 殳 鼎 |
| 水 | 水 氵 雨 鱼 风 冫 子 壬 癸 亥 女 月 贝 鼠 豕 气 血 黑 鬼 |

CSV 中约 5915 字有直接五行标注，余下 2190 字通过部首法覆盖 95.4%（7734/8105）。

纯函数，查表。无 I/O。

```go
func LookupCharacterElement(char string) Element
```

### 3.8 康熙字典笔画

起名领域事实标准。繁体字形笔画，非简体。所有竞品使用此标准，保持一致。

```go
func LookupKangxiStroke(char string) int
```

### 3.9 姓氏五行

仅信息展示，不参与三才五格以外的计算。部首法判定，无部首则查静态姓氏表。

---

## 4. 汉字数据库

### 4.1 数据来源

```
ben-hua/general_standard_chinese
  └── gsc_pinyin_with_tone.csv     ← 唯一基础数据源（8105 字，11 列）
自建
  ├── sancai_numbers.yaml          ← 1-81 数理吉凶（从 James88/qiming constants.py 提取）
  ├── sancai_configs.yaml          ← 三才配置词条（从 James88/qiming sancai.txt 解析）
  └── zodiac.yaml                  ← 生肖宜忌部首表（业界共识）
```

| 数据项 | 来源 | 说明 |
|--------|------|------|
| 汉字 + 部首 + 拼音 + 声调 | ben-hua `gsc_pinyin_with_tone.csv` | 通用规范汉字表 8105 字，11 列：num/word/pinyin/radical/stroke_count/wuxing/traditional/wubi/initial/final/tone |
| 五行判定 | CSV wuxing 列 + 部首法推断 | ~5915 字 CSV 有直接五行标注；~2190 字通过 radicalToElement 映射表推断，覆盖率 95.4%（7734/8105） |
| 三才五格数理吉凶 | 自建 `sancai_numbers.yaml` | 1-81 数理五行 + 吉/凶 |
| 三才配置词条 | 自建 `sancai_configs.yaml` | 125 种天人地组合（5×5×5）× 吉凶判定 |
| 生肖宜忌部首 | 自建 `zodiac.yaml` | 12 地支 → 宜用/忌用部首，业界共识 |

### 4.2 笔画策略

当前使用 ben-hua CSV 自带的 `stroke_count` 列（简体笔画）。对于起名场景，简体笔画与康熙笔画在绝大多数情况下一致（简繁同形的 ~7335 字）。

康熙笔画补丁（James88/qiming `wuxing_dict_fanti.json`，770 繁简异体字）待后续集成。

### 4.3 规模与覆盖面

从 8105 字全集中，通过 CSV wuxing 列 + 部首法推断，得到 7734 字（覆盖率 95.4%）。

每元素分布（部首法覆盖后）：
- 木部约 1486 字，火部约 1520 字，土部约 1605 字，金部约 1588 字，水部约 1535 字

未覆盖的 371 字为部首无明确五行提示的罕见字，不影响起名场景。

### 4.4 存储格式

`data/gsc_pinyin_with_tone.csv`，构建时 `//go:embed` 嵌入二进制，`init()` 时加载到内存。CSV 解析时五行缺失的条目自动调用 `inferElementFromRadical()` 通过部首法补全。

每字符存入 `CharByRune map[rune]CharacterEntry`（O(1) 单字查询）和 `CharByElement map[Element][]CharacterEntry`（按元素批量取字）。

### 4.5 三才数理数据

`internal/mingli/qiming/data/sancai_numbers.yaml`（从 James88/qiming `constants.py` 提取）：

```yaml
# 1-81 数理 → 五行 + 吉凶 + 分类标签
numbers:
  1: {element: wood, fortune: ji, desc: "天地开泰，万物起始"}
  2: {element: wood, fortune: xiong, desc: "分离破败，进退失据"}
  # ... 3-81
```

### 4.6 三才配置词条

`internal/mingli/qiming/data/sancai_configs.yaml`（从 James88/qiming `sancai.txt` 解析）：

```yaml
# 天人地三才五行 → 吉凶 + 解释
configs:
  "木木木": {fortune: da_ji, desc: "三木成林，生机无限，基础稳固，成功顺调..."}
  "木木火": {fortune: da_ji, desc: "木火通明，得长辈庇荫，顺调成功..."}
  # ... ~125 种组合 (5×5×5)
```

### 4.7 生肖宜忌表

`data/zodiac.yaml`（自建，业界公知数据），12 地支完整覆盖。数据由 `loadZodiac()` 解析为 `ZodiacByBranch map[int]zodiacBranchEntry`，`ZodiacFromYearBranch()` 查询。

```yaml
子:  # 鼠
  animal: 鼠
  preferred: ["米", "豆", "宀", "口", "田", "禾", "艹", "木"]
  forbidden: ["午", "马", "火", "日", "灬", "南"]
丑:  # 牛
  animal: 牛
  preferred: ["艹", "田", "禾", "宀", "车", "土", "巳", "酉"]
  forbidden: ["未", "羊", "月", "肉", "心", "忄"]
# ... 寅(虎) 至 亥(猪)
```

---

## 5. Engine 函数签名

```go
// engine/naming.go

// 汉字五行判定
func LookupCharacterElement(char string) Element
func LookupSurnameElement(surname string) Element

// 笔画
func LookupKangxiStroke(char string) int

// 生肖
func ZodiacFromYearBranch(branch Branch) ZodiacHint

// 三才五格
func ComputeWuGe(surname string, givenChars []string) WuGe
func ComputeSanCai(tian, ren, di Element) SanCai

// 数理五行（1-81 数 → 五行 + 吉凶）
func StrokeFortune(stroke int) (Element, string, string) // element, fortune, description

// 音韵
func AnalyzePhonetic(chars []CharacterEntry) PhoneticMark

// 流水线
func GenerateCandidates(surname string, analysis NamingAnalysis, limit int) []NameCandidate
```

---

## 6. API 契约 → `docs/API.md`

HTTP 契约的**唯一权威源**为 `docs/API.md`（Naming 段）。本节仅为设计引用。

| 端点 | 用途 | Auth |
|------|------|------|
| `POST /api/qiming/analyze` | 八字起名分析，返回 chart + 喜用神 + 生肖提示 | 🔒 |
| `GET /api/qiming/characters` | 按元素取候选字 | 🔒 |
| `POST /api/qiming/sancai` | 三才五格分析——用户拼的名字查吉凶 | 🔒 |

完整起名方案（Top 20 推荐名字 + 详细解读）通过 Reports 系统生成：`POST /api/reports` with `scene=naming`。

---

## 7. 前端交互

### 7.1 起名流程

1. **输入**：姓氏 + 出生日期时间
2. `POST /api/qiming/analyze` → 展示八字分析 + 喜用神 + 生肖宜忌
3. 用户可选浏览字库（`GET /api/qiming/characters?element=water`）
4. 用户如果自己拼名字 → `POST /api/qiming/sancai` → 三才五格吉凶即时展示
5. 付费生成完整方案 → Reports SSE 流式输出 Top 20 + 详细解读

### 7.2 展示方式

结果页展示 Top 20 名字，每个名字带属性标记（不是分数）：

```
李沐泽  [三才全吉] [五行补益到位] [平仄交替]
  沐（水/8画）泽（水/17画）— 双水补命局缺水
  天格8(金/吉) 人格25(土/吉) 地格25(土/吉) 外格8(金/吉) 总格32(木/吉)
  三才配置：金土土 — 天人地相生，全吉

李涵远  [三才一吉] [五行补益到位]
  涵（水/12画）远（土/17画）— 水补命局，土生金为次
  ...
```

---

## 8. 错误

| Code | HTTP | 含义 |
|------|------|------|
| `invalid_request` | 400 | surname 缺失或非中文单字 |
| `invalid_birth_info` | 400 | 出生信息字段缺失或非法（复用 BaZi 校验） |
| `not_found` | 404 | 未保存 birth_info |

---

## 9. 输入校验

- `surname`: 1-2 个中文字符，非空
- `birth_info`: 与 `POST /api/bazi/chart` 相同校验规则
- `name`（sancai）: 2-4 个中文字符（含姓氏）

---

## 10. 与现有系统的关系

| | BaZi | Naming |
|------|------|------|
| 输入 | 出生时间 + 地点 | 出生时间 + 姓氏 |
| 核心计算 | engine.ComputeChart | engine.GenerateCandidates |
| 静态数据 | cities.json | gsc_pinyin_with_tone.csv + sancai_numbers.yaml + sancai_configs.yaml + zodiac.yaml |
| 流派依据 | 子平八字 | 喜用神 + 生肖喜忌 + 三才五格 |
| 持久化 | 不持久化（值对象） | 不持久化引擎结果；起名历史走 Reports |
| 解读 | mingge 免费直接返回 | 完整方案走 Reports SSE |
| 认证 | 排盘匿名可用 | 全部需登录（🔒） |
