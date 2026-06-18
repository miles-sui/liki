# Bazhai API (八宅风水)

Base: `POST /api/bazhai/*`

All responses wrapped in `{"data":{...}}`.

## MingGua

`POST /api/bazhai/minggua` — Compute fate trigram (命卦).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth_year | int | yes | Birth year, 1900-2100 |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/bazhai/minggua \
  -H 'Content-Type: application/json' \
  -d '{"birth_year":2000,"gender":"male"}'
```

## Bazhai Chart

`POST /api/bazhai/chart` — Combined 八宅合参: fate trigram, eight stars directions, annual purple-white flying stars, pillar bagua.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/bazhai/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2000-01-01T12:00:00+08:00","longitude":116.4},"gender":"male"}'
```

---

# Xuankong API (玄空飞星)

## SanYuan JiuYun

`GET /api/xuankong/sanyuan` — San Yuan Jiu Yun (三元九运) timetable.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | no | Query year (default: 2024) |

```bash
curl -s 'https://liki.hk/api/xuankong/sanyuan?year=2026'
```

## Xuankong Chart

`POST /api/xuankong/chart` — 玄空飞星排盘: 三元九运, 坐向, nine-palace flying stars, 旺山旺向 evaluation, 双星加会.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| sit_mountain | int | yes | 坐山 index, 0-23 |
| face_mountain | int | yes | 朝向 index, 0-23 |

```bash
curl -s -X POST https://liki.hk/api/xuankong/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2026-06-17T12:00:00+08:00","longitude":116.4},"sit_mountain":0,"face_mountain":11}'
```
