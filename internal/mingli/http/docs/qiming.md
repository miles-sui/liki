# Qiming API (起名)

Base: `POST /api/qiming/*`, `GET /api/qiming/*`

## Qiming Generate

`POST /api/qiming/generate` — Generate Chinese name candidates from surname, yong shen, xi shen, and zodiac. Engine uses 8105-character wuxing-annotated database.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname, 1-2 Chinese characters |
| yong_shen | string | yes | Yong shen element (金/木/水/火/土) |
| xi_shen | [string] | no | Xi shen element array |
| zodiac | int | yes | Zodiac branch, 1-12 (子=1..亥=12) |
| gender | string | no | "male" or "female" |
| limit | int | no | Candidates to return, 1-200 |

Returns: wu ge (五格) scores, san cai (三才) config, phonetic analysis, wuxing match highlights.

```bash
curl -s -X POST https://api.tokflux.com/api/qiming/generate \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","yong_shen":"水","xi_shen":["金"],"zodiac":5,"limit":10}'
```

## Qiming Evaluate

`POST /api/qiming/evaluate` — Evaluate a specific Chinese name (测名) against wuxing requirements and zodiac compatibility.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname, 1-2 Chinese characters |
| given_name | string | yes | Given name, 1-2 Chinese characters |
| yong_shen | string | yes | Yong shen element |
| zodiac | int | yes | Zodiac branch, 1-12 |

Returns: wu ge scores, san cai config, phonetic mark, wuxing match status, zodiac notes.

```bash
curl -s -X POST https://api.tokflux.com/api/qiming/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","given_name":"沐泽","yong_shen":"水","zodiac":5}'
```

## Qiming Characters

`GET /api/qiming/characters` — List usable characters by element and stroke range.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| element | string | yes | Element (wood/fire/earth/metal/water or 木/火/土/金/水) |
| stroke_min | int | no | Minimum stroke count |
| stroke_max | int | no | Maximum stroke count |
| limit | int | no | Results to return, 1-200 |

```bash
curl -s 'https://api.tokflux.com/api/qiming/characters?element=wood&limit=10'
```
