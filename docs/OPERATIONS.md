# 25types — 运维与基础设施

> 部署架构、数据库运维、安全策略、可观测性、测试、构建流水线。

---

## 1. 部署架构

### 1.1 总览

Cloudflare 橙云（代理模式）作为外部 HTTPS 入口，Caddy 通过 DNS-01 挑战获取 Let's Encrypt 证书。

```
用户 → Cloudflare (橙云, HTTPS) → Caddy (:443, DNS-01 证书) → Go API (:8080)

Docker Compose:
  ├─ caddy (:443, DNS-01 自动 HTTPS) — 总入口
  │    ├─ /*                 → /app/frontend/dist/ (静态 HTML/JS/CSS，含 en/ 和 zh-CN/ 语言子目录)
  │    └─ /api/*             → backend:8080 (反向代理，含路由级速率限制)
  │    └─ 数据卷: caddy_data (TLS 证书持久化)
  │
  ├─ backend (:8080) — 纯 JSON API
  │    └─ 数据卷: ./db:/var/lib/25types (SQLite)
  │
  └─ 数据卷: caddy_data, ./db (宿主持久化)
```

Go 不渲染 HTML，不 serve 静态文件（生产环境）。Caddy 负责全部静态资源 + Clean URL 映射 + `/api/` 反向代理 + 速率限制。

**本地开发（无 Docker）**：`scripts/dev-local.sh` 启动 Go API (:8081) + Caddy (:8080)，使用 `deploy/caddy/Caddyfile.local`（与生产共享 `static_routes`，无速率限制、无 TLS）。前端构建产物 `web/dist/` 由 Caddy 直接 serve，路由行为与生产一致。

### 1.2 Cloudflare 橙云与 DNS-01

橙云（代理模式）下 Cloudflare 边缘节点拦截所有 HTTP 流量并代行 TLS 终结。传统 HTTP-01 挑战的验证请求同样被拦截，Caddy 无法完成验证 → 证书申请失败。

改用 DNS-01 挑战：Caddy 通过 Cloudflare API 在 DNS 区域自动创建 `_acme-challenge` TXT 记录来完成域名验证，不依赖 HTTP 通路。

**Cloudflare 设置：**
- 域名需在 Cloudflare Dashboard 中**添加为站点**并激活（NS 记录需指向 Cloudflare 分配的 nameserver）
- DNS A 记录指向服务器 IP，开启橙云（Proxied）
- SSL/TLS 模式选 **Full (strict)**——Cloudflare 边缘与源站均为有效证书
- API Token：在 Cloudflare Dashboard → 我的个人资料 → API 令牌 创建，权限选 **Zone:DNS:Edit**，区域资源选 `25types.com`。**必须使用 API Token**（格式 `cfat_...`），Global API Key（32 位 hex）会被 `libdns/cloudflare` 模块拒绝

**Caddy 自定义镜像**：官方 `caddy:2-alpine` 不含 DNS 和速率限制插件。通过 `xcaddy build --with github.com/caddy-dns/cloudflare --with github.com/mholt/caddy-ratelimit` 构建。Dockerfile 见 `deploy/caddy/Dockerfile`。

**环境变量**：`CF_API_TOKEN` 在 `.env` 中设置，不进入仓库。Caddy 通过 `{env.CF_API_TOKEN}` 语法引用。

### 1.3 Caddyfile

Caddy 负责全部静态资源 + Clean URL 映射 + 语言路由 + `/api/` 反向代理 + 速率限制。`root` 指向 Eleventy 构建产物 `/app/frontend/dist`。

**语言路由**：首页 `/` 通过 `Accept-Language` header 嗅探，302 重定向到 `/en/` 或 `/zh-CN/`（默认 `/en/`）。所有页面均有 locale 路径前缀，`try_files` 在 `dist/en/` 和 `dist/zh-CN/` 子目录中查找。

**SPA 路由**：`/en/users/\d+`、`/zh-CN/users/\d+` 等动态路由通过 `path_regexp` 匹配后 rewrite 到对应 locale 的静态 HTML 壳。`/me/*` 通过 `handle_path` 剥离前缀后服务。

**共享资源**：`js/`、`css/`、`fonts/`、`img/` 位于 `dist/` 根，无 locale 前缀，所有语言共用。

完整 Caddyfile 见 `deploy/Caddyfile`。本地开发使用 `deploy/caddy/Caddyfile.local`（导入 `static_routes`，无速率限制、无 TLS，root 指向本地 `web/dist/`）。

### 1.4 Docker Compose (服务器部署)

生产部署使用 `deploy/docker-compose.prod.yml`，包含 backend 和 caddy 两个服务。Backend 挂载 `./db:/var/lib/25types`（SQLite），Caddy 挂载 `./deploy/caddy/Caddyfile:/etc/caddy/Caddyfile:ro`（生产 Caddy 配置，含 DNS-01 证书 + 速率限制）。环境变量通过 `.env` 文件注入。详细配置见 `deploy/docker-compose.prod.yml`。

数据库文件 (`./db/`) 和 TLS 证书 (`caddy_data` 卷) 独立于容器生命周期。`.env` 文件在部署时上传，容器重建不影响数据。

### 1.5 静态文件目录

```
/app/frontend/dist/          # Eleventy 构建产物 (Caddy serve 根目录)
  js/                        # 客户端 JS（共享，无 locale 前缀）
    common.js, alpine.min.js, echarts.min.js
    assess.js ... bond.js    # 21 个页面专属组件
  css/                       # Tailwind + daisyUI（共享）
  img/, fonts/               # 静态资源（共享）
  en/                        # 英文全站页面
    index.html, about.html, assess.html ...
    types/W/index.html ...   # 25 类型详情页
    me.html, history.html ...
  zh-CN/                     # 中文全站页面（同上结构）
    index.html, about.html ...
    types/W/index.html ...
```


---

## 2. 速率限制

速率限制在 Caddy 代理层实施（Go 应用层不设限流）。Caddy 通过 `caddy-ratelimit` 模块，以 `{remote_host}` 为 key 对真实客户端 IP 进行限流。

| 限流 zone | 速率 | 覆盖路由 |
|-----------|------|----------|
| `auth` | 20 req/min | `route /api/auth/*`（注册、登录、忘记密码、重置密码） |
| `assess` | 30 req/min | `route /api/assessments`（评估提交） |

其他 `/api/*` 路由不限流，Caddy 直接反向代理。限流命中时 Caddy 返回 429，含 `Retry-After` header，Go 无感知。

---

## 3. 环境变量

全部通过环境变量。缺必需配置时 `FATAL` + `os.Exit(1)`。

| 变量 | 必须 | 默认 | 说明 |
|------|:--:|------|------|
| `DATABASE_PATH` | ✓ | — | SQLite 文件路径 |
| `JWT_SECRET` | ✓ | — | 32+ 字节 base64 |
| `LISTEN_ADDR` | | `:8080` | 监听地址 |
| `LOG_LEVEL` | | `info` | debug/info/warn/error |
| `BACKUP_DIR` | | `./backups` | 每日备份目录 |
| `EMAIL_PROVIDER` | | `resend` | 邮件服务商：`resend` 或 `tencent` |
| `RESEND_API_KEY` | | — | Resend API key |
| `TENCENT_SECRET_ID` | | — | 腾讯云 SecretId |
| `TENCENT_SECRET_KEY` | | — | 腾讯云 SecretKey |
| `TENCENT_REGION` | | `ap-hongkong` | 腾讯 SES 区域 |
| `EMAIL_FROM` | 生产必填 | — | 发件人，格式 `显示名 <地址>`，如 `25types <noreply@mail.25types.com>`；为空则禁用邮件 |
| `TENCENT_FROM` | | `EMAIL_FROM` 的值 | 腾讯 SES 发件人（通常无需单独设，与 `EMAIL_FROM` 一致即可）
| `CF_API_TOKEN` | ✓ | — | Cloudflare API Token (Zone:DNS:Edit，Caddy DNS-01 证书用) |
| `DODO_API_KEY` | ✓ | — | Dodo Payments API key（Dashboard > Developer > API Keys） |
| `DODO_TEST_MODE` | | — | 设为 `1` 使用 Test 模式，不设或 `0` 使用 Live |
| `DODO_PRODUCT_DONATION` | ✓ | — | 捐赠产品 Product ID |
| `DODO_WEBHOOK_KEY` | ✓ | — | Webhook 签名密钥（Dashboard > Developer > Webhooks） |

---

## 4. 数据库设计

### 4.1 表总览

共 10 张业务表 + 1 张迁移追踪表。

| 表 | 聚合归属 | 行数规模 | 说明 |
|---|---------|---------|------|
| `users` | User (聚合根) | 百~千 | 注册用户。11 列 |
| `user_tokens` | User | ≤ users×2 | 邮箱验证 / 密码重置 token，每个用户每种最多一条 |
| `user_subscriptions` | Commerce | ≤ users | 通行证状态。与 users 1:1 |
| `subscription_events` | Commerce | 百~千 | 订阅审计日志（购买、激活码兑换） |
| `redemption_codes` | Commerce | 十~百 | 激活码模板。管理员通过 `/api/admin/codes` 生成 |
| `code_redemptions` | Commerce | ≤ users | 激活码兑换记录。与 redemption_codes N:1，与 users N:1 |
| `assessments` | Assessment | 千~万 | 每次测量一条记录。不可变（有触发器保护） |
| `review_links` | ReviewLink (聚合根) | 百~千 | 他评邀请链接。可软删除 |
| `match_links` | MatchLink (聚合根) | 百~千 | 匹配链接。可软删除 |
| `bond_events` | Bond | 百~千 | 每次 Instant Compare 或 match link 应答生成一条 |
| `frontend_errors` | 运维 | 百 | 前端 JS 错误收集。30 天 TTL |
| `schema_migrations` | 系统 | ≤20 | 迁移版本追踪 |

### 4.2 聚合到表映射

每个 DDD 聚合根对应一张主表，聚合内实体/值对象经 JSON 列或关联表存储。

```
User 聚合
  ├── users                    — 聚合根（id, name, email, password_hash, token_version…）
  └── user_tokens              — email_verify / password_reset token（user_id, token_type PK）

Assessment 实体
  └── assessments              — 不可变测量事件。profile_json 存 PersonalityProfile 值对象

ReviewLink 聚合
  └── review_links             — 聚合根（token 唯一，subject_user_id FK→users）

MatchLink 聚合
  └── match_links              — 聚合根（user_id, token, is_deleted 软删除）

Bond 事件
  └── bond_events              — 事件存储（link_id 可空=即时对比，bond_json 快照）

Flow                         — 值对象，不建表。每次请求从 PersonalityProfile 实时计算
```

### 4.3 表详细定义

**users**（最终态，经 009 清理后）：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `name` | TEXT | NOT NULL UNIQUE | 用户名 |
| `email` | TEXT | NOT NULL DEFAULT '' | 邮箱（可为空字符串表示未设） |
| `password_hash` | TEXT | NOT NULL | argon2id 哈希 |
| `token_version` | INTEGER | NOT NULL DEFAULT 1 | JWT 失效计数器，密码变更或注销时 +1 |
| `is_public` | INTEGER | NOT NULL DEFAULT 0, CHECK (0,1) | 是否公开展示 |
| `email_verified_at` | TEXT | | 邮箱验证时间（NULL = 未验证） |
| `pending_email` | TEXT | NOT NULL DEFAULT '' | 待验证的新邮箱 |
| `deactivated_at` | TEXT | | 注销时间（NULL = 活跃）。7 天冷静期后可重新激活 |
| `created_at` | TEXT | DEFAULT strftime | |
| `updated_at` | TEXT | DEFAULT strftime | |

索引：`idx_users_deactivated`（WHERE deactivated_at IS NOT NULL）、`idx_users_email_unique`（UNIQUE, WHERE email != ''）。

**assessments**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `user_id` | INTEGER | FK→users(id) | 可为 NULL（匿名评估） |
| `assessment_type` | TEXT | NOT NULL, CHECK ('self','peer') | |
| `identity_id` | TEXT | NOT NULL | 25 型 ID，如 "WF"（自评不可为空） |
| `answers_json` | TEXT | NOT NULL DEFAULT '[]' | 原始答案 JSON |
| `profile_json` | TEXT | NOT NULL DEFAULT '{}' | PersonalityProfile JSON（自评不可为空） |
| `created_at` | TEXT | NOT NULL DEFAULT strftime | |
| `review_link_id` | INTEGER | FK→review_links(id) | peer 评估必填 |
| `reviewer_name` | TEXT | NOT NULL DEFAULT '' | peer 评估者昵称 |
| `legacy_user_token` | TEXT | NOT NULL DEFAULT '' | 匿名用户 token，用于注册后认领 |

触发器 `trg_assessments_no_update` 保护核心字段不可变（identity_id, answers_json, profile_json, assessment_type, review_link_id）。仅允许更新 user_id 和 legacy_user_token（用于认领）。

索引：`idx_assessments_user`（user_id, created_at DESC）、`idx_assessments_review_link`（review_link_id）、`idx_assessments_user_self`（user_id, id DESC WHERE type='self'）、`idx_assessments_anonymous_token`（legacy_user_token WHERE != ''）。

**review_links**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `subject_user_id` | INTEGER | NOT NULL FK→users(id) | 被评价者 |
| `token` | TEXT | UNIQUE NOT NULL | URL token |
| `expires_at` | TEXT | | 过期时间（30 天） |
| `created_at` | TEXT | DEFAULT strftime | |
| `deleted_at` | TEXT | | 软删除标记 |

索引：`idx_review_links_subject`（subject_user_id）。

**match_links**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `user_id` | INTEGER | NOT NULL FK→users(id) | 创建者 |
| `token` | TEXT | NOT NULL UNIQUE | URL token |
| `created_at` | TEXT | DEFAULT strftime | |
| `is_deleted` | INTEGER | NOT NULL DEFAULT 0 | 软删除标记 |

**bond_events**（迁移 011+013 最终态）：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `link_id` | INTEGER | FK→match_links(id) | NULL = 即时对比 |
| `initiator_user_id` | INTEGER | NOT NULL FK→users(id) | 发起方 |
| `other_user_id` | INTEGER | FK→users(id) | 对方（NULL = 匿名） |
| `other_name` | TEXT | NOT NULL DEFAULT '' | 匿名对方显示名 |
| `assessment_id` | INTEGER | FK→assessments(id) | 关联的匿名评估 |
| `bond_json` | TEXT | NOT NULL DEFAULT '{}' | Bond 计算快照 {self, other, delta_a, delta_b} |
| `created_at` | TEXT | DEFAULT strftime | |

插入前去重：同对人 (initiator_user_id, other_user_id) 无论方向，先 DELETE 旧记录再 INSERT，保证每对人只保留最新一条。
索引：`idx_bond_events_initiator`（initiator_user_id）、`idx_bond_events_link_id`（link_id）。

**user_subscriptions**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `user_id` | INTEGER | PK FK→users(id) | |
| `passport_expires_at` | TEXT | | 通行证过期时间（NULL = 无） |
| `plan` | TEXT | NOT NULL DEFAULT '', CHECK ('','monthly','yearly','code') | 获取渠道 |
| `bond_count` | INTEGER | NOT NULL DEFAULT 0 | 已使用的免费 Bond 次数 |

**subscription_events**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `user_id` | INTEGER | NOT NULL FK→users(id) | |
| `event_type` | TEXT | NOT NULL CHECK ('purchase','redeem') | |
| `plan` | TEXT | NOT NULL DEFAULT '' | monthly / yearly / code |
| `expires_at` | TEXT | | 本次订阅到期时间 |
| `created_at` | TEXT | NOT NULL DEFAULT strftime | |

索引：`idx_subscription_events_user`（user_id, created_at DESC）。

**redemption_codes**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `code` | TEXT | NOT NULL UNIQUE | 激活码字符串 |
| `duration_d` | INTEGER | NOT NULL, CHECK (≥0) | 有效天数（0 = 永久） |
| `max_uses` | INTEGER | NOT NULL DEFAULT 1 | 最大使用次数（0 = 无限） |
| `created_by` | TEXT | NOT NULL DEFAULT '' | 创建者 |
| `notes` | TEXT | NOT NULL DEFAULT '' | |
| `expires_at` | TEXT | | 激活码自身过期时间（NULL = 永不过期） |
| `created_at` | TEXT | NOT NULL DEFAULT strftime | |

**code_redemptions**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `code_id` | INTEGER | NOT NULL FK→redemption_codes(id) | |
| `user_id` | INTEGER | NOT NULL FK→users(id) | |
| `created_at` | TEXT | DEFAULT strftime | |

唯一约束：`idx_code_redemptions_unique`（code_id, user_id）——同一用户不能重复兑换同一激活码。

**user_tokens**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `user_id` | INTEGER | NOT NULL FK→users(id) | |
| `token_type` | TEXT | NOT NULL CHECK ('email_verify','password_reset') | |
| `token` | TEXT | NOT NULL | |
| `expires_at` | TEXT | NOT NULL | |
| `created_at` | TEXT | NOT NULL DEFAULT strftime | |

主键：(user_id, token_type)。每个用户每种 token 最多一条，重新签发时 REPLACE。
索引：`idx_user_tokens_token`（token）——支持 "谁拥有这个 token？" 查找。

**frontend_errors**：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `id` | INTEGER | PK AUTOINCREMENT | |
| `message` | TEXT | NOT NULL DEFAULT '' | 错误消息 |
| `filename` | TEXT | NOT NULL DEFAULT '' | 源文件 |
| `lineno` | INTEGER | NOT NULL DEFAULT 0 | 行号 |
| `colno` | INTEGER | NOT NULL DEFAULT 0 | 列号 |
| `stack` | TEXT | NOT NULL DEFAULT '' | 堆栈 |
| `url` | TEXT | NOT NULL DEFAULT '' | 页面 URL |
| `created_at` | TEXT | NOT NULL DEFAULT strftime | |

索引：`idx_frontend_errors_created_at`（created_at）——支持 TTL 清理。

### 4.4 索引设计

| 索引 | 表 | 用途 |
|------|-----|------|
| `idx_users_deactivated` (WHERE) | users | 活跃用户查询（绝大多数查询排除已注销用户） |
| `idx_users_email_unique` (UNIQUE WHERE) | users | 邮箱唯一性 + 按邮箱查找（登录、密码重置） |
| `idx_assessments_user` | assessments | 用户历史列表（按时间倒序） |
| `idx_assessments_user_self` (WHERE) | assessments | 查找最新自评 Profile（最高频查询） |
| `idx_assessments_review_link` | assessments | 按他评链接聚合所有 peer 评估 |
| `idx_assessments_anonymous_token` (WHERE) | assessments | 匿名认领——注册时匹配遗留评估 |
| `idx_review_links_subject` | review_links | 用户的他评链接列表 |
| `idx_bond_events_initiator` | bond_events | 按发起方查询 bonds |
| `idx_bond_events_link_id` | bond_events | 按 match link 查询关联 bond |
| `idx_user_tokens_token` | user_tokens | token 查找（验证邮件、重置密码） |
| `idx_subscription_events_user` | subscription_events | 用户订阅历史审计 |
| `idx_code_redemptions_unique` (UNIQUE) | code_redemptions | 防重复兑换 |
| `idx_frontend_errors_created_at` | frontend_errors | TTL 清理（DELETE WHERE created_at < …） |

设计原则：每个索引对应至少一条查询路径。WHERE 部分索引（partial index）用于缩小索引体积——例如 `idx_users_deactivated` 仅索引已注销用户（通常 <1% 行），`idx_assessments_user_self` 仅索引自评（最高频查询）。无冗余索引——`idx_users_email_verified` 在 008 中删除，因其查询路径已被 `idx_users_email_unique` 覆盖。

### 4.5 运行时配置

```sql
PRAGMA journal_mode=WAL;           -- 支持并发读写（多读一写）
PRAGMA journal_size_limit=67108864; -- WAL 文件上限 64MB
PRAGMA mmap_size=33554432;         -- 32MB 内存映射
PRAGMA auto_vacuum=INCREMENTAL;    -- 增量 vacuum，空间回收
PRAGMA busy_timeout=5000;          -- 写冲突等待 5s，不立即返回 SQLITE_BUSY
PRAGMA foreign_keys=ON;            -- 强制 FK 约束
PRAGMA cache_size=-8000;           -- 8MB 缓存（负数 = KB）
```

所有 PRAGMA 在 `db.Open()` 中统一设置。`journal_size_limit` 防 WAL 文件无限增长；`mmap_size` 加速只读查询；`auto_vacuum=INCREMENTAL` 允许 `PRAGMA incremental_vacuum` 逐步回收空间；`busy_timeout` 保证写冲突排队而非报错。

### 4.6 Go 连接池

```go
db.SetMaxOpenConns(1)  // SQLite 单写者，多连接无益
db.SetMaxIdleConns(1)  // 保持一个长连接
```

单连接保证 PRAGMA 设置在连接级生效（`foreign_keys=OFF` 仅对当前连接有效）。并发由 WAL 模式读并发 + busy_timeout 写排队保证。一个实例一个 SQLite 文件——不跨网络、不共享文件系统。

### 4.7 数据生命周期

| 操作 | 策略 | 说明 |
|------|------|------|
| 用户注销 | 软删除（`deactivated_at`） | 7 天冷静期后可重新激活。数据不清除 |
| 用户级联 | 注销时：assessments 保留（不可变），review_links 软删除，match_links 软删除，bond_events 保留 | |
| 前端错误 | 30 天 TTL | 每日定时清理 `DELETE FROM frontend_errors WHERE created_at < date('now', '-30 days')` |
| 订阅事件 | 永久保留 | 审计追踪，不自动清理 |
| WAL 文件 | `journal_size_limit=64MB` | 达到上限后自动截断 |
| 数据库文件 | `auto_vacuum=INCREMENTAL` | 删除操作后逐步回收空间 |

---

## 5. 备份与恢复

### 5.1 策略

**备份**：宿主机的 cron 每日凌晨 3:00 执行 `sqlite3 /opt/25types/db/25types.db ".backup /opt/25types/backups/25types-$(date +%Y%m%d).db"`。SQLite WAL 模式下 `.backup` 在线热备，写入不阻塞。

**保留策略**：保留最近 7 天日备 + 最近 3 个月月备。超过的自动清理。

**恢复**：停止 Go API → 替换 `/opt/25types/db/25types.db` → 启动 Go API。恢复后旧 JWT token 全部失效（JWT secret 可能已变），用户需重新登录。

**验证**：每周日 cron 执行 `sqlite3 /opt/25types/backups/25types-$(date +%Y%m%d).db "PRAGMA integrity_check"`，结果写入日志。失败则告警。

**关键触发备份**：用户注销（7 天冷静期后数据匿名化操作）前自动备份一次。

### 5.2 Docker 卷挂载

数据库文件通过 volume 挂载到宿主机，独立于容器生命周期。静态前端文件在构建时 `COPY web/dist/ /app/frontend/dist/` 入 Caddy 镜像，无需运行时挂载。TLS 证书通过 `caddy_data` 命名卷持久化。

---

## 6. Schema Migration

### 6.1 机制

用 `embed` 将迁移 SQL 文件嵌入 Go binary。启动时按序执行：

```go
//go:embed migrations/*.sql
var migrations embed.FS
```

### 6.2 版本追踪

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version   INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
```

每执行一条 migration 记录一条，已执行的跳过。迁移仅向前（forward-only），不回滚。

### 6.3 约束

- 迁移脚本与 Go 代码同仓库，版本一致
- 不做自动回滚——down 迁移仅开发环境手动执行
- 新增非空列必须提供默认值（SQLite 不支持 `ALTER TABLE ADD COLUMN ... DEFAULT` 的事务内行为一致，需额外小心）
- 首次部署统一走 migration runner。`CREATE TABLE IF NOT EXISTS` 只在 migration 文件中使用——migration runner 是唯一真相源。

### 6.4 迁移清单

| 版本 | 文件 | 内容 |
|------|------|------|
| 001 | `001_init.sql` | users, assessments（含不可变触发器）, review_links, match_requests（含交叉阻止触发器）, schema_migrations |
| 002 | `002_redemption_codes.sql` | user_subscriptions, redemption_codes, code_redemptions |
| 003 | `003_email_verification_expires.sql` | users 加 email_verification_token / email_verification_expires_at 列 |
| 004 | `004_code_redemptions_unique.sql` | code_redemptions 唯一约束——防重复兑换 |
| 005 | `005_frontend_errors.sql` | frontend_errors 表（JS 错误收集，30 天 TTL） |
| 006 | `006_index_cleanup.sql` | 索引管理：添加 email 查找索引，添加自评查询部分索引，删除低选择性索引 |
| 007 | `007_user_tokens.sql` | 提取 user_tokens 表——token 存储从 users 表分离 |
| 008 | `008_subscription_cleanup.sql` | 删除 has_passport 冗余列 + 添加 subscription_events 审计表 + assessments 质量 CHECK 加固 |
| 009 | `009_remove_old_token_columns.sql` | 删除 users 表 4 个死 token 列 + 死索引 + code_redemptions 补 created_at |
| 010 | `010_remove_matches_commerce.sql` | 删除 match_requests + 全部 commerce 表。新增 match_links + bond_events |
| 011 | `011_nullable_other_user_id.sql` | bond_events.other_user_id 改为可空（支持匿名） |
| 012 | `012_default_public.sql` | users.is_public 默认值改为 1（默认公开） |
| 013 | `013_bond_json.sql` | bond_events 添加 bond_json / other_name / assessment_id 列 |

### 6.5 @no_fk 标记约定

涉及表重建（`CREATE TABLE new → INSERT → DROP → RENAME`）的迁移在执行 `DROP TABLE` 时可能触发其他表的 FK 约束失败。SQLite 的 `PRAGMA foreign_keys=OFF` 仅在连接级生效——事务内 `tx.Exec("PRAGMA foreign_keys=OFF")` 被忽略。

**约定**：需要临时关闭 FK 检查的迁移文件在顶部声明 `-- @no_fk` 标记。`execMigrationTx` 检测到此标记时在 `db.Begin()` 之前于连接级执行 `PRAGMA foreign_keys=OFF`，事务结束后恢复 `ON`。

```go
// internal/db/db.go — execMigrationTx
noFK := strings.Contains(sqlText, "@no_fk")
if noFK {
    db.Exec("PRAGMA foreign_keys=OFF")
    defer db.Exec("PRAGMA foreign_keys=ON")
}
```

当前使用 `@no_fk` 的迁移：`009_remove_old_token_columns.sql`（重建 users 表，被 assessments / review_links / match_requests 等 FK 引用）。

此标记是精确的、显式的、自文档化的——默认所有迁移在 FK 保护下运行，仅明确声明的迁移例外。

---

## 7. 邮件基础设施

**邮件服务**：双提供商——Resend（默认）和腾讯云 SES。通过 `EMAIL_PROVIDER` 环境变量切换（`resend` 或 `tencent`）。

**发件人配置**（`EMAIL_FROM`）：
- 格式：`显示名 <地址>`，如 `25types <noreply@mail.25types.com>`
- 发送域名（`@` 后面部分）需在对应邮件服务商控制台完成 SPF/DKIM 验证
- 不设默认值——为空则整个邮件子系统禁用，启动日志输出 `[email] EMAIL_FROM not set — disabling email`
- 腾讯 SES 可通过 `TENCENT_FROM` 单独覆盖（通常不需要）

**Resend（`EMAIL_PROVIDER=resend`）**：
- 免费层 100 封/天

**腾讯云 SES（`EMAIL_PROVIDER=tencent`）**：
- 未设 `TENCENT_FROM` 时使用 `EMAIL_FROM`
- 使用腾讯云 API v3 TC3-HMAC-SHA256 签名，国际站 `ap-hongkong` 支持 Simple 纯文本模式
- 国内站需在控制台配置模板——将模板 ID 写入 `TENCENT_TEMPLATE_VERIFY` / `TENCENT_TEMPLATE_RESET` / `TENCENT_TEMPLATE_WELCOME` 环境变量（当前未实现模板模式，走国际站 Simple）

**其他邮件配置**：
- 回复地址：不设 Reply-To（noreply）
- 格式：纯文本（plain text）。不设 HTML 邮件——交付率更高，开发更简单
- 退订：不适用——仅事务性邮件，非营销邮件
- 链接格式：`https://25types.com/verify-email?token={token}` 等
- 所有邮件同时提供 EN/ZH 两个语言版本，按用户请求时的 `X-Locale` 选择

### 发送失败处理

- 邮件 API 返回非 2xx → 记录 ERROR 日志，用户侧返回 500（不暴露邮件服务商信息）
- 邮件服务不可用（API key 未设）→ 忘记密码和邮箱验证功能不可用，其余正常运行
- 邮件发送失败不回滚业务流程（best-effort）
- 不重试——用户可重新触发发送

---

## 8. 支付基础设施

**支付服务**：Dodo Payments — Merchant of Record 型。Dodo 作为法律上的卖家，处理全球税务、退款争议、PCI 合规。支持中国大陆身份注册（KYC 用身份证/护照）。数字产品专用。

**Checkout 流程**：
1. 前端 `POST /api/payments/checkout` → 后端调 Dodo REST API (`POST /checkouts`) 创建 checkout session → 返回 Dodo 托管支付页 URL
2. 用户跳转 Dodo checkout 页面完成付款
3. Dodo webhook 回调 `POST /api/payments/webhook`（`payment.succeeded` 事件）→ 后端验签 + 解析 metadata → 写入 `donations` 表并 best-effort 发送感谢邮件

**捐赠档位**：$9.90 / $19.90 / $29.90（美分：990 / 1990 / 2990）。Dodo Dashboard 创建一个 `Pay What You Want` 产品，后端透传金额。

**Webhook 验签**：Dodo 使用 Standard Webhooks 规范（HMAC-SHA256）。签名内容为 `webhook-id.webhook-timestamp.raw_body`，密钥通过 `DODO_WEBHOOK_KEY` 环境变量注入。验签失败返回 401。

**收款与提现**：Dodo 定期结算到绑定的银行账户/Payoneer（支持 170+ 国家含中国）。提现细节在 Dashboard > Payouts 配置。

**费率**：4% + $0.40/笔。无月费。

**所需账号清单**：
- [ ] Dodo Payments 账号 — [dodopayments.com](https://dodopayments.com)，中国身份证可注册
- [ ] 在 Dashboard 创建一个 `Pay What You Want` 产品用作捐赠
- [ ] Dashboard > Developer > API Keys 获取 API Key
- [ ] Dashboard > Developer > Webhooks 创建 Webhook，URL 设为 `https://25types.com/api/payments/webhook`，订阅 `payment.succeeded` 事件，获取 Signing Key
- [ ] `.env` 写入：`DODO_API_KEY`、`DODO_PRODUCT_DONATION`、`DODO_WEBHOOK_KEY`

**测试模式**：设 `DODO_TEST_MODE=1`，所有请求走 `https://test.dodopayments.com`。Dashboard 切换到 Test Mode 后可创建测试产品和模拟 webhook。

---
## 9. SEO

### 9.1 多语言 URL 结构

公开页面采用 locale 路径前缀——`/en/types/WF`、`/zh-CN/types/WF`。这是 Google 推荐的多语言 URL 模式。默认 locale (`en`) 可通过 `Accept-Language` 嗅探自动跳转，但也保留 `/en/` 前缀作为规范 URL。

### 9.2 静态页面架构

不设服务端渲染。所有页面为静态 HTML + JS fetch：

- **类型详情页**：构建时预生成静态 HTML，含完整内容。SEO 标签嵌入 HTML。
- **公开用户页**：静态 HTML 壳 + JS fetch。OG 图片在每日 sitemap 重生成时预构建。
- **他评入口**：静态 HTML 壳 + JS fetch。OG 标签使用社交化文案。
- **工具页**：静态 HTML 壳含内联初始 JSON 数据。

### 9.3 Meta 与结构化数据

- 每页输出独立 `<title>`、`<meta name="description">`、`og:*`
- `hreflang`：每个页面标注自身和所有兄弟语言版本的 `<link rel="alternate" hreflang="..." href="...">`
- Canonical：每个页面标注 `<link rel="canonical" href="...">`
- 结构化数据：25 型详情页嵌入 JSON-LD `DefinedTerm` + `BreadcrumbList`，用户公开页嵌入 `Person`

### 9.4 Sitemap 与 Robots

构建时生成静态 `/sitemap.xml`，含全部公开页面——类型页 × 25 × 2 (en/zh-CN)、工具页 × 2、所有公开用户页。部署时写入磁盘，每日凌晨定时重生成一次。请求始终返回磁盘文件，零运行时查询。`/robots.txt` 指向 sitemap。

### 9.5 OG 图片自动生成

每个公开页的 `og:image` 图片在构建时生成静态 PNG，存至 content 目录：

- **类型页**（25 个 × 2 locale = 50 张）——五行色块 + 类型名 + 一句话描述
- **用户公开页**——比例条形图（从 p 向量）+ identity 标签。每日凌晨随 sitemap 一并重新生成
- **工具页**（/explore, /calendar）——固定品牌图

图片尺寸 1200×630，Google Discover 推荐比例。

### 9.6 Core Web Vitals

| 指标 | 目标 | 说明 |
|------|------|------|
| LCP | < 2.5s | 最大内容绘制——首屏文字应首先渲染 |
| FID | < 100ms | 首次输入延迟——骨架标签不阻塞交互 |
| CLS | < 0.1 | 累计布局偏移——预留图片和图表尺寸，避免加载后跳动 |

### 9.7 URL 规范

Content 文件路径稳定不变，版本更新通过 `stale-while-revalidate` 缓存策略自动刷新。业务页面的任何重命名或移动均设置 301。自定义 404 页面返回有用导航链接。

---

## 10. 安全

### 10.1 认证

JWT 30 天有效。HMAC-SHA256 签名，密钥来自环境变量 `JWT_SECRET`。密码变更或账户注销后 token 即时失效（token_version + 1）。密码 argon2id 哈希（m=47104, t=1, p=4）。

**密钥轮换**:
1. 旧密钥验证新请求，新密钥签发新 token。持续 30 天（覆盖最长 JWT 有效期）
2. 30 天后仅新密钥验证，旧 token 全部失效
3. 已签发 token 仍含 `token_version`，双重失效保护（密钥 + 版本）

### 10.2 传输安全

生产环境强制 HTTPS。`Strict-Transport-Security: max-age=31536000; includeSubDomains`。

**防护**：
- `Content-Security-Policy` — 使用构建时计算的内联脚本 SHA-256 hash，写入 Caddyfile CSP 的 `script-src` 列表。不使用动态 nonce（Go 不 serve HTML）
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- 所有用户输入在服务端验证和转义。API 错误响应的 `message` 不含堆栈、SQL、文件路径

### 10.3 CORS

v1 同域部署——前端和 API 由同一 Caddy 实例代理，不需要 CORS header。

若后续前端分离到独立域名（如 CDN），仅对 `/api/` 下的公开端点开放 CORS，🔒 端点永远仅同域。CORS header 由 Caddy 添加，Go 不处理。

---

## 11. 可观测性

### 11.1 日志

Go `log/slog`，JSON 格式输出 stdout。字段: `time`, `level`, `msg`, `method`, `path`, `status`, `duration`, `user_id`（如有认证）, `error`（如有）。

**日志级别**:
- 5xx → ERROR（含堆栈）
- 4xx → WARN（不含堆栈）
- 2xx → 不记录（或 DEBUG）
- API 响应 > 1s → WARN

### 11.2 PII 清洗

`slog.ReplaceAttr` 中全量替换 `password`、`token`、`email` 字段值为 `[REDACTED]`。记录字段名但不暴露值。

### 11.3 健康检查

`GET /api/health` — 存活探针。检查 DB 可读性（`db.Ping()`）。

```
Response 200:
  { "data": { "status": "ok", "db": "ok" } }

DB 不可达时（status 仍为 200）:
  { "data": { "status": "ok", "db": "error" } }
```

---

## 12. 测试策略

### 12.1 分层

| 层 | 范围 | 工具 | 运行频率 |
|------|------|------|---------|
| Go 单元测试 | 引擎计算（ComputeD、Bond、Flow、分类）、数据验证、JWT 签发/验证 | `go test` | 每次 push |
| Go 集成测试 | HTTP handler + SQLite 内存库 + 真实中间件链。含用户旅程：注册→评估→flow→history→peers | `go test` + `httptest` | 每次 push |
| 前端数据契约测试 | JS 源码模式检查、d/p 转换单元测试、dist 构建产物验证 | `test-frontend-data.sh` | 每次 push |
| 前端单元测试 | Alpine 组件（slider、bond view）、工具函数（locale、格式化） | Vitest | 每次 push |
| E2E 冒烟 | 全端点 API 冒烟 + 静态页面 HTTP 200 | `smoke.sh` | main 分支合并后 |
| 安全 | CSP header 存在性、JWT 过期行为、速率限制触发 | 自定义脚本 | 每次 push |

### 12.2 覆盖率目标

| 层 | 目标 | 说明 |
|------|------|------|
| 引擎函数 | 100% 行覆盖 | 函数小且纯，边界明确 |
| HTTP handler | >= 80% 行覆盖 | 含错误路径 |
| 中间件 | >= 90% 行覆盖 | Auth、速率限制——安全关键 |
| 前端 JS | >= 60% 行覆盖 | 图表和 DOM 操作天然难测 |

不要求全项目 x%——不同层风险不同，分别设目标。

### 12.3 Go 集成测试设置

`sql.Open("sqlite3", ":memory:")` + 运行迁移 → 注入 handler。每个测试独立建库，不共享状态。

**覆盖场景**：
- 注册→登录→带 token 访问 🔒 端点
- 评估提交流程：匿名→带 token→`data_complete` 跃迁
- 完整用户旅程：注册→30题评估→验证 profile (d Σ=0, p Σ=1, d→p 转换)→flow yearly→assessment history→peers（`internal/app/http/user_journey_test.go`）
- 匿名评估→注册自动关联（anonymous_token claim）
- 匹配全流程：请求→接受→Bond→取消
- 他评全流程：创建链接→提交评价→聚合查询
- 并发冲突：两人几乎同时互发匹配请求
- 速率限制：连续请求触发 429

### 12.4 E2E 冒烟三条核心旅程

| 旅程 | 步骤 | 验证点 |
|------|------|--------|
| 匿名→注册 | 首页→评估→答 5 题→看结果→注册→结果归属 | 匿名 token 传递、注册后评估归属、雷达图渲染 |
| 匹配→Bond | 登录→浏览用户→发起匹配→切换用户接受→查看 Bond | 力流图渲染、首次完整免费 |
| 他评 | 登录→创建链接→匿名打开→提交评价→登录用户查看聚合 | 链接有效态、答案聚合、自评 vs 他评对比 |

### 12.5 CI 流程

```
push → go vet → go test unit → go test integ → frontend src check → smoke (server required)
```

`scripts/test-all.sh` 执行完整本地测试（Phase 1: go vet + unit + integ + frontend；Phase 2: smoke.sh 88 项 HTTP 测试）。`scripts/test-all.sh --e2e` 追加 Playwright E2E（54 项浏览器测试）。

E2E 不阻塞 push——独立 job，main 合并后跑，失败只告警不回滚。

---

## 13. 实施说明

### 13.1 资产处置

| 旧资产 | 处置 | 原因 |
|--------|------|------|
| `backend/internal/engine/` | 重构 | S/C 矩阵、25 原型向量可复用；Bond/Flow/分类公式需对齐新模型 |
| `backend/internal/service/` | 重构 | JWT 签发/校验、DB 查询骨架可复用；查询语句需对齐新 schema |
| `backend/internal/handler/` | 重写 | 新路由完全重新注册，旧路由全删，不复用 |
| `backend/internal/db/` | 重构 | WAL 配置可复用；schema migration 需重写 |
| `backend/internal/model/` | 重写 | Go types 对齐 DDD 领域模型文档 |
| `backend/internal/content/` | 迁移 | YAML 加载器可简化——内容改由 Caddy file_server 直出 |
| `content/` (YAML) | 保留 + 复制到 `static/content/` | 直接复用，Caddy serve |
| `docs/` | 保留 | 实施依据，不动 |
| `Caddyfile` | 重写 | 新路由 + 新静态目录 + CSP header |

### 13.2 实施顺序

```
1. engine/      d 模型 + ReLU + 新 Bond/Flow/分类
2. db/          migration + argon2id 替换 bcrypt
3. model/       Go types 对齐 DDD 领域模型文档
4. handler/     路由全删后重新注册，按 API.md 路由总表逐条实现
5. service/     新查询 + 支付 stub
6. content/     YAML → static/content/，Caddy 直出
7. web/         新前端 (Eleventy 构建 → dist/)，逐页按 FRONTEND.md 验收
```

第 6 步完成后，7 可与 5 并行——前后端互不阻塞。

**为什么这个顺序**：
- engine 是计算核心，handler 和 service 都依赖 model/types，所以 1→2→3 必须先走
- handler 路由注册决定了 API 契约——前端必须等它稳定
- content 迁移是纯文件操作，不依赖代码，可随时做
- 前端 7 依赖 4 的 API 契约稳定，但不依赖 5（支付）完成

### 13.3 关键验节点

| 阶段 | 验收标准 |
|------|---------|
| engine 完成 | `go test ./internal/ganzhi/... ./internal/tianwen/... ./internal/25types/...` 100% pass，覆盖 ComputeD/ReLU/ComputeBond/ClassifyIdentity |
| handler 完成 | 每条新路由 curl 返回正确 envelope；旧路由全部 404 |
| content 完成 | `curl localhost/content/en/types.json` 返回 JSON，`Cache-Control` header 存在 |
| 前端第一批 | 匿名用户可完成评估→看结果→分享卡片→注册的完整闭环 |
| 前端第二批 | 已登录用户可查看历史/他评/Flow，修改资料，注销 |
| 前端第三批 | 匹配→Bond→付费→解锁的完整闭环 |

### 13.4 构建流水线

**输入 → 输出：**

| 步骤 | 工具 | 输入 | 输出 |
|------|------|------|------|
| HTML (en) | Eleventy (`LOCALE=en`) | `pages/*.html` + `content/en/` | `dist/en/*.html` |
| HTML (zh-CN) | Eleventy (`LOCALE=zh-CN`) | `pages/*.html` + `content/zh-CN/` | `dist/zh-CN/*.html` |
| JS bundle | esbuild | `web/scripts/build-echarts.js` | `js/vendor/echarts.min.js` |
| CSS | Tailwind + daisyUI | 全站使用的 class | `dist/css/tailwind.min.css` |
| 指纹化 | `web/scripts/fingerprint.js` | dist/ 下 JS/CSS 文件 | 版本哈希文件名 + manifest.json |
| 翻译检查 | `web/scripts/check-translations.js` | 构建产物 + `content/en/translations.json` | 差集报告 |

产物构建到 `web/dist/`，Caddy Docker 镜像构建时 `COPY web/dist/ /app/frontend/dist/` 将源码树中已构建好的静态文件直接复制到容器内。

### 13.5 CI/CD 概要

```
代码推送 (main) → 构建镜像 → 冒烟测试 → docker save → scp 上传服务器 → docker load → 部署
```

不使用 registry，镜像通过 `docker save` / `scp` / `docker load` 传输。

**阶段分解：**

| 阶段 | 做了什么 | 产出 |
|------|---------|------|
| Lint + Test | `go vet` + `go test ./...`；前端 `eslint` + `prettier` | 通过/失败 |
| 构建后端镜像 | `docker compose -f deploy/docker-compose.yml build backend` | `25types-backend:latest` |
| 构建 Caddy 镜像 | `docker compose -f deploy/docker-compose.prod.yml build caddy` | `25types-caddy:latest` (含 DNS 插件) |
| 冒烟测试 | 启动容器 → `smoke.sh` 全端点验证 | 通过/失败 |
| 部署 | `scripts/deploy.sh` → build + save + scp + load + docker compose up | 线上更新 |

**回滚**：保留最近 3 个版本的镜像 tar 包。线上异常时 `docker load < 上一版本.tar.gz && docker compose up -d`，秒级回滚。数据库向下兼容——回滚不涉及 schema 降级。

**环境隔离**：
- `local` — `scripts/dev-local.sh`（Go :8081 + Caddy :8080，无 Docker，无速率限制）
- `dev` — 本地 `docker compose -f deploy/docker-compose.yml`（HTTP only，仅 80 端口），`JWT_SECRET` 手设
- `prod` — 目标机器，`docker compose -f deploy/docker-compose.prod.yml`，`.env` 注入 `CF_API_TOKEN`、`JWT_SECRET`、`RESEND_API_KEY`，不进入仓库

---

## 14. 可访问性 (a11y)

所有公开页面和已登录核心页面达到 WCAG 2.1 AA 级：

- 色彩对比度 >= 4.5:1（文字）/ 3:1（大文本）。五行色块搭配文字标签，不只靠颜色区分元素
- 所有交互控件支持键盘操作（Tab/Enter/Escape）
- 图表（元素雷达图、25 型图谱）提供等价的文本描述——`aria-label` 或旁注文字
- 表单有 `<label>` 关联，错误提示为 `<output>` 或 `aria-describedby`
- 页面有 skip-link 跳至主要内容

验收：
- 五行色块存在色盲模拟下可区分的非颜色标识
- 无需鼠标即可完成评估流程
- Lighthouse Accessibility 得分 >= 95

---

## 15. 国际化

当前支持 en/zh-CN。采用 URL-based locale 路由：所有页面以 `/{locale}/` 为路径前缀（`/en/about`、`/zh-CN/about`），Caddy 在 `/` 根路径通过 `Accept-Language` 302 重定向。

构建时通过两次 Eleventy build（`LOCALE=en` / `LOCALE=zh-CN`）分别产出各语言 HTML，每页仅注入当前语言数据。共享 JS/CSS/fonts/img 无 locale 前缀，所有语言共用。

API 通过 `X-Locale` header 传递语言偏好。数字和日期按 locale 格式化。UI 布局不做 RTL 假设——为未来阿拉伯语等 RTL 语言留空间，但不在当前范围。

---

## 16. Caddy 路由约定

Caddy 使用 profile 匹配器将 `/profile/{name}` 路径路由到 profile 页面。匹配规则在 `deploy/caddy/static_routes` 中定义，由 `deploy/caddy/routes`（生产/开发 Docker）和 `deploy/caddy/Caddyfile.local`（本地开发）通过 `import static_routes` 共享使用。

### 16.1 Profile 匹配器机制

Profile URL 为 `/profile/{name}`，通过 `/profile/` 前缀与静态页面路径天然区分，不需要排除列表。

| 匹配器 | 正则 | 匹配范围 | 重写目标 |
|--------|------|----------|----------|
| `@profileEn` | `^/en/profile/[^/]+$` | 英文 locale 下的 `/profile/{name}` | `/en/profile.html` |
| `@profileZh` | `^/zh-CN/profile/[^/]+$` | 中文 locale 下的 `/profile/{name}` | `/zh-CN/profile.html` |
| `@profileRaw` | `^/profile/[^/]+$` | 无 locale 前缀的 `/profile/{name}` | `/en/profile.html` |

`try_files` 在匹配器**之后**运行，处理所有其他路径的 .html 回退。

### 16.2 新增页面 Checklist

每次新增静态页面时，必须同步更新以下文件：

1. **`scripts/smoke.sh`** Phase 0 — 添加 `check_page` 调用（含内容校验）
2. **`web/e2e/`** — 添加 Playwright E2E 测试覆盖关键用户流程

### 16.3 豁免页面

`error_404` / `error_500` 由 Caddy `handle_errors` 块处理，不依赖常规路由。`profile` 是重写目标本身，无需额外处理。

### 16.4 邮件链接与 Locale

从邮件发送的链接（密码重置、邮箱验证、Bond 通知）**必须包含 locale 前缀**：

```
https://25types.com/{locale}/reset-password?token=xxx
https://25types.com/{locale}/verify-email?token=xxx
https://25types.com/{locale}/profile/{name}
```

否则 Caddy 的 `try_files` 回退到 `/en/` 版本，中文用户看到英文页面。

---

## 参见

- [INDEX](INDEX.md) — 聚合边界、共享内核、路由总览
- [domain/user](domain/user.md) — User 聚合
- [domain/assessment](domain/assessment.md) — 评估领域、ComputeD、Schema
- [domain/match](domain/match.md) — Bond 计算
- [domain/commerce](domain/commerce.md) — 支付
- [API](API.md) — HTTP 契约
- [FRONTEND](FRONTEND.md) — 前端规范
