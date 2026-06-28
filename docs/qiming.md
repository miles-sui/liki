# 起名 API · Qiming API

## 架构

四个原子端点，供 agent/LLM 按 liki.md 规定的流程编排。确定性计算归后端，偏好判断归 LLM。

```
wuge → LLM 筛字库（去生僻/拗口/消极）→ compose → LLM 选 8 名（首要性别、覆盖多风格、典故加分）→ detail → 读 report-naming.md 生成报告
```

设计原则：
- 五行和音韵由算法处理，LLM 不重复判断
- LLM 专注于算法做不了的事：字义筛选、性别适配、典故查找、风格多样性
- 报告模板（report-naming.md）是起名报告格式参考，LLM 对话内按此格式输出报告

## API

### qiming.wuge

枚举吉数笔画组合 + 候选字。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓，1-2 字 |
| yong_shen | string | yes | 用神（木/火/土/金/水） |
| xi_shen | [string] | no | 喜神 |

返回 `{surname, combos: [{stroke1, stroke2, sancai, fortune}], yong_chars, xi_chars}`。

字数据为 `CharLite{char, tone}`，按笔画分组 `map[int][]CharLite`。

### qiming.compose

字池笛卡尔积 + 平仄过滤，返回名字列表。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓 |
| combos | [StrokeCombo] | yes | wuge 返回的 combos |
| yong_chars | {stroke: [string]} | yes | 用神字池 |
| xi_chars | {stroke: [string]} | yes | 喜神字池 |

组合规则：yong+yong、yong+xi、xi+yong（不允许 xi+xi）。

### qiming.detail

批量查询五格三才音韵。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓 |
| names | [string] | yes | 名字列表，1-50 |

返回 `{results: [{name, characters: [Character], wu_ge: WuGe, san_cai: SanCai, phonetic: Phonetic}]}`。

### qiming.evaluate

单名评测。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓 |
| given_name | string | yes | 名 |
| yong_shen | string | yes | 用神 |

返回 `Evaluation{surname, given_name, characters, wu_ge, san_cai, phonetic, wuxing_match}`。

## Engine 公开 API

类型：`WuGeData`, `NameCandidate`, `Evaluation`, `WuGe`, `SanCai`, `StrokeCombo`, `Character`, `CharLite`, `Phonetic`

函数：`PrepareWuGe`, `ComposeNames`, `DetailNames`, `EvaluateName`
