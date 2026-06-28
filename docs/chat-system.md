# 聊天系统设计 · Chat System Design

## 概述

灵机聊天子系统是 AI 起名顾问的唯一入口。用户先付费，进入 7 天聊天窗口，与 AI 磋商起名，最后输出完整报告。

与旧架构的区别：
- 单产品（仅起名），不再有八字/合盘 chart
- JWT cookie 认证，不再有服务端 SessionStore
- LLM 在对话中直接输出报告，不再有 GenerateFromData 后台生成
- 消息持久化到 SQLite `chat_messages`，跨设备可续聊

目标用户以 30-55 岁为主，移动端占比高。界面需大字、大触摸区。

## 用户旅程

```
用户打开首页
  │
  ├─ 输入邮箱 → POST /api/orders（创建 pending 订单）
  ├─ POST /api/payments/checkout → 跳支付页面
  │
  ├─ [微信/支付宝/国际卡] 支付
  │
  ├─ 支付成功 → webhook → 标记 paid + chat_expires_at = now+7d
  ├─ 重定向 /chat → JWT cookie
  │
  ▼
[聊天中]  POST /api/agent/naming（SSE）
  │
  │  收集出生信息（城市、时间、性别、姓氏）
  │    → query_city → compute_time
  │    → compute_chart → compute_ziwei
  │
  │  磋商起名
  │    → compute_naming_wuge → compose → detail → evaluate
  │    → 多轮讨论，反复调整
  │
  │  生成报告（用户说"生成报告"）
  │    → LLM 直接输出 markdown 报告
  │    → handler 识别 "# 起名报告"
  │    → 存 llm_json
  │    → report_ready 事件 → 前端跳转 /report/{id}
  │
  ▼
[报告页]  /report/{orderID}
  │  完整起名报告
  │  可分享链接（无需登录）
```

### 7 天内续聊

```
输邮箱 → POST /api/auth/login
  │
  ├─ FindActiveOrdersByEmail → 查 paid 且未过期订单
  ├─ JWT cookie
  └─ 重定向 /chat
       │
       └─ POST /api/agent/naming
            ├─ LoadChatHistory → 全部历史消息
            └─ LLM 看到完整上下文，无缝续聊
```

### 多订单选择

```
第一单 7 天过期前又买一单：

FindActiveOrdersByEmail → 2 个订单
  → 返回订单列表（order_id + 到期时间）
  → 用户选择其中一个
  → JWT cookie 绑定所选 order_id
```

## 认证与会话

```
无服务端 session，纯 JWT cookie：

  登录                         每个请求
  ┌──────────┐                ┌──────────────┐
  │ setJWTCookie │             │ jwtAuth(r)   │
  │            │                │              │
  │ JWT payload:               │ 读取 cookie   │
  │  email      │              │ 验签 HS256    │
  │  order_id   │              │ 提取 email    │
  │  exp: +24h  │              │ 提取 order_id │
  │            │                │              │
  │ Cookie:    │                │ handler 用    │
  │  liki_token│               │ order_id 查 DB│
  │  HttpOnly  │                │ 获取完整状态  │
  │  Secure    │                └──────────────┘
  │  SameSite:Lax
  │  MaxAge:86400
  └──────────┘

JWT 24h 过期 → 重新输邮箱登录即可（无需重新支付）
无用户表、无注册流程
```

## 核心数据流

### 单次请求完整时序

```
Browser                    namingHandler              NamingChatAgent          LLM (DeepSeek)
  │                             │                         │                      │
  │─ POST /api/agent/naming ──→│                         │                      │
  │  {message, lang}            │                         │                      │
  │                             │─ jwtAuth → order_id     │                      │
  │                             │─ GetOrder → 校验        │                      │
  │                             │─ LoadChatHistory        │                      │
  │                             │─ CreateChatMessage(user)│  ← DB 写入（立即）   │
  │                             │                         │                      │
  │                             │─ NamingChat(ctx, locale, msgs, onEvent) →      │
  │                             │                         │─ ChatStreamWithTools →│
  │  ←══ SSE: thinking ═══════│←══ onEvent ════════════│← thinking          │
  │  ←══ SSE: phase ══════════│←══ onEvent ════════════│← phase             │
  │  ←══ SSE: text-delta ═════│←══ onEvent ════════════│← text-delta        │
  │                             │                         │                      │
  │                             │                         │── tool_call ──────→  │
  │                             │                         │   Execute(query_city)│
  │                             │                         │← tool_result ────   │
  │                             │                         │── text-delta ────→  │
  │                             │                         │   ...               │
  │                             │                         │                      │
  │                             │← result (all messages)  │                      │
  │                             │                         │                      │
  │                             │─ IsNamingReport(last)？ │                      │
  │                             │   YES → UpdateLlmJSON   │                      │
  │                             │   YES → report_ready    │                      │
  │  ←══ SSE: report-ready ═══│                         │                      │
  │                             │─ BatchCreateChatMessages│  ← DB 写入（批量）   │
  │                             │─ flushSSE               │                      │
  │  ←══ SSE: :ok ════════════│                         │                      │
  │                             │─ return 200             │                      │
  │                             │                         │                      │
  │  [前端收到 report-ready → setTimeout → location.href = /report/{id}]           │
```

### 中断处理

```
主动取消（用户点停止）：
  ctx cancel → NamingChat 返回 ctx.Err()
  → handler 不写后续消息，但已有的 CreateChatMessage(user) 已入库
  → 前端回到 idle 状态

主动取消（用户关 tab）：
  同上。消息已在 DB，下次续聊可恢复。

LLM 超时（120s）：
  SSE error 事件 → 前端显示错误提示 → 用户可重发

订单过期（7 天后）：
  handler 返回 403 "聊天已过期" → 前端显示过期提示
```

## SSE 事件协议

### 事件类型

| type | 含义 | content | 触发时机 |
|------|------|------|------|
| `text-delta` | LLM 流式 token | 增量文本 | LLM 输出中 |
| `thinking-delta` | LLM 推理 | 推理增量 | DeepSeek V4 Pro reasoning |
| `thinking` | 思考开始 | — | LLM 开始推理 |
| `phase` | 阶段提示 | 可读描述 | tool 执行中 |
| `report-ready` | 报告已生成 | "/report/{orderID}" | 流结束，检测到报告 |
| `error` | 错误 | 错误描述 | 异常 |

不再有 `done` 事件。旧架构用 done 表示报告完成 + 订单创建，新架构 report-ready 替代。

### 事件流示例

**首轮对话（收集出生信息）：**

```
data: {"type":"thinking"}
data: {"type":"text-delta","content":"您好"}
data: {"type":"text-delta","content":"！请"}
data: {"type":"text-delta","content":"问您的"}
data: {"type":"text-delta","content":"出生年月日？"}
```

**Tool calling 轮（计算八字）：**

```
data: {"type":"thinking"}
data: {"type":"text-delta","content":"好的，我先核实一下您的出生时间。"}
data: {"type":"phase","content":"正在计算命理数据…"}
data: {"type":"text-delta","content":"您的八字是…"}
```

**报告生成轮：**

```
data: {"type":"thinking"}
data: {"type":"text-delta","content":"# 起名报告\n\n"}
data: {"type":"text-delta","content":"## 命理分析\n\n"}
...
data: {"type":"text-delta","content":"以上是完整的起名报告。"}
data: {"type":"report-ready","content":"/report/xxx-yyy-zzz"}
data: ": ok\n\n"
```

## 前端状态机

### UI 状态

```
welcome ──(send)──→ chatting(loading) ──(text-delta)──→ streaming
                         │                    │
                         │                    ├──(stop)──→ chatting(idle)
                         │                    │
                         │                    └──(report-ready)──→ redirect /report/{id}
                         │
                         ├──(question)──→ chatting(idle) ──(send)──→ chatting(loading)
                         │
                         └──(error)──→ 错误提示（可重发）
```

### 状态定义

| 状态 | 含义 | 输入框 | 停止按钮 |
|------|------|------|------|
| `welcome` | 空会话 | 显示 | 隐藏 |
| `chatting` | 活跃对话，等待输入或响应 | 显示 | 隐藏 |
| `streaming` | LLM 流式输出中 | 隐藏 | 显示 |

### chatting 子态

| 子态 | 含义 |
|------|------|
| `idle` | 等待用户输入 |
| `loading` | POST 已发，等待 SSE 首字节 |

### 防重复发送

```
sendMessage() 开头同步设 pending = true
finally 清除
```

## 前端渲染性能

### 节流渲染

```
text-delta → append to accumulator
             if (time since last render > 80ms) → render
             or if (accumulated chars > 20) → render
```

### 滚动锚定

```
scrollDown() 仅在用户处于底部时执行：
  scrollTop + clientHeight >= scrollHeight - 64px
```

## 关键设计决策

### JWT 替代 Session Store

旧架构用服务端 `map[string]*Session` + `sync.RWMutex`。新架构用 JWT cookie：
- 无状态，无并发 map 操作
- 进程重启不丢会话
- 不需要 TTL 清理 goroutine
- 消息持久化到 SQLite，不依赖内存 session

### 单请求内完成持久化

每轮 POST 独立完成：读历史 → LLM → 写消息。不维持跨轮长连接：
- 用户打字间隔可能几十秒，长连接浪费资源
- 独立请求更容易做错误处理和超时控制
- 消息在 DB 中，跨设备可续聊

### LLM 直接输出报告

旧架构支付后 webhook 触发生成完整报告（GenerateFromData）。新架构：
- LLM 在对话中直接输出 markdown 报告
- handler 识别 `# 起名报告` 标题头 → 存 llm_json → 发 report-ready
- 报告模板（`web/skills/report-naming.md`）保留为对外参考文档

理由：
- DeepSeek V4 Pro 2M 上下文窗口，数据全在上下文里
- 减少一条 webhook → 后台任务 → 轮询的链路
- 用户看到报告生成过程（逐 token 流式），体验更好

### 上下文策略

每轮 POST 发完整历史消息。不裁剪：
- 起名对话最多几十条消息，远低于 DeepSeek 2M 上下文
- 裁剪逻辑复杂且容易丢关键上下文
- system prompt 按 locale 缓存在 `sync.Map`，避免每次 ReplaceAll

### 错误分级

| 级别 | 场景 | 处理 |
|------|------|------|
| 可恢复 | 城市未收录、时辰未知 | LLM 对话中自然追问 |
| 致命 | LLM 超时、引擎异常、订单过期 | SSE error 事件 → 前端提示 |
