# Qiming API (起名)

Base: `POST /api/qiming/*`

All responses wrapped in `{"data":{...}}`.

## WuGe

`POST /api/qiming/wuge` — Enumerate auspicious stroke combinations and character candidates for a surname.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname, 1-2 Chinese characters |
| yong_shen | string | yes | YongShen element: 木/火/土/金/水 |
| xi_shen | [string] | no | XiShen element array |

```bash
curl -s -X POST https://liki.hk/api/qiming/wuge \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","yong_shen":"水","xi_shen":["金"]}'
```

## Compose

`POST /api/qiming/compose` — Compose name candidates from character pools. Only yong+yong, yong+xi, and xi+yong pairs are allowed. Names that fail ping-ze validation are filtered.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname |
| combos | [object] | yes | Stroke combos from wuge: `[{stroke1, stroke2}]` |
| yong_chars | object | yes | Yong-shen chars `{stroke: [char]}` |
| xi_chars | object | yes | Xi-shen chars `{stroke: [char]}` |

Returns a string array of name candidates.

```bash
curl -s -X POST https://liki.hk/api/qiming/compose \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","combos":[{"stroke1":5,"stroke2":8}],"yong_chars":{"5":["沐","沛"],"8":["洪","涛"]},"xi_chars":{"5":["圣"],"8":["恩","轩"]}}'
```

## Detail

`POST /api/qiming/detail` — Batch query five-grid, three-talent, and phonetic details.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname |
| names | [string] | yes | Given names to query, 1-50 |

Returns an array of NameCandidate objects.

```bash
curl -s -X POST https://liki.hk/api/qiming/detail \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","names":["沐洪","沐涛","沛恩"]}'
```

## Evaluate

`POST /api/qiming/evaluate` — Evaluate a specific Chinese name.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| surname | string | yes | Surname, 1-2 Chinese characters |
| given_name | string | yes | Given name, 1-2 Chinese characters |
| yong_shen | string | yes | YongShen element: 木/火/土/金/水 |

```bash
curl -s -X POST https://liki.hk/api/qiming/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"surname":"李","given_name":"沐泽","yong_shen":"水"}'
```
