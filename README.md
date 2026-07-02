# 灵机 Liki

[![Wikidata](https://img.shields.io/badge/Wikidata-Q140329242-blue?logo=wikidata)](https://www.wikidata.org/wiki/Q140329242)
[![License](https://img.shields.io/badge/license-AGPL--3.0-green)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)

**[liki.hk](https://liki.hk)** · [GitHub](https://github.com/miles-sui/liki) · [X](https://x.com/liki_hk) · [Telegram](https://t.me/liki_naming) · [知乎](https://zhihu.com/people/liki.hk) · [小红书](https://www.xiaohongshu.com/user/profile/liki_hk) · [邮箱](mailto:hi@liki.hk)

> 全栈命理计算引擎，JSON-RPC 调用。驱动 [liki.hk](https://liki.hk) AI 起名顾问。

## AI 起名顾问

[liki.hk](https://liki.hk) — 基于八字用神、五行喜忌、三才五格，AI 取名 Chat 产品。

## 命理引擎

让 AI agent 调用八字、紫微、奇门、六爻、起名、风水、黄历——全部通过 `POST /jsonrpc`，不含 LLM，纯计算。

### 安装 Skill

Agent 发现 `https://liki.hk/llms.txt`，互动后安装：

```
/skills install https://liki.hk/skills/liki.md
```

不安装也可直接调 API，Skill 作用是为 agent 注入灵机角色身份和工作流。

### API 调用

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

### Method 列表（31 个）

**八字 (bazi):** `bazi.chart` `bazi.bond` `bazi.liunian` `bazi.liuyue` `bazi.liuri` `bazi.liushi` `bazi.xiaoyun` `bazi.xiaoxian`

**紫微斗数 (ziwei):** `ziwei.chart` `ziwei.daxian` `ziwei.liunian` `ziwei.liuyue` `ziwei.liuri` `ziwei.bond`

**奇门遁甲 (qimen):** `qimen.pan`

**起名 (qiming):** `qiming.sancai` `qiming.chars` `qiming.compose` `qiming.evaluate`

**六爻 (liuyao):** `liuyao.qigua` `liuyao.chart`

**黄历 (huangli):** `huangli.date` `huangli.month` `huangli.bond.date` `huangli.bond.month`

**八宅 (bazhai):** `bazhai.minggua` `bazhai.chart`

**玄空 (xuankong):** `xuankong.sanyuan` `xuankong.chart`

**元数据:** `rpc.discover` — 返回 OpenRPC 1.4.1 文档，含所有 method 的 params/result JSON Schema

## 技术栈

Go 1.26 + SQLite (WAL) + Caddy · 前端 HTML + Vue 3 · DeepSeek V4 Pro (流式 tool-calling + SSE) · Dodo Payments + 虎皮椒 · Resend

## 项目结构

```
cmd/liki/           Entry point
internal/
  agent/              NamingChatAgent (8 tool) + RPCRegistry (31 method)
  engine/             计算引擎 — 11 个包，纯 Go，无 I/O 依赖
  http/               Handler + 中间件 + 路由 + JWT
  llm/                DeepSeek 客户端
  payment/            支付服务 + Store (SQLite)
  dodo/ xunhu/        支付渠道
  email/ i18n/        基础设施
web/                前端 + i18n + e2e + wiki 静态站点
scripts/            构建 / 部署 / 测试
deploy/             Docker + Caddy 配置
docs/               设计文档
```

## License

AGPL-3.0
