# 聊天系统设计 · Chat System Design

## 概述

灵机聊天子系统用单页对话界面替代原有的 3 个独立表单页（chart.html、bond.html、naming.html）。用户用自然语言描述需求，Agent 检测意图、收集参数、排盘、生成报告，最后引导至支付。

核心思路：**引导式收集 + 一次性报告生成**，不是自由闲聊。对话的归宿是排盘出报告。

目标用户以 30-55 岁为主，移动端占比高。界面需大字、大触摸区、每次只问一个问题以降低认知负荷。

## 用户旅程

```
用户打开 /chat
  │
  ▼
[欢迎]  空状态
  │     3 个快捷 chip：八字排盘 / 合盘配对 / 起名分析
  │     也可以直接输入自然语言
  │     用户输入或点击 chip → 创建 session，进入收集中
  │
  ▼
[收集中]  多轮对话（3-6 轮）
  │
  ├─ chart:  检测产品 → 收集参数（年/月/日/时/城市/性别）
  │
  ├─ naming: 检测产品 → 收集参数 + 姓氏
  │          姓氏必填（用于三才五格计算），缺则追问
  │
  └─ bond:   检测产品 → 收集甲方参数 → 追问乙方 → 收集乙方参数
  │
  │       用户不知道时辰？→ LLM 追问一次 → 仍不知则降级继续
  │       城市查不到？→ LLM 转问用户时区 → 用户手动输入
  │       每轮用户看到 phase 事件，知道 agent 在做什么
  │
  ▼
[生成中]  参数确认后 LLM 调 compute_* tool
  │       engine 计算 + LLM 流式输出 teaser 报告
  │       前端节流渲染（80ms/20字窗口）→ 丝滑不卡顿
  │       随时可点停止
  │
  ▼
[预览]   报告完成后前端自动展示购买卡片
  │       卡片含价格（货币自适应）和"查看完整报告"按钮
  │       LLM 不再生成 CTA 文案——购买引导由前端 done 事件统一处理
  │       卡片浮在消息区底部，独立于聊天气泡
  │
  ▼
[支付]   跳转 /report/{orderID}
         order 在 done 事件前已创建
         支付成功 → webhook → 标记 paid + 生成完整报告 → 报告可访问
```

## 状态机

### 后端 Session 状态（服务端权威）

```
                        ┌────────────┐
        session         │ COLLECTING │  参数收集，Agent 追问
        created         └─────┬──────┘
                              │ LLM purchase tool 触发
                              │ Chat → purchase → done
                              ▼
                        ┌────────────┐
                        │   CLOSED   │  终态，done 事件已发出
                        └────────────┘

任一状态 ──(TTL 超时)──→ 从 store 删除（移出 map）
```

session 创建即进入 COLLECTING。没有独立的 NEW 状态——首条 POST 既创建 session 也携带首条消息。

**状态转换触发点：**

| 转换 | 触发条件 | 位置 |
|---|---|---|
| (创建) → COLLECTING | 首条用户消息，session 创建 | `http/agent.go` handler |
| COLLECTING → COLLECTING | LLM 追问/teaser/Q&A | `agent/chat_agent.go` Chat() |
| COLLECTING → CLOSED | purchase tool + CreateOrder + done 事件 | `agent/chat_agent.go` Chat() |

**并发保护：**

- COLLECTING：允许新消息（多轮收集）
- CLOSED：拒绝新消息，返回 400（session 已关闭）

### 前端 UI 状态（展示层）

```
welcome ──(send)──→ chatting(loading) ──(text-delta)──→ streaming ──(done)──→ preview
                         │                    │                            │
                         │                    └──(stop)──→ chatting(idle)  │
                         │                           │                     │
                         │                           │ 停止后显示轻量引导   │
                         │                           │ "报告未完成，可重新生成 │
                         │                           │  或继续对话"          │
                         │                                                 │
                         ├──(question)──→ chatting(idle) ──(send)──→ chatting(loading)
                         │
                         ├──(recoverable error)──→ chatting(idle)
                         │   瞬态错误 toast + 自动恢复，不丢上下文
                         │
                         └──(fatal error/http err)──→ error ──(newChat)──→ welcome
                              session 过期 / 服务崩溃 / 引擎异常
```

**状态定义：**

| 状态 | 含义 | 输入框 | 消息区 | 欢迎区 | 停止按钮 |
|---|---|---|---|---|---|
| `welcome` | 空会话，未创建 session | 显示 | 隐藏 | 显示 | 隐藏 |
| `chatting` | 活跃对话，等待输入或等待响应 | 显示 | 显示 | 隐藏 | 隐藏 |
| `streaming` | LLM 流式输出中，输入框替换为停止按钮 | 隐藏 | 显示 | 隐藏 | 显示 |
| `preview` | 报告预览卡片 + 支付按钮 | 隐藏 | 显示 | 隐藏 | 隐藏 |
| `error` | 致命错误，只能重来 | 显示 | 显示 | 隐藏 | 隐藏 |

**chatting 子态（`substate` 字段）：**

| 子态 | 含义 | 发送按钮 | loading dots |
|---|---|---|---|
| `idle` | 等待用户输入 | 可用 | 无 |
| `loading` | POST 已发，等待 SSE 首字节 | 禁用 | 最后一个 asst bubble 显示 |

**状态转换详解：**

| 转换 | 触发 | 说明 |
|---|---|---|
| welcome → chatting(loading) | sendMessage() | 首条消息，同步设 `pending = true` 防重 |
| chatting(loading) → streaming | 收到首个 text-delta | 流式开始 |
| chatting(loading) → chatting(idle) | 收到 question | agent 追问，等待回复 |
| chatting(loading) → chatting(idle) | 收到 error 且 recoverable=true | toast 提示后恢复，用户可重发 |
| chatting(loading) → error | 收到 error 且 recoverable=false | 致命错误，只能 newChat() |
| chatting(loading) → error | HTTP 错误 / 网络异常 | 同上 |
| chatting(idle) → chatting(loading) | sendMessage() | 用户回复追问 |
| streaming → chatting(idle) | stopStream() | 用户停止，显示引导文案 |
| streaming → preview | 收到 done | 报告完成 |
| streaming → chatting(idle) | 收到 error 且 recoverable=true | 流式中的可恢复错误 |
| streaming → error | 收到 error 且 recoverable=false | 流式中的致命错误 |
| preview → (页面跳转) | goPayment() | 重定向到 /report/{orderID} |
| error → welcome | newChat() | 用户清空重来 |
| 任意 → welcome | newChat() | 开始新会话 |

## 会话通道：POST + SSE

### 为什么是 POST + SSE 而不是 WebSocket

| 维度 | POST + SSE | WebSocket |
|---|---|---|
| 调试 | curl 直接测，HTTP 工具链成熟 | 需要专用客户端 |
| 代理/负载均衡 | 天然兼容 | 需要协议升级 |
| 连接模型 | 每轮一个短连接，用完即关 | 长连接，需心跳保活 |
| 适用场景 | 用户间歇输入（10-60s 间隔），响应需流式 | 高频双向实时通信 |
| 断线恢复 | session ID 恢复历史 | 重连 + 状态同步 |
| 实现复杂度 | net/http + Flusher，~50 行 | gorilla/websocket 或 nhooyr.io/websocket |

灵机聊天：用户打字间隔 10-60 秒，服务端需要流式推送 LLM token。POST + SSE 是最佳匹配。

### 通道模型

```
Turn 1:
  Browser ──POST {session_id:"", message:"想看八字"}──→ Go
  Browser ←══SSE: tool-progress → question ←══════════ Go

Turn 2:
  Browser ──POST {session_id:"abc123", message:"1990年5月 北京"}──→ Go
  Browser ←══SSE: tool-progress → question ←══════════════════════ Go

  ...

Turn N:
  Browser ──POST {session_id:"abc123", message:"申时 男"}──→ Go
  Browser ←══SSE: tool-progress → computing → text-delta... → done ←══ Go
```

每次 POST 都是一个完整的 HTTP 请求-响应周期。session_id 把多轮链接成一个会话。

### Session 即通道

```
Session 结构：
  ID          string           // crypto/rand hex，通道标识符
  Phase       Phase            // collecting | closed
  Messages    []llm.Message    // 完整对话历史（system + user + assistant + tool）
  CreatedAt   time.Time
  ExpiresAt   time.Time        // 30min TTL，每次 POST/Touch 刷新
```

- 前端存 `sessionStorage.getItem('chatSessionID')`，关闭 tab 即丢
- 服务端 `map[string]*Session` + `sync.RWMutex`，5min 清理过期
- 每次 POST 刷新 ExpiresAt
- 每条请求携带 `lang` 字段（zh/hk/en），后端映射为 BCP 47 locale → 替换 system prompt 中的 `{locale}` 占位符，控制 LLM 输出语言

## SSE 事件协议

### 事件类型

| type | 含义 | content | 其他字段 | 触发阶段 |
|---|---|---|---|---|
| `text-delta` | LLM 流式 token | 增量文本 | — | Phase 1 / Phase 3 |
| `thinking-delta` | LLM 推理内容 | 推理增量 | — | Phase 1 / Phase 3 |
| `thinking` | 思考中指示 | — | — | Phase 1 开始 / Phase 3 开始 |
| `phase` | 阶段进度描述 | 可读描述 | Data (CollectProgress) | Phase 1 / Phase 2 / Phase 3 |
| `done` | 报告完成 | — | product, order_id, amount | Phase 3 末尾 |
| `error` | 错误 | 错误描述 | — | 任意阶段 |

error 事件表示致命错误（session 过期、引擎崩溃、服务 500），前端进入 error 终态，只能 newChat()。可恢复的降级场景（城市未收录、时辰未知）由 LLM 在对话中处理，不发 error 事件。

### 事件流示例

**Chart 完整流程（多轮收集 + 排盘）：**

```
data: {"type":"thinking"}
data: {"type":"text-delta","content":"您好，我来为您排盘"}
data: {"type":"text-delta","content":"。请问您的出生年月日是什么？"}

[用户下一轮 POST：1990年5月20日 北京]

data: {"type":"thinking"}
data: {"type":"phase","content":"正在查询地理信息…","data":{"phase":"collect","tool":"query_city","status":"tool_done"}}
data: {"type":"text-delta","content":"请问出生时辰和性别？"}

[用户下一轮 POST：下午3点 男]

data: {"type":"thinking"}
data: {"type":"text-delta","content":"参数已确认，正在为您排盘，请稍候。"}
data: {"type":"phase","content":"正在计算命理数据…"}
data: {"type":"phase","content":"正在生成分析报告…"}
data: {"type":"thinking"}
data: {"type":"text-delta","content":"\n## 八字排盘\n\n"}
data: {"type":"text-delta","content":"您的日主为"}
...
data: {"type":"done","data":{"order_id":"xxx","amount":990,"product":"chart"}}
```

**Bond 双人收集：**

```
data: {"type":"thinking"}
data: {"type":"text-delta","content":"请提供您的出生年月日时、出生城市和性别"}

[用户 POST：甲方信息]

data: {"type":"thinking"}
data: {"type":"phase","content":"正在查询地理信息…","data":{"phase":"collect","tool":"query_city","status":"tool_done"}}
data: {"type":"text-delta","content":"请提供另一方的出生信息"}

[用户 POST：乙方信息]

data: {"type":"thinking"}
data: {"type":"phase","content":"正在查询地理信息…","data":{"phase":"collect","tool":"query_city","status":"tool_done"}}
data: {"type":"phase","content":"正在计算命理数据…"}
data: {"type":"phase","content":"正在生成分析报告…"}
data: {"type":"thinking"}
data: {"type":"text-delta","content":"\n## 合盘分析\n\n"}
...
data: {"type":"done","data":{"order_id":"xxx","amount":1990,"product":"bond"}}
```

**致命错误：**

```
data: {"type":"error","content":"会话已过期，请重新开始"}
```

## 核心数据流

### 每轮对话详细时序

```
Browser                    http/agent.go                 agent.Chat()            LLM
  │                             │                         │                      │
  │─ POST /api/agent/chat ─────→│                         │                      │
  │  {session_id, message, lang}│                         │                      │
  │                             │─ load/create session    │                      │
  │                             │─ sess.AppendMessage(user msg)                  │
  │                             │─ chat.Chat(ctx, locale, msgs, onEvent, …) →    │
  │  ←══ SSE: thinking ════════│←══ onEvent(ev) ════════│─ ChatStreamWithTools() →│
  │  ←══ SSE: text-delta ══════│←══ onEvent(ev) ════════│  ← text-delta ────────│
  │  ←══ SSE: phase ═══════════│←══ onEvent(ev) ════════│  ← phase ─────────────│  query_city
  │                             │                         │                      │                  │
  │                             │                         │  no tool call:       │                  │
  │                             │─ sess.SetMessages(msgs) │  return, no purchase │                  │
  │                             │─ return 200 (stream closed)                     │                  │
  │                             │                         │                      │                  │
  │  [用户看到追问，输入回答，下一轮 POST...]              │                      │                  │
  │                             │                         │                      │                  │
  │─ POST /api/agent/chat ─────→│                         │                      │                  │
  │  {same session_id, answer}  │                         │                      │                  │
  │                             │─ chat.Chat(ctx, locale, msgs, onEvent, …) →    │                  │
  │  ←══ SSE: text-delta ══════│←══ onEvent(ev) ════════│  ← text-delta ────────│  compute_* tool → engine│
  │  ←══ SSE: phase ═══════════│←══ onEvent(ev) ════════│  ← phase ─────────────│─ ChatStreamWithTools() →│
  │  ←══ SSE: text-delta ══════│←══ onEvent(ev) ════════│  ← text-delta ────────│  teaser 报告流
  │  [前端节流渲染: 80ms/20字窗口]                        │                      │  teaser 报告流        │
  │                             │                         │                      │                  │
  │  [用户追问 Q&A, 继续对话...]│                         │                      │                  │
  │                             │                         │                      │                  │
  │─ POST /api/agent/chat ─────→│  (purchase intent)      │                      │                  │
  │  ←══ SSE: text-delta ══════│←══ onEvent(ev) ════════│  ← text-delta ────────│  purchase tool       │
  │                             │                         │─ handlePurchase      │                  │
  │                             │                         │─ CreateOrder(chartJSON + Q&A)              │
  │  ←══ SSE: done ════════════│←══ onEvent(ev) ════════│─ done {order_id, amount, product}          │
  │                             │─ sess.SetPhase(PhaseClosed)                   │                  │
  │                             │─ return 200              │                      │                  │
  │  [用户看到购买栏，点击支付]  │                         │                      │                  │
```

### 中断与异常处理

```
主动取消（用户点停止）：
  AbortController.abort() → fetch 取消
  r.Context().Done() → handler 检测到 ctx 取消
  ├─ 已收到的 assistant content 保存到 session
  ├─ goroutine 退出
  └─ 前端回到 chatting(idle)，显示引导："报告未完成，可重新生成或继续对话"

主动取消（用户关 tab）：
  同上，但前端状态丢失（session 还在服务端，下次同 session_id POST 可继续）

可恢复的降级场景（城市未收录 / 时辰未知）：
  LLM 在 Phase 1 对话中自然追问，不发 error 事件
  ├─ 前端保持 chatting(idle)
  └─ session 保持，用户可继续回答

致命错误（session 过期 / 引擎崩溃 / 服务 500）：
  ├─ SSE error 事件
  ├─ 前端进入 error 终态
  └─ 只能 newChat() 重新开始

流中断（网络波动，SSE 连接断开）：
  fetch 抛异常 → 前端按致命错误处理
  已收到的部分 content 未保存 → 丢失
  用户需重新发送消息触发重新生成

Phase 2 引擎错误：
  发送 SSE error 事件
  bazi compute 失败 → 可能是参数问题，用户可重新对话
  Go panic → 服务异常
```

## 前端渲染性能

### 核心问题：O(N²) 的流式渲染

当前 `handleEvent` 每次 text-delta 的处理路径：

```
asst.content += evt.content                // 字符串追加
asst.html = this.renderMD(asst.content)    // 全量 marked.parse + DOMPurify → innerHTML
this.scrollDown()                          // DOM 读 + 写
```

对于一个 2000 字的报告，DeepSeek 每块吐 5-15 个中文字符，总计 **130-400 次 text-delta 事件**。每次事件全量解析累积 markdown + DOMPurify 净化 + innerHTML 替换。总字符处理量 = Σ(1..N) = O(N²)，移动端中低端机型上肉眼可见卡顿。

### 优化：节流渲染

不每个 token 都 render。累积 token，定时刷新：

```
text-delta → append to accumulator (asst.content)
             if (time since last render > 80ms) → render + commit
             or if (accumulated chars > 20) → render + commit
```

- `asst.content` 始终实时追加（保持数据完整）
- `asst.html` 只在节流窗口到期时更新
- `done` / `question` / `error` 事件到达时强制最后一次 render
- 80ms 对人眼不可感知（人眼在 100ms 内的变化视为即时）

效果：400 次 DOM 操作 → ~25 次，减少 94%。

### 滚动锚定

```
scrollDown() 仅在用户处于底部时执行：
  scrollTop + clientHeight >= scrollHeight - 64px
用户手动上滚查看已生成内容时，新 token 静默追加不抢滚动位。
```

### 防重复发送

```
sendMessage() 开头同步设 this.pending = true
finally 清除
```

当前用 `phase === 'loading'` 检查有竞态——await 期间用户连点两次，第一次还没设 phase，第二次也通过。

### CDN 预热 + SRI

```html
<link rel="preconnect" href="https://cdn.jsdelivr.net">
<script src="..." integrity="sha384-..." crossorigin="anonymous">
```

preconnect 提前建立 TLS 连接节省 1 RTT。SRI 保证 CDN 资源未被篡改。

### 性能基线

| 指标 | 当前 | 优化后 |
|---|---|---|
| text-delta → 屏幕可见 | 同步（微任务内） | ≤80ms 节流窗口 |
| 2000 字报告 DOM 操作 | 130-400 次 innerHTML | ~25 次 |
| 移动端总 CPU 时间 | ~4s（分散在流式过程中） | ~250ms |
| 滚动帧率 | 可能 < 30fps | 稳定 60fps |

## Agent 系统提示词要求

ChatSystemPrompt（`data/prompts/chat.txt`）需覆盖以下领域约束：

```
产品检测：
  - 八字/命盘/运势/排盘 → chart
  - 合婚/合盘/两人/配对 → bond
  - 取名/起名/改名 → naming

必要参数（必须由用户明确提供，禁止假设）：
  - 出生年、月、日（公历）
  - 出生时辰（性别决定大运顺逆，结果完全不同）
  - 出生城市（用于经纬度和时区）
  - 性别（男命顺排大运、女命逆排大运）

城市查找：
  query_city 查经纬度 → 根据国家代码和日期推时区
  查不到 → 请用户提供时区

时辰降级：
  必须先追问一次，解释时辰对排盘的重要性
  用户仍不知道 → 填 12:00，告知准确度降低，继续排盘

起名特殊要求：
  - 姓氏必填（用于三才五格计算），缺则追问
  - 性别影响用字选择
```

## 关键设计决策

### Session 不持久化到 SQLite

Session 存内存，不落盘。原因：
- 聊天是短暂的（30min TTL），不是持久数据
- 一旦出报告，order 已存 SQLite，order 是持久记录
- 减少 SQLite 写入压力
- 进程重启丢 session 可接受（聊天流程短，几分钟出报告）

### 每轮是独立 HTTP 请求

不维持长连接跨轮。原因：
- 用户可能间隔几十秒才回复，长连接浪费资源
- 独立的请求更容易做错误处理、超时控制
- session.Messages 承载完整上下文，请求间无依赖

### Agent 流式 tool-calling

agent 用 `ChatStreamWithTools`（SSE 流式），事件模型：
- `text-delta`：LLM 追问文本实时流式输出
- `phase`：工具执行进度
- `thinking-delta`：LLM 推理内容（DeepSeek V4 Pro reasoning）

工具参数 JSON Schema 从 `openapi.json` 提取，编译时嵌入，运行时注入 tool calling 的 `parameters` 字段。Agent 使用 5 个 tool（`get_city_coords`、`compute_chart`、`compute_bond`、`compute_naming`、`purchase`），HTTP API 共 28 个端点。`openapi.json`（v1.1.0）同时包含所有端点的响应 schema，供外部 AI agent 服务发现。

### Agent 流式 + 节流渲染

agent.Chat 用 `ChatStreamWithTools`（流式），报告长（几百到上千字），用户等不了。前端不每个 token 都 render，用 80ms/20 字窗口节流，减少 94% DOM 操作，移动端流畅不卡。

### 错误分级：可恢复 vs 致命

可恢复的降级场景（城市未收录、时辰未知、LLM 追问）由 LLM 在 Phase 1 对话中自然处理，不发 error 事件。error 事件仅用于致命错误（session 过期、引擎崩溃、服务 500），前端进入 error 终态只能 newChat()。

这样用户花 3 轮收集的参数不会因为一个瞬态错误全丢。

### 不做 session 持久化

Session 在内存，不落盘。原因：
- 聊天是短暂的（30min TTL），不是持久数据
- 一旦出报告，order 已存 SQLite，order 是持久记录
- 减少 SQLite 写入压力
- 进程重启丢 session 可接受（聊天流程短，几分钟出报告）
- `GET /api/agent/session` 端点可恢复当前 session 的消息历史，但不跨进程重启
- `GET /api/agent/greeting` 端点返回 greeting 消息

### LLM 消息历史策略

每轮 POST 发完整 `sess.Messages`（含 system prompt + 所有历史轮次）。不裁剪。原因：
- 一轮对话最多十几条消息（3-6 轮收集 + system + tool），远低于 DeepSeek 64K 上下文
- 裁剪逻辑复杂且容易丢关键上下文

## 实施计划

| 优先级 | 内容 | 涉及文件 | 状态 |
|---|---|---|---|
| **P0** | 前端状态机对齐 | `web/js/chat.js`，`web/chat.html` | ✅ 完成 |
| **P0** | 流式渲染节流 | `web/js/chat.js`（80ms/20 字窗口） | ✅ 完成 |
| **P0** | Agent 流式 tool-calling | `internal/agent/chat_agent.go`（ChatStreamWithTools），`internal/llm/client.go` | ✅ 完成 |
| **P1** | Phase 进度事件统一 | `internal/agent/chat_agent.go`（phase 事件替代 tool-progress/computing） | ✅ 完成 |
| **P1** | sess.Phase 驱动 + 流程编排 | `internal/http/agent.go`（收集 → agent → closed） | ✅ 完成 |
| **P1** | 滚动锚定 | `web/js/chat.js`（scrollDown 加 isNearBottom） | ✅ 完成 |
| **P1** | 防重复发送 | `web/js/chat.js`（sendMessage 加同步 pending 标记） | ✅ 完成 |
| **P1** | 异步 greeting 生成 | `cmd/lingji/main.go`（goroutine + fallback） | ✅ 完成 |
| **P2** | 活动刷新 ExpiresAt | `internal/http/agent.go`（每次 POST 刷新） | ✅ 完成 |
| **P2** | 错误分级（可恢复 vs 致命） | `web/js/chat.js`（error 事件 + recoverable） | ✅ 完成 |
| **P2** | 城市查找降级 | ChatSystemPrompt（tool error 时 LLM 问时区） | ✅ 完成 |
| **P2** | 性别参数显式要求 | ChatSystemPrompt（加性别必要说明） | ✅ 完成 |
| **P3** | 时辰未知降级 | ChatSystemPrompt（追问一次后降级继续） | ✅ 完成 |
| **P3** | `GET /api/agent/session` | `internal/http/agent.go` | ✅ 完成 |
| **P3** | 报告页支付按钮 + 布局优化 | `web/report.html`，`web/js/report.js` | ✅ 完成 |
| **P3** | SSE 截断修复（frontend + server flush） | `web/js/chat.js`，`internal/http/agent.go` | ✅ 完成 |
