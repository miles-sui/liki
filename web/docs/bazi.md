# BaZi API (八字)

Base: `POST /api/bazi/*`

Input conventions:
- `birth` is `{"time": "RFC3339", "longitude": float}`, e.g. `{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}`
- `date` is `"YYYY-MM-DD"`
- `gender` is `"male"` or `"female"`

All responses wrapped in `{"data":{...}}`.

## BaZi Chart

`POST /api/bazi/chart` — Compute full BaZi chart.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| gender | string | yes | "male" or "female" |

```bash
curl -s -X POST https://liki.hk/api/bazi/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4},"gender":"male"}'
```

## BaZi Bond

`POST /api/bazi/bond` — Cross-chart analysis (合盘).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| a | object | yes | `{birth, gender}` |
| b | object | yes | `{birth, gender}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/bond \
  -H 'Content-Type: application/json' \
  -d '{"a":{"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4},"gender":"male"},"b":{"birth":{"time":"1985-08-15T12:00:00+08:00","longitude":116.4},"gender":"female"}}'
```

## BaZi LiuNian

`POST /api/bazi/liunian` — Yearly fortune (流年).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Target year, 1900-2100 |
| birth | object | yes | `{time, longitude}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liunian \
  -H 'Content-Type: application/json' \
  -d '{"year":2026,"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}}'
```

## BaZi LiuYue

`POST /api/bazi/liuyue` — Monthly fortune (流月).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Year, 1900-2100 |
| month | int | yes | Month, 1-12 |
| birth | object | yes | `{time, longitude}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liuyue \
  -H 'Content-Type: application/json' \
  -d '{"year":2026,"month":6,"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}}'
```

## BaZi LiuRi

`POST /api/bazi/liuri` — Daily fortune (流日).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| date | string | yes | YYYY-MM-DD |
| birth | object | yes | `{time, longitude}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liuri \
  -H 'Content-Type: application/json' \
  -d '{"date":"2026-06-17","birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}}'
```

## BaZi LiuShi

`POST /api/bazi/liushi` — Hourly fortune (流时).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| date | string | yes | YYYY-MM-DD |
| hour | int | yes | Hour, 0-23 |
| birth | object | yes | `{time, longitude}` |

```bash
curl -s -X POST https://liki.hk/api/bazi/liushi \
  -H 'Content-Type: application/json' \
  -d '{"date":"2026-06-17","hour":14,"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}}'
```

## BaZi XiaoYun

`POST /api/bazi/xiaoyun` — Minor fortune (小运).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| gender | string | yes | "male" or "female" |
| count | int | yes | Number of years, 1-120 |

```bash
curl -s -X POST https://liki.hk/api/bazi/xiaoyun \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4},"gender":"male","count":10}'
```

## BaZi XiaoXian

`POST /api/bazi/xiaoxian` — Minor limit (小限).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| gender | string | yes | "male" or "female" |
| count | int | yes | Number of years, 1-120 |

```bash
curl -s -X POST https://liki.hk/api/bazi/xiaoxian \
  -H 'Content-Type: application/json' \
  -d '{"gender":"male","count":10}'
```
