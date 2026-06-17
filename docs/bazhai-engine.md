# 八宅风水引擎

## Public API

| 类型 | 说明 |
|------|------|
| `MingGua` | 命卦（卦名/卦数/东西四命） |
| `Chart` | 八宅合参（命卦+八宅方位+流年星+四柱八卦） |

| 函数 | 说明 |
|------|------|
| `ComputeMingGua(gender, birthYear)` | 命卦 |
| `ComputeChart(solarTime, gender)` | 八宅合参 |

### HTTP Routes

| 路由 | Handler |
|------|---------|
| `POST /api/fengshui/minggua` | 命卦 |
| `POST /api/fengshui/chart` | 八宅合参 |
