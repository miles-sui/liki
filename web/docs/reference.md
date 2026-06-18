# Agent API

## Agent Chat

`POST /api/agent/chat` — SSE streaming chat endpoint.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| session_id | string | no | Resume existing session |
| message | string | yes | User message |
| country | string | no | IP country code hint |
| city | string | no | IP city hint |
| lang | string | no | Frontend language: zh/hk/en |

Response: SSE stream with ChatEvent JSON objects (`data: {...}\n\n`).

```bash
curl -s -X POST https://liki.hk/api/agent/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"我想算八字"}'
```

## Agent Greeting

`GET /api/agent/greeting` — Cached LLM-generated greeting.

```bash
curl -s https://liki.hk/api/agent/greeting
```

## Session Restore

`GET /api/agent/session` — Restore session history.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| session_id | string | yes | Session ID |

```bash
curl -s 'https://liki.hk/api/agent/session?session_id=xxx'
```

---

# Payment API

## Checkout

`POST /api/payments/checkout` — Create payment checkout.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| order_id | string | yes | Order ID from chat done event |
| email | string | no | Email for receipt |

```bash
curl -s -X POST https://liki.hk/api/payments/checkout \
  -H 'Content-Type: application/json' \
  -d '{"order_id":"ord_xxx","email":"user@example.com"}'
```

## Order Status

`GET /api/orders/{id}/status` — Get order status.

```bash
curl -s https://liki.hk/api/orders/ord_xxx/status
```

## Order Retry

`POST /api/orders/{id}/retry` — Retry report generation for paid orders missing llm_json.

```bash
curl -s -X POST https://liki.hk/api/orders/ord_xxx/retry
```

## Report

`GET /api/reports/{id}` — Get generated report.

```bash
curl -s https://liki.hk/api/reports/ord_xxx
```

---

# Misc

## Location

`GET /api/location` — IP-based geolocation. Returns country, city, currency.

```bash
curl -s https://liki.hk/api/location
```

## Health

`GET /api/health` — Service health check.

```bash
curl -s https://liki.hk/api/health
```

## Version

`GET /api/version` — Build time of the running binary.

```bash
curl -s https://liki.hk/api/version
```

## Stats

`GET /api/stats` — Anonymous page view and conversion counters.

```bash
curl -s https://liki.hk/api/stats
```
