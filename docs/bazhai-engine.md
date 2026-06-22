# 八宅风水引擎

## Public API

| 类型 | 说明 |
|------|------|
| `MingGua` | 命卦（卦名/卦数/东西四命） |
| `Chart` | 八宅合参（命卦+八宅方位+流年星+四柱八卦） |

| 函数 | 说明 |
|------|------|
| `ComputeMingGua(gender ganzhi.Gender, birthYear int)` | 命卦 |
| `ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) → Chart` | 八宅合参（编排入口，api.go） |

编排层 `api.go` 收 `tianwen.SolarTime` → `ComputeBazi` → 引擎 `computeChart(bz ganzhi.Bazi, gender ganzhi.Gender, year int)` 收精确实体。

### HTTP Routes

| 路由 | Handler |
|------|---------|
| `POST /api/fengshui/minggua` | 命卦 |
| `POST /api/fengshui/chart` | 八宅合参 |
