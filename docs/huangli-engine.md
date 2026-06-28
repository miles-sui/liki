# 黄历引擎

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
| `QueryDate(date string, event string) → (Day, error)` | 单日查询 |
| `QueryMonth(yearMonth string, event string) → (Month, error)` | 整月查询 |
| `ComputeBondDay(st tianwen.SolarTime, eventType string, dateStr string) → (BondDay, error)` | 合日查询（编排入口，api.go） |
| `ComputeBondMonth(st tianwen.SolarTime, eventType string, yearMonth string) → (BondMonth, error)` | 合月查询（编排入口，api.go） |

编排层 `api.go` 收 `tianwen.SolarTime` → `ComputeBazi` → 引擎 `computeBondDay(bz ganzhi.Bazi, …)` / `computeBondMonth(bz ganzhi.Bazi, …)` 收精确实体。

### JSON-RPC Methods

| 路由 | Handler |
|------|---------|
| `huangli.date` | 单日查询 |
| `huangli.month` | 整月查询 |
| `huangli.bond.date` | 合日查询 |
| `huangli.bond.month` | 合月查询 |
