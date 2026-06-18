# Huangli API (黄历)

All responses wrapped in `{"data":{...}}`.

## Huangli Date

`GET /api/huangli/date` — Query almanac for a specific date.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| date | string | yes | YYYY-MM-DD |
| event | string | yes | Event filter (嫁娶, 入宅, 开业, 出行, etc.) |

Returns `{"data":{"entry":{...}}}`.

```bash
curl -s 'https://liki.hk/api/huangli/date?date=2026-06-17&event=嫁娶'
```

## Huangli Month

`GET /api/huangli/month` — Query almanac for a month.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| month | string | yes | YYYY-MM |
| event | string | yes | Event filter (嫁娶, 入宅, 开业, 出行, etc.) |

Returns `{"data":{"entries":[...]}}`.

```bash
curl -s 'https://liki.hk/api/huangli/month?month=2026-06&event=嫁娶'
```

## Huangli Bond Date

`POST /api/huangli/bond/date` — Cross-reference birth with huangli for single-day selection (择日).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| event_type | string | yes | Event type (嫁娶, 入宅, 开业, 出行, etc.) |
| date | string | yes | YYYY-MM-DD |

Returns `{"data":{"entry":{...}}}`.

```bash
curl -s -X POST https://liki.hk/api/huangli/bond/date \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2000-03-15T14:00:00+08:00","longitude":116.4},"event_type":"嫁娶","date":"2026-06-17"}'
```

## Huangli Bond Month

`POST /api/huangli/bond/month` — Cross-reference birth with huangli for month-range selection (择月).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| event_type | string | yes | Event type (嫁娶, 入宅, 开业, 出行, etc.) |
| month | string | yes | YYYY-MM |

Returns `{"data":{"entries":[...]}}`.

```bash
curl -s -X POST https://liki.hk/api/huangli/bond/month \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2000-03-15T14:00:00+08:00","longitude":116.4},"event_type":"嫁娶","month":"2026-06"}'
```
