# 玄空风水引擎

## Public API

| 类型 | 说明 |
|------|------|
| `SanYuanYun` | 三元九运当前运信息 |
| `Chart` | 玄空飞星排盘（运星+山星+向星+旺衰+双星加会+收山出煞） |

| 函数 | 说明 |
|------|------|
| `ComputeSanYuanYun(year int)` | 查当前三元九运 |
| `ComputeChart(st SolarTime, sitMountain int, faceMountain int) → Chart` | 玄空飞星排盘（编排入口，api.go） |

编排层 `api.go` 收 `SolarTime` → 提取年份 → 引擎 `computeChart(sitMountain int, faceMountain int, year int)` 收精确实体。

### HTTP Routes

| 路由 | Handler |
|------|---------|
| `GET /api/fengshui/sanyuan` | 三元九运查询 |
| `POST /api/xuankong/chart` | 玄空飞星排盘 |
