# Fengshui API (风水)

Base: `GET /api/fengshui/*`, `POST /api/fengshui/*`

## Fengshui MingGua

`POST /api/fengshui/minggua` — Compute fate trigram (命卦) via Eight Mansions (八宅) method. Returns personal trigram, gua number, East/West group classification, and all 8 trigrams for reference.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Birth year, 1900-2200 |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://api.tokflux.com/api/fengshui/minggua \
  -H 'Content-Type: application/json' \
  -d '{"year":2000,"gender":"male"}'
```

## Fengshui HeCan

`POST /api/fengshui/hecan` — Combined Feng Shui reference (风水合参). Returns: fate trigram (命卦), Eight Mansions four-auspicious-four-inauspicious directions (八宅四吉四凶), annual purple-white flying stars (年紫白飞星), pillar bagua (四柱纳甲卦), yong-shen pass-through. No scoring — each system speaks for itself.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth_year | int | yes | Birth year, 1900-2200 |
| gender | string | yes | "male" or "female" |
| bazi | object | yes | Four pillars as named fields `{"year":{stem,branch},"month":{stem,branch},"day":{stem,branch},"hour":{stem,branch}}` |
| yong_shen | object | yes | Full YongShenResult from chart: `{"fuyi":{"strength","pattern","yong","xi","ji"},"tiaohou":{"season","yong","xi","ji","detail"}}` |
| year | int | yes | Reference year, >=1864 |

```bash
curl -s -X POST https://api.tokflux.com/api/fengshui/hecan \
  -H 'Content-Type: application/json' \
  -d '{"birth_year":2000,"gender":"male","bazi":{"year":{"stem":7,"branch":5},"month":{"stem":5,"branch":3},"day":{"stem":3,"branch":1},"hour":{"stem":9,"branch":11}},"yong_shen":{"fuyi":{"strength":"身强","pattern":"建禄格","yong":"金","xi":"土","ji":"火"},"tiaohou":{"season":"春","yong":"丙","xi":"癸","ji":"土"}},"year":2026}'
```

## San Yuan Jiu Yun

`GET /api/fengshui/san-yuan` — Get San Yuan Jiu Yun (三元九运) timetable. Current period: 九运 (2024-2043).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Query year, >=1864 |

```bash
curl -s 'https://api.tokflux.com/api/fengshui/san-yuan?year=2026'
```
