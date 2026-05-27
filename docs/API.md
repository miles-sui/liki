# 25types — HTTP 契约

> 所有 API 端点的请求/响应 JSON Schema 和错误码。唯一权威源。

---

## 约定

- **Base URL**: `/api`（Go 业务 API）; 静态内容构建时注入前端，评估题库由后端 API serve
- **Envelope — 成功单对象**: `{ "data": { ... } }`
- **Envelope — 成功列表**: `{ "data": { "items": [...], "total": 42 } }`
- **Envelope — 错误**: `{ "error": { "code": "...", "message": "..." } }`
- **Auth 标记**: 🔒 = 需认证（401 若未认证），🔓 = 可选认证
- **付费标记**: 💰 = 需付费（402 若配额不足/未订阅）
- **SSE**: 报告生成使用 Server-Sent Events（`text/event-stream`），非请求-响应模式
- **Swagger**: Go handler 用 swag 注解声明契约，`swag init` 生成 `swagger.yaml`，CI 检测注解同步。`openapi-typescript` 从 YAML 生成前端 TS 类型
- **域命名**: API 路由按领域分组（`/api/bazi/` `/api/huangli/` `/api/fengshui/` `/api/qiming/` `/api/reference/`），每组管理自己的域
- **引擎与解读分离**: 域端点返回结构化引擎数据（数字、枚举、类型标签），自然语言解读统一交由 LLM 处理

---

## Auth (8)

### `POST /api/auth/register`

```
Request:  { "name": "alice", "email": "alice@example.com", "password": "secret123", "anonymous_token": "..." }

Response 201:
  { "data": {
      "token": "eyJhbGciOiJIUzI1NiIs...",
      "user": { "id": 1, "name": "alice", "email": "", "email_verified": false, "is_public": false }
  } }
| 400 | invalid_request | Name, email, and password are required |
| 400 | invalid_request | Invalid email format |
| 400 | invalid_request | Password must be at least 8 characters |
| 400 | invalid_request | Password must not contain username |
| 409 | conflict | Username already exists |
| 409 | conflict | Username is reserved |
| 409 | conflict | Email already registered |
```

`email` 必填。邮箱写入 `pending_email`，同时生成验证 token 并发送验证邮件（`SendVerificationEmail`）。注册成功直接返回 JWT，不阻塞功能——未验证邮箱也能使用全部功能。`email_verified` 始终返回 `false`，直到用户点击验证链接。

带 `anonymous_token` 时注册成功即自动认领匿名评估。

注册时拒绝保留名（见 URL 保留名列表）。

### `POST /api/auth/login`

```
Request:  { "name": "alice", "password": "secret123" }
          或 { "name": "alice@example.com", "password": "secret123" }

Response 200:
  { "data": {
      "token": "eyJhbGciOiJIUzI1NiIs...",
      "user": { "id": 1, "name": "alice", "email": "...", "email_verified": true, "is_public": false }
  } }
| 400 | invalid_request | Name and password are required |
| 401 | unauthorized | Invalid username or password |
```

`name` 字段同时接受用户名和已验证的邮箱。输入含 `@` 时按邮箱查询（仅查 `email_verified_at IS NOT NULL` 的记录），否则按用户名查询。都找不到统一返回 401（防枚举）。

注销恢复：7 天冷静期内登录 → 清除 `deactivated_at`，token_version + 1。超 7 天返回 401。

### `POST /api/auth/logout` 🔒

```
Request:  (empty)
Response 200: { "data": { "status": "logged_out" } }
| 401 | unauthorized | Token missing or invalid |
```

### `PUT /api/auth/password` 🔒

```
Request:  { "current_password": "old12345", "new_password": "new_secret" }

Response 200: { "data": { "token": "eyJhbGciOiJIUzI1NiIs...", "status": "password_changed" } }
| 400 | invalid_request | Password must be at least 8 characters |
| 401 | incorrect_password | Current password is incorrect |
```

成功后 token_version + 1，返回新 token。

### `GET /api/auth/verify-email`

```
Query:    ?token=xxx

Response 200: { "data": { "email_verified": true } }
| 400 | invalid_request | Invalid or expired token |
```

Token 有效期 24 小时。若 `pending_email` 非空则将其移至 `email`。

邮件中的验证链接格式：`https://25types.com/{locale}/verify-email?token=xxx`，语言前缀按用户请求时的 `X-Locale` 决定。

### `POST /api/auth/resend-verification` 🔒

```
Response 200: { "data": { "email": "suiqiang@foxmail.com", "status": "sent" } }
| 400 | no_email         | No email to verify |
| 409 | already_verified | Email is already verified |
| 401 | unauthorized     |
```

目标邮箱选取规则：优先 `pending_email`，其次未认证的 `email`。若两者都无则返回 400，若已认证且无待验证邮箱则返回 409。Token 有效期 24 小时，覆盖旧 token。

本地开发无邮件服务时，验证链接打印到服务器日志：
`[email] no sender — verification link: http://localhost:8080/en/verify-email?token=xxx`

### `POST /api/auth/forgot-password`

```
Request:  { "email": "alice@example.com" }

Response 200: { "data": { "status": "ok" } }
// 无论邮箱是否存在，始终返回 200（防枚举）
```

邮件中的重置链接格式：`https://25types.com/{locale}/reset-password?token=xxx`，语言前缀按用户请求时的 `X-Locale` 决定。Token 有效期 15 分钟。

### `POST /api/auth/reset-password`

```
Request:  { "token": "xxx", "password": "newpassword" }

Response 200: { "data": { "status": "password_changed" } }
| 400 | invalid_request | Password must be at least 8 characters |
| 400 | invalid_request | Invalid or expired token |
```

Token 有效期 15 分钟。成功后 token_version + 1。

---

## Users (4)

### `GET /api/users/me` 🔒

```
Response 200:
  { "data": { "id": 1, "name": "alice", "email": "...", "email_verified": true,
              "is_public": false, "supporter_since": null } }
| 401 | unauthorized |
```

`supporter_since`: 首次捐赠时间（ISO 8601），`null` 表示未曾捐赠。前端据此派生 `isSupporter` 布尔值。纯身份标识，不给功能特权。

### `PATCH /api/users/me` 🔒

```
Request:  { "name": "new_alice", "is_public": true }  // 全字段可选
          { "email": "new@example.com" }
          { "birth_info": { "year": 1995, "month": 8, "day": 21, "hour": 14, "minute": 30,
              "longitude": 116.4, "timezone": 120, "gender": "male" } }

Response 200: { "data": { "id": 1, "name": "new_alice", "email": "...", "email_verified": false } }
| 400 | invalid_request | At least one field is required |
| 400 | invalid_request | Name cannot be empty |
| 401 | unauthorized |
| 409 | conflict | Username is reserved |
| 409 | conflict | Name already taken |
| 409 | conflict | Email already in use |
```

`birth_info` 保存后，`GET /api/profiles/{name}` 实时计算并返回 `mingli_chart`。

改邮箱：新地址写入 `pending_email`，发送验证邮件。`email` 保持旧值至验证完成。

### `DELETE /api/users/me` 🔒

```
Request:  (empty)
Response 200: { "data": { "status": "deactivated", "reactivate_by": "2026-05-13T12:00:00Z" } }
| 401 | unauthorized |
```

软删除 + 7 天冷静期。立即 token_version + 1。

### `GET /api/users/me/export` 🔒

```
Response 200:
  { "data": {
      "user": { "id": 1, "name": "alice", "email": "..." },
      "assessments": [
        { "id": 42, "type": "self", "identity_id": "WF",
          "profile": { "d": {...}, "p": {...} },
          "answers_json": [{ "qid": "Q01", "selections": ["W","F"] }, ...],
          "created_at": "..." }
      ],
      "review_links": [
	        { "id": 1, "token": "...", "expires_at": "...", "created_at": "..." }
	      ],
      "exported_at": "..."
  } }
| 401 | unauthorized |
```

包含用户信息、评估数据、review link 记录。

---

## Profile (2)

### `GET /api/profiles/{name}` 🔓

```
Response 200 (owner):
  { "data": {
      "user": { "id": 1, "name": "alice" },
      "profile": { "d": {...}, "p": {...}, "identity": { "label": "...", "id": "WF", "category": "composite" } },
      "flow_month": { "month_id": "巳月", "month_en": "Si Yue", "generates": 1, "restrains": 2 },
      "peers": { "d": {...}, "count": 3 },
      "is_owner": true,
      "is_public": true
  } }

Response 200 (visitor, public):
  { "data": { "user": {...}, "profile": {...}, "flow_month": {...},
      "peers": {...}, "is_owner": false, "is_public": true } }

Response 200 (owner, with birth_info — includes bazi_chart):
  { "data": { "user": {...}, "profile": {...}, "flow_month": {...},
      "peers": {...}, "is_owner": true, "is_public": true,
      "bazi_chart": { /* ChartOutput 结构（与 POST /api/bazi/chart 一致） */ },
      "birth_info": { "year": 1995, "month": 8, "day": 21, "hour": 14, "minute": 30,
        "longitude": 116.4, "timezone": 8.0, "gender": "male" }
  } }

Response 200 (no assessment):
  { "data": { "user": {...}, "profile": null, "is_owner": ..., "is_public": ... } }
| 404 | not_found | User not found (also for private profiles viewed by non-owner) |
```

私密用户对非所有者返回 404，与用户不存在统一处理。

用户保存 `birth_info` 后，profile 响应自动包含 `bazi_chart`（实时计算）和 `birth_info`（已存值）。`bazi_chart` 字段与 `POST /api/bazi/chart` 响应格式一致，包含完整的四柱、藏干、十神、纳音、十二长生、大运。

### `GET /api/profiles/{name}/bonds` 🔒

```
Response 200:
  { "data": {
      "items": [
        { "other_user": { "name": "bob", "identity_label": "Pioneer", "identity_id": "W" },
          "bond": { "self": {...}, "other": {...}, "delta_a": {...}, "delta_b": {...} },
          "source": "instant" | "match_link",
          "created_at": "..." }
      ],
      "total": 3
  } }
| 401 | unauthorized |
| 404 | not_found | (非本人查看时返回，与用户不存在统一处理) |
```

仅 Profile 主人可查看。

---

## BaZi (8)

八字命理域。全部 POST，显式传参（不依赖 auth 隐式传递）。`POST /api/bazi/chart` 是主入口——包含完整排盘、大运详表、喜用神、五行分布，可直接交由 LLM 解读。

### `POST /api/bazi/chart`

排盘。提交出生信息，返回完整八字命盘。

```
Request:
  {
    "year": 1990, "month": 6, "day": 15, "hour": 14, "minute": 30,
    "longitude": 121.47, "timezone": 8.0, "gender": "male"
  }

Response 200 { "data": {
  "year_pillar": {
    "stem": 7, "branch": 7, "nayin": "庚申",
    "hidden_stems": { "main": 7, "mid": 9, "minor": 3 },
    "ten_gods": [
      { "stem": 7, "ten_god": "比肩", "source": "stem" },
      { "stem": 9, "ten_god": "食神", "source": "main_qi" },
      { "stem": 3, "ten_god": "七杀", "source": "mid_qi" }
    ],
    "life_stages": [{ "stage": "帝旺", "stem": 7, "hidden_stem": 0 }],
    "shensha": [{ "name": "天乙贵人", "type": "吉" }],
    "is_void": false, "is_self_he": false, "is_kui_gang": true
  },
  "month_pillar": { ... },
  "day_pillar":   { ... },
  "hour_pillar":  { ... },
  "day_master": "壬",
  "dayun": {
    "start_age": 5, "direction": "forward",
    "pillars": [{
      "stem": 7, "branch": 7, "age_start": 5, "age_end": 14,
      "name": "庚申", "element": "金", "ten_god": "七杀运",
      "stem_rels": [...], "branch_rels": [...], "shensha": [...], "is_current": false
    }, ...],
    "current_pillar_index": 2
  },
  "element_count": { "木": 2, "火": 1, "土": 1, "金": 3, "水": 1 },
  "life_stages": [{ "name": "帝旺", "branch": 7 }, ...],
  "solar_time_minutes": 851.5,
  "solar_datetime": "1990-06-15T14:11:30+08:00",
  "bazi_datetime": "1990-06-15 未时",
  "full_he_hui": [...],
  "gong_jia": [...],
  "tai_yuan_ming_gong": { ... },
  "nayin_relations": [...],
  "sanqi_name": "",
  "zodiac": "马",
  "season": "夏",
  "lunar_month": "五月",
  "hour_range": "13:00-15:00",
  "xun_name": "甲子旬",
  "wang_shuai": { "木": "休", "火": "旺", "土": "相", "金": "死", "水": "囚" },
  "day_mansion": { "index": 10, "name": "虚日鼠", ... },
  "yong_shen": {
    "fuyi": { "strength": "身强", "pattern": "建禄格", "yong": "金", "xi": "土", "ji": "火" },
    "tiaohou": { "season": "夏", "yong": "壬", "xi": "庚", "ji": "土", "detail": "..." }
  }
} }
| 400 | invalid_birth_info | 字段缺失或值非法 |
```

`timezone` 默认 8 (UTC+8)，`longitude` 默认 120.0。`gender` 必填。

### `POST /api/bazi/bond`

合八字。两组出生信息，返回全量 chart_a/chart_b + 五维交叉分析。

```
Request: {
  "a": { "year": 1990, "month": 6, "day": 15, "hour": 14, "minute": 30, "longitude": 121.47, "timezone": 8.0, "gender": "male" },
  "b": { "year": 1992, "month": 3, "day": 8, "hour": 10, "minute": 0, "longitude": 116.40, "timezone": 8.0, "gender": "female" }
}

Response 200 { "data": {
  "chart_a": { /* 完整 ChartOutput（与 POST /api/bazi/chart 一致） */ },
  "chart_b": { /* 完整 ChartOutput */ },
  "bond": {
    "pillar_cross": { /* 16 对柱交互 */ },
    "ten_god_cross": { /* 十神互视 */ },
    "nayin_cross": { /* 纳音五行关系 */ },
    "shensha_cross": { /* 神煞互现 */ },
    "structure": { /* 结构比较 */ }
  }
} }
| 400 | invalid_birth_info / chart_required |
```

### `POST /api/bazi/liunian`

流年排盘。传入四柱和目标年份，可选大运柱。

```
Request: { "bazi": { "year": {"stem": 8, "branch": 10}, "month": {"stem": 1, "branch": 1}, "day": {"stem": 3, "branch": 5}, "hour": {"stem": 7, "branch": 7} }, "year": 2026, "current_dayun": { "stem": 1, "branch": 1 } }
Response 200: { "data": { "year": 2026, "year_stem": 3, "year_branch": 7, ... } }
| 400 | invalid_request | year out of range |
```

`bazi` 为 `{"year": Pillar, "month": Pillar, "day": Pillar, "hour": Pillar}` 的命名字段对象。`current_dayun` 可选。

### `POST /api/bazi/liuyue`

流月运势。

```
Request: { "bazi": { "year": {...}, "month": {...}, "day": {...}, "hour": {...} }, "year": 2026, "month": 5 }
Response 200: { "data": { "year": 2026, "month": 5, "month_name": "巳月", ... } }
| 400 | invalid_request |
```

### `POST /api/bazi/liuri`

流日分析。

```
Request: { "bazi": { "year": {...}, "month": {...}, "day": {...}, "hour": {...} }, "date": "2026-05-25", "dayun_pillar": { ... }, "liunian_pillar": { ... } }
Response 200: { "data": { "date": "2026-05-25", "day_name": "...", ... } }
```

`dayun_pillar` 和 `liunian_pillar` 可选。

### `POST /api/bazi/liushi`

流时分析。

```
Request: { "bazi": { "year": {...}, "month": {...}, "day": {...}, "hour": {...} }, "date": "2026-05-25", "hour": 14 }
Response 200: { "data": { "time": "14:00", "hour_name": "...", ... } }
```

### `POST /api/bazi/xiao-yun`

小运（hour-branch-based yearly fortune）。

```
Request: { "birth": { "year": 1990, ... }, "count": 12 }
Response 200: { "data": [ { "age": 1, "stem": 1, "branch": 1, "name": "甲子", ... }, ... ] }
```

`count` 默认 12，最大 60。

### `POST /api/bazi/xiao-xian`

小限（year-branch-based yearly cycle）。

```
Request: { "birth": { "year": 1990, ... }, "count": 12 }
Response 200: { "data": [ { "age": 1, "branch": 1 }, ... ] }
```

`count` 默认 12，最大 60。

### 已消除端点

`mingge` / `dayun` / `deficiency` / `tiaohou` → 数据已含在 `POST /api/bazi/chart` 响应中，由 LLM 解读。

---

## Huangli (4)

万年历域。GET 查询公开数据（无个人出生信息），POST 对敲需传出生信息做个性化交叉比对。日期 + 日历 + 节气。

### `GET /api/huangli/query?date=2026-05-25`

单日黄历查询。返回日柱干支、建除十二神（含宜忌）、黄道黑道、喜财福神方位、彭祖禁忌、二十八宿值宿。

```
Query:    ?date=2026-05-25
          ?month=2026-05            — 整月黄历
          ?month=2026-05&event_type=wedding  — 整月 + 建除按事件判定宜忌

Response 200 { "data": {
  "days": [{
    "date": "2026-05-25",
    "day_pillar": { "gan": 1, "zhi": 1, "na_yin": "海中金" },
    "jian_chu": "建",
    "suitable": true,
    "marks": ["宜嫁娶", "宜出行"],
    "warnings": ["忌开市"],
    "huangdao": { "index": 0, "name": "青龙", "path": "黄道" },
    "directions": { "xi_shen": "东北", "cai_shen": "西南", "fu_shen": "东南" },
    "taboos": { "stem": "甲不开仓财物耗散", "branch": "子不问卜自惹祸殃" },
    "mansion": { "index": 10, "name": "虚日鼠", ... }
  }],
  "year_month": "2026-05"
} }
| 400 | invalid_request | date or month is required |
```

`event_type` 可选：`wedding`/`engage`/`open`/`sign`/`move` 等。传入后建除十二神自动按事件判定 `suitable`/`marks`/`warnings`。

### `GET /api/huangli/jieqi?year=2026&month=5&day=25`

节气深度 + 人元司令。

```
Response 200 { "data": { "jieqi_depth": {...}, "ren_yuan": {...} } }
```

### `POST /api/huangli/bond`

对敲（个性化择日）。传入出生信息 + 日期/月份，返回黄历数据 + 个人干支/地支/太岁关系标注。

```
Request: {
  "birth_info": { "year": 1990, "month": 6, "day": 15, "hour": 12, "gender": "male" },
  "date": "2026-05-25"
}
或 { "birth_info": {...}, "month": "2026-05", "event_type": "wedding" }

Response 200 { "data": {
  "days": [{
    "date": "2026-05-25",
    "day_pillar": { "gan": 1, "zhi": 1, "na_yin": "海中金" },
    "jian_chu": "建", "suitable": true, "marks": [...], "warnings": [...],
    "huangdao": {...}, "directions": {...}, "taboos": {...}, "mansion": {...},
    "gan_relation": "比肩",
    "zhi_relation": "六合",
    "tai_sui_relation": "值太岁"
  }],
  "year_month": "2026-05"
} }
| 400 | invalid_request | date or month is required |
```

`gan_relation`: 日干与日主十神关系。`zhi_relation`: 日支与出生日支关系（六合/三合/六冲/六害/相刑等）。`tai_sui_relation`: 日支与年支关系。

---

## FengShui (3)

风水域。

### `GET /api/fengshui/san-yuan?year=2026`

三元九运。

```
Response 200 { "data": { "current": { "year": 2026, "yuan": "...", ... }, "all_periods": [...] } }
```

### `POST /api/fengshui/minggua`

命卦计算（八宅）。

```
Request: { "year": 1990, "gender": "male" }
Response 200 { "data": { "ming_gua": { "gua": {...}, "gua_number": 1, "group": "东四命" }, "all_trigrams": [...] } }
| 400 | invalid_request | year out of range (1900-2200) / gender must be 'male' or 'female' |
```

### `POST /api/fengshui/hecan`

风水合参。八字用神 + 八宅命卦 + 年紫白飞星三层数据并列输出，不评分。

```
Request: { "ming_gua": 1, "yong_shen": "水", "xi_shen": "金", "year": 2026 }
Response 200 { "data": {
  "ming_gua": { "gua": { ... }, "gua_number": 1, "group": "东四命" },
  "ba_zhai_dirs": { "sheng_qi": ["南"], "tian_yi": ["东"], ... },
  "year_stars": { "year": 2026, "center_star": { ... }, "palaces": [...] },
  "yong_shen_elem": "水",
  "xi_shen_elem": "金"
} }
| 400 | invalid_request | ming_gua must be 1-9 |
```

---

## Bonds (1)

### `POST /api/bonds` 🔒

```
Request:  { "with_user_id": 2 }  或 { "with_name": "alice" }

Response 200:
  { "data": {
      "self": {...}, "other": {...},
      "delta_a": {...}, "delta_b": {...},
      "other_user": { "name": "bob", "identity_label": "Pioneer", "identity_id": "W" }
  } }
| 400 | invalid_request |
| 401 | unauthorized |
| 404 | not_found | No profile found for one or both users |
```

即时对比。双方都有 Profile 即可调用，不需要匹配关系。计算后自动写入 `bond_events`（`link_id=NULL`, `source=instant`）。不限次数，无需付费。

---

## Match Links (6)

匹配分享链接。统一入口，`type` 区分评估匹配（`"assessment"`）和八字匹配（`"mingli"`）。同一资源，同一 token 路径——消除重复的 CRUD 端点组。

### `POST /api/match-links` 🔒

创建匹配链接。

```
Request:  { "type": "assessment" }
          { "type": "mingli" }

Response 201:
  { "data": {
      "id": 1,
      "type": "assessment",
      "token": "d4f8a1c2b3e4f5a6c7d8e9f0a1b2c3d4",
      "url": "/m/d4f8a1c2b3e4f5a6c7d8e9f0a1b2c3d4"
  } }
| 400 | invalid_request | type is required and must be "assessment" or "mingli" |
| 400 | invalid_request | mingli type requires saved birth_info |
| 401 | unauthorized |
```

`type=mingli` 时创建者需已保存 `birth_info`。不限次数，无需付费。

### `GET /api/match-links` 🔒

```
Response 200:
  { "data": {
      "items": [
        { "id": 1, "type": "assessment", "token": "...", "match_count": 2, "created_at": "..." },
        { "id": 2, "type": "mingli", "token": "...", "match_count": 1, "created_at": "..." }
      ],
      "total": 2
  } }
```

### `DELETE /api/match-links/{id}` 🔒

```
Response 200: { "data": { "status": "deleted" } }
| 404 | not_found |
```

### `GET /api/m/{token}`

```
Response 200 (type=assessment):
  { "data": { "type": "assessment", "token": "d4f8a1c2...", "creator_name": "alice", "valid": true } }

Response 200 (type=mingli):
  { "data": { "type": "mingli", "token": "d4f8a1c2...", "creator_name": "alice", "valid": true,
      "chart_a": { /* ChartOutput 结构（与 POST /api/bazi/chart 一致） */ }
  } }
| 404 | not_found | (链接不存在或已删除) |
```

`type=mingli` 时返回创建者的八字概览。创建者无出生信息时 `chart_a` 为空。

### `POST /api/m/{token}` 🔓

根据链接类型提交对应数据。

```
Request (type=assessment, anonymous):
  { "answers": [
      { "qid": "Q01", "selections": ["W","F"] },
      ...(全部 30 题)
    ],
    "anonymous_token": "...",
    "other_name": "name shown to link creator" }

Request (type=assessment, use existing profile):
  { "use_existing": true }

Request (type=mingli, with birth info):
  { "birth_info": {
      "year": 1990, "month": 6, "day": 15,
      "hour": 12, "minute": 0,
      "longitude": 120.0, "timezone": 120.0,
      "is_dst": false, "gender": "female"
    },
    "other_name": "name shown to link creator" }

Request (type=mingli, use existing):
  { "use_existing": true }

Response 201 (type=assessment):
  { "data": {
      "profile": { "d": {...}, "p": {...}, "identity": {...} },
      "assessment_id": 42,
      "bond": { "self": {...}, "other": {...}, "delta_a": {...}, "delta_b": {...} }
  } }

Response 201 (type=mingli):
  { "data": {
      "chart_a": { /* 完整 ChartOutput */ },
      "chart_b": { /* 完整 ChartOutput */ },
      "bond": {
        "pillar_cross": { /* 16 对柱交互 */ },
        "ten_god_cross": { /* 十神互视 */ },
        "nayin_cross": { /* 纳音五行关系 */ },
        "shensha_cross": { /* 神煞互现 */ },
        "structure": { /* 结构比较 */ }
      }
  } }
| 400 | invalid_request | Either use_existing or data is required |
| 400 | invalid_request | answers is required (type=assessment) |
| 400 | invalid_birth_info | (type=mingli, 出生信息校验失败) |
| 404 | not_found |
```

提交全部 30 题（assessment）或出生信息（mingli），计算对应结果。`type=assessment` 时计算 Bond 并写入 `bond_events`；`type=mingli` 时计算合盘并写入 `mingli_match_events`。`anonymous_token` 可用于后续注册认领。

---

## Assessments (6)

### `POST /api/assessments` 🔓

```
Request:
  { "answers": [
      { "qid": "Q01", "selections": ["W","F"] },
      { "qid": "Q02", "selections": ["E","R"] }
    ],
    "anonymous_token": "550e8400-..." }

Response 201 (已登录):
  { "data": {
      "id": 42,
      "profile": { "d": {...}, "p": {...} },
      "identity": { "label": "Pioneer-Luminary", "id": "WF", "category": "composite" },
      "complete": false
  } }

Response 200 (匿名):
  { "data": { "profile": {...}, "identity": {...}, "complete": false } }

Response 200 (匿名 + token):
  { "data": { "profile": {...}, "identity": {...}, "anonymous_token": "...", "complete": false } }
| 400 | invalid_request | answers is required |
| 400 | invalid_request | Invalid qid: xxx |
```

前端累积提交：每次 POST 发送全部已有答案（第 1 组 5 题、第 2 组 10 题……第 6 组 30 题）。后端无状态，仅计算收到的答案。`complete: true` 时表示 30 题齐全，评估完成。

### `GET /api/assessments/questions`

```
Query:    ?locale=en

Response 200:
  { "data": {
      "rounds": [
        { "id": 1, "questions": [
            { "qid": "Q01", "text": "...", "options": [{ "element": "W", "text": "..." }, ...] },
            ...
          ] },
        { "id": 2, "questions": [...] },
        ...
      ]
  } }
```

返回全部 30 题（6 轮 × 5 题）。题文与选项从 `assessment.yaml` 加载，按 locale 返回对应语言。无认证要求。

### `GET /api/assessments` 🔒

```
Query:    ?page=1

Response 200:
  { "data": {
      "items": [
        { "id": 42, "profile": { "d": {...}, "p": {...} },
          "identity": { "label": "...", "id": "WF", "category": "composite" },
          "created_at": "..." }
      ],
      "total": 5
  } }
```

### `GET /api/assessments/{id}` 🔓

```
Response 200:
  { "data": {
      "id": 42, "user_name": "alice",
      "profile": { "d": {...}, "p": {...} },
      "identity": { "label": "...", "id": "WF", "category": "composite" },
      "created_at": "..."
  } }
| 404 | not_found |
```

### `GET /api/assessments/peers` 🔒

```
Response 200:
  { "data": {
      "self": { "profile": { "d": {...}, "p": {...} }, "identity": {...} },
      "peers_aggregated": { "profile": { "d": {...}, "p": {...} }, "identity": {...} },
      "combined": { "profile": { "d": {...}, "p": {...} }, "identity": {...} },
      "peer_count": 3
  } }
| 401 | unauthorized |
| 404 | not_found | No profile found |
```

All peer answers are aggregated — no per-link filtering. This protects reviewer anonymity (individual reviews cannot be singled out).

### `POST /api/assessments/claim` 🔒

```
Request:  { "anonymous_token": "550e8400-..." }
Response 200: { "data": { "claimed": 3 } }
| 400 | invalid_request |
| 401 | unauthorized |
```

---

## Reviews (8)

### `POST /api/reviews` 🔒

```
Request:  (empty)

Response 201:
  { "data": {
      "id": 1,
      "token": "d4f8a1c2b3e4f5a6c7d8e9f0a1b2c3d4",
      "url": "/r/d4f8a1c2b3e4f5a6c7d8e9f0a1b2c3d4",
      "expires_at": "2026-06-05T10:00:00Z"
  } }
```

链接默认 30 天有效。

### `GET /api/reviews` 🔒

```
Response 200:
  { "data": {
      "items": [
        { "id": 1, "token": "...", "subject_name": "alice",
          "submission_count": 5, "created_at": "..." }
      ],
      "total": 2
  } }
```

已删除链接不出现在列表中。

### `GET /api/reviews/{id}` 🔒

```
Response 200:
  { "data": {
      "id": 1, "token": "...", "url": "/r/...",
      "subject_name": "alice", "submission_count": 5,
      "expires_at": "...", "created_at": "...",
      "submissions": [
        { "reviewer_name": "bob", "answered_count": 8, "last_submitted_at": "..." }
      ]
  } }
| 401 | unauthorized |
| 404 | not_found | (非本人创建的链接或链接不存在) |
```

### `DELETE /api/reviews/{id}` 🔒

```
Response 200: { "data": { "status": "deleted" } }
| 404 | not_found |
```

软删除。已提交的他评数据保留。

### `POST /api/reviews/{id}/renew` 🔒

```
Response 200:
  { "data": { "id": 1, "token": "d4f8a1c2...", "expires_at": "2026-07-05T10:00:00Z" } }
| 404 | not_found | (链接不存在或已删除) |
```

延长 30 天。有效或已过期链接均可续期。

### `GET /api/reviews/given` 🔓

```
Query:    ?anonymous_token=...

Response 200:
  { "data": {
      "items": [ { "subject_name": "alice", "answered_count": 8, "created_at": "..." } ],
      "total": 5
  } }
```

### `GET /api/r/{token}`

```
Query:    ?locale=en

Response 200 (有效):
  { "data": {
      "subject_name": "alice",
      "valid": true,
      "recommended_qids": ["Q07", "Q12", "Q18", "Q23", "Q29"],
      "questions": [
        { "qid": "Q07", "text": "...", "options": [{ "element": "W", "text": "..." }, ...] },
        ...
      ]
  } }

Response 200 (过期):
  { "data": { "subject_name": "alice", "valid": false, "expired": true, "questions": [] } }
| 404 | not_found | (链接不存在或已删除) |
```

服务端自动排除当前评审者已答题（基于 reviewer token / 登录身份），问题文本按 `locale` 返回对应语言。前端无需额外请求。

### `POST /api/r/{token}`

```
Request:
  { "reviewer_name": "bob",
    "anonymous_token": "550e8400-...",
    "answers": [ { "qid": "Q01", "selections": ["W","F"] }, ... ] }

Response 201 (登录):
  { "data": { "subject_identity": { "label": "...", "id": "WF", "category": "composite" } } }

Response 201 (匿名):
  { "data": { "subject_identity": {...}, "anonymous_token": "..." } }
| 400 | invalid_request | answers is required / reviewer_name is required |
| 400 | invalid_request | Invalid qid: xxx |
| 404 | not_found |
```

Peer review is one-shot submission — no `data_complete` tracking needed (unlike self-assessment which is incremental).

---

## Flow (3)

### `GET /api/flow` 🔒

```
Response 200:
  { "data": { "month_id": "寅月", "month_en": "Early Spring", "generates": 1, "restrains": 2 } }
| 401 | unauthorized |
| 404 | not_found | No profile found |
```

### `GET /api/flow/yearly` 🔒

```
Response 200:
  { "data": {
      "months": [
        { "id": "寅月", "name_en": "Early Spring", "generates": 1, "restrains": 2 },
        ...
      ],
      "current": "寅月"
  } }
| 401 | unauthorized |
| 404 | not_found | No profile found |
```

全部用户返回 12 个月，无需付费。

### `GET /api/solar-terms`

公开端点，始终返回当前年份数据。响应格式不变。

---

## Reports (7)

报告是核心付费产品。一问一报告，无对话历史。引擎数据由各域端点返回，完整解读通过 Reports 生成。SSE 流式生成。

`scene` 对齐 API 域前缀（`mingli` / `naming` / `huangli` / `relationship` / `career` / `general`），子场景通过 `sub_scene` 区分（如 `scene=mingli, sub_scene=dayun`）。命格解读（`/api/bazi/mingge`）免费且不走报告系统——它是"认识自己"的获客基础设施。

### `POST /api/reports` 🔒

创建报告。提交问题场景和引擎数据，返回报告 ID。始终返回 engine_data 回显。

```
Request:
  { "scene": "mingli" | "naming" | "huangli" | "relationship" | "career" | "general",
    "sub_scene": "dayun" | "liunian" | "parenting" | undefined,
    "question": "我适合学医还是学计算机？",
    "engine_data": { /* 引擎计算的 ground truth。结构与对应域端点返回一致 */ },
    "locale": "zh-CN" }

Response 201:
  { "data": {
      "id": "rpt_abc123",
      "scene": "career",
      "question": "我适合学医还是学计算机？",
      "engine_data": { /* 回显 */ },
      "status": "pending",
      "created_at": "2026-05-23T10:30:00Z"
  } }
| 400 | invalid_request | scene/sub_scene 缺失或 scene 非法 |
| 401 | unauthorized |
```

`scene` 决定了 system prompt 模板和 LLM 输出结构。`engine_data` 是 JSON 对象，结构随 scene 不同——服务端不校验内部结构，仅透传。

### `GET /api/reports` 🔒

报告历史列表。

```
Query:    ?scene=mingge&limit=20&offset=0  (全部可选)

Response 200:
  { "data": {
      "items": [
        { "id": "rpt_abc123", "scene": "career", "question": "我适合学医还是学计算机？",
          "status": "completed", "preview": "你的能量方向是 Wood-Fire，适合需要开创性...",
          "created_at": "2026-05-23T10:30:00Z" }
      ],
      "total": 12
  } }
| 401 | unauthorized |
```

`preview` 为 LLM 输出的第一段（~150 字），供列表卡片展示。

### `GET /api/reports/{id}` 🔒

单份报告完整内容。

```
Response 200:
  { "data": {
      "id": "rpt_abc123",
      "scene": "career",
      "question": "我适合学医还是学计算机？",
      "engine_data": { /* 完整引擎数据 */ },
      "content": "## 职业方向分析\n\n### 你的能量方向\n你的能量方向是 Wood-Fire...",
      "traces": [
        { "text": "你的能量方向是 Wood-Fire，适合需要开创性和持续动力的领域。",
          "source": "D={wood:0.33,fire:0.33,...} → identity=WF" }
      ],
      "status": "completed",
      "created_at": "2026-05-23T10:30:00Z",
      "updated_at": "2026-05-23T10:30:45Z"
  } }
| 401 | unauthorized |
| 404 | not_found |
```

`content` 为 Markdown 格式，前端用 Markdown 渲染器展示。`traces` 为溯源列表，每个溯源块锚定一句解读到一条引擎数据。

### `GET /api/reports/{id}/stream` 🔒 💰

SSE 流式生成报告。LLM 逐 token 推送，前端实时渲染。

```
Headers: Accept: text/event-stream

SSE events:
  event: token      data: "你的能量方向是 Wood-Fire，适合需要开创性"
  event: token      data: "和持续动力的领域。"
  event: trace      data: {"text":"...","source":"D={wood:0.33,fire:0.33} → identity=WF"}
  event: done       data: {"report_id":"rpt_abc123","status":"completed"}
  event: error      data: {"code":"llm_error","message":"LLM service unavailable"}

| 401 | unauthorized |
| 402 | quota_exceeded | 免费配额用尽（3 次/天），需付费 |
| 404 | not_found |
```

**免费配额**：每用户每天 3 次流式生成。已付费用户（月/年订阅）无限制。
**超时**：30 秒无 token 推送则服务端断开，前端可重连。
**重连**：支持 `Last-Event-ID` header，服务端从断点续推。

若报告已有完成的 `content`（重复请求），直接推送 `done` 事件并带完整内容。

### `PUT /api/reports/{id}` 🔒

修改重问。用户修改问题后更新同一份报告。

```
Request:
  { "question": "我适合学工科还是文科？",
    "feedback": "上一版太笼统，想要更具体的专业推荐" }   // 可选

Response 200:
  { "data": {
      "id": "rpt_abc123",
      "question": "我适合学工科还是文科？",
      "status": "pending",
      "updated_at": "2026-05-23T11:00:00Z"
  } }
| 400 | invalid_request | question 缺失 |
| 401 | unauthorized |
| 404 | not_found |
```

更新后 `status` 回到 `pending`，需重新调用 `/stream` 生成新内容。服务端记录 `feedback` 和上一版摘要注入 LLM system prompt。

### `DELETE /api/reports/{id}` 🔒

```
Response 200: { "data": { "status": "deleted" } }
| 401 | unauthorized |
| 404 | not_found |
```

### `GET /api/reports/{id}/share` 🔒

生成报告的公开分享快照。

```
Response 200:
  { "data": {
      "share_url": "/report/rpt_abc123",
      "expires_at": null
  } }
| 401 | unauthorized |
| 404 | not_found |
```

公开快照为只读 HTML 页面（`/report/{id}`），显示引擎数据 + 完整 LLM 文案，无需登录。用户可在设置中选择是否允许分享。

---

## Naming (3)

起名域。`generate` 自动生成候选名，`evaluate` 评测已有名，`characters` 按五行浏览用字。

### `POST /api/qiming/generate`

自动生成候选名列表。

```
Request: {
  "surname": "张",
  "yong_shen": "金",
  "xi_shen": ["土"],
  "zodiac": 7,
  "gender": "male",
  "limit": 20
}

Response 200 { "data": {
  "surname": "张", "surname_element": "火", "yong_shen": "金", "xi_shen": ["土"],
  "zodiac_hint": { "animal": "马", "preferred_radicals": [...], "forbidden_radicals": [...] },
  "candidates": [
    { "name": "张钧硕", "characters": [...], "wu_ge": {...}, "san_cai": {...}, "phonetic": {...}, "highlights": [...] }
  ]
} }
| 400 | invalid_request | surname / yong_shen required |
```

`zodiac` 为年柱地支 1-12。`limit` 默认 20，最大 50。

### `POST /api/qiming/evaluate`

评测指定名字。返回五格、三才、音韵、五行匹配、生肖部首检查。

```
Request: { "surname": "张", "given_name": "三", "yong_shen": "金", "zodiac": 7 }

Response 200 { "data": {
  "surname": "张", "given_name": "张钧",
  "characters": [...], "wu_ge": {...}, "san_cai": {...}, "phonetic": {...},
  "wuxing_match": true, "zodiac_notes": ["钧含宜用部首钅"]
} }
| 400 | invalid_request | surname / given_name required |
```

### `GET /api/qiming/characters?element=金&stroke_min=0&stroke_max=0&limit=50`

按五行元素浏览用字。公开，无需认证。支持中文元素名（木/火/土/金/水）。

```
Response 200 { "data": { "items": [ { "char": "钧", "element": "金", "stroke": 12, ... } ], "total": 86 } }
| 400 | invalid_request | element 非法 |
```

---

## Reference (10)

参考表域。全部 GET，公开。给 LLM function calling 提供结构化 lookup 数据。

### `GET /api/reference/stems`

天干全表（甲~癸）。

### `GET /api/reference/branches`

地支全表（子~亥），含藏干、生肖、时辰段。

### `GET /api/reference/nayin`

60 甲子纳音全表。

### `GET /api/reference/shensha`

神煞规则表（天乙贵人、月德、天德等）。

### `GET /api/reference/zodiac`

生肖合害表（六合、三合、三会、六冲、六害、相刑）。

### `GET /api/reference/mansions`

二十八宿全表。

### `GET /api/reference/trigrams`

八卦全表（乾兑离震巽坎艮坤）。

### `GET /api/reference/huangdao`

黄道黑道十二神（青龙、明堂等）。

### `GET /api/reference/24-shan`

二十四山全表（静态枚举）。

```
Response 200 { "data": { "mountains": [ { "index": 0, "name": "子", "angle": 0, "element": "水", ... }, ... ] } }
```

### `GET /api/reference/cities?q=北京`

城市经纬度搜索。无 `?q=` 返回全部，带 `?q=` 前缀搜索。

```
Response 200: { "data": { "items": [...], "total": 4500 } }
```

---

## Career (1)

## Career (1)

### `GET /api/career/matches` 🔒

职业方向匹配。基于 25types 类型 ID，返回匹配的学科和职业方向。

```
Request:  ?type=WF

Response 200:
  { "data": {
      "type_id": "WF",
      "type_name": "Wood-Fire",
      "majors": ["Computer Science", "Design", "Entrepreneurship", ...],
      "careers": ["Tech Entrepreneur", "UX Designer", "Creative Director", ...],
      "avoid": ["Heavily regulated industries", "Repetitive operational roles"],
      "advice": "Wood creates, Fire expresses. You're a builder who can also sell the vision..."
  } }
| 400 | invalid_request | type query parameter is required |
| 401 | unauthorized |
| 404 | not_found | type not found |
```

类型映射表定义在 `internal/app/http/career.go`，覆盖全部 25 个类型。可用 `GET /api/career/types` 获取所有类型 ID 和名称列表。

---

## Daily (2)

### `GET /api/daily/suggestion` 🔒

每日建议。基于今日干支 + 用户命盘，自动推送轻量建议。免费。

```
Response 200:
  { "data": {
      "date": "2026-05-23",
      "day_stem": 7, "day_branch": 7,
      "day_name": "庚申日",
      "element": "金",
      "suggestion_type": "action",
      "suggestion": "今日金旺，是你的用神日。适合做重要决策——特别是财务和职业方向。",
      "color": "白色、银色",
      "direction": "西方"
  } }
| 401 | unauthorized |
| 404 | not_found | 未保存 birth_info |
```

`suggestion_type` 轮换：`action`（行事方向）、`color`（穿搭颜色）、`emotion`（情绪提示）、`health`（健康小贴士）。≤150 字。用户每日首次打开 App 时展示。

### `POST /api/daily/question` 🔒

每日提问。从每日建议卡片触发，仅限"今天"范围。轻量回答，~150 字。每天 3 次免费。

```
Request:
  { "question": "今天适合出门谈客户吗？",
    "suggestion_date": "2026-05-23" }

Response 200:
  { "data": {
      "question": "今天适合出门谈客户吗？",
      "answer": "今日庚申日，金旺。金是你的用神——今天谈客户、签合同的能量都对。建议上午 9-11 点（巳时，火生金），穿白色/银灰色，坐西边位。",
      "remaining_free": 2
  } }
| 400 | invalid_request | question 缺失 |
| 401 | unauthorized |
| 404 | not_found | 未保存 birth_info |
| 429 | quota_exceeded | 今日免费次数用尽（3 次/天） |
```

问题必须和"今天"相关——超出范围引导到报告系统。`remaining_free` 为今日剩余次数。

---

## Location (1)

### `GET /api/location`

```
Response 200:
  { "data": { "city": "Changsha", "country": "China", "lat": 28.2014, "lng": 112.9611 } }

// Private IP fallback (dev/localhost):
  { "data": { "city": "Beijing", "country": "China", "lat": 39.9042, "lng": 116.4074 } }
```

IP-based geolocation using ip-api.com (free, no API key). Reads client IP from `X-Forwarded-For` header (set by Caddy), falls back to `RemoteAddr`. Private/loopback IPs return Beijing as default. Frontend supplements with browser Geolocation API when backend falls back to Beijing.

---

## Payments (5)

### `POST /api/payments/checkout` 🔒

```
Request:  { "amount": 990 }   // 美分：990 = $9.90, 1990 = $19.90, 2990 = $29.90

Response 200:
  { "data": { "url": "https://checkout.dodopayments.com/..." } }
| 400 | invalid_request | Amount is required and must be 990, 1990, or 2990 |
| 401 | unauthorized |
| 503 | service_unavailable | Payment service not configured |
```

创建 Dodo Payments checkout session，返回支付页 URL。amount 必须是 990 / 1990 / 2990 之一。

### `POST /api/payments/webhook`

```
公共端点，Dodo 回调。无 auth。
签名验证失败返回 401。仅处理 payment.succeeded 事件。
成功后写入 donations 表并 best-effort 发送感谢邮件。
```

### `POST /api/payments/confirm` 🔒

```
Request:  { "payment_id": "pay_xxx" }

Response 200 (confirmed):
  { "data": { "confirmed": true } }
Response 200 (pending / not confirmed):
  { "data": { "confirmed": false } }
| 400 | invalid_request | payment_id is required |
| 401 | unauthorized |
```

用户从 Dodo 支付页返回后调用此端点。服务端向 Dodo API 查询支付状态，若 `status = succeeded` 则创建捐赠记录（通过 `payment_id` 唯一约束实现幂等）。与 webhook 互补 — confirm 提供即时反馈，webhook 提供异步兜底。

### `GET /api/payments/plans`

公开端点。返回当前可用的订阅计划和报告单价。

```
Response 200:
  { "data": {
      "single_report_price": 9.9,
      "currency": "CNY",
      "plans": [
        { "id": "monthly", "name": "月度无限", "name_en": "Monthly Unlimited",
          "amount": 29.0, "interval": "month",
          "features": ["无限生成报告", "报告历史保存"] },
        { "id": "yearly", "name": "年度方案", "name_en": "Yearly Plan",
          "amount": 99.0, "interval": "year",
          "features": ["无限报告", "起名/择日完整方案", "优先新功能"] }
      ]
  } }
```

### `POST /api/payments/subscribe` 🔒

创建订阅。

```
Request:  { "plan_id": "monthly" }

Response 200:
  { "data": { "url": "https://checkout.dodopayments.com/sub/..." } }
| 400 | invalid_request | plan_id 缺失或非法 |
| 401 | unauthorized |
| 409 | conflict | 已有活跃订阅 |
| 503 | service_unavailable | Payment service not configured |
```

订阅状态和配额由服务端通过 webhook 和 API 查询维护，不依赖客户端上报。

```

---

## Health (1)

### `GET /api/health`

```
Response 200: { "data": { "status": "ok", "db": "ok" } }
Response 200 (DB down): { "data": { "status": "ok", "db": "error" } }
Response 200 (no DB): { "data": { "status": "ok", "db": "unavailable" } }
```

存活探针。检查 DB 可读性（`db.Ping()`）。

---

## Stats (1)

### `GET /api/stats`

```
Response 200: { "data": { "total_assessments": 5678 } }
```

公开端点，无需认证。返回全站自评数量（包含 +1000 基础值）。

---

## 错误码总表

| 码 | HTTP | 含义 |
|----|------|------|
| `invalid_request` | 400 | 参数缺失、格式错误或语义无效 |
| `unauthorized` | 401 | Token 缺失或无效 |
| `token_expired` | 401 | JWT 过期——客户端应触发刷新或跳转登录 |
| `payment_required` | 402 | 免费配额用尽，需付费或订阅 |
| `quota_exceeded` | 402 | 操作次数超限（如每日免费流式生成 3 次） |
| `forbidden` | 403 | 已认证但无权限 |
| `not_found` | 404 | 资源不存在、已删除、或不可见 |
| `conflict` | 409 | 资源冲突（重复名、保留名、重复邮箱、已有活跃订阅） |
| `invalid_birth_info` | 400 | 出生信息字段缺失、值非法或 JSON 格式错误（八字专用） |
| `chart_required` | 400 | 合八字请求 a 或 b 缺失（八字专用） |
| `llm_error` | 500 | LLM 服务不可用——SSE 内以 event:error 推送 |
| `rate_limited` | 429 | 请求频率超限，含 `Retry-After` header |
| `internal` | 500 | 服务端错误——message 不含堆栈/SQL/路径 |

---

## 参见

- [INDEX](INDEX.md) — 共享内核、JWT 策略、Auth 标记、Base URL、多语言
- [PRODUCT](PRODUCT.md) — 产品定位、用例矩阵、报告模型、定价
- [FRONTEND](FRONTEND.md) — 前端架构、技术栈、页面清单
- [domain/user](domain/user.md) — User 聚合行为
- [domain/assessment](domain/assessment.md) — Assessment/ReviewLink 聚合行为
- [domain/match](domain/match.md) — Match Link / Bond 聚合行为
- [domain/bazi](domain/bazi.md) — BaZi 八字领域规格
- [domain/naming](domain/naming.md) — 命名引擎领域规格
- [domain/huangli](domain/huangli.md) — 择日引擎领域规格
- [domain/career](domain/career.md) — 职业匹配领域规格
- [appendix/errors](appendix/errors.md) — 完整错误码参照
