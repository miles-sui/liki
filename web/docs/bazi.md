# BaZi API (八字)

Base: `POST /api/bazi/*`

## BaZi Chart

`POST /api/bazi/chart` — Compute full BaZi chart from birth information.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Birth year, 1900-2200 |
| month | int | yes | Birth month, 1-12 |
| day | int | yes | Birth day, 1-31 |
| hour | int | no | Birth hour, 0-23 (default: 0) |
| minute | int | no | Birth minute, 0-59 (default: 0) |
| longitude | float | no | Birthplace longitude, -180~180 (default: 120) |
| timezone | float | no | UTC offset hours, -12~14 (default: 8) |
| gender | string | yes | "male" or "female" |

Returns (wrapped in `{"data":{...}}`): `nianzhu`/`yuezhu`/`rizhu`/`shizhu` (each with `gan`/`zhi`/`nayin`/`canggan`/`shishen`/`changsheng`/`shensha`/`is_void`/`is_self_he`/`is_kui_gang`), `riyuan`, `wuxing` (string keys: "木"/"火"/"土"/"金"/"水"), `dayun`, `yong_shen` (nested `fuyi` + `tiaohou`), `wang_shuai`, `changsheng`, `hehui`, `gong_jia`, `tai_yuan_ming_gong`, `nayin_rel`, `xiu`, `zodiac`, `season`, `lunar_month`, `hour_range`, `xun_name`, `sanqi_name`, `solar_time_minutes`, `solar_datetime`, `bazi_datetime`.

```bash
curl -s -X POST https://liki.hk/api/bazi/chart \
  -H 'Content-Type: application/json' \
  -d '{"year":2000,"month":3,"day":15,"hour":14,"minute":0,"gender":"male","longitude":116.4,"timezone":8}'
```

## Solar Time

`POST /api/tianwen/solartime` — Compute true solar time from birth parameters. Useful for verifying hour branch boundaries and timezone accuracy.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Birth year, 1900-2200 |
| month | int | yes | Birth month, 1-12 |
| day | int | yes | Birth day, 1-31 |
| hour | int | no | Birth hour, 0-23 (default: 0) |
| minute | int | no | Birth minute, 0-59 (default: 0) |
| longitude | float | no | Birthplace longitude, -180~180 (default: 120) |
| timezone | float | no | UTC offset hours, -12~14 (default: 8) |

Returns: `{solar_time, hour_branch, hour_branch_name}`

```bash
curl -s -X POST https://liki.hk/api/tianwen/solartime \
  -H 'Content-Type: application/json' \
  -d '{"year":2000,"month":3,"day":15,"hour":14,"minute":0,"longitude":116.4,"timezone":8}'
```

## BaZi Bond

`POST /api/bazi/bond` — Cross-chart analysis (合盘). Returns both full charts + five-dimensional bond: zhuzhu_rel (16 pillar-pair interactions), shishen_rel (mutual ten-god perspectives), nayin_rel (nayin wuxing relations), shensha_rel (shensha co-occurrence), structure (taiyuan/minggong/dayun/xunkong comparisons). No scoring — structured facts only.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| a | object | yes | Person A birth parameters (same as chart params) |
| b | object | yes | Person B birth parameters (same as chart params) |

```bash
curl -s -X POST https://liki.hk/api/bazi/bond \
  -H 'Content-Type: application/json' \
  -d '{"a":{"year":2000,"month":3,"day":15,"hour":14,"gender":"male","longitude":116.4,"timezone":8},"b":{"year":1998,"month":7,"day":22,"hour":9,"gender":"female","longitude":121.5,"timezone":8}}'
```

## BaZi LiuNian

`POST /api/bazi/liunian` — Yearly fortune (流年运势) for a given year.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields: `{"nian":{gan,zhi},"yue":{gan,zhi},"ri":{gan,zhi},"shi":{gan,zhi}}` |
| year | int | yes | Target year, 1900-2100 |
| current_dayun | object | no | Current dayun pillar `{gan, zhi}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liunian \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"nian":{"gan":7,"zhi":5},"yue":{"gan":5,"zhi":3},"ri":{"gan":3,"zhi":1},"shi":{"gan":9,"zhi":11}},"year":2026}'
```

## BaZi LiuYue

`POST /api/bazi/liuyue` — Monthly fortune for a given year+month.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields (same format as liunian) |
| year | int | yes | Year, 1900-2200 |
| month | int | yes | Month, 1-12 |

```bash
curl -s -X POST https://liki.hk/api/bazi/liuyue \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"nian":{"gan":7,"zhi":5},"yue":{"gan":5,"zhi":3},"ri":{"gan":3,"zhi":1},"shi":{"gan":9,"zhi":11}},"year":2026,"month":6}'
```

## BaZi LiuRi

`POST /api/bazi/liuri` — Daily fortune for a specific date.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields |
| date | string | yes | Date, YYYY-MM-DD |
| dayun_pillar | object | no | Current dayun pillar `{gan, zhi}` |
| liunian_pillar | object | no | Current liunian pillar `{gan, zhi}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liuri \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"nian":{"gan":7,"zhi":5},"yue":{"gan":5,"zhi":3},"ri":{"gan":3,"zhi":1},"shi":{"gan":9,"zhi":11}},"date":"2026-05-27"}'
```

## BaZi LiuShi

`POST /api/bazi/liushi` — Hourly fortune for a specific date+hour.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields |
| date | string | yes | Date, YYYY-MM-DD |
| hour | int | yes | Hour, 0-23 |

```bash
curl -s -X POST https://liki.hk/api/bazi/liushi \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"nian":{"gan":7,"zhi":5},"yue":{"gan":5,"zhi":3},"ri":{"gan":3,"zhi":1},"shi":{"gan":9,"zhi":11}},"date":"2026-05-27","hour":14}'
```

## BaZi XiaoYun

`POST /api/bazi/xiao-yun` — Minor fortune (小运), auxiliary yearly cycle.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | Birth parameters (same as chart params: year, month, day, hour, minute, longitude, timezone, gender) |
| count | int | yes | Number of years to return, 1-120 |

```bash
curl -s -X POST https://liki.hk/api/bazi/xiao-yun \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"year":2000,"month":3,"day":15,"hour":14,"gender":"male","longitude":116.4,"timezone":8},"count":10}'
```

## BaZi XiaoXian

`POST /api/bazi/xiao-xian` — Minor limit (小限), birthday-bounded annual cycle.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| gender | string | yes | "male" or "female" |
| count | int | yes | Number of years to return, 1-120 |

```bash
curl -s -X POST https://liki.hk/api/bazi/xiao-xian \
  -H 'Content-Type: application/json' \
  -d '{"gender":"male","count":10}'
```
