# Fengshui API (风水)

Base: `GET /api/fengshui/*`, `POST /api/fengshui/*`, `POST /api/xuankong/*`

## Bazhai MingGua

`POST /api/fengshui/minggua` — Compute fate trigram (命卦) via Eight Mansions (八宅) method. Returns personal trigram, gua number, East/West group classification.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth_year | int | yes | Birth year, 1900-2200 |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/fengshui/minggua \
  -H 'Content-Type: application/json' \
  -d '{"birth_year":2000,"gender":"male"}'
```

## Bazhai Chart

`POST /api/fengshui/chart` — Combined八宅合参. Returns: fate trigram (命卦), Eight Mansions four-auspicious-four-inauspicious directions (八宅四吉四凶), annual purple-white flying stars (年紫白飞星), pillar bagua (四柱纳甲卦). No scoring — each system speaks for itself.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| solar_time | object | yes | SolarTime `{"year","month","day","hour","minute","longitude","offset"}` |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/fengshui/chart \
  -H 'Content-Type: application/json' \
  -d '{"solar_time":{"year":2000,"month":1,"day":1,"hour":12,"minute":0,"longitude":120,"offset":480},"gender":"male"}'
```

## Xuankong Chart

`POST /api/xuankong/chart` — 玄空飞星排盘. Returns: 三元九运, 坐向, nine-palace flying stars (period/mountain/facing), 旺山旺向/上山下水/反吟伏吟 evaluation, 双星加会, 收山出煞.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| solar_time | object | yes | SolarTime `{"year","month","day","hour","minute","longitude","offset"}` |
| sit_mountain | int | yes | 坐山 index, 0-23 |
| face_mountain | int | yes | 朝向 index, 0-23 |

```bash
curl -s -X POST https://liki.hk/api/xuankong/chart \
  -H 'Content-Type: application/json' \
  -d '{"solar_time":{"year":2026,"month":6,"day":16,"hour":12,"minute":0,"longitude":120,"offset":480},"sit_mountain":0,"face_mountain":11}'
```

## San Yuan Jiu Yun

`GET /api/fengshui/sanyuan` — Get San Yuan Jiu Yun (三元九运) timetable. Current period: 九运 (2024-2043).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Query year, >=1864 |

```bash
curl -s 'https://liki.hk/api/fengshui/sanyuan?year=2026'
```
