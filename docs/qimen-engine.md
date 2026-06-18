# 奇门遁甲 Engine 设计

## 概述

奇门遁甲以时辰/日/月/年为驱动单位排盘，输出一个 9 宫盘（洛书九宫），每宫含地/天/人/神四层信息，附带克应、格局、旺衰、应期等分析层。

## 类型系统

### PalaceIndex — 九宫

洛书数 1-9 → 宫位。5 为中宫，寄坤二宫。

```
8 1 6   →   艮 坎 乾
3 5 7       震 中 兑
4 9 2       巽 离 坤
```

| 常量 | 洛书数 | 方位 |
|------|--------|------|
| PalaceKan | 1 | 坎 |
| PalaceKun | 2 | 坤 |
| PalaceZhen | 3 | 震 |
| PalaceXun | 4 | 巽 |
| PalaceZhong | 5 | 中 |
| PalaceQian | 6 | 乾 |
| PalaceDui | 7 | 兑 |
| PalaceGen | 8 | 艮 |
| PalaceLi | 9 | 离 |

### StarIndex — 九星

顺时针序：

| 常量 | 星名 | 五行 |
|------|------|------|
| StarTianPeng | 天蓬 | 水 |
| StarTianRui | 天芮 | 土 |
| StarTianChong | 天冲 | 木 |
| StarTianFu | 天辅 | 木 |
| StarTianQin | 天禽 | 土 |
| StarTianXin | 天心 | 金 |
| StarTianZhu | 天柱 | 金 |
| StarTianRen | 天任 | 土 |
| StarTianYing | 天英 | 火 |

### DoorIndex — 八门

顺时针序：

| 常量 | 门名 | 五行 | 吉凶 |
|------|------|------|------|
| DoorXiu | 休 | 水 | 吉 |
| DoorSheng | 生 | 土 | 吉 |
| DoorShang | 伤 | 木 | 凶 |
| DoorDu | 杜 | 木 | 平 |
| DoorJing | 景 | 火 | 平 |
| DoorSi | 死 | 土 | 凶 |
| DoorJingMen | 惊 | 金 | 凶 |
| DoorKai | 开 | 金 | 吉 |

### SpiritIndex — 八神

| 常量 | 阳遁 | 阴遁 |
|------|------|------|
| SpiritZhiFu | 值符 | 值符 |
| SpiritTengShe | 螣蛇 | 螣蛇 |
| SpiritTaiYin | 太阴 | 太阴 |
| SpiritLiuHe | 六合 | 六合 |
| SpiritGouChen | 勾陈 | 白虎 |
| SpiritZhuQue | 朱雀 | 玄武 |
| SpiritJiuDi | 九地 | 九地 |
| SpiritJiuTian | 九天 | 九天 |

### Palace — 单宫信息

```go
type Palace struct {
    EarthStem  ganzhi.Gan  // 地盘干
    HeavenStem ganzhi.Gan  // 天盘干
    Star       StarIndex   // 九星
    Door       DoorIndex   // 八门
    Spirit     SpiritIndex // 八神
    HiddenStem ganzhi.Gan  // 暗干
}
```

### Chart — 全盘输出

```go
type Chart struct {
    Pan              pan               // 排盘（天地人神四层）
    StemInteractions [9]StemInteraction // 十干克应
    DoorInteractions [9]DoorInteraction // 八门克应
    StarInteractions [9]StarInteraction // 九星克应
    WangShuai        [9]WangShuai       // 旺衰
    MenPo            []PalaceIndex      // 门迫
    MenZhi           []PalaceIndex      // 门制
    Patterns         []Pattern          // 格局
    YingQi           YingQi             // 应期
}
```

分析层类型说明：

| 类型 | 含义 |
|------|------|
| StemInteraction | 十干克应：地盘干+天盘干的组合吉凶 |
| DoorInteraction | 八门克应：门+宫的组合意义 |
| StarInteraction | 九星克应：星+宫的组合吉凶 |
| WangShuai | 旺衰：星在宫的旺/相/休/囚/废状态 |
| MenPo | 门迫：门克宫 |
| MenZhi | 门制：宫克门 |
| Pattern | 格局：全盘层面的吉凶格局（如三奇得使、天遁、伏吟等） |
| YingQi | 应期：马星/空亡/值符值使的应期判断 |

### 五行常量

```
WuxingMu=1, WuxingHuo=2, WuxingTu=3, WuxingJin=4, WuxingShui=5
```

## 排盘算法

### 定局 (determineJuShu)

输入年月日和日柱干支，输出局数 + 阴阳遁。24 节气各有上/中/下元三局，根据日柱在 60 甲子中的位置判定三元归属。

阳遁 (冬至→夏至)、阴遁 (夏至→冬至)。

### 四层排盘 (computePan)

1. **地盘** (placeDiPan): 三奇六仪 (戊己庚辛壬癸丁丙乙)，戊起局数宫，阳顺阴逆排布
2. **值符值使** (findDuty): 时柱找旬首 → 六仪 → 地盘宫 → 该宫星/门为值符值使
3. **天盘** (placeTianPan): 值符星加时干宫，其余八星顺时针排布，天盘干随星携带
4. **人盘** (placeRenPan): 值使门加时支宫，其余七门顺时针排布
5. **神盘** (placeShenPan): 值符神跟值符星走，阳顺阴逆排布八神
6. **暗干** (placeAnGan): 时干加值使门宫，八干顺排
7. **马星** (findMaXing): 时支定马星：寅午戌马在申、亥卯未马在巳、申子辰马在寅、巳酉丑马在亥
8. **空亡** (findKongWang): 时柱旬空 → 两空亡地支所在宫

### 盘类型 (kind)

| kind | 驱动柱 | 说明 |
|------|--------|------|
| shi | 时柱 | 时盘（默认） |
| ri | 日柱 | 日盘 |
| yue | 月柱 | 月盘 |
| nian | 年柱 | 年盘 |

排盘算法相同，仅驱动柱不同。局数始终按日柱定。

## Public API

| 函数 | 说明 |
|------|------|
| `ComputeChart(st SolarTime, kind string) → Chart` | 奇门排盘 + 全部分析层（编排入口，api.go） |

编排层 `api.go` 收 `SolarTime` → `ComputeBazi` → 引擎 `computeChart(bz Bazi, kind string, y, m, d int)` 收精确实体。

### HTTP Route

```
POST /api/qimen/pan
```

请求体：

```json
{
    "solar_time": "2026-06-16T14:30:00+08:00",
    "kind": "shi"
}
```

`kind` 可选值: `"shi"` / `"ri"` / `"yue"` / `"nian"`，默认 `"shi"`。
