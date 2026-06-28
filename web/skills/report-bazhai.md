# 八宅风水报告模板

基于 chart 和 minggua API 返回数据生成八宅风水报告。

## 数据来源

- chart（`bazhai.chart`）：`ming_gua`（`gua` 卦象 — `name`/`wuxing`/`yin_yang`、`gua_number` 卦数、`group` 东西四命）、`ba_zhai_dirs`（`sheng_qi`/`tian_yi`/`yan_nian`/`fu_wei` 四吉方、`huo_hai`/`wu_gui`/`liu_sha`/`jue_ming` 四凶方，每方为方位名数组）、`year_stars`（`year`/`center_star`/`palaces[]`，每星含 `number`/`color`/`name`/`wuxing`/`auspicious`）、`pillar_bagua`（四柱各配八卦，每卦 `name`/`wuxing`/`yin_yang`）
- minggua（`bazhai.minggua`）：`ming_gua` 同上

只引用数据中实际存在的字段。若某字段数据中不存在，跳过该分析维度，不编造。

## 报告结构

### 一、命卦定位

列出命卦信息：`ming_gua.gua.name`（五行 `ming_gua.gua.wuxing`，`ming_gua.gua.yin_yang`）+ `ming_gua.group`。简述该命卦的性格倾向和五行特征。

### 二、四吉四凶方位

以命卦为基准，列出 `ba_zhai_dirs` 八方：
- 吉方：`sheng_qi`/`tian_yi`/`yan_nian`/`fu_wei` 逐条列出方位名，说明各星吉应和最佳用途
- 凶方：`huo_hai`/`wu_gui`/`liu_sha`/`jue_ming` 逐条列出方位名，说明各星凶应和化解方法

方位名直接使用 API 返回的字符串，不推测方位。

### 三、年飞星分析

引述 `year_stars.year` 年份，列出九宫飞星分布。重点展开：
- `center_star` 中宫星及其影响
- 五黄煞和二黑病符落宫，给出化解建议（如金属物泄土气）
- 吉星（一白/六白/八白/九紫）落宫，给出催旺建议
- 叠加分析：飞星吉凶与八宅方位吉凶的综合判断

### 四、四柱八卦

逐柱列出 `pillar_bagua` 四卦：年柱配→何卦、月柱配→何卦、日柱配→何卦、时柱配→何卦。各卦五行与命卦五行的生克关系，柱对应宫位的提示。

## 边界处理

- `ba_zhai_dirs` 某方位数组为空 → 跳过该方，不提
- `year_stars` 为空 → 跳过飞星章节
- `pillar_bagua` 为空 → 跳过四柱八卦章节
- 用户仅调 minggua → 只输出命卦定位，告知可进一步排八宅盘获取完整分析
- 宅向不确定 → 先看命卦和年飞星，建议确定坐向后补看玄空
