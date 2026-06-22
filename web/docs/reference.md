# Access Policy

## Public APIs

以下 API 对外开放，无限制：

| 端点 | 方法 | 说明 |
|---|---|---|
| `/api/bazi/*` | POST | 八字排盘/合盘/流运 |
| `/api/ziwei/*` | POST | 紫微斗数 |
| `/api/qimen/*` | POST | 奇门遁甲 |
| `/api/qiming/*` | POST | 智能起名 |
| `/api/liuyao/*` | POST | 六爻 |
| `/api/huangli/*` | GET/POST | 黄历 |
| `/api/bazhai/*` | POST | 八宅风水 |
| `/api/xuankong/*` | GET/POST | 玄空飞星 |
| `/api/payments/checkout` | POST | 支付下单 |
| `/api/payments/webhook` | POST | 支付回调（Dodo IP 仅） |
| `/api/payments/return/*` | GET | 支付跳转 |
| `/api/orders/*/status` | GET | 订单状态 |
| `/api/reports/*` | GET | 报告 |
| `/api/health` | GET | 健康检查 |
| `/api/version` | GET | 版本 |
| `/api/stats` | GET | 统计 |
| `/api/location` | GET | IP 定位 |

## Restricted APIs

以下 API 仅限 Liki 前端（liki.hk）访问，外部调用返回 403：

| 端点 | 说明 |
|---|---|
| `/api/agent/*` | Agent 对话/问候/会话（LLM token 成本） |
| `/api/orders/*/retry` | 报告重试（LLM token 成本） |

## Notes

- Restricted API 通过 Caddy 检查 Origin 头实现，无 Origin 或 Origin 为 liki.hk 放行
- 免费计算 API 有 rate limit（60/min, burst 10），无需认证

---
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

## Errors

All errors use a unified envelope: `{"error":{"code":"...","message":"..."}}`.

| HTTP status | code | Description |
|-------------|------|-------------|
| 400 | `invalid_request` | JSON parse failure, invalid business parameter, or engine compute error |
| 413 | `too_large` | Request body exceeds 1MB limit |
| 422 | `validation_error` | Structured validation failure with field details |
| 404 | `not_found` | Resource does not exist |
| 500 | `internal_error` | Unexpected server error |

```bash
# Example error response (422)
curl -s -X POST https://liki.hk/api/bazi/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"1984-02-04T06:00:00+08:00","longitude":116.4}}'
# → {"error":{"code":"validation_error","message":"gender: cannot be blank."}}
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
