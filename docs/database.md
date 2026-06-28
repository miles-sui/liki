# 数据库设计 · Database Schema

SQLite，单文件，WAL 模式。代码位置：`internal/payment/store.go`。

## 连接配置

| 参数 | 值 | 说明 |
|------|-----|------|
| 驱动 | `modernc.org/sqlite` | 纯 Go |
| 日志模式 | WAL | 预写式日志 |
| 忙超时 | 5000ms | 锁等待超时 |
| 外键 | ON | 已启用 |
| MaxOpenConns | 1 | 串行化所有数据库操作 |

默认路径：`/var/lib/liki/liki.db`（可通过 `DB_PATH` 环境变量覆盖）。

## 表：orders

```sql
CREATE TABLE IF NOT EXISTS orders (
    order_id        TEXT PRIMARY KEY,
    product         TEXT NOT NULL,
    amount          INTEGER NOT NULL,
    currency        TEXT NOT NULL,
    provider        TEXT NOT NULL DEFAULT '',
    email           TEXT NOT NULL DEFAULT '',
    chart_json      TEXT NOT NULL DEFAULT '',
    llm_json        TEXT NOT NULL DEFAULT '',
    birth_info      TEXT NOT NULL DEFAULT '',
    chat_expires_at TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    payment_id      TEXT,
    locale          TEXT NOT NULL DEFAULT 'zh-Hans',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_orders_email ON orders(email, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_payment_id ON orders(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_stale ON orders(status, created_at);
```

### 列说明

| 列 | 类型 | 说明 |
|------|------|------|
| order_id | TEXT PK | UUID v4，`agent.NewOrderID()` 生成 |
| product | TEXT | 产品类型，目前固定 `naming` |
| amount | INTEGER | 金额，单位：分（CNY=2990 / USD=2990） |
| currency | TEXT | 货币：CNY（虎皮椒）/ USD（Dodo） |
| provider | TEXT | 支付通道：dodo / xunhu |
| email | TEXT | 用户邮箱，结账时填写 |
| chart_json | TEXT | 引擎计算结果 JSON（预留） |
| llm_json | TEXT | LLM 生成的起名报告 markdown，对话中直接输出后存入 |
| birth_info | TEXT | 出生信息 JSON（`{raw, geo, timeset}`），按需写入 |
| chat_expires_at | TEXT | 聊天截止时间（支付成功 + 7 天），格式 `2006-01-02 15:04:05` |
| status | TEXT | 状态：pending / paid |
| payment_id | TEXT UNIQUE | 支付网关返回的支付 ID |
| locale | TEXT | BCP 47 locale（zh-Hans / zh-Hant / en） |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 更新时间，每次修改自动刷新 |

所有 datetime 使用 UTC。

## 表：chat_messages

```sql
CREATE TABLE IF NOT EXISTS chat_messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id   TEXT NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_order ON chat_messages(order_id, created_at);
```

### 列说明

| 列 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 |
| order_id | TEXT | 关联 orders.order_id |
| role | TEXT | user / assistant / tool |
| content | TEXT | 消息内容 |
| created_at | TEXT | 创建时间，按序排列即为对话顺序 |

## 数据生命周期

### 订单状态机

```
CreateOrder(order_id, product, amount, currency, email, chartJSON, llmJSON, provider)
  │
  ▼
┌─────────┐     支付成功        ┌─────────┐
│ pending │ ───────────────→  │  paid   │
└────┬────┘                   └────┬────┘
     │                             │
     │ 24h 过期自动清理            │ chat_expires_at = now + 7d
     │ CleanStale()                │
     ▼                             ▼
   删除                     7 天后 chat_expires_at 过期
                               handler 返回 "聊天已过期"
```

### 消息持久化时序

```
POST /api/agent/naming
  │
  ├─ CreateChatMessage(order_id, "user", msg)   ← 用户消息立即入库
  │
  ├─ SSE 流式返回 LLM 回复
  │    （此时 assistant 消息仅在内存中）
  │
  └─ 流结束
       │
       ├─ BatchCreateChatMessages(order_id, msgs)  ← assistant 消息批量入库
       │
       └─ 检测到报告
            └─ UpdateLlmJSON(order_id, content)     ← 报告 markdown 存入 orders.llm_json
```

### 老用户续聊

```
POST /api/auth/login { email }
  │
  ├─ FindActiveOrdersByEmail(email)
  │   → SELECT * FROM orders
  │     WHERE email = ? AND status = 'paid'
  │       AND chat_expires_at > datetime('now')
  │
  ├─ 单订单 → JWT cookie
  └─ 多订单 → 用户选择 → JWT cookie

POST /api/agent/naming
  │
  ├─ jwtAuth → order_id
  ├─ LoadChatHistory(order_id)
  │   → SELECT role, content FROM chat_messages
  │     WHERE order_id = ? ORDER BY created_at ASC
  │
  └─ 历史消息 + 新消息拼成完整上下文 → LLM 续聊
```

## 查询操作

| 操作 | 方法 | 关键 SQL |
|------|------|------|
| 创建订单 | CreateOrder | `INSERT INTO orders (order_id, product, amount, currency, email, chart_json, llm_json, provider) VALUES (...)` |
| 查询订单 | GetOrder | `SELECT * FROM orders WHERE order_id = ?` |
| 更新邮箱 | UpdateEmail | `UPDATE orders SET email = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 更新支付通道 | UpdateProvider | `UPDATE orders SET provider = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 标记已付 | MarkPaidIdempotent | `UPDATE orders SET status = 'paid', payment_id = ?, updated_at = datetime('now') WHERE order_id = ? AND status = 'pending' RETURNING email, product` |
| 设置聊天期限 | SetChatExpiresAtIfEmpty | `UPDATE orders SET chat_expires_at = ? WHERE order_id = ? AND chat_expires_at = ''` |
| 写入出生信息 | UpdateBirthInfoIfEmpty | `UPDATE orders SET birth_info = ? WHERE order_id = ? AND birth_info = ''` |
| 查询有效订单 | FindActiveOrdersByEmail | `SELECT * FROM orders WHERE email = ? AND status = 'paid' AND chat_expires_at > datetime('now')` |
| 存报告 | UpdateLlmJSON | `UPDATE orders SET llm_json = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 写聊天消息 | CreateChatMessage | `INSERT INTO chat_messages (order_id, role, content) VALUES (?, ?, ?)` |
| 批量写消息 | BatchCreateChatMessages | 多行 INSERT |
| 加载历史 | LoadChatHistory | `SELECT role, content FROM chat_messages WHERE order_id = ? ORDER BY created_at ASC` |
| 清理过期订单 | CleanStale | `DELETE FROM orders WHERE status = 'pending' AND created_at < datetime('now', ?)` |

## 迁移策略

无迁移框架。模式在应用启动时通过 `CREATE TABLE IF NOT EXISTS` 自动应用。新增列通过 `ALTER TABLE ... ADD COLUMN` 渐进迁移（带 DEFAULT，兼容已有数据）。

## 备份

生产环境直接复制 `.db` 文件即可。WAL 模式下 SQLite 保证一致性读。
