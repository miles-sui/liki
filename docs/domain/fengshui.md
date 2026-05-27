# FengShui（风水）— 实现规格

> 二十四山、三元九运、命卦。纯计算值对象，不建表，实时计算。

理论依据：`docs/theory/`（待补充）

---

## 1. 聚合边界

```
三元九运 ──→ 当前运
二十四山 ──→ 山向表
BirthInfo ──→ MingGua（命卦）
```

| 概念 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| MingGua | 值对象 | 实时计算 | 出生年 + 性别 → 命卦 |
| SanYuan | 值对象 | 实时计算 | 三元九运周期 |
| 24Shan | 值对象 | 查表 | 二十四山方位 |
| Trigrams | 值对象 | 静态 | 八卦信息 |

**不建表理由**：所有计算从出生信息和静态表确定性派生，遵循 DPD-013。

---

## 2. 领域类型

```go
type Trigram int // 1=坎 2=坤 3=震 4=巽 5=乾 6=兑 7=艮 8=离
```

## 3. 行为规则

### 3.1 八宅命卦 (Ming Gua)

出生年 + 性别 → 命卦数 (1-9)。男命 (100 - shortYear) % 9，女命 (shortYear - 4) % 9。余 5 男 → 2(坤)，女 → 8(艮)。0 → 9。

### 3.2 大游年诀 (Ba Zhai Directions)

命卦数 → 四吉四凶方。纯查表，不评分。

| 卦 | 生气 | 天医 | 延年 | 伏位 | 祸害 | 五鬼 | 六煞 | 绝命 |
|----|------|------|------|------|------|------|------|------|
| 坎1 | 6乾 | 8艮 | 9离 | 1坎 | 2坤 | 3震 | 7兑 | 4巽 |
| 坤2 | 9离 | 1坎 | 6乾 | 2坤 | 8艮 | 7兑 | 4巽 | 3震 |
| 震3 | 8艮 | 6乾 | 1坎 | 3震 | 7兑 | 2坤 | 9离 | 4巽 |
| 巽4 | 1坎 | 9离 | 8艮 | 4巽 | 5中 | 6乾 | 3震 | 2坤 |
| 乾6 | 4巽 | 3震 | 2坤 | 6乾 | 8艮 | 7兑 | 9离 | 1坎 |
| 兑7 | 3震 | 4巽 | 6乾 | 7兑 | 8艮 | 9离 | 1坎 | 2坤 |
| 艮8 | 2坤 | 7兑 | 3震 | 8艮 | 9离 | 1坎 | 5中 | 6乾 |
| 离9 | 8艮 | 3震 | 4巽 | 9离 | 6乾 | 5中 | 2坤 | 1坎 |

### 3.3 年紫白飞星 (Annual Flying Stars)

下元甲子(1984)七赤入中，每年中宫星递减 1。九星按洛书轨迹飞布九宫。详见 `fengshui_feixing.go`。

### 3.4 合参 (Hecan)

命卦 + 八宅 + 飞星 + 八字用神三层并列输出，不评分不分析。`ComputeHeCan(mingGuaNum, yongShen, xiShen, year)` 组装四个结果。

## 4. Engine 函数

- `ComputeMingGua(gender, year)` → `MingGuaResult`
- `ComputeYearStars(year)` → `YearStarResult`
- `EightMansionDirs(guaNum)` → `(auspicious [4]int, inauspicious [4]int)`
- `BaZhaiDirectionsForGua(guaNum)` → `BaZhaiDirections`
- `ComputeHeCan(mingGuaNum, yongShen, xiShen, year)` → `HeCanResult`
- `AllSanYuanYun()` → `[]SanYuanYun`
- `All24Shan()` → `[]Shan24`
- `AllTrigrams()` → `[9]Trigram`

## 5. API 契约

参见 `docs/API.md` — FengShui 部分。
