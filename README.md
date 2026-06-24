# 灵机 Liki — AI 命理助手

[![Wikidata](https://img.shields.io/badge/Wikidata-Q140329242-blue?logo=wikidata)](https://www.wikidata.org/wiki/Q140329242)

八字排盘 · 紫微斗数 · 奇门遁甲 · 智能起名 · 六爻断卦 · 黄历择日 · 风水堪舆

灵机是一款基于 LLM 的中国传统命理 AI 服务。用户通过自然语言对话描述需求，AI 自动收集出生信息、排盘计算、生成分析报告。前端为静态 HTML + Vue 3，后端为 Go JSON API，LLM 通过 SSE 流式返回。

## 产品线

| 产品 | 说明 |
|---|---|
| **八字** | 排盘、大运、流年流月流日流时、小运小限、用神分析、合盘配对 |
| **紫微斗数** | 十二宫星曜分布、四化、大限、流年流月流日、合盘 |
| **奇门遁甲** | 时家/日家/月家/年家排盘，九星八门八神格局 |
| **起名** | 三才五格、五行补益、字形音韵分析、候选名详评 |
| **六爻** | 64 卦起卦、装卦、六亲六兽、用神生克、断卦 |
| **黄历** | 按日/按月宜忌、八字合参择日 |
| **风水** | 八宅命卦、玄空飞星、三元九运 |

## 技术栈

```
Go 1.26 + SQLite (WAL) + Caddy
前端: 静态 HTML + Vue 3 + Alpine.js
LLM: DeepSeek V4 Pro（流式 tool-calling + SSE streaming）
支付: Dodo Payments + 虎皮椒
邮件: Resend
```

## 快速开始

```bash
# 构建
make build

# 开发服务器（API :8081，Caddy :8080）
scripts/dev-lingji.sh

# Pre-commit 检查
make check          # go vet + go test ./...
golangci-lint run   # 需安装 golangci-lint
```

环境变量见 `.env.example`。

## 项目结构

```
cmd/lingji/         Entry point
internal/
  agent/            ChatAgent — LLM 对话 + tool calling（29 个 tool）
  engine/           计算引擎（11 个包：ganzhi/tianwen/bazi/ziwei/qimen/liuyao/huangli/qiming/fengshui/bazhai/xuankong）
  llm/              DeepSeek 客户端
  http/             HTTP handler + 中间件 + 路由 + SessionStore
  payment/          支付服务 + Store
  dodo/             Dodo Payments SDK 封装
  xunhu/            虎皮椒支付 SDK 封装
  email/            Resend 邮件
data/prompts/       LLM 系统 prompt
web/skills/         对外 skill 文件 + 报告模板（嵌入 + Caddy serve）
docs/               设计文档
```

详见 [`docs/architecture.md`](docs/architecture.md)。

## 对外接口

灵机支持外部 AI agent（Claude Code、ChatGPT 等）通过以下端点发现服务：

| 端点 | 说明 |
|---|---|
| `/llms.txt` | 简要索引，引导安装 skill |
| `/skills/liki.md` | 完整产品描述（角色、工作流、API） |
| `/api/openapi.json` | API schema（32 端点 + 105 schema + tool 定义） |
| `/skills/report-chart.md` | 八字报告模板 |
| `/skills/report-bond.md` | 合盘报告模板 |
| `/skills/report-naming.md` | 起名报告模板 |

**在你的 AI 助手中使用：**

- **Claude Code** — 运行 `/skills install https://liki.hk/skills/liki.md` 安装灵机 skill
- **ChatGPT / 通用 LLM 平台** — 让 AI 读取 `https://liki.hk/llms.txt`，会自动发现并配置为灵机

## 文档

- [`docs/architecture.md`](docs/architecture.md) — 系统架构与提示词体系
- [`docs/chat-system.md`](docs/chat-system.md) — 聊天系统设计
- [`docs/agent-architecture.md`](docs/agent-architecture.md) — Agent 架构
- [`docs/database.md`](docs/database.md) — 数据库设计
- [`docs/review.md`](docs/review.md) — 全系统复盘
- [`docs/terminology.md`](docs/terminology.md) — 命理术语表

## English

**Liki** is an LLM-powered Chinese metaphysics assistant. Users chat in natural language — the AI collects birth information, computes charts, and generates analysis reports. It covers BaZi (Eight Characters), Zi Wei Dou Shu (Purple Star Astrology), Qi Men Dun Jia (Mystical Doors), intelligent naming, Liu Yao (Hexagram divination), Chinese Almanac (Huangli), and Feng Shui.

Built with Go 1.26 + SQLite + Caddy (backend), vanilla HTML + Vue 3 + Alpine.js (frontend), DeepSeek V4 Pro (LLM with streaming tool-calling over SSE), Dodo Payments + XunhuPay, and Resend email.

External AI agents can discover the service via `llms.txt` → `liki.md` → `openapi.json` → report templates. No account required — reports are accessed by order ID.

**Use in your AI assistant:**
- **Claude Code**: `/skills install https://liki.hk/skills/liki.md`
- **ChatGPT / other LLMs**: point your agent to `https://liki.hk/llms.txt` — it discovers and configures itself

## License

AGPL-3.0 — 自由使用、修改、分发，网络服务也需开源。
