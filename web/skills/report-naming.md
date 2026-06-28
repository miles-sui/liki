# 起名报告模板

基于 chart 和 detail API 返回数据生成起名报告。

## 数据来源

- chart（`bazi.chart`）：`ri_yuan` 日主、`fu_yi`（`qiangruo` 身强身弱、`geju` 格局、`yong` 用神、`xi` 喜神、`ji` 忌神）、`tiao_hou`（`yong` 调候用神、`xi` 调候喜神）
- detail（`qiming.detail`）：`surname` 姓氏、`candidates[]` 候选列表，每项含 `name`、`characters[]`（`char`/`wuxing`/`stroke`/`radical`/`pinyin`/`tone`）、`wu_ge`（五格：`tian_ge`/`ren_ge`/`di_ge`/`wai_ge`/`zong_ge`，每格 `stroke`/`wuxing`/`fortune`/`description`）、`san_cai`（`configuration`/`fortune`/`description`）、`phonetic`（`tones`）

只引用数据中实际存在的字段，不编造。

## 报告结构

### 一、命理基础与用神

引述 chart 数据中的 `ri_yuan`、`fu_yi.qiangruo`、`fu_yi.geju`。说明用神喜忌和起名五行方向。提姓氏五行及其与方向的关系。

### 二、候选名字速览

每个名字一行，含五行属性、风格定位（儒雅/刚健/灵动/古朴等）、亮点（典故或字义一句话）。让读者先有全景再读详析。

### 三、候选名字逐一分析

先一句话说明筛选标准（性别适配、风格多样、典故出处）。然后每个名字按以下维度展开：

- 字形字义：逐字 `char`/`radical`/`stroke`/`wuxing`，解释字义和整体寓意。
- 典故出处：有古文诗词出处的，引原文 + 出处（作者/篇名）+ 释义。没有则跳过。
- 五行：各字五行与用神方向的匹配。
- 三才：`san_cai.configuration` + `san_cai.fortune` + `san_cai.description`。
- 五格：逐格引述 `stroke`/`fortune`/`description`，重点人格。
- 音韵：`phonetic.tones`，加谐音检查。
- 实用性：生僻字/多音字/笔画/性别适配/不良联想。

### 四、横向对比与推荐

对比维度：五行匹配 · 典故内涵 · 三才配置 · 五格数理 · 音韵美感 · 实用性。

首先给出综合最优的名字及理由。然后按不同偏好给出推荐：
- 命理契合 → 五行匹配最优者
- 文化内涵 → 典故最丰富者
- 音韵美感 → 平仄搭配最佳者
- 日常实用 → 笔画最少、无生僻字者

## 边界处理

- 候选列表为空 → 告知用户当前条件无合适匹配，建议调整用神方向或放宽筛选条件
- 用户不满意所有候选 → 建议调整 xi_shen 或换一批五格组合重新 compose
- 用户提供自选名字 → 跳过筛选和候选人分析，直接 evaluate → 输出命理基础 + 详细评估 + 建议
- 数据中缺少某分析字段 → 跳过该维度，不臆测
