
---

## Public API

### Types

| 类型 | 说明 |
|------|------|
| `Day` | 日课（日柱/建除/宜忌/黄道黑道/纳音/五行/吉方/彭祖百忌/二十八宿/节气/人元） |
| `Month` | 月课（月干支+全月日课） |
| `Year` | 年课（年干支/生肖/太岁） |
| `BondDay` | 合日（Day + 干支关系/太岁关系） |
| `BondMonth` | 合月（Month + 全月合日） |
| `BondYear` | 合年（Year + 干支关系） |

### Functions

| 函数 | 说明 |
|------|------|
| `QueryDate(date, event)` | 单日查询 |
| `QueryMonth(yearMonth, event)` | 整月查询 |
| `QueryYear(year)` | 年查询 |
| `CrossDate(dm, dz, date, event)` | 合日查询 |
| `CrossMonth(dm, dz, yearMonth, event)` | 合月查询 |
| `CrossYear(dm, dz, year)` | 合年查询 |

### HTTP Routes

| 路由 | Handler |
|------|---------|
| `GET /api/huangli/query` | 日/月查询 |
| `POST /api/huangli/bond` | 合日/合月/合年 |
