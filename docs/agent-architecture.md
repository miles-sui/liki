# Agent 架构文档

## 共享内核

所有术数体系共用的基础，不可分叉：

```
ganzhi/  → 天干地支、五行生克刑冲合害
tianwen/ → 节气、农历、真太阳时、干支历
```

## Engine 层（11 个包）

```
                  共享内核 (ganzhi + tianwen)
                      │
  ┌───────┬───────┬───┴───┬───────┬───────┬───────┐
  ▼       ▼       ▼       ▼       ▼       ▼       ▼
┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
│ 八字 ││ 紫微 ││ 奇门 ││ 六爻 ││ 风水 ││ 择日 ││ 起名 │
│bazi  ││ziwei ││qimen ││liuyao││feng  ││huang ││qiming│
│      ││      ││      ││      ││shui  ││li    ││      │
│排盘  ││命盘  ││时盘  ││起卦  ││24山  ││宜忌  ││五行  │
│大运  ││大限  ││日盘  ││装卦  ││命卦  ││建除  ││三才  │
│用神  ││流年  ││月盘  ││用神  ││飞星  ││星宿  ││五格  │
│合盘  ││四化  ││年盘  ││月建  ││挨星  ││节气  ││字形  │
│      ││合盘  ││克应  ││应期  ││旺衰  ││人元  ││音韵  │
│      ││      ││格局  ││      ││城门  ││      ││      │
│      ││      ││应期  ││      ││      ││      ││      │
└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
  bazhai/  xuankong/
```

每个 Engine 遵循原子化原则：单一输入→单一输出，纯 Go 计算，无 I/O 依赖。

## Agent 层

```
NamingChatAgent (单一实例)
  │
  ├─ NamingChat(ctx, locale, messages, onEvent)
  │    └─ SSE 流式对话 → 8 个 tool → engine 计算 → 磋商起名
  │         │
  │         └─ 用户要求时，LLM 直接在对话中输出 markdown 报告
  │              handler 识别（IsNamingReport）→ 存 llm_json → report_ready
  │
  └─ prompt 缓存：sync.Map（locale → 编译后 system prompt），避免重复 ReplaceAll
```

### Tool Calling 流程

```
用户消息
  │
  ▼
NamingChat(ctx, locale, messages, onEvent)
  │
  ├─ ensureNamingPrompt() → system prompt 注入（若 messages[0] 不是 system）
  │
  ├─ ChatStreamWithTools(messages, tools) → LLM
  │    │
  │    ├─ text_delta  → onEvent(EventTextDelta) → SSE 推送
  │    ├─ thinking    → onEvent(EventThinking)  → SSE 推送
  │    ├─ tool_call   → Execute(tool, args)
  │    │    │
  │    │    ├─ query_city            → 城市→经纬度+时区
  │    │    ├─ compute_time          → 公历+经纬度→Timeset（真太阳时）
  │    │    ├─ compute_chart         → 八字排盘
  │    │    ├─ compute_ziwei         → 紫微斗数
  │    │    ├─ compute_naming_wuge   → 三才五格
  │    │    ├─ compute_naming_compose → 候选名组合
  │    │    ├─ compute_naming_detail → 单名详析
  │    │    └─ compute_naming_evaluate → 候选名评估
  │    │
  │    └─ tool_result → LLM 继续推理
  │
  └─ 返回 messages（含所有轮次），handler 持久化 + 检测报告
```

### 消息持久化

```
namingHandler 每次 POST：

  LoadChatHistory(order_id)  ──→  DB 读取
  [system] + history + [user]  ──→  LLM
  CreateChatMessage(user)     ──→  DB 写入（立即）
  SSE streaming...
  BatchCreateChatMessages(new) ──→  DB 写入（流结束后）

  检测到报告 → UpdateLlmJSON → report_ready 事件 → 前端跳转
```

## 工具注册

### NamingChatAgent（8 tools）

`NewNamingToolRegistry()` 注册：

| 工具 | 用途 |
|------|------|
| query_city | 城市名→经纬度+时区 |
| compute_time | 公历+经纬度→Timeset（真太阳时、农历） |
| compute_chart | Timeset+性别→八字排盘 |
| compute_ziwei | Timeset+性别→紫微斗数命盘 |
| compute_naming_wuge | 姓氏→三才五格分析 |
| compute_naming_compose | 八字+姓氏→候选名列表 |
| compute_naming_detail | 单个名字→字形音韵详析 |
| compute_naming_evaluate | 候选名列表→综合评分排序 |

### RPCRegistry（29 tools，外部 API）

`NewRPCRegistry()` 注册全部引擎能力，供 JSON-RPC API（`POST /jsonrpc`，29 个 method）使用。包含八字、紫微、奇门、六爻、风水、黄历等全部计算工具。与 NamingChatAgent 的工具完全独立。

## 提示词组织

| Prompt | 来源 | 注入点 | 用途 |
|---|---|---|---|
| 系统 prompt | `agent.NamingPrompt`（`internal/agent/naming.txt`） | `ensureNamingPrompt()` | 角色定义、工具使用规则、报告格式约定 |
| 工具 schema | `agent.ToolsJSON`（`internal/agent/tools.json`） | `NewNamingToolRegistry()` | LLM tool calling 的 `parameters` 字段 |

**系统 prompt** 在每次 NamingChat 调用时作为 messages[0] 注入。启动时预编译（`{locale}` 占位符替换），按 locale 缓存在 `sync.Map`。

**无需报告模板**。报告由 LLM 在对话中直接输出 markdown，handler 识别 `# 起名报告` 标题后存入 `llm_json`。起名报告模板（`web/skills/report-naming.md`）保留为对外公开参考文档。

## 启动验证

```
main() → ValidateTools()
           │
           └─ sync.Once 触发 tools.json 解析
              格式错误 → 启动失败 fast fail
              解析成功 → 缓存结果
```

## 加新 Tool 路径

1. `internal/agent/tools_qiming.go`（或相应文件）— 注册 tool handler
2. `internal/agent/tools.json` `x-agent-tools` — 加 tool 的 JSON Schema 定义
3. `internal/agent/tools.go` `NewNamingToolRegistry()` — 调 `registerTool()`
