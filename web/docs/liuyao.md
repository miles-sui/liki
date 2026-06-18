# LiuYao API (六爻)

Base: `POST /api/liuyao/*`

Input conventions:
- `birth` is `{"time": "RFC3339", "longitude": float}`

All responses wrapped in `{"data":{...}}`.

## LiuYao Chart

`POST /api/liuyao/chart` — Compute LiuYao divination chart (六爻起卦排盘).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| birth | object | yes | `{time, longitude}` |
| yong_shen | string | no | 用神: 父母/兄弟/官鬼/妻财/子孙/世爻 (default: 世爻) |
| fixed | [6]int | no | Fixed yao values for manual hexagram setting |

```bash
curl -s -X POST https://liki.hk/api/liuyao/chart \
  -H 'Content-Type: application/json' \
  -d '{"birth":{"time":"2026-06-17T12:00:00+08:00","longitude":116.4},"yong_shen":"世爻"}'
```
