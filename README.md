# 灵机 Liki — AI 起名顾问 · AI 命理引擎

[![Wikidata](https://img.shields.io/badge/Wikidata-Q140329242-blue?logo=wikidata)](https://www.wikidata.org/wiki/Q140329242)
[![License](https://img.shields.io/badge/license-AGPL--3.0-green)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)

**[liki.hk](https://liki.hk)** · [GitHub](https://github.com/miles-sui/liki) · [X](https://x.com/liki_hk) · [Telegram](https://t.me/liki_naming) · [知乎](https://zhihu.com/people/liki.hk) · [小红书](https://www.xiaohongshu.com/user/profile/liki_hk) · [邮箱](mailto:hi@liki.hk)

> 灵机（Liki）融合人工智能与中国传统命理学的AI起名顾问，也为 AI agent 提供可直接安装的命理计算 Skill 与 JSON-RPC 命理计算引擎。

灵机（Liki）提供两件事：

1. **AI 起名顾问** — 付费 Chat 产品，基于八字用神、五行喜忌、三才五格，推荐中文名字
2. **AI 命理引擎** — 免费 JSON-RPC API，供 AI agent 与开发者调用（八字、紫微、奇门、六爻、风水、黄历）

## 外部 Agent 使用指南

外部 AI agent（Claude Code、ChatGPT、Cursor 等）可通过 JSON-RPC API 直接调用命理计算，无需 LLM 中间层。

### 入口

| 资源 | 地址 | 用途 |
|---|---|---|
| Product skill | `https://liki.hk/skills/liki.md` | 角色定义、工作流、参数收集规则 |
| JSON-RPC | `POST https://liki.hk/jsonrpc` | 所有计算入口 |
| Method 发现 | 调 `rpc.discover` | 获取全部可用 method 及 JSON Schema |
| 报告模板 | `https://liki.hk/skills/report-*.md` | 八字/合盘/起名/紫微/八宅/玄空 |

### 调用方式

```
POST /jsonrpc
Content-Type: application/json

{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "bazi.chart",
  "params": {
    "birth": {
      "time": "2000-06-15T12:00:00+08:00",
      "longitude": 116.4
    },
    "gender": "male"
  }
}
```

成功响应：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "_product": "chart",
    "data": { ... }
  }
}
```

错误响应：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32000,
    "message": "参数校验: ..."
  }
}
```

### Method 列表（29 个）

**八字 (bazi):** `bazi.chart` `bazi.bond` `bazi.liunian` `bazi.liuyue` `bazi.liuri` `bazi.liushi` `bazi.xiaoyun` `bazi.xiaoxian`

**紫微斗数 (ziwei):** `ziwei.chart` `ziwei.daxian` `ziwei.liunian` `ziwei.liuyue` `ziwei.liuri` `ziwei.bond`

**奇门遁甲 (qimen):** `qimen.pan`

**起名 (qiming):** `qiming.wuge` `qiming.compose` `qiming.detail` `qiming.evaluate`

**六爻 (liuyao):** `liuyao.chart`

**黄历 (huangli):** `huangli.date` `huangli.month` `huangli.bond.date` `huangli.bond.month`

**八宅 (bazhai):** `bazhai.minggua` `bazhai.chart`

**玄空 (xuankong):** `xuankong.sanyuan` `xuankong.chart`

**元数据:** `rpc.discover` — 返回 OpenRPC 1.4.1 文档，含所有 method 的 params/result JSON Schema

### 安装到 AI 助手

- **Claude Code** — `/skills install https://liki.hk/skills/liki.md`
- **ChatGPT / 通用 LLM** — 让 AI 读取 `https://liki.hk/llms.txt`，自动发现并配置


## 技术栈

Go 1.26 + SQLite (WAL) + Caddy · 前端 HTML + Vue 3 · DeepSeek V4 Pro (流式 tool-calling + SSE) · Dodo Payments + 虎皮椒 · Resend

## 快速开始

```bash
make build                          # 构建
scripts/dev-liki.sh               # 开发服务器 (API :8081, Caddy :8080)
make check                          # golangci-lint + go vet + go test -race ./...
make test-deploy URL=http://localhost:8080  # 部署后四层测试
```

环境变量见 `.env.example`。

## 项目结构

```
cmd/liki/         Entry point
internal/
  agent/            NamingChatAgent (8 tool) + RPCRegistry (29 method)
  engine/           计算引擎 — 11 个包，纯 Go 计算，无 I/O 依赖
  llm/              DeepSeek 客户端
  http/             Handler + 中间件 + 路由 + JWT auth
  payment/          支付服务 + Store (SQLite)
data/prompts/       LLM 系统 prompt (内嵌)
web/skills/         对外 skill + 报告模板
docs/               设计文档
```

## License

AGPL-3.0
