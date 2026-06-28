# 六爻 Engine 设计

## 概述

六爻（纳甲筮法）以摇钱起卦为起点，输出 64 卦中一个本卦，若爻位有变则产出变卦。每爻含纳支、六亲、世应、六兽四层信息，附带用神、旺衰、日建关系、应期等分析层。

## 类型系统

### YaoType — 爻类型

| 值 | 常量 | 含义 |
|----|------|------|
| 6 | LaoYin | 老阴 ⚋→⚊ 动爻 |
| 7 | ShaoYang | 少阳 ⚊ 静爻 |
| 8 | ShaoYin | 少阴 ⚋ 静爻 |
| 9 | LaoYang | 老阳 ⚊→⚋ 动爻 |

### LiuQin — 六亲

| 常量 | 名称 | 关系 |
|------|------|------|
| QinXiongDi | 兄弟 | 同我 |
| QinFumu | 父母 | 生我 |
| QinZiSun | 子孙 | 我生 |
| QinGuanGui | 官鬼 | 克我 |
| QinQiCai | 妻财 | 我克 |

### LiuShou — 六兽

| 常量 | 名称 |
|------|------|
| ShouQingLong | 青龙 |
| ShouZhuQue | 朱雀 |
| ShouGouChen | 勾陈 |
| ShouTengShe | 螣蛇 |
| ShouBaiHu | 白虎 |
| ShouXuanWu | 玄武 |

从日干起青龙，按甲乙起青龙→朱雀→勾陈→螣蛇→白虎→玄武顺序排布。

### YongShen — 用神

| 常量 | 名称 | 问事范围 |
|------|------|----------|
| YongFumu | 父母 | 长辈、文书、房屋 |
| YongXiongDi | 兄弟 | 朋友、同事、竞争 |
| YongGuanGui | 官鬼 | 工作、官运、疾病 |
| YongQiCai | 妻财 | 财运、妻子、物品 |
| YongZiSun | 子孙 | 子女、健康、宠物 |
| YongShiYao | 世爻 | 自身、求问人 |

### WangShuai — 旺衰

| 常量 | 名称 |
|------|------|
| WSWang | 旺 |
| WSXiang | 相 |
| WSXiu | 休 |
| WSQiu | 囚 |
| WSSi | 死 |

### Line — 单爻

```go
type Line struct {
    Position int           // 1-6, 初爻→上爻
    Type     YaoType       // 老阴/少阳/少阴/老阳
    Gan      ganzhi.Gan    // 纳甲天干（预留）
    Zhi      ganzhi.Zhi    // 纳支
    Wuxing   ganzhi.Wuxing // 地支五行
    LiuQin   LiuQin        // 六亲
    ShiYing  string        // "世"/"应"/""
    LiuShou  LiuShou       // 六兽
}
```

### FuShen — 伏神

```go
type FuShen struct {
    Position int    // 爻位 1-6
    LiuQin   LiuQin // 伏神六亲
    Zhi      string // 伏神地支
}
```

### YongShenResult — 用神分析

```go
type YongShenResult struct {
    Type     YongShen // 用神类型
    Position int      // 用神爻位 1-6, 0 为不上卦
    FuShen   *FuShen  // 伏神，不上卦时填充
}
```

### DayRelation — 日建关系

```go
type DayRelation struct {
    Relation string // 生/扶/克/冲/合/平
    Strength string // 旺/衰/平
}
```

### YingQi — 应期

```go
type YingQi struct {
    YongShen   string // 用神名称
    DongYaoPos int    // 动爻位置
    YingTime   string // 应期描述
    Assessment string // 综合判断
}
```

### Chart — 卦盘 + 分析

```go
type Chart struct {
    Name         string         // 卦名
    BenGua       int            // 本卦索引 0-63
    BianGua      int            // 变卦索引，0=无变
    Palace       string         // 宫名（乾兑离震巽坎艮坤）
    PalaceWuxing ganzhi.Wuxing  // 宫五行
    Lines        [6]Line        // 本卦六爻
    BianLines    [6]Line        // 变卦六爻
    DayGan       ganzhi.Gan     // 日干
    DayZhi       ganzhi.Zhi     // 日支
    MonthZhi     ganzhi.Zhi     // 月建
    DongYao      []int          // 动爻位置 1-6

    YongShen     YongShenResult // 用神分析（含飞伏）
    WangShuai    [6]WangShuai   // 月建旺衰
    DayRelations [6]DayRelation // 日建关系
    YingQi       YingQi         // 应期推算
}
```

## 64 卦数据

八宫排序，每宫 8 卦（本宫、一世、二世、三世、四世、五世、游魂、归魂）。纳支表按八宫固定。

内部类型：`guaIndex`、`guaMeta`、`guaTable`、`naZhiTable`、`palaceNames`、`palaceWuxing`。

## 排盘流程 (ComputeChart)

1. **起卦**: 默认随机 3 枚铜钱 6 次（6/7/8/9），支持 `fixed` 手动指定
2. **排盘**: 本卦 → 变卦（翻转老阴/老阳爻位）→ 装卦（纳支+六亲+世应+六兽）
3. **月建日建**: 从用事时间的八字月柱和日柱计算
4. **取用神**: 根据问事类型找对应六亲的爻，不上卦则找伏神
5. **旺衰**: 月建对六爻的旺/相/休/囚/死
6. **日建关系**: 日支对六爻的生/扶/克/冲/合/平
7. **应期**: 动爻临值、静爻待冲、伏神待冲出

## Public API

| 函数 | 说明 |
|------|------|
| `ComputeChart(st tianwen.SolarTime, yongShen YongShen, fixed [6]int) → Chart` | 起卦+排盘+用神+旺衰+日建关系+应期（编排入口，api.go） |

编排层 `api.go` 收 `tianwen.SolarTime` → `ComputeBazi` → 引擎 `computeChart(bz ganzhi.Bazi, yongShen YongShen, fixed [6]int)` 收精确实体。`computeGuaPan` 收 `ganzhi.Zhu` 而非 `time.Time`。

## JSON-RPC Method

```
liuyao.chart
```

请求体：

```json
{
    "solar_time": "2026-06-16T14:30:00+08:00",
    "yong_shen": "官鬼",
    "fixed": [6, 7, 8, 9, 7, 6]
}
```

- `solar_time`: 用事时间，从 `bazi.chart` 获取的 `solar_time` 字段
- `yong_shen`: "父母"/"兄弟"/"官鬼"/"妻财"/"子孙"/"世爻"（默认 "世爻"）
- `fixed`: 可选，手动指定 6 个爻（6/7/8/9），不传则为随机起卦
