# BaZi（八字）— 实现规格

> 真太阳时计算、八字四柱排盘、合八字匹配。纯计算值对象，不建表，实时计算。

理论依据：`docs/theory/bazi-theory.md`

---

## 1. 聚合边界

```
BirthInfo ──输入──> SolarTime ──────> Pillar(时柱)
                    │
                    └──> Calendar ──> Pillar(年/月/日柱)
                                           │
                                           └──> MingliChart ──> MingliMatch(合八字)
```

| 概念 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| BirthInfo | 输入值对象 | 一次请求 | 年月日时分必填，经度可选（默认 120°=北京时间） |
| SolarTime | 值对象 | 实时计算 | Spencer 公式 + 经度修正 + 夏令时回拨 |
| Pillar | 值对象 | 实时计算 | 一根柱 = Stem + Branch，不可变 |
| MingliChart | 值对象 | 实时计算 | 四柱 + 藏干 + 十神 + 纳音 + 十二长生 + 大运 |
| MingliMatch | 值对象 | 实时计算 | 两份 Chart 的合分析结果 |

**不建表理由**：排盘结果纯计算值对象，无持久化需求。参考 DPD-013（BaZi 不建表）和 DPD-010（Bond/Flow 不建表）。用户通过 profile 保存 birth_info，profile API 返回实时计算的 bazi_chart。

---

## 2. 领域类型

### 2.1 基础枚举

```go
type Stem int    // 1=甲 2=乙 3=丙 4=丁 5=戊 6=己 7=庚 8=辛 9=壬 10=癸
type Branch int  // 1=子 2=丑 3=寅 4=卯 5=辰 6=巳 7=午 8=未 9=申 10=酉 11=戌 12=亥
type Element int // 1=木 2=火 3=土 4=金 5=水
type YinYang bool // true=阳 false=阴
type Gender string // "male" | "female"
type Strength int  // 身弱=0, 中和=1, 身强=2
```

包位置：全部基础类型定义在 `internal/ganzhi/`，`bazi` 包通过类型别名引用。

### 2.2 输入

出生信息由各传输层定义（HTTP `birthParams`、MCP `BirthProfile`），字段相同：

| 字段 | 类型 | 说明 |
|------|------|------|
| year | int | 公历年份，1900-2200 |
| month | int | 1-12 |
| day | int | 1-31 |
| hour | int | 0-23（当地时间） |
| minute | int | 0-59，默认 0 |
| longitude | float64 | 出生地经度，默认 120.0（北京时间） |
| timezone | float64 | 时区小时偏移，默认 8（UTC+8） |
| gender | string | "male" / "female" |

夏令时由后端 `IsDST(year, month, day)` 自动检测，无需前端传递。

入库入口：`ComputeChartFromBirth(year, month, day, hour, minute int, longitude, timezone float64, gender Gender) ChartResult` — 从原始出生数据到完整命盘的一步计算。

### 2.3 支柱

```go
type Pillar struct {
    Stem   Stem   `json:"stem"`   // 天干
    Branch Branch `json:"branch"` // 地支
}

type Bazi struct {
    Year  Pillar
    Month Pillar
    Day   Pillar
    Hour  Pillar
}
```

`Bazi` 使用命名字段代替 `[4]Pillar` 位置数组，提供 `Slice() [4]Pillar` 方法兼容索引访问。

### 2.4 排盘

**内部类型**（`ComputeChart` 产出，不直接序列化）：

```go
// HiddenStemsOut — 单柱藏干。
type HiddenStemsOut struct {
    Main  Stem  `json:"main"`  // 本气（值类型，永远存在）
    Mid   *Stem `json:"mid"`   // 中气（可能为 nil）
    Minor *Stem `json:"minor"` // 余气（可能为 nil）
}

// PillarInfo — 单柱的所有派生数据。
type PillarInfo struct {
    Stem        Stem             `json:"stem"`
    Branch      Branch           `json:"branch"`
    NaYin       string           `json:"nayin"`
    HiddenStems HiddenStemsOut   `json:"hidden_stems"`
    TenGods     []TenGodEntry    `json:"ten_gods"`
    LifeStages  []LifeStageEntry `json:"life_stages"`
    ShenSha     []ShenShaEntry   `json:"shensha"`
    IsVoid      bool             `json:"is_void"`
    IsSelfHe    bool             `json:"is_self_he"`
    SelfHeName  string           `json:"self_he_name,omitempty"`
    IsKuiGang   bool             `json:"is_kui_gang"`
}

// ChartResult — 完整命盘。PillarInfo 取代旧的 PillarDerived，一步产出全部派生数据。
type ChartResult struct {
    Year         PillarInfo
    Month        PillarInfo
    Day          PillarInfo
    Hour         PillarInfo
    SolarTime    float64
    SolarDate    time.Time
    BaziDate     time.Time
    LifeStages   TwelveStages
    Dayun        DayunPillars
    DayMaster    Stem
    ElementCount map[Element]int
}

// TwelveStages — 十二长生宫（命名字段 + Slice() 转换）。
type TwelveStages struct { ... } // ChangSheng/MuYu/.../Yang
type StageOut struct { Name string; Branch Branch }
```

`ChartResult` 便捷方法：
- `ToBazi() Bazi` — 提取四柱
- `NaYinArray() [4]string` / `HiddenStemsArray() [4]HiddenStemsOut` / `TenGodsArray() [4][2]string`

**API 输出类型**（HTTP/MCP 统一使用）：

```go
// ChartOutput — 排盘 API 的完整输出，HTTP/MCP 共享类型。
type ChartOutput struct {
    YearPillar       PillarInfo       `json:"year_pillar"`
    MonthPillar      PillarInfo       `json:"month_pillar"`
    DayPillar        PillarInfo       `json:"day_pillar"`
    HourPillar       PillarInfo       `json:"hour_pillar"`
    DayMaster        string           `json:"day_master"`       // 天干中文名
    LifeStages       [12]StageOut     `json:"life_stages"`
    Dayun            *DayunResult     `json:"dayun"`
    ElementCount     map[string]int   `json:"element_count"`    // "木":3, "火":1, ...
    SolarTimeMinutes float64          `json:"solar_time_minutes"`
    SolarDatetime    string           `json:"solar_datetime"`   // RFC3339
    BaziDatetime     string           `json:"bazi_datetime"`    // "2006-01-02 子时"
    FullHeHui        []TripleHeFull   `json:"full_he_hui"`
    GongJia          []GongJiaEntry   `json:"gong_jia"`
    TaiYuanMingGong  TaiYuanMingGong  `json:"tai_yuan_ming_gong"`
    NayinRelations   []NaYinRelation  `json:"nayin_relations"`
    SanQiName        string           `json:"sanqi_name"`
    Zodiac           string           `json:"zodiac"`
    Season           string           `json:"season"`
    LunarMonth       string           `json:"lunar_month"`
    HourRange        string           `json:"hour_range"`
    XunName          string           `json:"xun_name"`
    WangShuai        map[string]string `json:"wang_shuai"`
    DayMansion       DayMansion       `json:"day_mansion"`
    YongShen         YongShenResult   `json:"yong_shen"`
}

// BondOutput — 合盘 API 输出。
type BondOutput struct {
    ChartA ChartOutput `json:"chart_a"`
    ChartB ChartOutput `json:"chart_b"`
    Bond   BondResult  `json:"bond"`
}
```

`BuildChartOutput(chart ChartResult, birthYear, birthMonth, birthHour int) ChartOutput` 将内部 `ChartResult` 转为 API 输出。

### 2.5 合八字

无评分，结构化事实数据输出。`ComputeBond` 对两份命盘进行五维交叉分析：

```go
func ComputeBond(a, b ChartResult, aBirthYear, aBirthMonth, aBirthHour, bBirthYear, bBirthMonth, bBirthHour int) BondResult

type BondResult struct {
    PillarCross  PillarCross    `json:"pillar_cross"`   // 16 对柱交互
    TenGodCross  TenGodCross    `json:"ten_god_cross"`   // 十神互视
    NayinCross   NayinCross     `json:"nayin_cross"`     // 纳音五行关系
    ShenshaCross ShenshaCross   `json:"shensha_cross"`   // 神煞互现
    Structure    StructureCross `json:"structure"`       // 结构比较（胎元/命宫/大运/旬空）
}
```

### 2.6 命格分析

```go
// YongShenResult — 用神分析，包含扶抑与调候两套独立体系。
type YongShenResult struct {
    FuYi    FuYiResult    `json:"fuyi"`    // 扶抑取用（日主旺衰）
    TiaoHou TiaoHouResult `json:"tiaohou"` // 调候取用（月令气候）
}

type FuYiResult struct {
    Strength string `json:"strength"` // 身强/身弱/中和
    Pattern  string `json:"pattern"`  // 格局
    Yong     string `json:"yong"`     // 用神
    Xi       string `json:"xi"`       // 喜神
    Ji       string `json:"ji"`       // 忌神
}

type TiaoHouResult struct {
    Season string `json:"season"`           // 季候
    Yong   string `json:"yong"`             // 调候用神
    Xi     string `json:"xi"`               // 调候喜神
    Ji     string `json:"ji"`               // 调候忌神
    Detail string `json:"detail,omitempty"` // 调候说明
}

func ComputeYongShen(chart ChartResult) YongShenResult
```

---

## 3. 行为规则

### 3.1 真太阳时计算

```
func ComputeSolarTime(birth BirthInfo) float64:
    lstMinutes := birth.Hour * 60 + birth.Minute                     // ① 当地标准时间
    if birth.IsDST:
        lstMinutes -= 60                                              // ② 夏令时回拨
    lonOffset := 4 * (birth.Longitude - birth.Timezone)               // ③ 经度修正（分）
    n := dayOfYear(birth.Year, birth.Month, birth.Day)                // ④ 日序数
    B := 360 * (n - 81) / 365.0                                       // ⑤ Spencer 公式
    eot := 9.87*sin(2*B*π/180) - 7.53*cos(B*π/180) - 1.5*sin(B*π/180)
    ast := lstMinutes + lonOffset + eot                                // ⑥ 真太阳时（分钟）
    // 归一化到 [0, 1440)
    return ((ast % 1440) + 1440) % 1440
```

**夏令时规则**（中国 1986-1991）：
- 1986: 5/4 – 9/14
- 1987: 4/12 – 9/13
- 1988: 4/10 – 9/11
- 1989: 4/16 – 9/17
- 1990: 4/15 – 9/16
- 1991: 4/14 – 9/15

`birth.IsDST` 由前端自动检测（`isChinaDST()` 函数，与后端 `IsDST()` 逻辑一致），用户也可通过 DST 开关手动覆盖。

### 3.2 时辰确定

```
func HourBranchFromSolarTime(astMinutes float64) Branch:
    hourIndex := int((astMinutes + 120) / 120) % 12  // +120 = 子时起始偏移
    return Branch(hourIndex + 1)                      // 1=子 ... 12=亥
```

### 3.3 四柱推算

**年柱**：
```
若月份 < 2 或 (月份 == 2 且 日 < 立春日): year -= 1
stem  = (year - 3) % 10    // 0→10=癸, 1=甲...
branch = (year - 3) % 12   // 0→12=亥, 1=子...
```

**月柱**：先确定节气月（monthIndex = 节气索引），再：
```
monthStem  = (yearStem*2 + monthIndex) % 10
monthBranch = monthIndex + 2   // 正月=寅(3)
```

**日柱**：儒略日法，1900-01-01 为甲戌（0-based index=10）：
```
jd := JulianDay(year, month, day)
diff := jd - baseJD(1900, 1, 1)
gzIndex := (10 + diff) % 60
dayStem   := gzIndex % 10       // 0=甲...9=癸
dayBranch := gzIndex % 12       // 0=子...11=亥
```

**时柱**：
```
hourBranch := HourBranchFromSolarTime(solarTime)
hourStem := (dayStem*2 + hourBranch - 2) % 10
```

### 3.4 排盘扩展

**入口**：`ComputeChartFromBirth(year, month, day, hour, minute int, longitude, timezone float64, gender Gender) ChartResult`
— 默认 longitude=120, timezone=8，自动检测夏令时，一步产出完整命盘。

内部三步渐进计算：
```
① ast = ComputeSolarTime(year, month, day, hour, minute, longitude, timezone, isDST) → 真太阳时（分钟）
② bz  = ComputeBazi(ast, year, month, day, hour, minute, timezone, isDST) → BaziResult（四柱 + 真太阳时）
③ chart = ComputeChart(bz, year, month, day, gender) → ChartResult（完整命盘，PillarInfo 一步产出全部派生数据）
```

**藏干**：查表，12 个 Branch 各对应 1-3 个 Stem。`HiddenStemsForBranch(b Branch) HiddenStemsQi`

**十神**：以日干为中心，对其他柱的天干+地支藏干本气计算关系。五行生克链：木→火→土→金→水→木。
```
关系判断: 同五行(比劫) / 生我(印) / 我生(食伤) / 我克(财) / 克我(官杀)
阴阳维度: 同阴阳=偏, 异阴阳=正
```

**纳音**：四柱的 StemBranch 序数查表得纳音五行名称。

**十二长生**：日干对四柱地支各查十二长生表。阳干顺排、阴干逆排，火土同宫。`computeLifeStages` 返回 `TwelveStages`（命名字段结构体），`TwelveStages.Slice()` 转为 `[12]StageOut` 供 API 输出。

**大运**：
```
阳年 = 年干为甲丙戊庚壬
顺排 = (男 && 阳年) || (女 && !阳年)
起运岁 = 出生日到下一个顺向/逆向"节"的天数 / 3
大运柱 = 从月柱开始，顺排取下一柱，逆排取上一柱，重复 8 次
```

**大运格式化**：`ComputeDayunResult(DayunPillars, dayMaster Stem, birthYear int, currentYear int, pillars [4]Pillar) *DayunResult` — 将原始大运柱包装为含十神、关系、神煞的完整输出。

**干支关系函数**（位于 `internal/ganzhi/branch_relation.go`）：
- `IsStemHe(a, b Stem) bool` — 天干五合
- `IsBranchHe(a, b Branch) bool` — 地支六合
- `IsTripleHe(a, b Branch) bool` — 地支三合
- `IsTripleHui(a, b Branch) bool` — 地支三会
- `IsLiuChong(a, b Branch) bool` — 六冲
- `IsXing(a, b Branch) bool` — 相刑
- `IsHai(a, b Branch) bool` — 六害

这些函数仅依赖 `Stem`/`Branch` 基础类型，因此从 `bazi` 包下沉到 `ganzhi` 包，所有包通过 `ganzhi.Is*` 共享调用。

### 3.5 合八字

`ComputeBond(a, b ChartResult, ...)` 无评分，纯结构化事实数据。输出五维交叉分析：

| 维度 | 内容 |
|------|------|
| pillar_cross | 16 对柱交互（天干五合、地支六合/三合/三会/六冲/相刑/六害） |
| ten_god_cross | 十神互视（双方日干对对方四柱天干的十神视角） |
| nayin_cross | 纳音五行关系（相生/相克/相同 + 详情） |
| shensha_cross | 神煞互现（一方神煞在对方柱中的出现） |
| structure | 结构比较（胎元/命宫/大运/旬空/三元） |

不设分数和等级——各系统各自说话，由 LLM 解读。

### 3.6 城市查找

`configs/cities.json` 预置中国城市经纬度数据，启动加载到内存：

```go
type City struct {
    Name    string  `json:"name"`
    Country string  `json:"country"`
    Lat     float64 `json:"lat"`
    Lng     float64 `json:"lng"`
}
```

数据来源 `joelacus/world-cities`（GeoNames，WGS-84），取人口 ≥ 100,000 城市，`configs/cities.json` 嵌入。

---

## 4. Engine 函数签名

包分布：
- **`internal/ganzhi/`** — 基础类型、干支关系判断、五行生克
- **`internal/tianwen/`** — 真太阳时、节气、儒略日、城市查找
- **`internal/mingli/bazi/`** — 八字排盘、合八字、神煞、大运、命格、流年流月流日流时

### 4.1 基础层 (`internal/ganzhi/`)

```go
// tiangan_dizhi.go
func StemElement(s Stem) Element
func StemYinYang(s Stem) YinYang
func BranchElement(b Branch) Element
func Sheng(from, to Element) bool
func Ke(from, to Element) bool
func SixtyCycleName(stem Stem, branch Branch) int
func StemNameStr(s Stem) string
func BranchNameStr(b Branch) string
func BranchSeason(b Branch) string
func BranchLunarMonth(b Branch) string
func BranchHourRange(b Branch) string
func Zodiac(b Branch) string
func ElementFromChinese(s string) Element

// branch_relation.go
func IsStemHe(a, b Stem) bool
func IsBranchHe(a, b Branch) bool
func IsTripleHe(a, b Branch) bool
func IsTripleHui(a, b Branch) bool
func IsLiuChong(a, b Branch) bool
func IsXing(a, b Branch) bool
func IsHai(a, b Branch) bool
```

### 4.2 天文层 (`internal/tianwen/`)

```go
func ComputeSolarTime(year, month, day, hour, minute int, longitude, timezone float64, isDST bool) float64
func HourBranchFromSolarTime(astMinutes float64) Branch
func IsDST(year, month, day int) bool
func JulianDay(year, month, day int) int
func DayOfYear(year, month, day int) int
func SolarMonthIndex(t time.Time) int
func LiChunDay(year int) (int, int)
func SearchCities(q string) []City
```

### 4.3 八字核心管线 (`internal/mingli/bazi/`)

```go
// --- 管线入口 ---
func ComputeChartFromBirth(year, month, day, hour, minute int, longitude, timezone float64, gender Gender) ChartResult
func ComputeSolarTime(year, month, day, hour, minute int, longitude, timezone float64, isDST bool) float64
func ComputeBazi(solarTime float64, year, month, day, hour, minute int, timezone float64, isDST bool) BaziResult
func ComputeChart(bz BaziResult, year, month, day int, gender Gender) ChartResult
func BuildChartOutput(chart ChartResult, birthYear, birthMonth, birthHour int) ChartOutput

// --- 四柱基础 ---
func YearPillar(year, month, day int) Pillar
func MonthPillar(birthTime time.Time, yearStem Stem) Pillar
func DayPillar(year, month, day int) Pillar
func HourPillar(solarTime float64, dayStem Stem) Pillar
```

### 4.4 排盘扩展 (`internal/mingli/bazi/`)

```go
// --- 藏干、十神、纳音 ---
func HiddenStemsForBranch(b Branch) HiddenStemsQi
func ComputeTenGodsTable(dayMaster Stem, bz ganzhi.Bazi, hs [4]HiddenStemsOut) [4][]TenGodEntry
func ComputeLifeStageTable(bz ganzhi.Bazi, hiddenStems [4]HiddenStemsOut) [4][]LifeStageEntry
func NaYinString(s Stem, b Branch) string
func ComputeNaYinRelations(nayin [4]string) []NaYinRelation
func TenGodName(tg int) string
func TenGodType(dmElem Element, dmYY YinYang, otherElem Element, otherYY YinYang) int

// --- 神煞 ---
func ComputeShenSha(bz ganzhi.Bazi, dayMaster Stem, monthBranch Branch) [4][]ShenShaEntry
func ComputeKongWang(dayPillar Pillar, bz Bazi) []int
func ComputeDynamicShenSha(b Branch, yearBranch Branch, dayMaster Stem) []ShenShaEntry

// --- 大运 ---
func ComputeDayunResult(bf DayunPillars, dayMaster Stem, birthYear, currentYear int, bz Bazi) *DayunResult
func ComputeDayunInteractions(dayunPillars []DayunPillar, bz Bazi) []PillarInteraction

// --- 合会局 ---
func ComputeFullTripleHeHui(bz ganzhi.Bazi) []TripleHeFull

// --- 宫位 ---
func ComputeGongJia(bz ganzhi.Bazi) []GongJiaEntry
func ComputeTaiYuanMingGong(monthPillar Pillar, yearStem Stem, birthMonth, birthHour int) TaiYuanMingGong

// --- 合八字 ---
func ComputeBond(a, b ChartResult, aBirthYear, aBirthMonth, aBirthHour, bBirthYear, bBirthMonth, bBirthHour int) BondResult

// --- 命格分析 ---
func ComputeYongShen(chart ChartResult) YongShenResult
func ComputeDayMasterStrength(elementCount map[Element]int, dayMaster Stem, monthBranch Branch) Strength
func ComputePattern(dayMaster Stem, monthBranch Branch, monthTenGodStem string) string
func DayMasterNameString(s Stem) string
func MonthBranchNameString(monthBranch Branch) string

// --- 调候 ---
func ComputeTiaohou(dayMaster Stem, monthBranch Branch) (TiaoHouResult, bool)

// --- 旺衰 ---
func ComputeWangShuaiMap(monthBranch Branch) map[string]string
func MonthWangShuai(elem Element, monthBranch Branch) string

// --- 流年/流月/流日/流时 ---
func ComputeLiunian(year int, dayMaster Stem, bz Bazi, currentDayun *DayunPillar) *LiunianResult
func ComputeLiuyue(year, month int, dayMaster Stem, bz ganzhi.Bazi) *LiuyueResult
func ComputeLiuri(date string, dayMaster Stem, bz ganzhi.Bazi, dayunPillar *Pillar, liunianPillar *Pillar) *LiuriResult
func ComputeLiushi(date string, hour int, dayMaster Stem, bz ganzhi.Bazi) *LiushiResult
func ComputeFuYinFanYin(flow Pillar, bz ganzhi.Bazi) []FuYinFanYinEntry

// --- 小运/小限 ---
func ComputeXiaoYun(gender Gender, dayMaster Stem, maxAge int) []XiaoYunPillar
func ComputeXiaoXian(gender Gender, maxAge int) []XiaoXianEntry

// --- 二十八宿 ---
func MansionForDay(dayPillar Pillar) DayMansion
func AllMansions() [28]DayMansion

// --- 旬空 ---
func XunName(dayPillar Pillar) string
func XunIndex(dayPillar Pillar) int

// --- 特殊属性 ---
func IsSelfHe(p Pillar) bool
func SelfHeName(p Pillar) string
func IsKuiGang(p Pillar) bool
func SanQiType(bz ganzhi.Bazi) string
func SanQiName(typ string) string

// --- 干支交互 ---
func AnalyzeStemRelation(a, b Stem) StemRelation
func AnalyzeBranchRelation(a, b Branch) BranchRelation
func AnalyzePillarWithBazi(pillar Pillar, bz Bazi) ([]StemRelation, []BranchRelation)

// --- 工具 ---
func ElementCountStrings(ec map[Element]int) map[string]int
func ElementThatGenerates(e Element) Element
func ElementThatControls(e Element) Element
```

### 4.5 神煞计算

`ComputeShenSha` 拆为 22 个独立函数，每种神煞一个私有函数，可独立测试：

```go
func ComputeShenSha(bz ganzhi.Bazi, dayMaster Stem, monthBranch Branch) [4][]ShenShaEntry
// 内部调用：addTianYi, addWenChang, addXueTang, addTianDe, addYueDe,
// addJiangXing, addJinYu, addYangRen, addYiMa, addTaoHua, addHongLuan,
// addTianXi, addGuChen, addHuaGai, addJieSha, addZaiSha, addYueEn,
// addXueRen, addKuiGang, addGouJiao, addTianLuoDiWang, addShiEDaBai
```

### 4.6 调候 (Tiaohou)

《穷通宝鉴》气候调候用神，优先于扶抑用神。

```go
type TiaoHouResult struct {
    Season string `json:"season"`
    Yong   string `json:"yong"`
    Xi     string `json:"xi"`
    Ji     string `json:"ji"`
    Detail string `json:"detail,omitempty"`
}

func ComputeTiaohou(dayMaster Stem, monthBranch Branch) (TiaoHouResult, bool)
```

日主(10天干) × 月令(12月) → 调候用神。当季节寒暖燥湿失衡时，以调候用神为准，不取扶抑用神。

### 4.7 城市查找

```go
type City struct { Name, NameZh, Country string; Lat, Lng, Population float64 }
func SearchCities(q string) []City
```

---

## 5. 前端交互

### 5.1 位置获取

三种方式，按优先级降级：

1. **浏览器自动定位** — `navigator.geolocation.getCurrentPosition()`，WGS-84 坐标，精度 < 50m。用户同意即获经纬度。无需后端参与。
2. **城市搜索** — `GET /api/bazi/cities?q=xxx` 服务端前缀搜索。返回匹配的城市列表（name 和 name_zh 均参与匹配，最多 50 条），前端展示下拉选项。
3. **手动输入** — 用户直接填经度值（高级模式）。

自动定位成功后，前端将经纬度填入请求的 `longitude` 字段，`timezone` 由前端按经度自动推算（`round(longitude/15)*15`）。

---

## 6. API 契约 → `docs/API.md`

HTTP 契约的**唯一权威源**为 `docs/API.md`（BaZi 段）。本节仅为设计引用，不重复定义。

| 端点 | 用途 | Auth |
|------|------|------|
| `POST /api/bazi/chart` | 排盘 | 无 |
| `POST /api/bazi/bond` | 合八字 | 无 |
| `GET /api/bazi/cities` | 城市列表 | 无 |

---

## 6. 错误

| Code | HTTP | 含义 |
|------|------|------|
| `invalid_birth_info` | 400 | 出生信息字段缺失、值非法或 JSON 格式错误 |
| `chart_required` | 400 | 合八字请求缺少 a 或 b |

所有 BaZi 错误统一使用 `invalid_birth_info` 码（包含年份超范围、月份/小时越界、性别非法等），handler 层 `validateBirthInfo` 返回英文 message 描述具体原因。

---

## 7. 输入校验

- `year`: 1900 ≤ year ≤ 2200
- `month`: 1 ≤ month ≤ 12
- `day`: 1 ≤ day ≤ 31（按月份校验有效日期）
- `hour`: 0 ≤ hour ≤ 23
- `minute`: 0 ≤ minute ≤ 59
- `longitude`: −180 ≤ longitude ≤ 180
- `timezone`: 默认 120.0（如未传）
- `gender`: "male" | "female"（合八字必填，排盘可选）
- `locale`: "en" | "zh-CN"，解读文本语言

---

## 8. 解读模板

解读文本不硬编码在 Engine 中。前端 `web/content/{locale}/mingli-templates.yaml` 通过 Eleventy `_data/mingliTemplates.js` 加载，嵌入 `window.MINGLI_TEMPLATES` 全局对象。模板按 locale 分文件（en / zh-CN）。

**日主解读**（`day_masters`，按 stem 数字 key 索引）：

```yaml
day_masters:
  "1":
    stem: "甲"
    element: "木"
    name: "甲木"
    portrait: "如参天大树——正直刚毅，有领导气质。甲木之人以德服人..."
  # ... "2" 乙木 至 "10" 癸水，每个含 stem / element / name / portrait
```

**合婚等级解读**（`match_levels`，按 level 字符串索引）：

```yaml
match_levels:
  "上等匹配":
    label: "上等匹配"
    desc: "两命局五行互相呼应——天干相合、地支相助..."
  "中等匹配":
    label: "中等匹配"
    desc: "整体契合度良好，关键维度有明确的连接..."
  "一般":
    label: "一般"
    desc: "吉凶参半——有些维度表现出和谐，有些则显示摩擦..."
  "不合":
    label: "不合"
    desc: "多个维度存在显著的五行冲突..."
```

前端 `mingli-chart-card.njk` 使用 `window.MINGLI_TEMPLATES.day_masters[String(chart.day_master)]` 显示日主名称和肖像；`mingli-match-card.njk` 使用 `window.MINGLI_TEMPLATES.match_levels[matchResult.level]` 显示等级标签和描述。

---

## 9. 与现有系统的关系

| | 现有 25types | 新增 BaZi |
|------|-------------|-----------|
| 理论基础 | 五元素心理学 | 子平八字 |
| 输入 | 30 题问卷 | 出生时间 + 地点 |
| 计算 | engine.ComputeD/P/Identity | bazi.ComputeSolarTime/ComputeBazi/ComputeChart + bazi.ComputeBond |
| 持久化 | assessments 表 | 不持久化（值对象） |
| 认证 | 混合（匿名/登录） | 全部匿名可用 |
| 解读 | types.yaml 类型描述 | mingli-templates.yaml 解读模板（前端加载） |
