# QiMen API (奇门遁甲)

Base: `POST /api/qimen/*`

Input conventions:
- `birth` is `{"time": "RFC3339", "longitude": float}`

All responses wrapped in `{"data":{...}}`.

## QiMen Pan

`POST /api/qimen/pan` — Compute QiMen DunJia chart (奇门遁甲排盘).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| kind | string | no | 局类型: shi/ri/yue/nian (default: shi) |

```bash
curl -s -X POST https://liki.hk/api/qimen/pan \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2026-06-17T12:00:00+08:00","longitude":116.4},"kind":"shi"}'
```
