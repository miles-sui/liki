# 数据库设计 · Database Schema

SQLite，单文件，WAL 模式。代码位置：`internal/payment/store.go`。

## 连接配置

| 参数 | 值 | 说明 |
|------|-----|------|
| 驱动 | `github.com/mattn/go-sqlite3` | CGo 绑定 |
| 日志模式 | WAL | 预写式日志 |
| 忙超时 | 5000ms | 锁等待超时 |
| 外键 | ON | 已启用，但未定义外键约束 |
| MaxOpenConns | 1 | 串行化所有数据库操作 |

默认路径：`/var/lib/lingji/lingji.db`（可通过 `DB_PATH` 环境变量覆盖）。

## 表：orders

```sql
CREATE TABLE IF NOT EXISTS orders (
    order_id    TEXT PRIMARY KEY,
    product     TEXT NOT NULL,
    amount      INTEGER NOT NULL,
    currency    TEXT NOT NULL,
    provider    TEXT NOT NULL DEFAULT '',
    email       TEXT NOT NULL DEFAULT '',
    chart_json  TEXT NOT NULL,
    llm_json    TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending',
    payment_id  TEXT,
    locale      TEXT NOT NULL DEFAULT 'zh-Hans',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_payment_id ON orders(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_stale ON orders(status, created_at);
```

### 列说明

| 列 | 类型 | 约束 | 说明 |
|------|------|------|------|
| order_id | TEXT | PRIMARY KEY | UUID v4，由服务层生成 |
| product | TEXT | NOT NULL | 产品类型：chart / bond / naming |
| amount | INTEGER | NOT NULL | 金额，单位：分 |
| currency | TEXT | NOT NULL | 货币：CNY（虎皮椒）/ USD（Dodo） |
| provider | TEXT | NOT NULL DEFAULT '' | 支付通道：dodo / xunhu |
| email | TEXT | NOT NULL DEFAULT '' | 用户邮箱，结账时填写 |
| chart_json | TEXT | NOT NULL | 引擎计算结果 JSON |
| llm_json | TEXT | NOT NULL DEFAULT '' | LLM 解读文本，webhook 收到支付确认后生成，报告页兜底重试 |
| status | TEXT | NOT NULL DEFAULT 'pending' | 状态：pending / paid |
| payment_id | TEXT | 可空 | 支付网关返回的支付 ID |
| locale | TEXT | NOT NULL DEFAULT 'zh-Hans' | 报告语言 locale，结账时写入 |
| created_at | TEXT | NOT NULL | 创建时间，格式 `2006-01-02 15:04:05` |
| updated_at | TEXT | NOT NULL | 更新时间，格式同上 |

### 索引

| 索引名 | 列 | 用途 |
|--------|-----|------|
| idx_orders_stale | (status, created_at) | 过期订单清理 |

### 状态机

```
pending ──(支付成功)──→ paid
  │
  └──(24h 过期)──→ 删除
```

## 查询操作

| 操作 | 方法 | SQL |
|------|------|-----|
| 创建订单 | CreateOrder | `INSERT INTO orders (order_id, product, amount, currency, chart_json, llm_json, locale, provider) VALUES (?, ?, ?, ?, ?, ?, ?, ?)` |
| 查询订单 | GetOrder | `SELECT order_id, product, amount, currency, provider, email, chart_json, llm_json, status, COALESCE(payment_id,''), locale, created_at, updated_at FROM orders WHERE order_id = ?` |
| 更新邮箱 | UpdateEmail | `UPDATE orders SET email = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 更新支付通道 | UpdateProvider | `UPDATE orders SET provider = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 标记已付(幂等) | MarkPaidIdempotent | 先 UPDATE pending→paid WHERE status='pending'，已 paid 则 SELECT 返回 (幂等，支持 webhook 重发) |
| 缓存报告 | UpdateLlmJSON | `UPDATE orders SET llm_json = ?, updated_at = datetime('now') WHERE order_id = ?` |
| 缓存报告(首次) | UpdateLlmJSONIfEmpty | `UPDATE orders SET llm_json = ? WHERE order_id = ? AND llm_json = ''` |
| 清理过期 | CleanStale | `DELETE FROM orders WHERE status = 'pending' AND created_at < ?` |

## 迁移策略

无迁移框架。模式在应用启动时通过 `CREATE TABLE IF NOT EXISTS` 自动应用。

修改 schema 的方式：直接编辑 `store.go` 中的 `schema` 常量。添加新列时需确保兼容现有数据（使用 DEFAULT 或允许 NULL）。

旧版 schema 中的 `pdf_path` 列在启动时通过 `ALTER TABLE orders DROP COLUMN pdf_path` 自动移除（忽略列不存在的错误）。

## 备份

生产环境建议定期备份 SQLite 文件。WAL 模式下直接复制 `.db` 文件安全（SQLite 保证一致性读）。
