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

当前注册的 tool：`get_city_coords`, `compute_chart`, `compute_bond`, `compute_naming`, `purchase`。

## 注入策略

当前（Phase 1）：统一注入，单次 `ChatStreamWithTools`，所有 prompt + tool 一次性发给 LLM。
未来（Phase 2）：多框架时考虑分阶段注入，按意图分段注入 prompt。

## 加新 Framework 路径

1. `internal/engine/{name}/` — 实现 engine 计算
2. `internal/agent/tools.go` — 注册 tool handler
3. `internal/llm/data/tools/` — 加 tool schema JSON
4. `web/skills/` — 加报告模板 prompt（如需要），同时嵌入 + 对外公开
