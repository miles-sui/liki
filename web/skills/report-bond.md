# 合盘报告模板

双方各自的八字解读参考 https://liki.hk/skills/report-chart.md，合盘报告只引述结论（日主、身强身弱、格局），不做完整展开。

## 数据结构

- `chart_a`, `chart_b` — 双方八字排盘（字段参见 https://liki.hk/skills/report-chart.md 数据结构章节）
- `zhu_cross` — 四柱交互
  - `pairs`: [{a_zhu, b_zhu, a_stem, b_stem, a_branch, b_branch, stem, branch}]
    - `a_zhu`, `b_zhu`: 哪一柱（`"年"`/`"月"`/`"日"`/`"时"`）
    - `a_stem`, `b_stem`: 双方天干
    - `a_branch`, `b_branch`: 双方地支
    - `stem`: {stem_a, stem_b, type, relation} — 天干关系。`type` 为 `"合"`/`"生"`/`"克"`/`"相同"`，`relation` 为解读文字
    - `branch`: {branch_a, branch_b, type, detail} — 地支关系。`type` 为 `"六合"`/`"三合"`/`"三会"`/`"六冲"`/`"六害"`/`"三刑"`/`"暗合"`/`"相同"`，`detail` 为解读文字
- `shi_shen_cross` — 十神交互
  - `a_to_b`: {nian_stem, yue_stem, ri_stem, shi_stem} — 甲方各柱天干对乙方日主的十神
  - `b_to_a`: 同上，乙方对甲方
- `nayin_cross` — 纳音与五行交互
  - `pairs`: [{a_zhu, b_zhu, a_na_yin, b_na_yin, relation}] — 同柱纳音关系。`relation` 为 `"相生"`/`"相克"`/`"相同"`
  - `elements`: {a: {木,火,土,金,水}, b: {木,火,土,金,水}} — 双方五行计数
  - `yong_shen`: {a: {yong, ji, yong_in_other, ji_in_other}, b: {…}} — 用忌神交叉。`yong_in_other` 为己方用神在对方八字中的出现次数，`ji_in_other` 为忌神出现次数
- `shensha_cross` — 神煞交互
  - `tian_yi`: {a_in_b, b_in_a} — 天乙贵人（对方是否为己方贵人）
  - `lu`: {a_in_b, b_in_a} — 禄
  - `tao_hua`: {a_in_b, b_in_a} — 桃花
  - `yi_ma`: {a_in_b, b_in_a} — 驿马
  - `kong_wang`: {a_in_b, b_in_a} — 空亡
  - `kui_gang`: {a_in_b, b_in_a} — 魁罡
  - `ri_de`: {a_in_b, b_in_a} — 日德
  - `ri_gui`: {a_in_b, b_in_a} — 日贵
- `structure` — 结构交互
  - `da_yun`: {a_current, b_current, stem_rel, branch_rel} — 当前大运互动。`a_current`/`b_current` 为 {gan, zhi, name, shi_shen}
  - `xun_gong`: {same_xun, same_gong} — 同旬/同宫

只引用数据中实际存在的字段。若某字段数据中不存在，跳过该分析维度，不要编造。

## 报告结构

### 一、双方八字概览

简述双方基本信息：
- 甲方：`ri.gan` + `fu_yi.qiangruo` + `fu_yi.geju` + 五行偏颇（引述 `nayin_cross.elements.a` 最旺/最弱五行）
- 乙方：同上
- 双方日主关系初判：引述日柱 pair 的 `stem.type`，一句话点出关系基调

引述自 chart 模板，不做深度展开。

### 二、天干互动分析

逐条解读 `zhu_cross.pairs` 中每条天干关系（`stem`）：
- 涉及哪柱（年柱=家庭背景/祖辈，月柱=事业/性格，日柱=感情核心，时柱=价值观/子女）
- `stem.type` 关系类型 + `stem.relation` 解读文字
- 日主之间的关系重点展开（日柱 pair 的 `stem`），这是感情核心指标

### 三、地支配合分析

逐条解读 `zhu_cross.pairs` 中每条地支关系（`branch`）：
- 涉及哪柱：月支=内在性格/家庭观念，日支=夫妻宫（重点），时支=生活习惯/晚年
- `branch.type` + `branch.detail`
- 六合/三合/三会优先视为正面加分，六冲不一定是负面，六害/三刑如实说明但不渲染
- **夫妻宫分析**：单独一节展开日柱 pair 的 `branch` 关系，这是合盘最关键的部分

### 四、十神互动分析

基于 `shi_shen_cross`：
- 甲方对乙方的十神（`a_to_b`）：各柱十神含义，重点看日柱（`ri_stem`）
- 乙方对甲方的十神（`b_to_a`）：同上
- 总结双方互动模式：谁更主动/被动，谁付出更多，权力关系如何

### 五、五行与用神互补

基于 `nayin_cross`：
- 列出双方五行分布（`elements.a` / `elements.b`），不编造数字
- 用神交叉：引述 `yong_shen.a` 和 `yong_shen.b`
  - `yong_in_other` 高 → 己方所需对方能补，五行互补佳
  - `ji_in_other` 高 → 己方所忌对方却有，存在五行冲突
- 纳音配合：引述 `nayin_cross.pairs`，年/日纳音相生为吉

### 六、神煞互动

基于 `shensha_cross`：
- 逐项列出双方神煞交互（`tian_yi`/`lu`/`tao_hua`/`yi_ma`/`kong_wang`/`kui_gang`/`ri_de`/`ri_gui`）
- 重点展开 `tian_yi`（贵人）、`tao_hua`（桃花）、`kong_wang`（空亡）
- 正面的（互贵）→ 加分项。负面的 → 如实说明，建议如何化解

### 七、大运同步与结构

基于 `structure`：
- 当前大运：双方各自 `da_yun.a_current` / `da_yun.b_current` 的干支和十神
- 大运互动：`da_yun.stem_rel` 和 `da_yun.branch_rel`（若存在）
- 同旬同宫：`xun_gong.same_xun` / `xun_gong.same_gong`，若为 true 说明根基契合度高
- 大运同步性：双方运势走势是否一致，不同步时说明哪方为上升期哪方为低谷期

### 八、综合建议

基于全部分析：
- 关系的核心优势和契合点
- 需要关注和沟通的差异点
- 具体建议：双方如何利用优势、化解冲突

## 边界处理

- 双方八字数据不全 → 告知缺少哪方数据，不强行分析
- `zhu_cross.pairs` 为空 → 告知合盘结果未生成，建议先分别排盘
- 某 `shensha_cross` 字段为 null → 跳过该项，不编造
- 日柱 pair 不存在 → 重点看月柱互动，注明日柱数据缺失
- `yong_shen` 某字段为 null → 跳过用神互补分析
- `da_yun` 字段为空 → 跳过大运同步章节
- 不要做"一定分手/一定不和"等绝对化断言
