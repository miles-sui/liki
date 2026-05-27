# 25types — 错误码参照

> **一份表，前后端共用。不再散落各文档。**

---

## 1. 错误码总览

| 错误码 | HTTP | 含义 |
|--------|------|------|
| `invalid_request` | 400 | 参数缺失、格式错误或语义无效 |
| `unauthorized` | 401 | Token 缺失或无效 |
| `token_expired` | 401 | JWT 过期 — 客户端应触发刷新或跳转登录 |
| `payment_required` | 402 | 需要通行证 — 客户端展示升级引导，不跳转 |
| `forbidden` | 403 | 已认证但无权限（私有用户、非匹配用户） |
| `not_found` | 404 | 资源不存在、已删除、或不可见 |
| `conflict` | 409 | 资源冲突（重复名、活跃请求、冷却期内） |
| `rate_limited` | 429 | 请求频率超限，含 `Retry-After` header |
| `internal` | 500 | 服务端错误 — message 不得含堆栈/SQL/路径 |

---

## 2. 完整条件索引

### invalid_request (400)

| 条件 | 端点 | message |
|------|------|---------|
| name 或 password 缺失 | `POST /api/auth/register` | `Name and password are required` |
| password 不足 8 字符 | `POST /api/auth/register` | `Password must be at least 8 characters` |
| name 或 password 缺失 | `POST /api/auth/login` | `Name and password are required` |
| answers 为空或缺失 | `POST /api/assessments` | `answers is required` |
| QID 不在题库中 | `POST /api/assessments` | `Invalid qid: xxx` |
| previous_qids 格式错误 | `GET /api/assessments/next-round` | `Invalid previous_qids format` |
| anonymous_token 缺失 | `POST /api/assessments/claim` | `anonymous_token is required` |
| with_user_id 或 with_name 缺失 | `POST /api/bonds` | `with_user_id or with_name is required` |
| 与自己对比 | `POST /api/bonds` | `Cannot compare with yourself` |
| answers 为空或缺失 | `POST /api/m/{token}` | `answers is required` |
| 请求体格式错误 | `POST /api/bonds` | `Invalid request body` |
| 请求体至少需要一个字段 | `PUT /api/users/me` | `At least one field is required` |
| reviewer_name 缺失（匿名评审） | `POST /api/r/{token}` | `reviewer_name is required` |
| answers 为空或缺失 | `POST /api/r/{token}` | `answers is required` |
| QID 不在题库中 | `POST /api/r/{token}` | `Invalid qid: xxx` |
| 验证 token 无效或过期 | `GET /api/auth/verify-email` | `Invalid or expired token` |
| 重置 token 无效或过期 | `POST /api/auth/reset-password` | `Invalid or expired token` |
| password 不足 8 字符 | `POST /api/auth/reset-password` | `Password must be at least 8 characters` |
| password 不足 8 字符 | `PUT /api/auth/password` | `Password must be at least 8 characters` |
| plan 缺失或无效 | `POST /api/payments/checkout` | `Plan is required and must be "monthly" or "yearly"` |
| email 缺失或格式错误 | `POST /api/payments/checkout` | `Valid email is required` |
| 签名缺失 | `POST /api/payments/webhook` | `Signature is required` |
| code 缺失 | `POST /api/payments/redeem` | `Code is required` |
| count 缺失或范围错误 | `POST /api/admin/codes` | `Count must be between 1 and 1000` |
| email 格式无效（密码找回） | `POST /api/auth/forgot-password` | (始终返回 200, 不暴露) |

### unauthorized (401)

| 条件 | 端点 | message |
|------|------|---------|
| Token 缺失或格式错误 | 全部 (L) 端点 | `Authentication required` |
| Token 无效（密码变更/登出） | 全部 (L) 端点 | `Token invalidated` |
| 当前密码错误 | `PUT /api/auth/password` | `Current password is incorrect` |
| 用户名或密码错误 | `POST /api/auth/login` | `Invalid username or password` |
| 账户已注销（冷静期外尝试登录） | `POST /api/auth/login` | `Invalid username or password` |
| 不是自己创建的链接 | `GET /api/reviews/{id}` | `Forbidden` |

### token_expired (401)

| 条件 | 端点 | message |
|------|------|---------|
| Token 过期（>30 天） | 全部 (L) 端点 | `Token expired` |

> 与 `unauthorized` 同为 401，区分为独立 code。客户端按 code 分流——`token_expired` 额外展示"登录已过期"toast，`unauthorized` 直接跳转登录页。

### payment_required (402)

| 条件 | 端点 | message |
|------|------|---------|
| 免费用户访问付费 Bond 功能 | `POST /api/bonds` | `Passport required` |
| 免费用户访问 12 月预报 | `GET /api/flow/yearly`（非当月） | `Passport required` |
| 免费用户访问他评深入分析 | `GET /api/assessments/peers`（深度字段） | `Passport required` |
| 免费用户访问 Bond 历史 | `GET /api/flow/monthly`（历史月份） | `Passport required` |

> 客户端展示通行证升级引导弹窗，不跳转页面。

### forbidden (403)

| 条件 | 端点 | message |
|------|------|---------|
| 目标用户设为私有 | `GET /api/users/{id}` | `This user's profile is private` |
| 非链接创建者查看详情 | `GET /api/reviews/{id}` | `Forbidden` |

> 注销用户与不存在用户返回同一 message（防用户枚举），但 HTTP 状态码不同：注销=403，不存在=404。前端按 code 分流。

### not_found (404)

| 条件 | 端点 | message |
|------|------|---------|
| 用户不存在或已注销 | `GET /api/users/{id}` | `User not found` |
| 评估不存在或所属用户已私有/注销 | `GET /api/assessments/{id}` | `Assessment not found` |
| 链接不存在或已删除 | `DELETE /api/reviews/{id}` | `Review link not found` |
| 链接不存在或已删除 | `POST /api/reviews/{id}/renew` | `Review link not found` |
| 链接不存在或已删除 | `GET /api/reviews/{id}` | `Review link not found` |
| 链接不存在或已删除 | `GET /api/r/{token}` | `Review link not found` |
| 链接已过期（提交时） | `POST /api/r/{token}` | `Review link has expired` |
| 无 profile（未做过评估） | `GET /api/flow` | `No profile found. Submit an assessment first.` |
| 无 profile（未做过评估） | `GET /api/flow/yearly` | `No profile found. Submit an assessment first.` |
| 双方缺一 profile | `POST /api/bonds` | `Both users must have completed an assessment` |
| 无 profile（未做过评估） | `GET /api/assessments/peers` | `No profile found. Submit an assessment first.` |
| 激活码不存在或已过期 | `POST /api/payments/redeem` | `Code not found or expired` |
| 链接不存在或已删除 | `GET /api/m/{token}` | `Match link not found or deleted` |
| 链接不存在或已删除 | `DELETE /api/match-links/{id}` | `Match link not found` |
| 创建者无 profile | `POST /api/m/{token}`（use_existing） | `Creator profile not found` |
| 用户无 profile | `POST /api/m/{token}`（use_existing） | `No existing profile found — complete an assessment first` |

### conflict (409)

| 条件 | 端点 | message |
|------|------|---------|
| 用户名已被注册 | `POST /api/auth/register` | `Username already exists` |
| 用户名已被占用 | `PUT /api/users/me` | `Username already taken` |
| 激活码已用完 | `POST /api/payments/redeem` | `Code has been fully redeemed` |
| 用户已兑换过此激活码 | `POST /api/payments/redeem` | `Code already redeemed by this user` |

### rate_limited (429)

速率限制在 Caddy 代理层实施，以真实客户端 IP 为粒度。Go 应用层不做限流。

| 条件 | 端点 | 速率 |
|------|------|------|
| 单 IP 认证操作超限 | `POST /api/auth/*` | 20 req/min |
| 单 IP 评估提交超限 | `POST /api/assessments` | 30 req/min |

响应含 `Retry-After` header（由 Caddy `caddy-ratelimit` 模块注入）。

### internal (500)

| 条件 | message |
|------|---------|
| DB 连接失败 | `An unexpected error occurred` |
| 计算引擎异常 | `An unexpected error occurred` |
| 序列化/反序列化异常 | `An unexpected error occurred` |
| 邮件发送失败 | `An unexpected error occurred` |
| 支付网关通信异常 | `An unexpected error occurred` |

message 不得包含堆栈跟踪、SQL 语句、文件路径、表名。仅通用描述。

---

## 3. 客户端调度映射

所有错误以统一 envelope 返回：

```json
{ "error": { "code": "invalid_request", "message": "answers is required" } }
```

前端按 `code` 分流处理：

| code | 客户端行为 |
|------|-----------|
| `invalid_request` | 表单字段提示，不重试。若含 `details` 数组则逐字段定位输入框 |
| `unauthorized` | 清除本地 token，跳转登录页 |
| `token_expired` | 同 `unauthorized`，额外 toast"登录已过期，请重新登录" |
| `payment_required` | 展示通行证升级引导弹窗，不跳转 |
| `forbidden` | 展示"该用户未公开"页面 |
| `not_found` | 展示 404 页面 |
| `conflict` | 展示冲突提示（名称已存在 / 已有请求 / 冷却期至 YYYY-MM-DD） |
| `rate_limited` | 展示限流提示，读取 `Retry-After` header 显示倒计时 |
| `internal` | 展示通用错误提示 + 重试按钮 |

### 全局拦截器

```js
// 所有 fetch 统一拦截 401/402
const origFetch = window.fetch;
window.fetch = async (...args) => {
  const res = await origFetch(...args);
  if (res.status === 401) {
    const body = await res.json().catch(() => ({}));
    const code = body?.error?.code;
    localStorage.removeItem('token');
    if (code === 'token_expired') showToast('登录已过期，请重新登录');
    window.location = '/login?redirect=' + encodeURIComponent(location.pathname);
  }
  if (res.status === 402) {
    showPassportUpgradeModal();
  }
  return res;
};
```

### 表单验证 details

`invalid_request` 响应可能携带 `details` 数组，客户端按 `field` 定位输入框：

```js
const err = data.error;
if (err.details) {
  err.details.forEach(d => {
    const el = document.querySelector(`[name="${d.field}"]`);
    if (el) el.classList.add('input-error');
  });
}
```

---

## 4. Envelope 格式

### 成功 — 单对象

```json
{ "data": { ... } }
```

### 成功 — 列表

```json
{ "data": { "items": [...], "total": 42 } }
```

### 错误

```json
{ "error": { "code": "invalid_request", "message": "human-readable detail" } }
```

### 表单验证错误（扩展 details）

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Validation failed",
    "details": [
      { "field": "name", "code": "already_taken" },
      { "field": "password", "code": "too_short" }
    ]
  }
}
```

---

## 5. 错误码覆盖速查

> 500 和 401（(L) 端点缺 token）为通用错误，不列入下表。下表仅列功能特有的错误码。

| 特征码 | 波及功能 (#) |
|--------|------------|
| 400 invalid_request | 1, 2, 5, 6, 7, 9, 15, 23, 29, 32, 44, 46, 47 |
| 401 unauthorized (密码/凭证) | 2, 6 |
| 401 token_expired | 所有 (L) 端点 — 客户端应触发刷新或跳转登录 |
| 402 payment_required | 33, 34, 37 — 需要通行证的功能 |
| 403 forbidden | 12, 29 |
| 404 not_found | 8, 12, 22, 23, 24, 26, 27, 29, 31, 32, 35, 36, 37, 46 |
| 409 conflict | 1, 15, 29, 46 |
| 429 rate_limited | 1, 2, 7, 29 |

功能编号见 [FRONTEND](../FRONTEND.md) §功能-页面-API 追溯矩阵。

---

> **服务端只设 code 和 message。界面文案和交互逻辑由客户端按 code 决定。**

## 参见

- [INDEX](../INDEX.md) 共享内核与错误码总览
- [API](../API.md) HTTP 契约
- [FRONTEND](../FRONTEND.md) 客户端错误处理实现
