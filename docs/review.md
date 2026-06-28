# 灵机全系统复盘：状态机 · 业务流 · 数据流

> 2026-06 更新：架构简化后复盘。从 6 产品多 Agent 简化为单一起名 NamingChatAgent。  
> 主要变化：JWT 替代 Session、LLM 对话内直接出报告替代 GenerateFromData、8 tool 替代 29 tool。

## 零、Engine 层现状

```
11 个包，全绿 go vet，全绿 go test -race：

ganzhi/     # 干支基础（天干地支、五行生克、合冲刑害、十神、藏干、长生十二宫、纳音）
tianwen/    # 天文历算（真太阳时、节气、干支历、公历农历互转）
bazi/       # 八字命理
ziwei/      # 紫微斗数
qimen/      # 奇门遁甲（时/日/月/年盘 + 克应 + 格局 + 应期）
liuyao/     # 六爻卜卦（64卦 + 装卦 + 用神 + 月建日建 + 应期）
fengshui/   # 风水基础（24山 + 飞星 + 旺衰 + 双星加会）
bazhai/     # 八宅风水（命卦 + 八宅方位 + 四柱八卦）
xuankong/   # 玄空风水（三元九运 + 飞星 + 挨星 + 旺山旺向 + 城门诀）
huangli/    # 黄历
qiming/     # 起名
```

原则：纯 Go 计算，无 I/O 依赖。每个函数原子化。

## 一、系统全景

```
                    浏览器
                      │
        ┌─────────────┼─────────────┐
        │             │             │
     Chat 流       Form 流       Free API
  (POST+SSE)    (POST+JSON)    (GET+JSON)
        │             │             │
   Agent+Engine   Engine only    Engine only
        │             │             │
   JWT cookie     无状态          无状态
        │
   Payment (Dodo/虎皮椒 webhook)
```

两条用户路径（Form 流已合并为 Free API）：
- **Chat 流**：购买 → JWT cookie → SSE 通道 → NamingChat（收集→计算→磋商→LLM 输出报告）→ report-ready → 跳转 /report/{id}
- **Free API**：直接调引擎 → 返回 JSON（无 LLM、无持久化、无支付）

---

## 二、状态机复盘

### 2.1 服务端 — 无状态

服务端无 session。JWT cookie（`liki_token`，HS256，24h 有效）携带 email + order_id。每个请求 handler 从 JWT 解出 order_id → 查 DB 获取订单状态。消息持久化在 SQLite `chat_messages` 表。

```
POST /api/auth/login
  │
  ├─ FindActiveOrdersByEmail → 查 paid 且 chat_expires_at > now
  ├─ 单订单 → setJWTCookie
  └─ 多订单 → 用户选择 → setJWTCookie

每个请求：
  jwtAuth(r) → email + order_id → GetOrder → 校验 paid + 未过期 → 业务逻辑
```

**评价：比旧 Session 方案更简洁。** 无内存 map、无 TTL 清理、进程重启不丢会话。

### 2.2 客户端 Chat UI

```
welcome ──(send)──→ chatting(loading) ──(text-delta)──→ streaming
  ↑                      │                                 │
  │                      │ (流结束, 无 report)               │ (stop)
  │                      └──→ chatting(idle) ←──────────────┘
  │                                 │
  │                                 │ (report-ready)
  │                                 └──→ redirect /report/{id}
  │
  └────(newChat)────────────────────┘
```

**与旧架构区别：** 没有 `preview`/`closed` 状态。报告在流中直接生成 → report-ready → 跳转。没有中间购买卡片。

### 2.3 支付状态机

```
pending ──(webhook)──→ paid (chat_expires_at = now + 7d)
  │
  └──(24h 过期清理)──→ deleted
```

---

## 三、数据流复盘

### 3.1 Chat 流完整链路

```
POST /api/agent/naming { message, lang }
  │
  ├─ jwtAuth → order_id
  ├─ GetOrder → 校验 paid + chat_expires_at
  ├─ LoadChatHistory → 全部历史消息
  │
  ├─ CreateChatMessage(user msg) → DB 立即写入
  │
  ├─ NamingChat(ctx, locale, messages, onEvent)
  │    │
  │    ├─ ensureNamingPrompt → system prompt (sync.Map 按 locale 缓存)
  │    ├─ ChatStreamWithTools → LLM
  │    │    ├─ text-delta → SSE 推送
  │    │    ├─ tool_call → Execute → engine 计算 → tool_result
  │    │    └─ 磋商 → 起名讨论
  │    │
  │    └─ 返回完整 messages
  │
  ├─ IsNamingReport(last msg)？
  │   YES → UpdateLlmJSON + report-ready SSE 事件
  │
  └─ BatchCreateChatMessages(new msgs) → DB 批量写入
```

### 3.2 数据层结构

| 层 | 存储位置 |
|---|---|
| 聊天消息（user/assistant/tool） | `chat_messages` 表 |
| LLM 报告（markdown） | `orders.llm_json` |
| 引擎计算结果（tool result JSON） | 仅在 LLM 上下文中，不单独存储 |
| 出生信息 | `orders.birth_info`（按需写入） |
| 订单状态/支付 | `orders` 表 |

### 3.3 与旧架构主要差异

| 旧（6 产品） | 新（仅起名） |
|---|---|
| `chart_json` 存 engine 计算结果 | 无需单独存储，结果在消息里 |
| `llm_json` 支付后 webhook 触发生成 | 对话中 LLM 直接输出，handler 识别后写入 |
| Session 内存 + session_id | JWT cookie + chat_messages |
| purchase tool → CreateOrder | 支付在前，对话在后 |
| done 事件 (含 order_id + amount) | report-ready 事件 |
| GenerateFromData 后台生成 | 内联报告 |

---

## 四、前端代码一致性

### 4.1 共享函数

| 函数 | 位置 | 调用方 |
|---|---|---|
| `goPay()` | `web/js/api.js` | chat.js, report.js |
| `apiGet` / `apiPost` | `web/js/api.js` | chat.js, report.js |

### 4.2 状态管理风格

| 页面 | 框架 | 状态方式 |
|---|---|---|
| `chat.js` | Vue 3 | `phase` + `substate` enum |
| `report.js` | Lit-html | `phase` enum |
| `index.html` | Vue 3 | 静态 |

### 4.3 i18n

| 页面 | 使用方式 |
|---|---|
| `index.html` | `$store.i18n.t()` |
| `report.html` | `$store.i18n.t()` |
| `chat.html` | `window.I18N.t()` |

---

## 五、行业标准实践对照

| 实践 | 灵机 | 评价 |
|---|---|---|
| POST + SSE 流式 LLM | ✅ | 最佳匹配，比 WebSocket 更简单 |
| 节流渲染 (80ms) | ✅ | 标准做法，减少 94% DOM 操作 |
| StreamRenderer 抽象 | ✅ | content→html 不变量的显式编码 |
| 响应式 UI (Vue 3) | ✅ | 声明式，框架自动 DOM 同步 |
| 统一错误 envelope | ✅ | `{"error":{"code":"...","message":"..."}}` |
| 分层架构 | ✅ | Handler → Service → Store |
| SQLite WAL + 单连接 | ✅ | 匹配低写入量场景 |
| 无状态 JWT 认证 | ✅ | 无需服务端 session |
| 消息持久化到 DB | ✅ | chat_messages 表，跨设备可续聊 |
| 幂等 webhook 验签 | ✅ | Dodo + 虎皮椒双通道签名验证 |
| 可恢复/致命错误分级 | ✅ | chat 流 SSE error 事件 |
| 并发保护 (pending guard) | ✅ | 防重复提交 |
| 流取消 (AbortController) | ✅ | 用户可停止流式输出 |
| LLM 内联报告 | ✅ | 减少 webhook→后台任务→轮询链路 |

---

## 六、改进优先级

| 优先级 | 项 | 涉及 | 状态 |
|---|---|---|---|
| P1 | report.js boolean soup → phase enum | `web/js/report.js` + `web/report.html` | ✅ 完成 |
| P1 | `goPay()` 提取到 `api.js` | `web/js/api.js` + `chat.js` + `report.js` | ✅ 完成 |
| P2 | `report.js` 改用 `apiGet` | `web/js/report.js` | ✅ 完成 |
| P2 | fetch 加 `AbortSignal.timeout()` | `web/js/api.js` | ✅ 完成 |
| P3 | chat.html i18n | `web/chat.html` | ✅ 完成 |
| P0 | bazi.Chart 子结构 JSON tag 补齐 | `bazi_chart.go`, `bazi_bond.go` 等 | ✅ 完成 |
| P0 | TimePoint 去掉 Lunar 字段 | `agent/tools.go`, `handler_helpers.go`, 13 个 tool schema JSON | ✅ 完成 |
| P1 | error code 标准化 | `handler_helpers.go`, `agent.go`, `bazi.go` | ✅ 完成 |
| P2 | Timeset 转换 14 处重复提取为 helper | `handler_helpers.go` + 6 个 handler 文件 | ✅ 完成 |
| P0 | 6 产品精简为单一起名 | 全项目 | ✅ 完成 |
