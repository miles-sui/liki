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
ChatAgent (单一实例)
├─ Chat(ctx, locale, messages, onEvent, orderCreator, amounts)
│    └─ LLM 对话 → tool 调用 → engine 计算 → 生成 teaser → purchase
└─ GenerateFromData(ctx, locale, product, chartJSON, onEvent)
     └─ LLM 生成完整报告（支付后 webhook 触发）
```

当前注册了 29 个 tool（八字 8 + 紫微 6 + 起名 4 + 奇门 1 + 八宅 2 + 玄空 2 + 六爻 1 + 黄历 4 + `query_city`），`purchase` 由 ChatAgent 硬编码处理。详见 `NewChatToolRegistry()`。

## 提示词组织

Agent 使用两种 prompt，来源和用途不同：

| Prompt | 来源 | 注入点 | 用途 |
|---|---|---|---|
| 系统 prompt | `doc.ChatPrompt`（`data/prompts/chat.txt`） | `ChatAgent.ensureSystemPrompt()` | 定义产品检测、参数收集规则、对话行为 |
| 报告模板 | `doc.ChartReportPrompt` 等（`web/skills/report-*.md`） | `GenerateFromData()` | 完整报告格式：数据结构、领域知识、章节规范 |

**系统 prompt** 在每次 Chat 调用时作为 messages[0] 注入，与对话历史一起发给 LLM。它只定义对话行为，不包含报告格式——teaser 报告由 LLM 自由生成简短摘要。

**报告模板** 仅在支付完成后使用。`GenerateFromData()` 将 engine 计算结果的 JSON 与对应产品的报告模板拼接，一次性发给 LLM 生成完整报告。模板中的数据结构必须严格对齐 engine 输出字段。

两种 prompt 分离的理由：
- 系统 prompt 短（~6KB），每次都发，需精炼
- 报告模板长（~10-20KB），只在最后用一次，可详尽
- 报告模板同时对外公开（`/skills/report-*.md`），系统 prompt 不对外

## 注入策略

当前（Phase 1）：统一注入，单次 `ChatStreamWithTools`，所有 prompt + tool 一次性发给 LLM。
未来（Phase 2）：多框架时考虑分阶段注入，按意图分段注入 prompt。

## 加新 Framework 路径

1. `internal/engine/{name}/` — 实现 engine 计算
2. `internal/agent/tools.go` — 注册 tool handler
3. `openapi.json` `x-agent-tools` — 加 tool schema JSON（`openapiParams()` 解析后注册到 `NewChatToolRegistry()`）
4. `web/skills/` — 加报告模板 prompt（如需要），同时嵌入 + 对外公开
