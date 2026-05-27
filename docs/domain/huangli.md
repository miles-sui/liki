# Dates（择日）— 实现规格

> 日柱对敲日历干支，建除十二神定宜忌，地支看冲合，天干定十神。传统流派驱动，算清楚，标出来，用户自己选。

---

## 1. 聚合边界

```
BirthInfo ──→ MingliChart ──→ 日柱（天干日主 + 地支）
                    │             │
                    │        流年→太岁
                    │             │
                    ▼             ▼
              指定月份每一天的干支
                    │
              ┌─────┼─────┬──────┐
              ▼     ▼     ▼      ▼
           建除   十神   冲合   神煞
           宜忌   (vs  (vs日柱 (吉/凶)
           标记   日主) &太岁)
              │     │     │      │
              └─────┼─────┼──────┘
                    ▼     ▼
                 每日标记 → 日历视图
```

| 概念 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| CalendarRequest | 值对象 | 请求级 | 出生信息 + 年月 + 事件类型 |
| DayPillar | 计算值 | 从 solar term 派生 | 任意公历日 → 干支日柱 |
| DayMark | 值对象 | 实时计算 | 单日完整标注 |
| Calendar | 值对象 | 实时计算 | 整月每日的 DayMark 列表 |

**不建表理由**：所有计算从 MingliChart 和日历数据确定性派生，遵循 DPD-013。择日历史记录由 Reports 系统管理。

---

## 2. 领域类型

### 2.1 查询（Query）— 公开数据

```go
type DayEntry struct {
    Date       string            `json:"date"`
    DayPillar  DayPillarInfo     `json:"day_pillar"`
    JianChu    string            `json:"jian_chu"`
    Suitable   bool              `json:"suitable"`
    Marks      []string          `json:"marks"`
    Warnings   []string          `json:"warnings"`
    HuangDao   HuangDaoStar      `json:"huangdao"`
    Directions DayStemDirections `json:"directions"`
    Taboos     DayTaboos         `json:"taboos"`
    Mansion    MansionEntry      `json:"mansion"`
}
```

Query 端点返回公开黄历数据（日柱、建除、黄道、方位、禁忌、值宿），不含个人出生信息。支持单日或整月查询。

### 2.2 对敲（Bond）— 个性化交叉比对

```go
type BondDayEntry struct {
    DayEntry                                              // 嵌入完整 DayEntry
    GanRelation    string `json:"gan_relation"`          // 日干 vs 日主 → 十神名
    ZhiRelation    string `json:"zhi_relation"`          // 日支 vs 出生日支 → 合/冲/刑/害
    TaiSuiRelation string `json:"tai_sui_relation"`      // 日支 vs 年支(太岁) → 合/冲/刑/害
}
```

Bond 端点 = Query 数据 + 个人出生信息交叉比对。传入出生信息和日期/月份，返回黄历数据叠加个人干支关系和地支冲合标注。

### 2.3 领域输入

```go
type BirthInfo struct {
    Year      int     `json:"year"`
    Month     int     `json:"month"`
    Day       int     `json:"day"`
    Hour      int     `json:"hour"`
    Minute    int     `json:"minute"`
    Longitude float64 `json:"longitude"`
    Timezone  float64 `json:"timezone"`
    IsDST     bool    `json:"is_dst"`
    Gender    string  `json:"gender"`
}
```

Query 不传 BirthInfo，Bond 传 BirthInfo 做个性化计算。

### 2.4 事件类型

```go
type EventType string

const (
    EventWedding   EventType = "wedding"    // 嫁娶/结婚
    EventEngage    EventType = "engage"     // 订婚/领证
    EventOpen      EventType = "open"       // 开业/开张
    EventSign      EventType = "sign"       // 签约/合同
    EventMove      EventType = "move"       // 搬家/入宅
    EventTravel    EventType = "travel"     // 出行/远行
    EventBuild     EventType = "build"      // 动土/装修
    EventExam      EventType = "exam"       // 考试/面试
    EventMedical   EventType = "medical"    // 手术/就医
    EventFuneral   EventType = "funeral"    // 丧葬
    EventGeneral   EventType = "general"    // 通用 — 不按建除过滤
)
```

### 2.5 输出包装

Query 和 Bond 统一使用 wrapper 包装单日或整月结果：

```go
type QueryOutput struct {
    Days      []DayEntry    `json:"days"`
    YearMonth string        `json:"year_month"`
}

type BondOutput struct {
    Days      []BondDayEntry `json:"days"`
    YearMonth string         `json:"year_month"`
}
```

**没有排序键。** 日历按日期顺序排列。用户自己看标签判断。

---

## 3. 行为规则

### 3.1 喜神/财神/福神方位 (Fangwei)

源出《协纪辨方书》。日干 → 三神方位，纯查表，不涉及个人八字。

| 日干 | 喜神 | 财神 | 福神 |
|------|------|------|------|
| 甲 | 东北 | 东北 | 东南 |
| 乙 | 西北 | 东北 | 东南 |
| 丙 | 西南 | 正西 | 西北 |
| 丁 | 正南 | 正西 | 正东 |
| 戊 | 东南 | 正北 | 正南 |
| 己 | 东北 | 正北 | 正南 |
| 庚 | 西北 | 正东 | 西南 |
| 辛 | 西南 | 正东 | 西南 |
| 壬 | 正南 | 正南 | 西北 |
| 癸 | 东南 | 正南 | 正西 |

### 3.2 建除十二神

按月支（查询月份对应的农历月支）确定该月第一个建日，按日支顺序排布十二神。

```
月支 → 建日地支：
寅月(正月) → 寅日为建
卯月(二月) → 卯日为建
...依此类推（月支 = 建日地支）
```

十二神序列：**建 → 除 → 满 → 平 → 定 → 执 → 破 → 危 → 成 → 收 → 开 → 闭**

建除宜忌表（传统规则）：

| 神 | 宜 | 忌 |
|----|-----|------|
| 建 | 出行 | 嫁娶、动土 |
| 除 | 治病、扫除 | 嫁娶、开业 |
| 满 | 祭祀 | 嫁娶、签约 |
| 平 | 修造 | 出行 |
| 定 | 签约、开市 | 动土 |
| 执 | 捕捉 | 出行、搬家 |
| 破 | — | 万事不宜 |
| 危 | 安床 | 开业、出行 |
| 成 | 嫁娶、开业、入宅、签约 | 诉讼 |
| 收 | 纳财 | 出行、嫁娶 |
| 开 | 开业、嫁娶、出行 | 动土、安葬 |
| 闭 | 埋葬 | 开业、签约 |

**事件类型 → 宜用建除神：**

| 事件 | 宜用神 | 忌用神 |
|------|--------|--------|
| 嫁娶 | 成、开 | 破、建、除、满、收 |
| 领证 | 定、成 | 破、满 |
| 开业 | 成、开 | 破、除、危、闭 |
| 签约 | 定、成 | 破、满、闭 |
| 搬家 | 成、平 | 破、执 |
| 出行 | 建、开 | 破、平、危、收 |
| 动土 | 平 | 破、定、开 |
| 考试 | 定、成 | 破 |
| 就医 | 除 | 破 |
| 丧葬 | 闭 | 破、开 |

- 建除 = 破 → `warnings` 标 "破日，万事不宜"
- 建除在忌用列表中 → `suitable = false`，`marks` 标建除神名
- 建除在宜用列表中 → `suitable = true`，`marks` 标 "成日宜嫁娶" 等

纯查表，传统公知数据，存 `internal/mingli/huangli/data/jianchu.yaml`。

### 3.3 天干对敲（十神判定）

候选日天干 vs 日主天干 → 十神。纯查表，已有 `engine/mingli.go`。

```
日主 = 甲（木）
日干 = 甲 → 比肩    日干 = 己 → 正财
日干 = 乙 → 劫财    日干 = 庚 → 七杀
日干 = 丙 → 食神    日干 = 辛 → 正官
日干 = 丁 → 伤官    日干 = 壬 → 偏印
日干 = 戊 → 偏财    日干 = 癸 → 正印
```

不评分。十神名直接放在 `GanRelation` 中展示。

### 3.4 地支对敲（vs 日柱）

日支 vs 日柱地支：

| 关系 | 判定 | marks/warnings |
|------|------|------|
| 六合 | 子丑、寅亥、卯戌、辰酉、巳申、午未 | marks "六合日" |
| 三合半 | 申子、亥卯、寅午、巳酉 | marks "三合半" |
| 无关系 | — | — |
| 相刑 | 子卯、寅巳申、丑戌未、辰午酉亥（自刑） | warnings "刑日柱" |
| 六害 | 子未、丑午、寅巳、卯辰、申亥、酉戌 | warnings "害日柱" |
| 六冲 | 子午、丑未、寅申、卯酉、辰戌、巳亥 | warnings "冲日柱" |

### 3.5 地支对敲（vs 太岁）

日支 vs 流年地支（太岁）：

规则同 §3.3。关系标注到 `TaiSuiRelation`，冲/刑/害进 `warnings`，合进 `marks`。

**冲太岁 → warnings 标 "冲太岁"。** 太岁为一年之主，冲犯为大忌。

### 3.6 神煞

日支 + 月支组合判定当日神煞，纯查表。

- 吉神：天乙贵人、天德、月德、月恩、天赦 → `marks` 标注
- 凶神：劫煞、灾煞、月煞、月破、四废 → `warnings` 标注

仅作标记，参考 `engine/shensha.go`。

### 3.7 完整流水线

```
① 排盘 + 定太岁           ② 逐日计算                    ③ 返回日历
────────────── → ─────────────────────── → ─────────────────
八字排盘取日柱      每月 28-31 天逐天：           按日期顺序排列
流年取太岁          建除→marks/warnings          不排序不推荐
                    天干十神→GanRelation         用户自己看
                    日支 vs 日柱→marks/warnings
                    日支 vs 太岁→marks/warnings
                    神煞→marks/warnings
```

### 3.8 前端展示

日历视图，每天一个格子，标注当日标签：

```
       2026 年 6 月 — 嫁娶择日
┌──────┬──────┬──────┬──────┬──────┬──────┬──────┐
│  一  │  二  │  三  │  四  │  五  │  六  │  日  │
├──────┼──────┼──────┼──────┼──────┼──────┼──────┤
│      │      │      │      │      │      │  1   │
│      │      │      │      │      │      │ 平日  │
│      │      │      │      │      │      │ 正印  │
├──────┼──────┼──────┼──────┼──────┼──────┼──────┤
│  2   │  3   │  4   │  5   │  6   │  7   │  8   │
│ 定日  │ 执日  │ 破日⚠️│ 危日  │ 成日✓ │ 收日  │ 开日✓ │
│ 劫财  │ 食神  │ 七杀  │ 偏印  │ 正财  │ 偏财  │ 比肩  │
│      │      │万事不宜│      │宜嫁娶 │      │宜嫁娶 │
│      │      │      │      │六合日🌟│      │天乙贵人│
├──────┼──────┼──────┼──────┼──────┼──────┼──────┤
│ ...  │ ...  │ ...  │ ...  │ ...  │ ...  │ ...  │
```

用户看到全貌，自己判断。

---

## 4. Engine 函数签名

```go
// engine/dates.go

// 公历日 → 干支日柱（基于 solar term 数据库）
func LookupDayPillar(date string) (DayPillar, error)

// 流年太岁地支
func TaiSui(year int) Branch

// 建除十二神
func LookupJianChu(date string) string

// 建除是否宜此事件
func JianChuSuitable(jianChu string, eventType EventType) (suitable bool, marks []string, warnings []string)

// 天干对敲 → 十神名
func EvaluateGan(dayGan Stem, dayMaster Stem) string

// 地支对敲 → 关系名 + marks/warnings
func EvaluateZhi(dayZhi Branch, refZhi Branch, label string) (relation string, marks []string, warnings []string)

// 当日神煞 → marks/warnings
func LookupShenSha(date string, dayZhi Branch) (marks []string, warnings []string)

// 单日黄历查询（公开数据，无出生信息）
func QueryDate(dateStr string, eventType string) (DayEntry, error)

// 整月黄历查询
func QueryMonth(yearMonth string, eventType string) ([]DayEntry, error)

// 单日对敲（黄历 + 个人出生信息交叉比对）
func CrossDate(birth BirthInfo, dateStr string, eventType string) (BondDayEntry, error)

// 整月对敲
func CrossMonth(birth BirthInfo, yearMonth string, eventType string) ([]BondDayEntry, error)
```

---

## 5. API 契约 → `docs/API.md`

HTTP 契约的**唯一权威源**为 `docs/API.md`（Huangli 段）。本节仅为设计引用。

| 端点 | 用途 | Auth |
|------|------|------|
| `GET /api/huangli/query` | 单日/整月黄历查询 `?date=` 或 `?month=`，可选 `event_type` | 🔓 |
| `POST /api/huangli/bond` | 对敲：黄历 + 个人出生信息交叉比对 | 🔓 |
| `GET /api/huangli/jieqi` | 节气深度 + 人元司令 | 🔓 |


Query/Bond 均支持单日（`date`）或整月（`month`）范围。Query 为公开数据（无 PII），Bond 含出生信息走 POST。完整择日方案（详细解读 + 时辰推荐）通过 Reports 系统生成：`POST /api/reports` with `scene=huangli`。

---

## 6. 错误

| Code | HTTP | 含义 |
|------|------|------|
| `invalid_request` | 400 | birth_info 缺失或非法；year_month 格式错误 |
| `invalid_birth_info` | 400 | 出生信息字段缺失或非法（复用 BaZi 校验） |
| `not_found` | 404 | year_month 超出日历数据覆盖范围 |

---

## 7. 输入校验

- `birth_info`: 与 `POST /api/bazi/chart` 相同校验规则
- `year_month`: "YYYY-MM" 格式，月份 01-12
- `event_type`: 可选，不传则 `general`

---

## 8. 与现有系统的关系

| | BaZi | Dates |
|------|------|------|
| 输入 | 出生时间 + 地点 | 出生时间 + 年月 + 事件类型 |
| 核心计算 | engine.ComputeChart | engine.MarkDay |
| 静态数据 | cities.json | solar_term 表（已有）+ jianchu.yaml（新增） |
| 流派依据 | 子平八字 | 建除十二神 + 日柱冲合 + 十神 + 太岁 |
| 持久化 | 不持久化（值对象） | 不持久化引擎结果；择日历史走 Reports |
| 解读 | mingge 免费直接返回 | 完整方案走 Reports SSE |
| 认证 | 排盘匿名可用 | 全部需登录（🔒） |
