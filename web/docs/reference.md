# Reference Data API

All `GET /api/reference/*`. Simple JSON lookup tables — no computation.

## Stems (天干)

`GET /api/reference/stems` — Ten heavenly stems (十天干): name, wuxing, yin-yang.

```bash
curl -s https://liki.hk/api/reference/stems
```

## Branches (地支)

`GET /api/reference/branches` — Twelve earthly branches (十二地支): name, wuxing, canggan, zodiac, hour range.

```bash
curl -s https://liki.hk/api/reference/branches
```

## Na Yin

`GET /api/reference/nayin` — Sixty JiaZi nayin five-wuxing table (六十甲子纳音).

```bash
curl -s https://liki.hk/api/reference/nayin
```

## Shen Sha

`GET /api/reference/shensha` — Shensha rules table (神煞规则): tianyi guiren, yuede, tiande.

```bash
curl -s https://liki.hk/api/reference/shensha
```

## Zodiac

`GET /api/reference/zodiac` — Zodiac relationships (生肖关系): liuhe (六合), sanhe (三合), sanhui (三会), liuchong (六冲), liuhai (六害), xiangxing (相刑).

```bash
curl -s https://liki.hk/api/reference/zodiac
```

## Mansions

`GET /api/reference/mansions` — Twenty-eight mansions reference (二十八宿).

```bash
curl -s https://liki.hk/api/reference/mansions
```

## Trigrams

`GET /api/reference/trigrams` — Eight trigrams reference (八卦): name, image, wuxing, direction.

```bash
curl -s https://liki.hk/api/reference/trigrams
```

## Huang Dao

`GET /api/reference/huangdao` — Huangdao twelve gods reference (黄道十二神).

```bash
curl -s https://liki.hk/api/reference/huangdao
```

## 24 Shan

`GET /api/reference/24-shan` — Twenty-four mountains reference (二十四山), fengshui compass directions.

```bash
curl -s https://liki.hk/api/reference/24-shan
```

## Location

`GET /api/location` — IP-based geolocation. Returns country, city, currency. Used by the chat agent to suggest default city and timezone.

```bash
curl -s https://liki.hk/api/location
```

## Solar Terms

`GET /api/solar-terms` — Current year's solar term months (12 months, each with start/end dates, plus current month).

```bash
curl -s https://liki.hk/api/solar-terms
```

## Health

`GET /api/health` — Service health check.

```bash
curl -s https://liki.hk/api/health
```
