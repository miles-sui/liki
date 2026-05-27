# BaZi API (八字)

Base: `POST /api/bazi/*`

## BaZi Chart

`POST /api/bazi/chart` — Compute full BaZi chart from birth information.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Birth year, 1900-2200 |
| month | int | yes | Birth month, 1-12 |
| day | int | yes | Birth day, 1-31 |
| hour | int | no | Birth hour, 0-23 (default 0) |
| minute | int | no | Birth minute, 0-59 (default 0) |
| longitude | float | no | Birthplace longitude, -180~180 (default 120.0 = Beijing) |
| timezone | float | no | UTC offset hours, -12~14 (default 8 = UTC+8) |
| gender | string | yes | "male" or "female" |

DST is auto-detected server-side (China 1986-1991). City name → coordinates: use `GET /api/reference/cities?q=<name>`, then `timezone = lng/15` rounded.

Returns: four pillars (each with stem, branch, nayin, hidden_stems, ten_gods table, life_stages table, shensha, is_void, is_self_he, is_kui_gang), day_master, element_count (string keys: "木"/"火"/"土"/"金"/"水"), dayun table, yong_shen (nested fuyi + tiaohou), wang_shuai map, life_stages array, full_he_hui, gong_jia, tai_yuan_ming_gong, nayin_relations, day_mansion, zodiac, season, lunar_month, hour_range, xun_name, sanqi_name, solar_time_minutes, solar_datetime, bazi_datetime.

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/chart \
  -H 'Content-Type: application/json' \
  -d '{"year":2000,"month":3,"day":15,"hour":14,"minute":0,"gender":"male","longitude":116.4,"timezone":8}'
```

## BaZi Bond

`POST /api/bazi/bond` — Cross-chart analysis (合盘). Returns both full charts + five-dimensional bond: pillar_cross (16 pillar-pair interactions), ten_god_cross (mutual ten-god perspectives), nayin_cross (nayin element relations), shensha_cross (shensha co-occurrence), structure (taiyuan/minggong/dayun/xunkong comparisons). No scoring — structured facts only.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| a | object | yes | Person A birth parameters (same as chart params) |
| b | object | yes | Person B birth parameters (same as chart params) |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/bond \
  -H 'Content-Type: application/json' \
  -d '{"a":{"year":2000,"month":3,"day":15,"hour":14,"gender":"male","longitude":116.4,"timezone":8},"b":{"year":1998,"month":7,"day":22,"hour":9,"gender":"female","longitude":121.5,"timezone":8}}'
```

## BaZi LiuNian

`POST /api/bazi/liunian` — Yearly fortune (流年运势) for a given year.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields: `{"year":{stem,branch},"month":{stem,branch},"day":{stem,branch},"hour":{stem,branch}}` |
| year | int | yes | Target year, 1900-2100 |
| current_dayun | object | no | Current dayun pillar `{stem, branch}` |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/liunian \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"year":{"stem":7,"branch":5},"month":{"stem":5,"branch":3},"day":{"stem":3,"branch":1},"hour":{"stem":9,"branch":11}},"year":2026}'
```

## BaZi LiuYue

`POST /api/bazi/liuyue` — Monthly fortune for a given year+month.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields (same format as liunian) |
| year | int | yes | Year, 1900-2200 |
| month | int | yes | Month, 1-12 |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/liuyue \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"year":{"stem":7,"branch":5},"month":{"stem":5,"branch":3},"day":{"stem":3,"branch":1},"hour":{"stem":9,"branch":11}},"year":2026,"month":6}'
```

## BaZi LiuRi

`POST /api/bazi/liuri` — Daily fortune for a specific date.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields |
| date | string | yes | Date, YYYY-MM-DD |
| dayun_pillar | object | no | Current dayun pillar `{stem, branch}` |
| liunian_pillar | object | no | Current liunian pillar `{stem, branch}` |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/liuri \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"year":{"stem":7,"branch":5},"month":{"stem":5,"branch":3},"day":{"stem":3,"branch":1},"hour":{"stem":9,"branch":11}},"date":"2026-05-27"}'
```

## BaZi LiuShi

`POST /api/bazi/liushi` — Hourly fortune for a specific date+hour.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| bazi | object | yes | Four pillars as named fields |
| date | string | yes | Date, YYYY-MM-DD |
| hour | int | yes | Hour, 0-23 |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/liushi \
  -H 'Content-Type: application/json' \
  -d '{"bazi":{"year":{"stem":7,"branch":5},"month":{"stem":5,"branch":3},"day":{"stem":3,"branch":1},"hour":{"stem":9,"branch":11}},"date":"2026-05-27","hour":14}'
```

## BaZi XiaoYun

`POST /api/bazi/xiao-yun` — Minor fortune (小运), auxiliary yearly cycle.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | Birth parameters (same as chart params: year, month, day, hour, minute, longitude, timezone, gender) |
| count | int | yes | Number of years to return, >=1 |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/xiao-yun \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"year":2000,"month":3,"day":15,"hour":14,"gender":"male","longitude":116.4,"timezone":8},"count":10}'
```

## BaZi XiaoXian

`POST /api/bazi/xiao-xian` — Minor limit (小限), birthday-bounded annual cycle.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| gender | string | yes | "male" or "female" |
| count | int | yes | Number of years to return, >=1 |

```bash
curl -s -X POST https://api.tokflux.com/api/bazi/xiao-xian \
  -H 'Content-Type: application/json' \
  -d '{"gender":"male","count":10}'
```
