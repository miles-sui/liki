# 起名 API · Qiming API

## 架构

四个原子端点，供 agent/LLM 自由编排。确定性计算归后端，偏好判断归 LLM。

```
PrepareWuGe → LLM 筛字 → ComposeNames → LLM 筛名 → DetailNames → LLM 精排
```

## API

### POST /api/qiming/wuge

枚举吉数笔画组合 + 候选字。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓，1-2 字 |
| yong_shen | string | yes | 用神（木/火/土/金/水） |
| xi_shen | [string] | no | 喜神 |

返回 `{surname, combos: [{stroke1, stroke2, sancai, fortune}], yong_chars, xi_chars}`。

字数据为 `CharLite{char, tone}`，按笔画分组 `map[int][]CharLite`。

### POST /api/qiming/compose

字池笛卡尔积 + 平仄过滤，返回名字列表。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓 |
| combos | [StrokeCombo] | yes | wuge 返回的 combos |
| yong_chars | {stroke: [string]} | yes | 用神字池 |
| xi_chars | {stroke: [string]} | yes | 喜神字池 |

组合规则：yong+yong、yong+xi、xi+yong（不允许 xi+xi）。

### POST /api/qiming/detail

批量查询五格三才音韵。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| surname | string | yes | 姓 |
| names | [string] | yes | 名字列表，1-50 |

返回 `{results: [{name, characters: [Character], wu_ge: WuGe, san_cai: SanCai, phonetic: Phonetic}]}`。

### POST /api/qiming/evaluate

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
