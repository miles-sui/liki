# 灵机架构设计 · Liki Architecture

## 概述

灵机（Liki）是一个中国命理 AI 服务，提供命理排盘、合盘分析、起名建议、问事卜卦、风水分析。前端为静态 HTML + Vue 3，后端为 Go JSON API，LLM 对话通过 SSE 流式返回，数据库为 SQLite。

## 组件图

```
                        浏览器
                          │
                   ┌──────┴──────┐
                   │   Caddy     │  TLS 终止 · 静态文件 · 反向代理
                   │  :8080      │  /zh/* /en/* → web/
                   │             │  /api/* → Go :8081
                   └──────┬──────┘
                          │
                   ┌──────┴──────┐
                   │  Go Server  │  cmd/lingji
                   │  :8081      │
                   └──────┬──────┘
                          │
      ┌───────────────────┼───────────────────┐
      │                   │                   │
      │  ┌────────────────┴────────────────┐  │
      │  │        HTTP Handlers            │  │  Free API + Agent Chat
      │  │  (package handler)              │  │
      │  └────────────────┬───────────────┘  │
      │                   │                  │
      └──────────┬────────┴────────┬─────────┘
                 │                 │
      ┌──────────┴──────────┐  ┌──┴──────────┐
      │  Agent              │  │ Engine      │  引擎层
      │   Chat + Purchase   │  │ Layer       │  纯 Go 计算
      └──────────┬──────────┘  └──┬──────────┘
                 │                 │
      ┌──────────┼─────────────────┼──────┐
      │          │                 │      │
 ┌────┴────┐ ┌──┴──────┐  ┌──────┴──────┐│
 │ Engine  │ │  LLM    │  │  Payment    ││
 │ bazi    │ │ DeepSeek│  │  Service    ││
 │ ziwei   │ └─────────┘  │  + Store    ││
 │ qimen   │              └──────┬──────┘│
 │ liuyao  │                     │       │
 │ fengshui│                ┌────┴────┐  │
 │ bazhai  │                │ SQLite  │  │
 │ xuankong│                │ orders  │  │
 │ qiming  │                └─────────┘  │
 │ huangli │                            │
 │ tianwen │                            │
 │ ganzhi  │                            │
 └─────────┘                            │
                                        │
 ┌───────────────────────────────────────┘
 │ SessionStore (内存)  Email (Resend)
 └───────────────────────────────────────
```

## Engine 层

11 个包，三层架构：

```
internal/engine/
├── ganzhi/     # 基础: 天干地支、五行、十神（无外部依赖）
├── tianwen/    # 基础: 天文历算（真太阳时、节气、干支历）
├── bazi/       # 命理: 八字排盘+大运+用神+合盘
├── ziwei/      # 命理: 紫微斗数（12宫+14主星+四化+格局+大限流年）
├── qimen/      # 命理/问事: 奇门遁甲（时/日/月/年盘+克应+格局+应期）
├── liuyao/     # 问事: 六爻卜卦（64卦+装卦+用神+月建日建+应期）
├── fengshui/   # 环境: 风水基础（24山+飞星+旺衰+双星加会）
├── bazhai/     # 环境: 八宅风水（命卦+八宅方位+四柱八卦）
├── xuankong/   # 环境: 玄空风水（三元九运+飞星+挨星+旺山旺向+城门诀）
├── huangli/    # 择日: 黄历（宜忌+二十八宿+时辰吉凶）
└── qiming/     # 命名: 起名（三才五格+五行补益+字形分析）
```

所有 Engine 包遵循同一原则：**纯 Go 计算，无 I/O 依赖**。每个函数原子化（单一输入→单一输出），SQL/HTTP/LLM 全部在上层。

### api.go 编排层

每个业务 Engine 包（bazi、ziwei、qimen、liuyao、bazhai、xuankong、huangli、tianwen）暴露 `api.go`：

- **公开入口**：接收 `ChartBase` 或 `SolarTime`，编排小写引擎函数。
- **流运函数**：`ComputeLiuNian(cb ChartBase, year)` / `ComputeLiuYue(cb ChartBase, year, month)` / `ComputeLiuRi(cb ChartBase, year, month, day)` / `ComputeLiuShi(cb ChartBase, year, month, day, hour)` — 统一签名，ChartBase 内部分离 DaYun 等值。
- **HTTP 和 Tool 共用**：同一套函数签名，tool handler 和 HTTP handler 无差异化。

```
api.go (public)                                engine file (private)
────────────────────────────────────           ──────────────────────
ComputeChart(st tianwen.SolarTime, gender)     → computeChart(bz ganzhi.Bazi, g) Chart
ComputeBond(a, b ChartBase) Bond              → computeBond(bzA, bzB ganzhi.Bazi) Bond
ComputeLiuNian(cb ChartBase, year)            → computeLiuNian(bz ganzhi.Bazi, year)
ComputeLiuYue(cb ChartBase, year, month)      → computeLiuYue(bz ganzhi.Bazi, year, month)
ComputeLiuRi(cb ChartBase, year, month, day)  → computeLiuRi(bz ganzhi.Bazi, year, month, day)
ComputeLiuShi(cb ChartBase, y, m, d, hour)    → computeLiuShi(bz ganzhi.Bazi, y, m, d, hour)
```

### 命名约定

- `Compute` 前缀：多步编排（如 `ComputeChart`、`ComputeLiuNian`）
- 无前缀：单公式推导（如 `RiZhu`、`NianZhu`）
- 所有领域概念使用类型化实体，不用裸 `int`
- JSON 字段用拼音 snake_case：`nian`/`yue`/`ri`/`shi`（四柱）、`yongshen`/`dayun`/`riyuan`
- ChartBase 仅 5 字段：`Nian`/`Yue`/`Ri`/`Shi`/`DaYun`，展示类字段（FuYi/TiaoHou/WuxingCount）在 Chart

### ChartBase 与 Chart

```go
type ChartBase struct {
    Nian, Yue, Ri, Shi  zhuInfo    // 四柱
    DaYun                *DaYun     // 大运
}
type Chart struct {
    ChartBase
    SolarTime, ChangSheng, FuYi, TiaoHou, WuxingCount, ...
}
```

ChartBase 被 Bond、流运函数共用。流运不依赖扶抑/调候（非经典用法），故放在 Chart 而非 ChartBase。

### OpenAPI / Agent Schema

工具和 HTTP 的参数定义统一在 `openapi.json`（OpenAPI 3.0）。29 个工具（compute_chart 等）的 JSON Schema 从 OpenAPI 提取，编译时嵌入，运行时传给 LLM 的 tool calling `parameters` 字段。

外部 Agent 通过 `GET /openapi.json` 获取完整 API 定义，`/skills/liki.md` 获取产品行为描述。

## 核心数据流

### Chat 流（Agent 对话）

```
POST /api/agent/chat  {session_id, message, lang}
  │
  └─ ChatAgent.Chat(messages, tools, onEvent, orderCreator, amounts)
       │  单 loop，tools (max 30 rounds, SSE 流式)
       │  工具: get_city_coords, compute_chart, compute_bond, compute_naming, purchase
       │
       ├─ 收集: LLM 追问出生信息
       ├─ 计算: compute_* → engine → LLM 生成 teaser
       ├─ Q&A: 用户追问 (~8 轮)，LLM 引导购买
       └─ purchase: 触发订单创建 → SQLite INSERT (status=pending)
            → SSE done {order_id, amount, product}
```

### Free API 流

```
POST /api/bazi/chart  (或 /api/bazi/bond, /api/qiming/wuge, /api/ziwei/chart 等)
  │
  ├─ 解析请求 → Engine 计算
  └─ 返回 JSON (命理数据，不经过 LLM/支付)
```

## 关键文件布局



## 关键设计决策

### 原子化引擎

所有 Engine 函数遵循正交原则：
- 单一输入 → 单一输出，无副作用
- 每个文件一个领域概念
- 类型安全（无 `map[string]any`）
- 公开函数 = API 契约，私有函数 = 内部实现

### 类型归属

打字系统按领域拆分：
- `llm.Message/Role/ToolCall` — LLM 线格式
- `agent.Product` — 报告产品类型
- `agent.TimePoint` — 出生时间点（RFC3339 公历 + 经度）
- `handler.BirthRequest` — HTTP 契约（统一出生+性别，八字/紫微/八宅复用）
- `ganzhi.Gan/Zhi/Wuxing/ShiShen/Zhu/Bazi` — 干支基础类型
- `tianwen.SolarTime/GregorianTime/LunarTime/Timeset` — 时间类型

所有领域概念用类型化实体传递，禁用裸 `int`。Map key 用 `ganzhi.Zhi` 而非 `int`，函数参数收 `ganzhi.Gan` 而非 `int`。

### JSON 契约

Engine 输出统一 snake_case，所有公开结构体显式声明 `json:"..."` tag。无 CamelCase 裸字段。

Error envelope 标准化：
- 400 `invalid_request` — JSON 解析失败 / 业务参数非法 / 引擎计算失败
- 422 `validation_error` — 结构化校验失败
- 404 `not_found` — 资源不存在
- 413 `too_large` — 请求体过大
- 500 `internal_error` — 服务内部错误

Handler 层提取了 `timesetOrRespond` helper，消除 14 处重复的 `Timeset()` 转换 + 错误响应模式。`decodeAndValidate[T]` 统一 JSON 解码 + 校验流程。

### 单连接 SQLite

`MaxOpenConns=1`，WAL 模式。所有数据库操作串行化。

### LLM 集成

- 模型：DeepSeek V4 Pro
- ChatAgent：流式 `ChatStreamWithTools`，120s 超时
- 单一 Agent 实例，`ReportPrompts` map 存产品报告 prompt

### 会话管理

服务端 Session（内存，30min TTL），`session_id` 存前端 sessionStorage。Phase: `collecting` → `closed`。

## 包依赖关系

```
cmd/lingji
  └─ internal/http         → handler 注册 + 中间件 + SessionStore
       ├─ internal/agent      → ChatAgent
       │    └─ internal/llm   → DeepSeek 客户端
       ├─ internal/engine     → Engine 层 (11 包)
       └─ internal/payment    → 支付服务 + Store
            ├─ internal/dodo   → Dodo Payments SDK
            └─ internal/email  → Resend 邮件客户端

无 orchestrator 包（已并入 agent），无 domain 包（类型归属到各自域）。
```
