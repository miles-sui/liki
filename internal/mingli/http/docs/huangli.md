# Huangli API (黄历)

Base: `GET /api/huangli/*` (query), `POST /api/huangli/*` (computation)

## Huangli Query

`GET /api/huangli/query` — Query almanac info for a date or month.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| date | string | one of | Date, YYYY-MM-DD |
| month | string | one of | Month, YYYY-MM |
| event | string | no | Event filter (wedding, open, sign, move, etc.) |

Returns: day pillar (stem/branch/na yin), jianchu god (建除十二神) with suitability marks/warnings, huangdao star, auspicious directions, Peng Zu taboos (彭祖百忌), day mansion (二十八宿).

```bash
curl -s 'https://api.tokflux.com/api/huangli/query?month=2026-05&event=wedding'
```

## Huangli Bond

`POST /api/huangli/bond` — Cross-reference birth info against huangli days. Returns everything from huangli/query plus: gan relation (day stem vs day master), zhi relation (day branch vs birth day pillar), tai sui relation (day branch vs year branch). Use for personalized date selection (择日).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth_info | object | yes | Birth parameters (same as bazi/chart params) |
| date | string | one of | Query date, YYYY-MM-DD |
| month | string | one of | Query month, YYYY-MM |
| event_type | string | no | Event type (wedding, open, sign, move, etc.) |

```bash
curl -s -X POST https://api.tokflux.com/api/huangli/bond \
  -H 'Content-Type: application/json' \
  -d '{"birth_info":{"year":2000,"month":3,"day":15,"hour":14,"gender":"male","longitude":116.4,"timezone":8},"month":"2026-05","event_type":"wedding"}'
```

## JieQi Depth

`GET /api/huangli/jieqi` — Solar term depth (节气深度), including RenYuan SiLing FenYe (人元司令分野).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| year | int | yes | Year, >=1900 |
| month | int | yes | Month, 1-12 |
| day | int | yes | Day, 1-31 |

```bash
curl -s 'https://api.tokflux.com/api/huangli/jieqi?year=2026&month=5&day=27'
```
