# 灵机全系统复盘：状态机 · 业务流 · 数据流

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
   Session(内存)  无状态          无状态
        │
   Payment (Dodo/虎皮椒 webhook)
```

三条用户路径：
- **Chat 流**：自然语言 → 收集参数（tool calling） → 排盘（compute_* tool） → LLM 流式 teaser → SSE done → 支付 → webhook 触发完整报告生成
- **Form 流**：表单提交 → Engine 计算 → 返回 JSON（无 LLM、无持久化）
- **Free API**：直接调引擎 → 返回 JSON（无 LLM、无持久化、无支付）

Form 流和 Free API 的 HTTP handler 都在 `internal/http`（bazi.go, qiming_handler.go 等），
直接调用 `bazi.ComputeChart()` 等 engine 函数，没有独立的 chart/bond/naming 包。

---

## 二、状态机复盘

### 2.1 服务端 Session（Chat 路径）

```
               POST 首次
                  │
                  ▼
          ┌──────────────┐
          │  COLLECTING  │◀── POST 追问
          └──────┬───────┘
                 │ purchase 调用 触发
                 ▼
          ┌──────────────┐
          │   CLOSED     │  终态，done 事件已发出
          └──────────────┘
```

**评价：简洁合理。** ChatAgent 单实例管理全流程，状态仅两态。LLM 通过 tool calling 驱动参数收集和排盘，无需中间状态（CONFIRMING/COMPUTING/GENERATING）的显式管理——这些是 agent 内部的过程，sess.Phase 只关心会话是否还能继续接收消息。

### 2.2 客户端 Chat UI

```
welcome ──(send)──→ chatting(loading) ──(text-delta)──→ streaming ──(done)──→ closed
  ↑                      │                                 │
  │                      │ (无 done, 流结束)                 │ (stop)
  │                      └──→ chatting(idle) ←─────────────┘
  │                                 │
  └────(newChat)────────────────────┘
```

**评价：合理。** Vue 3 reactive phase + substate 组合是标准实践。`streaming` 是统一态——收集和生成用同一 SSE 通道、同一 text-delta 事件，区别只在退出条件。

### 2.3 客户端 Report 页面 —— ✅ 已修复

```
当前：phase enum (loading/polling/ready/timeout/error)
  loading ──→ polling ──→ ready
                │             │
                └──→ timeout  │
                │             │
  error ←──────┴─────────────┘
```

已改为 `phase: 'loading' | 'polling' | 'ready' | 'timeout' | 'error'` enum，消除了 6 boolean 不一致状态的风险。

### 2.4 支付状态机

```
pending ──(webhook)──→ paid
  │
  └──(24h 过期清理)──→ deleted
```

**评价：合理。** 两态 + 定时清理，无需更复杂的状态。

---

## 三、数据流复盘

### 3.1 Chat 流完整链路

```
用户输入 (自然语言)
  │
  ├─ agent.Chat() → ChatStreamWithTools
  │
  ├─ LLM 追问收集参数
  │    └─ query_city tool → 城市经纬度确认
  │
  ├─ 参数齐全 → LLM 决定调 compute_* tool
  │    └─ tool handler 调用 engine.ComputeChart/ComputeBond 等
  │         └─ engine 返回 JSON → tool result 注入对话
  │
  ├─ LLM 基于 engine 结果流式生成 teaser 报告
  │    （Markdown，SSE text-delta 事件）
  │
  ├─ LLM 自然引导购买（最多 3 次）
  │
  └─ LLM 调 purchase 调用 → handlePurchase
       ├─ CreateOrder(chartJSON, product, amount) → SQLite (status=pending)
       └─ SSE done 事件 {order_id, amount, product}
            └─ sess.SetPhase(PhaseClosed)
```

### 3.2 数据层结构

| 层 | 类型 | 序列化目标 |
|---|---|---|
| Agent tool handler → Engine | tool args → `ComputeChart(st, …)` | — |
| Engine → Agent | 各 engine 产物（bazi.Chart, liuyao.Chart, qiming.NameCandidate 等） | `orders.chart_json` |
| LLM | 流式 string (Markdown) | `orders.llm_json` |

`chart_json` 存储 engine 完整计算结果的 JSON。`llm_json` 存储 LLM 生成的报告文本。

### 3.3 数据冗余点

| 冗余 | 说明 | 判断 |
|---|---|---|
| `chart_json` 存完整 engine 结果 | 包含用神、大运、流年等全部数据 | ✅ 合理，一份 JSON 含所有信息 |
| Bond 包含两个完整 Chart | 合盘需要双方完整命盘 | ✅ 正确，产品语义要求 |
| `llm_json` 创建时为空 | 支付后 webhook 触发 GenerateFromData + 缓存 | ✅ 合理，延迟生成 |
| Order 含 product 字段 | 标识 chart/bond/naming 产品类型 | ✅ 必要，报告页路由用 |

---

## 四、前端代码一致性

### 4.1 重复代码 —— ✅ 已修复

| 代码 | 位置 | 说明 |
|---|---|---|
| `goPay()` | `web/js/api.js` | 已提取为共享函数，chat.js 和 report.js 均调用 |
| `apiGet` | `web/js/api.js` | report.js 已改用 `apiGet` 统一请求 |
| Markdown 渲染 | `chat.js` + `report.js` | `marked.parse` + `DOMPurify` 仍有重复（两个页面不同的渲染逻辑） |

### 4.2 状态管理风格

| 页面 | 框架 | 状态方式 | 一致性 |
|---|---|---|---|
| `chat.js` | Vue 3 | `phase` + `substate` enum | ✅ |
| `report.js` | Lit-html | `phase` enum | ✅ |
| `index.html` | Vue 3 | 静态 | ✅ |

### 4.3 i18n

| 页面 | 使用 I18N | 一致性 |
|---|---|---|
| `index.html` | `$store.i18n.t()` | ✅ |
| `report.html` | `$store.i18n.t()` | ✅ |
| `chat.html` | `window.I18N.t()` | ✅ |

### 4.4 缺项

- ✅ `apiGet`/`apiPost` 已加 `AbortSignal.timeout()`（默认 30s）
- 表单无客户端校验（仅空值检查）
- 无离线/重连处理

---

## 五、行业标准实践对照

| 实践 | 灵机 | 评价 |
|---|---|---|
| POST + SSE 流式 LLM | ✅ | 最佳匹配，比 WebSocket 更简单 |
| 节流渲染 (80ms) | ✅ | 标准做法，减少 94% DOM 操作 |
| StreamRenderer 抽象 | ✅ | content→html 不变量的显式编码 |
| Phase enum 状态机 | ✅ (server + chat.js + report.js) | |
| 状态机显式不变量 | ✅ (flush 后 content≡html) | |
| 响应式 UI (Vue 3) | ✅ | 声明式，框架自动 DOM 同步 |
| 统一错误 envelope | ✅ | `{"error":{"code":"...","message":"..."}}` |
| 分层架构 | ✅ | Handler → Service → Store |
| Lazy generation + cache | ✅ | llm_json 首次付费时生成并缓存 |
| SQLite WAL + 单连接 | ✅ | 匹配低写入量场景 |
| Session 内存不落盘 | ✅ | 短生命周期，无需持久化 |
| 幂等 webhook 验签 | ✅ | Dodo + 虎皮椒双通道签名验证 |
| 可恢复/致命错误分级 | ✅ | chat 流 SSE error 事件 + recoverable 字段 |
| 并发保护 (pending guard) | ✅ | 防重复提交 |
| 流取消 (AbortController) | ✅ | 用户可停止流式输出 |

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
| P1 | 三条重复 {Birth, Gender} 请求 struct 合并为 BirthRequest | `bazi.go`, `ziwei.go`, `fengshui.go` | ✅ 完成 |
| P1 | error code 标准化（validation_error / invalid_request 区分） | `handler_helpers.go`, `agent.go`, `bazi.go` | ✅ 完成 |
| P1 | `liuyaoRequest.Fixed` 加校验、`validateEmail` 换标准库、`qiming_handler` 命名 struct | `liuyao.go`, `payment.go`, `qiming_handler.go` | ✅ 完成 |
| P2 | Timeset 转换 14 处重复提取为 `timesetOrRespond` helper | `handler_helpers.go` + 6 个 handler 文件 | ✅ 完成 |
| P2 | LLM tool schema 去 lunar 字段 | `openapi.json` x-agent-tools + path schema | ✅ 完成 |
| P2 | shishen→shi_shen / sancai→san_cai 命名统一 | `bazi_liunian.go` 等 5 文件, `qiming_types.go` | ✅ 完成 |
| P1 | error code 残余修正（server_error→internal_error, session_closed/invalid_surname→invalid_request） | `agent.go`, `qiming_handler.go` | ✅ 完成 |
| P1 | huangli bond 错误分类修正（422 validation_error → 400 invalid_request） | `huangli_handler.go` | ✅ 完成 |
| P1 | `agent.Person.Gender` 类型 `string` → `ganzhi.Gender`，消除 tool 层与 HTTP 层类型不一致 | `tools.go`, `tools_bazi.go`, `tools_ziwei.go`, `tools_other.go` | ✅ 完成 |
| P2 | API 文档更新（ziwei 补 gender、bazi 年份修正、error envelope 文档化） | `web/docs/ziwei.md`, `bazi.md`, `reference.md` | ✅ 完成 |
