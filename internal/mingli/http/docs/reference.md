# Reference Data API

All `GET /api/reference/*`. Simple JSON lookup tables — no computation.

## Stems

`GET /api/reference/stems` — Ten heavenly stems (十天干): name, element, yin-yang.

```bash
curl -s https://api.tokflux.com/api/reference/stems
```

## Branches

`GET /api/reference/branches` — Twelve earthly branches (十二地支): name, element, hidden stems, zodiac, hour range.

```bash
curl -s https://api.tokflux.com/api/reference/branches
```

## Na Yin

`GET /api/reference/nayin` — Sixty JiaZi na yin five-element table (六十甲子纳音).

```bash
curl -s https://api.tokflux.com/api/reference/nayin
```

## Shen Sha

`GET /api/reference/shensha` — Shensha rules table (神煞规则): tian yi gui ren, tao hua, yi ma, yang ren, etc.

```bash
curl -s https://api.tokflux.com/api/reference/shensha
```

## Zodiac

`GET /api/reference/zodiac` — Zodiac relationships (生肖关系): six he (六合), triple he (三合), six chong (六冲), mutual xing (相刑).

```bash
curl -s https://api.tokflux.com/api/reference/zodiac
```

## Mansions

`GET /api/reference/mansions` — Twenty-eight mansions reference (二十八宿).

```bash
curl -s https://api.tokflux.com/api/reference/mansions
```

## Trigrams

`GET /api/reference/trigrams` — Eight trigrams reference (八卦): name, image, element, direction.

```bash
curl -s https://api.tokflux.com/api/reference/trigrams
```

## Huang Dao

`GET /api/reference/huangdao` — Huangdao twelve gods reference (黄道十二神).

```bash
curl -s https://api.tokflux.com/api/reference/huangdao
```

## 24 Shan

`GET /api/reference/24-shan` — Twenty-four mountains reference (二十四山), fengshui compass directions.

```bash
curl -s https://api.tokflux.com/api/reference/24-shan
```

## Cities

`GET /api/reference/cities` — Geolocation lookup for ~300 Chinese cities. Returns longitude, latitude, and name.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| q | string | yes | City name in Chinese (e.g. 上海, 北京) |

```bash
curl -s 'https://api.tokflux.com/api/reference/cities?q=上海'
```

Use this to resolve `longitude` and `timezone` for BaZi chart requests. `timezone = lng/15` rounded to nearest integer.

## Solar Terms

`GET /api/solar-terms` — Current year's 24 solar terms timetable with precise dates.

```bash
curl -s https://api.tokflux.com/api/solar-terms
```

## Health

`GET /api/health` — Service health check.

```bash
curl -s https://api.tokflux.com/api/health
```

## Location

`GET /api/location` — Get city coordinates by request IP.

```bash
curl -s https://api.tokflux.com/api/location
```
