# ZiWei API (紫微斗数)

Base: `POST /api/ziwei/*`

Input conventions:
- `birth` is `{"time": "RFC3339", "longitude": float}`
- `gender` is `"male"` or `"female"`

All responses wrapped in `{"data":{...}}`.

## ZiWei Chart

`POST /api/ziwei/chart` — Compute ZiWei chart (紫微斗数排盘).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/ziwei/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4},"gender":"male"}'
```

## DaXian

`POST /api/ziwei/daxian` — Compute decade cycles (大限).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| chart | object | yes | Full ZiWei chart from /api/ziwei/chart |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/ziwei/daxian \
  -H 'Content-Type: application/json' \
  -d '{"chart":{},"gender":"male"}'
```

## LiuNian

`POST /api/ziwei/liunian` — Yearly cycle (流年).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| liu_year | int | yes | Target year, 1900-2100 |
| chart | object | yes | Full ZiWei chart |

```bash
curl -s -X POST https://liki.hk/api/ziwei/liunian \
  -H 'Content-Type: application/json' \
  -d '{"liu_year":2026,"chart":{}}'
```

## LiuYue

`POST /api/ziwei/liuyue` — Monthly cycle (流月).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| liu_year | int | yes | Target year, 1900-2100 |
| lunar_month | int | yes | Lunar month, 1-12 |
| chart | object | yes | Full ZiWei chart |

```bash
curl -s -X POST https://liki.hk/api/ziwei/liuyue \
  -H 'Content-Type: application/json' \
  -d '{"liu_year":2026,"lunar_month":5,"chart":{}}'
```

## LiuRi

`POST /api/ziwei/liuri` — Daily cycle (流日).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| liu_year | int | yes | Target year, 1900-2100 |
| lunar_month | int | yes | Lunar month, 1-12 |
| lunar_day | int | yes | Lunar day, 1-30 |
| chart | object | yes | Full ZiWei chart |

```bash
curl -s -X POST https://liki.hk/api/ziwei/liuri \
  -H 'Content-Type: application/json' \
  -d '{"liu_year":2026,"lunar_month":5,"lunar_day":15,"chart":{}}'
```

## Bond

`POST /api/ziwei/bond` — Cross-chart analysis (合盘).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| a | object | yes | Person A's full ZiWei chart |
| b | object | yes | Person B's full ZiWei chart |

```bash
curl -s -X POST https://liki.hk/api/ziwei/bond \
  -H 'Content-Type: application/json' \
  -d '{"a":{},"b":{}}'
```
