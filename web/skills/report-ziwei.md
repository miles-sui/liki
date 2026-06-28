# 紫微斗数报告模板

基于 chart、daxian、liunian、liuyue、liuri、bond API 返回数据生成紫微斗数报告。

## 数据来源

- chart（`ziwei.chart`）：
  - `palaces[]` — 十二宫。每宫含 `index`/`name`/`gan`/`zhi`/`is_body_palace`/`stars[]`。
  - `stars[]` — 星曜。每星含 `star`(索引)/`name`/`is_major`/`si_hua`(有则禄权科忌)/`brightness`(庙旺利平陷)。
  - `ming_gong` — 命宫索引（恒为0）、`shen_gong` — 身宫索引、`ju_shu`/`ju_shu_name` — 局数、`ziwei_pos` — 紫微星宫位。
  - `si_hua` — 本命四化（starIndex→禄权科忌）、`year_gan`/`hour_zhi` — 年干/时支、`birth_year`/`gender`。
  - `patterns[]` — 格局。每项 `name`/`description`/`score`（0下格、1中格、2上格）。
- daxian（`ziwei.daxian`）：`[{start_age, end_age, palace, name}]`，12项，每项为十年大限。
- liunian（`ziwei.liunian`）：`ming_gong`/`ming_gong_name`、`si_hua`（流年四化）、`si_hua_palace`（各化星落宫）、`minor_stars`（流年辅星位置）。
- liuyue（`ziwei.liuyue`）：`ming_gong`/`ming_gong_name`/`si_hua`。
- liuri（`ziwei.liuri`）：`ming_gong`/`ming_gong_name`/`si_hua`。
- bond（`ziwei.bond`）：`a_into_b`/`b_into_a`（双方命宫互入）、`star_cross[]`（`star`/`from_a`/`into_b`）、`sihua_cross[]`（`star`/`type`/`into_b`）。

只引用数据中实际存在的字段。若某字段数据中不存在，跳过该分析维度，不编造。

## 报告结构

### 一、命盘总览

以表格列出十二宫（`palaces[]`）：宫名/宫干/宫支/身宫标记/主要星曜及亮度。概述：
- 命宫主星（`palaces[ming_gong].stars` 中 `is_major` 的星）及亮度
- 身宫所在（`shen_gong` 对应 `palaces[shen_gong].name`）及后天重心
- `ju_shu_name` 五行局数
- 三方四正（命宫/财帛/官禄/迁移）星曜是否有力

### 二、命宫详解

命宫为全盘核心，单独展开：
- 坐哪些主星（`is_major`）和辅星，各星 `brightness`
- 命宫之星有无 `si_hua`，若化禄/权/科/忌各有什么含义
- 命宫干支与五行局数的关系
- 紫微星位置（`ziwei_pos`）与命宫的关系（同宫/三合/对宫/六合等）
- 性格特质、天赋优势、注意事项

### 三、十二宫逐一解读

每宫（命宫已详述，从兄弟宫开始）依次分析：
- 主星 `name` + `brightness` → 该宫基本含义
- 辅星 → 补充或消减
- `si_hua` → 该宫若为化星所在，说明动态影响
- 吉星（左辅右弼天魁天钺文昌文曲禄存）加分
- 煞星（擎羊陀罗火星铃星地空地劫）需注意，结合主星看是否化解
- 重点宫位：夫妻（婚姻）、财帛（财运）、官禄（事业）、迁移（外出）

若篇幅过长，父母/兄弟/交友/田宅可概述，命宫/夫妻/财帛/官禄/迁移详述。

### 四、四化飞布

基于 `chart.si_hua`（生年四化）：
- 逐条列出哪星化禄/权/科/忌，落在哪个宫（查 `palaces[]` 中该星所在）
- 化禄 → 该宫领域有福气和机会
- 化忌 → 该宫领域为人生课题，需用心经营
- 四化之间的联动：化禄和化忌的宫位如何平衡

### 五、格局

若 `patterns[]` 非空：
- 列出各项 `name` + `description`
- `score` 为 2 → 上格，积极展开
- `score` 为 1 → 中格，有条件的吉格
- `score` 为 0 → 下格，如实描述但不过度渲染

常见格局理解：杀破狼主动荡开创、紫府同宫主富贵、日月并明主贵气、机月同梁主幕僚。若 `patterns` 为空，不提格局。

### 六、大限

基于 daxian：
- 列出 `[{start_age, end_age, name}]` 当前及未来 1-2 个大限
- 当前大限所在的宫位（`name`），该宫原局星曜在大限中被激活
- 大限宫位与本命四化的互动
- 关键节点：何时入强宫（命/官禄/财帛）、何时入弱宫（疾厄/交友）

### 七、流年

若数据含 `liunian`：
- 流年命宫 `ming_gong_name` + 该宫原局星曜
- 流年四化（`liunian.si_hua`）与生年四化的叠加效应
- 流年辅星（`minor_stars`）位置及影响
- 各化星落宫（`si_hua_palace`）的领域提示

若含 `liuyue`/`liuri` → 按"流月/流日命宫+四化"简要列出，不逐宫展开。

## 边界处理

- `palaces` 为空 → 无法生成报告，提示先排盘
- 某宫 `stars` 为空 → 说明该宫无主星（借对宫看），不编造星
- `patterns` 为空 → 跳过格局章节
- `si_hua` 为空 → 跳过四化章节
- daxian 数据为空 → 跳过大限章节
- 流年/流月/流日数据不含 → 跳过对应章节
- 命宫无主星（空宫）→ 借对宫论，说明"借迁移宫之X星安命宫"
- 某星 `brightness` 为空 → 不提亮度，只论星性
- 煞星集中某宫 → 如实指出但不恐吓，建议以吉星方位化解
