# 灵机架构设计 · Liki Architecture

## 概述

灵机（Liki）是 AI 起名顾问。先付费，后进入 chat，7 天内反复磋商名字。前端为静态 HTML + Vue 3，后端为 Go JSON API，LLM 对话通过 SSE 流式返回，数据库为 SQLite。

同时对外提供 AI 命理引擎计算 API（八字、紫微、奇门、六爻、风水、黄历等），供 AI agent 与开发者使用。API 平台保持不变，起名 API 不对外开放。

## 组件图

```
                        浏览器
                          │
                   ┌──────┴──────┐
                   │   Caddy     │  TLS 终止 · 静态文件 · 反向代理
                   │  :8080      │  /zh-Hans/* /zh-Hant/* /en/* → web/
                   │             │  /api/* → Go :8080
                   └──────┬──────┘
                          │
                   ┌──────┴──────┐
                   │  Go Server  │  cmd/liki
                   │  :8080      │
                   └──────┬──────┘
                          │
      ┌───────────────────┼───────────────────┐
      │                   │                   │
      │  ┌────────────────┴────────────────┐  │
      │  │        HTTP Handlers            │  │  Free API + Naming Agent
      │  │  (package http)                 │  │
      │  └────────────────┬───────────────┘  │
      │                   │                  │
      └──────────┬────────┴────────┬─────────┘
                 │                 │
      ┌──────────┴──────────┐  ┌──┴──────────┐
      │  Agent              │  │ Engine      │  引擎层
      │  Naming (8 tools)   │  │ Layer       │  纯 Go 计算
      └──────────┬──────────┘  └──┬──────────┘
                 │                 │
      ┌──────────┼─────────────────┼──────┐
      │          │                 │      │
 ┌────┴────┐ ┌──┴──────┐  ┌──────┴──────┐│
 │ Engine  │ │  LLM    │  │  Payment    ││
 │ bazi    │ │ DeepSeek│  │  Service    ││
 │ ziwei   │ └─────────┘  │  + Store    ││
 │ qiming  │              └──────┬──────┘│
 │ tianwen │                     │       │
 │ ganzhi  │                ┌────┴────┐  │
 │ (其他   │                │ SQLite  │  │
 │  API用) │                │ orders  │  │
 └─────────┘                │ chat_ms │  │
                            └─────────┘  │
                                        │
 ┌───────────────────────────────────────┘
 │ Email (Resend)
 └───────────────────────────────────────
```

## Engine 层

11 个包，三层架构：

```
internal/engine/
├── ganzhi/     # 基础: 天干地支、五行、十神（无外部依赖）
├── tianwen/    # 基础: 天文历算（真太阳时、节气、干支历）
├── bazi/       # 命理: 八字排盘+大运+用神
├── ziwei/      # 命理: 紫微斗数（12宫+14主星+四化+格局）
├── qimen/      # 命理: 奇门遁甲
├── liuyao/     # 问事: 六爻卜卦
├── fengshui/   # 环境: 风水基础
├── bazhai/     # 环境: 八宅风水
├── xuankong/   # 环境: 玄空风水
├── huangli/    # 择日: 黄历
└── qiming/     # 命名: 起名（三才五格+五行补益+字形分析）
```

所有 Engine 包遵循同一原则：**纯 Go 计算，无 I/O 依赖**。每个函数原子化（单一输入→单一输出），SQL/HTTP/LLM 全部在上层。

起名 Agent 仅使用 `bazi`、`ziwei`、`qiming`、`tianwen`、`ganzhi`。其余 engine 包仅对 API 平台暴露。

### 命名约定

- `Compute` 前缀：多步编排（如 `ComputeChart`）
- 无前缀：单公式推导（如 `RiZhu`、`NianZhu`）
- 所有领域概念使用类型化实体，不用裸 `int`
- JSON 字段用拼音 snake_case：`nian`/`yue`/`ri`/`shi`（四柱）、`yongshen`/`dayun`/`riyuan`

### Agent Tool Schema

**内部 Naming Agent**：8 个 tool 的 JSON Schema 定义在 `internal/agent/tools.json`。编译时嵌入 `agent.ToolsJSON`，`NewNamingToolRegistry()` 在启动时解析并注册 handler 函数。运行时 tool schema 直接传给 LLM 的 tool calling `parameters` 字段。

**外部 API**：29 个 JSON-RPC method 通过 `RPCRegistry` 注册。`POST /jsonrpc` 的 `rpc.discover` 方法返回 OpenRPC 1.4.1 文档（动态生成），供外部 Agent 和开发者发现全部引擎计算能力。

外部 Agent 通过 `/skills/liki.md` 获取产品行为描述（含 API 调用说明），`/llms.txt` 获取服务索引。报告模板位于 `/skills/report-naming.md`。起名 API 不对外暴露。

## 核心数据流

### 购买 → 起名（Naming Agent）

```
首页输入邮箱 → POST /api/orders（创建 pending 订单）
  → POST /api/payments/checkout → 跳支付
  → 支付成功 → webhook: chat_expires_at = now+7d
  → 重定向 /chat → JWT cookie → POST /api/agent/naming（SSE）

Agent 单阶段（单 endpoint，工具收集+磋商一体）:

    → query_city（城市→经纬度）
    → compute_time（raw + geo → Timeset）
    → compute_chart → compute_ziwei
    → compute_naming_wuge → compose → detail → evaluate
    → 磋商讨论
    → 用户要求时，LLM 直接在对话里输出 markdown 报告
    → handler 识别报告 → 存 llm_json → 发 report_ready 事件 → 前端跳转 /report/{id}
```

### 消息持久化（per-request）

```
POST /api/agent/naming
  │
  ├─ 1. jwtAuth(r) → email + order_id
  ├─ 2. GetOrder(order_id) → 校验已支付、未过期
  ├─ 3. LoadChatHistory(order_id) → []ChatMessage（按 created_at ASC）
  ├─ 4. 拼 messages: [system prompt] + history + [user msg]
  │
  ├─ 5. CreateChatMessage(order_id, "user", msg)  ← 用户消息立即入库
  │
  ├─ 6. SSE 流式返回（http.Flusher）
  │       │
  │       ├─ text_delta  → 逐 token 推送
  │       ├─ thinking    → 推理过程（可选）
  │       ├─ phase       → 阶段提示
  │       ├─ tool_call   → 调用引擎计算
  │       └─ ...
  │
  └─ 7. 流结束后
         ├─ 检测最后一条 assistant 消息是否为报告
         │   └─ IsNamingReport(content) → UpdateLlmJSON(order_id, content)
         │                                → emit report_ready 事件
         └─ BatchCreateChatMessages(order_id, 新消息)  ← AI 回复批量入库
```

即使中途关闭网页：
- 用户消息已在步骤 5 入库
- 流结束后 AI 回复在步骤 7 入库
- 唯一丢的是正在流式输出的那轮 AI 回复（用户消息不丢）

### 老用户回来（断点续聊）

```
输入邮箱 → POST /api/auth/login
  │
  ├─ FindActiveOrdersByEmail(email)
  │   → SELECT * FROM orders
  │     WHERE email = ? AND status = 'paid'
  │       AND chat_expires_at > datetime('now')
  │
  ├─ 单订单 → 直接签发 JWT cookie
  └─ 多订单 → 返回订单列表，用户选择 → POST /api/auth/select-order → JWT cookie

重定向 /chat → POST /api/agent/naming
  │
  ├─ jwtAuth(r) → order_id
  ├─ LoadChatHistory(order_id) → 全部历史消息
  ├─ 拼入 system prompt 前
  └─ LLM 看到完整历史，无缝续聊
```

### 认证与会话

```
无服务端 session。纯 JWT cookie 方案：

  登录 → setJWTCookie(w, email, orderID)
         │
         └─ JWT payload: { email, order_id, exp: now+24h }
            Cookie: liki_token
              HttpOnly: true
              Secure: true
              SameSite: Lax
              MaxAge: 86400

  每个请求 → jwtAuth(r)
              │
              └─ 解析 liki_token cookie
                 → 验签（HS256, JWT_SECRET）
                 → 返回 email + order_id
                 → handler 用 order_id 查 DB 获取完整状态

JWT 有效期 24h，过期后重新登录（输邮箱即可，无需重新支付）。
```

### 数据切面

```
raw（用户输入）→ geo（query_city）→ timeset（compute_time）
                                          ↓
                               chat_messages 持久化 → 磋商 + 报告生成
```

### Free API 流

```
POST /jsonrpc（bazi.chart / ziwei.chart / qimen.pan 等 29 个 method）
  → 解析请求 → Engine 计算
  → 返回 JSON（命理数据，不经过 LLM/支付）
```

全部引擎能力通过 JSON-RPC 统一入口 `POST /jsonrpc` 暴露，供外部 AI agent 和开发者使用。

## 提示词体系

### 文件组织

```
internal/agent/data/
  naming.txt               ← Naming Agent 系统 prompt（内部）
  tools.json               ← Agent tool schema（内部，编译时嵌入）
web/skills/
  liki.md                   ← 产品 skill 文件（公开）
  report-naming.md          ← 起名报告模板（公开，LLM 对话内直接生成）
```

### Embed 清单

| 变量 | 来源 | 可见性 | 用途 |
|---|---|---|---|
| `agent.ToolsJSON` | `internal/agent/data/tools.json` | 内部 | `NewNamingToolRegistry()` 解析 tool schema → LLM tool calling |
| `agent.NamingPrompt` | `internal/agent/data/naming.txt` | 内部 | `NamingAgent` 注入 system message |

### 外部 Agent 发现路径

```
GET /llms.txt               → 简要索引，引导安装 skill
GET /skills/liki.md         → 完整产品描述（角色、工作流、API、错误处理）
GET /skills/report-naming.md → 起名报告模板
POST /jsonrpc rpc.discover → OpenRPC 1.4.1 文档（29 个 engine method）
```

## 关键设计决策

### 先付费后收集

无免费体验。支付在前，用户承诺付费后才进入 Agent 收集出生信息。避免用户在收集完信息后被突然要求付费的心理落差。

### 7 天 Chat

一次付费，7 天内可反复磋商。以 email 为主键，不建 users 表。JWT cookie 管理会话。聊天历史全量持久化到 `chat_messages`，每次 resume 重放。

### birth_info 数据切面

三层结构 `raw → geo → timeset`。`query_city` 和 `compute_time` 由 Agent 在对话中按需调用。出生信息随聊天消息自然收集，持久化到 `chat_messages`。无需单独的 confirm_birth 机制。

### 工具精简

Naming Agent 8 个 tool：`query_city`、`compute_time`、`compute_chart`、`compute_ziwei` + 起名域 4 个（`compute_naming_wuge`、`compute_naming_compose`、`compute_naming_detail`、`compute_naming_evaluate`）。API 平台的 29 个 JSON-RPC method 不变（含全部 Engine 包的计算能力）。

### 类型归属

- `llm.Message/Role/ToolCall` — LLM 线格式
- `agent.TimePoint` — 出生时间点（RFC3339 公历 + 经度）
- `handler.BirthRequest` — HTTP 契约（统一出生+性别）
- `ganzhi.Gan/Zhi/Wuxing/ShiShen/Zhu/Bazi` — 干支基础类型
- `tianwen.SolarTime/GregorianTime/LunarTime/Timeset` — 时间类型

### 单连接 SQLite

`MaxOpenConns=1`，WAL 模式。所有数据库操作串行化。

### LLM 集成

- 模型：DeepSeek V4 Pro
- Agent：流式 `ChatStreamWithTools`，120s 超时
- 单一 Agent 实例，单阶段流式对话，自然过渡到报告生成

### 认证

JWT cookie（含 email + order_id）。`POST /api/auth/login` 查询有效订单后签发。多订单时用户选择后签发。无用户表，无注册流程。
