# User 聚合 — 实现规格

> 注册、登录、登出、密码管理、邮箱验证、密码重置、资料编辑、账户注销、GDPR 数据删除。

---

## 1. 聚合边界

```
User ──owns──> Assessment(s) ──produces──> Profile(d, p) + Identity
  │
  ├──owns──> ReviewLink(s) ──receives──> Assessment(s) [type=peer]
  │
  └──participates──> MatchRequest ──accepted──> Match (派生视图, 不建表)
```

| 聚合根 | 生命周期 | 关键约束 |
|--------|---------|---------|
| User | 注册创建，自主注销 | name 唯一 |

---

## 2. 行为规则

### 2.1 注册

- 用户名 1-64 字符，非纯空白。密码 ≥ 8 字符，不能包含用户名（不区分大小写）。不要求混合字符类型（NIST SP 800-63B）
- 密码以 argon2id 哈希存储（m=47104, t=1, p=4），编码格式 `$argon2id$v=19$m=47104,t=1,p=4$<b64_salt>$<b64_hash>`
- 验证时检测前缀自动选择算法。登录成功时若算法或参数已升级，自动以新参数重哈希
- 注册成功即签发 JWT（30 天有效，HS256），`token_version` 初始值 1
- 若请求携带 `anonymous_token`，注册成功即自动认领该 token 关联的匿名评估——通过窄接口 `user.Claimer` 调用 `ClaimAnonymous()`。独立端点 `POST /api/assessments/claim` 保留用于跨设备场景
- 用户名冲突返回 409

### 2.2 登录

- 凭用户名 + 密码、已验证邮箱 + 密码、或待验证邮箱（`pending_email`）+ 密码登录。密码错误或用户不存在统一返回 401 "Invalid username or password"（防用户名枚举）
- 登录成功签发 JWT
- **注销恢复**：若用户处于 7 天冷静期（`deactivated_at` 非空且距当前 < 7 天），登录触发恢复：清除 `deactivated_at`，`token_version + 1`，发新 token。超 7 天则返回 401

### 2.3 登出

- 服务端 `token_version + 1`，所有设备的 JWT 即时失效
- 客户端清除 localStorage token。服务端调用失败时仍清除本地 token（客户端优先）

### 2.4 修改密码

- 需当前密码确认，错误返回 401
- 新密码最小 8 字符。成功后 `token_version + 1`，旧 token 全部失效，返回新 token。客户端需替换本地 token
- 登录时若算法参数已升级则自动重哈希，用户无感

### 2.5 改邮箱

非阻塞验证流程（`UpdateMeUseCase`）：
- 用户提交新邮箱 → `repo.UpdateFields()` 写入 `pending_email` + `email_verification_token`（`crypto/rand` 16 字节 hex，24h 有效）
- DB 更新成功后 → `sender.SendVerificationEmail()` 发送验证邮件（仅 `sender != nil` 时）
- `email` 字段保持旧值，`email_verified` 置 false——旧邮箱在验证完成前仍有效（登录/找回密码）
- 用户点击验证链接 → `GET /api/auth/verify-email?token=` → `pending_email` → `email`，清空 `pending_email`，`email_verified_at = now()`
- 若 `sender == nil`（未配置 Resend）：token 存入 DB 但不发送邮件，管理员可通过日志手动验证

### 2.6 密码重置

- `POST /api/auth/forgot-password` 返回 200——无论邮箱是否存在（防邮箱枚举）
- `ForgotPasswordUseCase(ctx, repo, sender, email, locale)` 完成完整流程：
  1. 生成 token（`crypto/rand` 16 字节 hex），有效期 15 分钟
  2. `repo.SetPasswordResetToken()` 存入 DB——匹配已验证邮箱或待验证邮箱（`pending_email`），均无则静默跳过
  3. 若 `sender != nil`（`RESEND_API_KEY` 已设置），调用 `sender.SendPasswordResetEmail()` 发送邮件
  4. 若 `sender == nil`，fallback 为 `log.Printf` 输出重置链接（管理员手动处理）
- `POST /api/auth/reset-password` 调用 `ResetPasswordUseCase` → 验证 token → 检查密码长度 → `repo.ResetPassword()` 写入新哈希
- 重置成功后 `token_version + 1`，旧 token 全部失效
- 邮件发送方：`EMAIL_FROM` 环境变量指定的地址（无默认值，为空则禁用邮件）。纯文本邮件，按 `X-Locale` header 选择 EN/ZH 模板

### 2.7 邮箱验证

两个场景共用 GET `/api/auth/verify-email?token=`：
- **首次验证**：注册或购买通行证时绑定邮箱
- **改邮箱重验证**：`PUT /users/me` 修改邮箱后

若 `pending_email` 非空：`pending_email` → `email`，清空 `pending_email`，`email_verified_at = now()`。否则仅置 `email_verified_at`。

**重新发送验证邮件** — `POST /api/auth/resend-verification` 🔒：
- 目标邮箱：优先 `pending_email`（待验证的新邮箱），其次 `email`（未认证的旧邮箱）
- 无邮箱可验证返回 `400 no_email`，已认证且无待验证邮箱返回 `409 already_verified`
- 生成新 token（覆盖旧 token），有效期 24 小时
- 若 `sender == nil`（本地开发），token 存入 DB，验证链接打印到服务器日志

### 2.8 账户注销

软删除 + 7 天冷静期。**立即生效**：
- `token_version + 1`：所有现有 token 立即失效
- `deactivated_at = now()`：7 天冷静期开始
- API 返回 `reactivate_by`（7 天后）

**7 天内登录**：清除 `deactivated_at`，账户恢复，`token_version + 1` 发新 token。数据完整保留。

**7 天后**（服务端定时或懒检查触发）。数据匿名化级联——操作顺序如下：

| 表 | 操作 | 理由 |
|----|------|------|
| `users` | name→`"deactivated_NNN"`, email→'', password_hash→'', `deactivated_at`=now | 匿名化，保留行以维护 FK 引用 |
| `assessments` (type=self) | `user_id`→NULL | 脱钩，保留供匿名聚合统计 |
| `assessments` (type=peer, subject) | `review_link_id`→NULL | 断链但保留匿名数据 |
| `assessments` (type=peer, reviewer) | `reviewer_name`→'', `user_id`→NULL | 清除评审者身份 |
| `review_links` | `deleted_at`=now | 软删除 |
| `match_links` (owner) | 软删除（is_deleted=1） | 链接失效但保留数据 |
| `bond_events` | 保留 | 对方仍可在 bonds 列表查看 |

匿名 peer 评估（`user_id IS NULL`）提交超过 90 天自动清理。注销满 2 年的无主评估每日 cron 清理。

### 2.9 数据导出

- 已登录用户可通过 `/me` 导出原始数据（评估记录 JSON）——GDPR 数据携带权，免费
- 通行证用户额外包含：bond_history、flow_records、peer_deep_analysis
- 匿名评审者提交时返回 `deletion_token`（`crypto/rand` hex），凭此 token 可删除自己的提交

### 2.10 公开用户浏览

- 按元素筛选（`?element=W`），分页展示。仅 `is_public=1` 且未注销用户
- 查看单个用户公开信息：若私有（`is_public=0`）返回 403，注销返回 404
- `identity_label` 返回 ID（如 "WF"），客户端从 types.json 查 locale 标签
- 页面须标注 ipsative 数据说明

---

## 3. 接口与错误模型

### 3.1 接口定义

| 接口 | 位置 | 职责 |
|------|------|------|
| `UserRepository` | `user/ports.go` | 写侧：CRUD、密码管理、token 管理、导出查询 |
| `UserReader` | `user/ports.go` | 读侧：公开用户列表/详情、profile JSON |
| `TokenValidator` | `user/ports.go` | JWT token_version 校验 |
| `PasswordHasher` | `user/ports.go` | argon2id 哈希/验证 |
| `Claimer` | `user/ports.go` | 匿名评估认领窄接口 |
| `EmailSender` | `user/ports.go` | transactional email — 验证邮件 + 密码重置 |
| `UpdateUserFields` | `user/ports.go` | 类型安全的字段更新 DTO（替代 `map[string]interface{}`）|

### 3.2 Sentinel Error

User 聚合相关错误全部在 `domain/errors.go` 定义，使用 `errors.Is()` 比较：

| Error | 触发条件 |
|-------|---------|
| `ErrNameAndPasswordRequired` | 注册/登录时姓名或密码为空 |
| `ErrPasswordTooShort` | 密码 < 8 字符 |
| `ErrPasswordContainsName` | 密码包含用户名（不区分大小写） |
| `ErrUsernameTaken` | 注册时用户名已存在 |
| `ErrInvalidCredentials` | 登录密码错误或用户不存在或注销超 7 天 |
| `ErrCurrentPasswordWrong` | 修改密码时当前密码错误 |
| `ErrTokenExpired` | 密码重置 token 过期 |
| `ErrNameEmpty` | 更新资料时 name 为空 |
| `ErrNoFields` | 更新资料时无任何字段 |

---

## 4. 并发与安全

- JWT `token_version` 计数器实现全设备即时踢出：密码变更、注销均 +1
- 密钥轮换：旧密钥验证新请求 30 天（覆盖最长 JWT 有效期），之后仅新密钥验证
- 速率限制在 Caddy 代理层：`auth` zone 20 req/min per IP，覆盖 `/api/auth/*`

---

## 参见

- [INDEX](../INDEX.md) — 共享内核、元素编码、JWT 策略
- [API](../API.md) — HTTP 契约
- [appendix/errors](../appendix/errors.md) — 错误码
- [domain/commerce](commerce.md) — 通行证订阅
- [theory/THEORY](../theory/THEORY.md) — 理论推导
