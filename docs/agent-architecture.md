# Agent 架构文档

## 共享内核

所有术数体系共用的基础，不可分叉：

```
ganzhi/  → 天干地支、五行生克刑冲合害
tianwen/ → 节气、农历、真太阳时、干支历
```

## Engine 层（6 个框架）

```
                    共享内核 (ganzhi + tianwen)
                        │
    ┌───────┬───────┬───┴───┬───────┬───────┐
    ▼       ▼       ▼       ▼       ▼       ▼
┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
│ 八字 ││ 紫微 ││ 奇门 ││ 六爻 ││ 风水 ││ 起名 │
│bazi  ││ziwei ││qimen ││liuyao││feng  ││qiming│
│      ││      ││      ││      ││shui  ││      │
│排盘  ││命盘  ││时盘  ││起卦  ││飞星  ││五行  │
│大运  ││大限  ││日盘  ││装卦  ││挨星  ││三才  │
│用神  ││流年  ││月盘  ││用神  ││旺衰  ││五格  │
│合盘  ││四化  ││年盘  ││月建  ││城门  ││字形  │
│      ││合盘  ││克应  ││应期  ││双星  ││      │
│      ││      ││格局  ││      ││收山  ││      │
│      ││      ││应期  ││      ││      ││      │
└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
```

每个 Engine 遵循原子化原则：单一输入→单一输出，纯 Go 计算，无 I/O 依赖。

## Agent 层

```
ChatAgent (单一实例)
├─ Chat(messages, tools, onEvent, orderCreator, amounts)
│    └─ LLM 对话 → tool 调用 → engine 计算 → 生成 teaser → purchase
└─ GenerateFromData(chartJSON, reportPrompt)
     └─ LLM 生成完整报告
```

当前注册的 tool：`compute_chart`, `compute_bond`, `compute_naming`。后续可加 `compute_qimen`, `compute_liuyao`, `compute_fengshui`。

## 注入策略

Phase 1（当前）：统一注入，单次 `ChatStreamWithTools`，所有 prompt + tool 一次性发给 LLM。
Phase 2（未来 5+ framework）：分阶段注入，按意图分段注入 prompt。
Phase 3（未来跨框架）：Plan-and-Execute，LLM 先出 plan 再逐步调度 framework。

## 加新 Framework 路径

1. `internal/engine/{name}/` — 实现 engine 计算
2. `internal/agent/tools.go` — 注册 tool handler
3. `cmd/lingji/main.go` — 声明 prompt
4. `data/prompts/` — 加解读 prompt
