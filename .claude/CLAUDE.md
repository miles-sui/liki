# Liki (灵机)

Go 1.26 + SQLite + Caddy。静态 HTML + Alpine.js 前端，DeepSeek LLM，Dodo Payments + 虎皮椒。
模型用 DeepSeek V4 Pro（当前最新），超时 120s，全流式 tool-calling + SSE streaming。

## Commands

```
# 构建
make build

# 开发服务器 (API :8081, Caddy :8080)
scripts/dev-lingji.sh

# Pre-commit — 按顺序跑
make check                         # go vet + go test ./...
golangci-lint run                  # 本地需装 golangci-lint

# 部署后测试 — 四层正交：API → 页面 → 渲染 → 流程
make test-api URL=http://localhost:8080       # API 层：57 项 HTTP 状态码 + 响应体字段
make test-pages URL=http://localhost:8080     # 页面层：所有语言×页面可访问，无 console error
make test-render URL=http://localhost:8080    # 渲染层：框架渲染错误（裸模板/损坏图片）
make test-flows URL=http://localhost:8080     # 流程层：用户操作链（chat→支付→报告），不含 pages/render
make test-deploy URL=http://localhost:8080    # 一键：四层按序全跑，任一步失败即停

# 部署
make deploy       # 两台
make deploy us    # 仅海外
make deploy cn    # 仅国内
```

## Architecture

```
doc.go              //go:embed openapi.json → doc.OpenAPIJSON
                    //go:embed data/prompts/chat.txt → doc.ChatPrompt
                    //go:embed web/skills/report-chart.md → doc.ChartReportPrompt
                    //go:embed web/skills/report-bond.md → doc.BondReportPrompt
                    //go:embed web/skills/report-naming.md → doc.NamingReportPrompt
cmd/lingji/         Entry point
data/prompts/       LLM 系统 prompt（嵌入，不对外）
  chat.txt          统一系统 prompt：收集 + 3 产品 teaser + Q&A 引导
web/skills/         报告模板（嵌入 + 对外公开）
  liki.md           产品 skill 文件（角色、工作流、API 调用）
  report-chart.md   八字完整报告模板
  report-bond.md    合盘完整报告模板
  report-naming.md  起名完整报告模板
internal/
  agent/            ChatAgent（Chat + GenerateFromData）。单 Agent，29 个 tool
  engine/           Gan-Zhi/Tianwen/BaZi/HuangLi/Fengshui/Qiming — 计算引擎
  payment/          支付服务（checkout/webhook/download/report）+ Store
  llm/              DeepSeek 客户端 + tool schema JSON
  dodo/             Dodo Payments SDK 封装
  xunhu/            虎皮椒支付 SDK 封装
  email/            Resend 邮件客户端
  http/             Handler (package handler) + 中间件 + 路由 + 编排 + SessionStore（含 Free API）
  i18n/             国际化工具

### 模块边界

| 模块 | 职责 | 依赖 | 禁止依赖 |
|---|---|---|---|
| ChatAgent | LLM 对话 + tool calling + purchase 处理 | LLMClient, ToolRegistry 接口 | engine, payment |
| Handler | 薄层: 参数绑定 + SSE 流 + 响应 | 以上所有 | 引擎逻辑, LLM 逻辑 |

分层原则：Handler 薄（只做参数绑定+响应），逻辑在 service 层。

## Conventions

### Go
- context.Context 作为所有 I/O 函数的第一参数。
- 错误用 fmt.Errorf("doing X: %w", err) 包装，不裸 return err。
- 不写 init()（注册驱动/flag 除外）。
- 不启动无生命周期的 goroutine（需有 cancelable context 控制）。
- 不用 interface{} / any 除非必要，优先泛型或具体类型。
- 导出符号必须有 doc comment，以符号名开头。
- Handler 薄：只做参数绑定+响应，逻辑在 service 层。

### 测试
- 表驱动，行命名用 name 字段，helper 调 t.Helper()。
- Integration 测试用 //go:build integration，go test -tags integration ./...。

### Git
- Commit 用英文，Conventional Commits: feat:/fix:/chore:/refactor:。
- PR 前跑 gofmt → lint → test -race（顺序执行，前一步不过不跑后一步）。

### API
- Envelope: {"data":{...}} (单条)｜{"data":{"items":[...],"total":N}} (列表)｜{"error":{"code":"...","message":"..."}} (错误)
- 路由不加 /v1/ 前缀。Caddy 处理 TLS + 静态文件 + 反向代理。

### 数据
- SQLite WAL 模式，单连接（MaxOpenConns=1）。

### 支付
- Dodo Payments（国际卡）+ 虎皮椒（微信/支付宝），双 provider 共存。
- 前端用户自选支付方式，相同数字金额不同币种（¥9.90 vs $9.90）。
- 订单表 `currency` 字段区分（CNY/USD），`provider` 字段记录支付通道。
- 不要跳过 webhook 签名验证。

### LLM
- 当前模型: DeepSeek V4 Pro。选型标准: 最新旗舰、支持 tool-calling + SSE streaming。
- 单一 Agent 架构: 1 个 ChatAgent，统一 prompt + 29 个 tool（八字/紫微/起名/奇门/六爻/风水/黄历 + query_city）。
  - tools 注册在 `NewChatToolRegistry()`，共 29 个，覆盖全部产品线
  - purchase 由 ChatAgent 硬编码处理，不在 tool registry 中
  - 流程: Chat（收集 → compute_* → teaser → Q&A(~8轮, 最多推荐购买 3 次) → purchase → done）
  - compute 工具返回 `{"_product":"...","data":{...}}`, LLM 根据 _product 选择报告格式
  - 购买由 LLM 自然引导，purchase tool 触发订单创建。
- Tool schema 从 `doc.OpenAPIJSON`（openapi.json）提取，`openapiParams()` 解析 `x-agent-tools` 和 path schema。
- 公开索引: `web/llms.txt`（Caddy 静态 serve，llms.txt spec 格式）。
- Go 代码中无 LLM prompt，只有 UI 进度文案。
- 多语言：前端 `lang` 字段传入 → `langToLocale()` 映射（zh→zh-Hans, hk→zh-Hant, en→en）→ `strings.ReplaceAll({locale})` 替换 prompt 中的 `{locale}` 占位符。报告页暂无语言选择，默认 zh-Hans。

### 流程
- **Form 流**: POST /api/bazi/chart (或 /api/bazi/bond, /api/qiming/generate) → engine+LLM → 预览 → 支付 → 报告页
- **Chat 流**: POST /api/agent/chat → SSE 通道 → ChatAgent.Chat（单流：收集 → compute_* → teaser → Q&A → purchase tool → 创建订单 → done 事件）→ 前端 buy card → 支付 → webhook 触发 GenerateFromData（完整报告）→ GET /api/reports/{id} 返回。购买引导由 LLM 自然完成，purchase tool 触发订单创建。

## Don't

- 不要对外暴露 API 8080 端口 → 外部访问经 Caddy:443 反向代理。
- 不要用 PATH 上的 grep（ugrep，$() 捕获有 bug） → 用 /bin/grep。
- 健康检查不要直连 API → 走 curl --resolve 完整 HTTPS 链路。
- 不要跳过 webhook 签名验证 → 每个 webhook 必验。
- 不要加用户系统/注册 → 产品定位无账号体系。
- 不要改项目结构不更新此文件 → 结构变化同步 CLAUDE.md。

## Pitfalls

| 陷阱 | 解法 |
|---|---|
| expose: ["8080"] 不对外发布 | 健康检查走 Caddy --resolve |
| Caddy 启动后 TLS 证书加载需时间 | 健康检查重试 6×5s |
| DOMAIN 默认值 compose 和 script 不同步 | 两边一致用 liki.hk |
| COPY vendor/ + COPY . . vendor 重复 | vendor 不能进 .dockerignore |
| Caddy depends_on lingji 等 30s healthcheck | interval: 10s + start_period: 5s |
| tar 追加 .env 用 gunzip→tar rf→gzip 三步 | 初始 tar 直接包含 .env |
| all 目标 image save 两次 | docker save 提到 deploy() 循环外 |

## Domain docs

- docs/architecture.md — 系统架构
- docs/chat-system.md — 聊天系统设计
- docs/database.md — 数据库设计
- docs/review.md — 全系统复盘
- docs/terminology.md — 命理术语表
- web/llms.txt — 公开 AI agent 服务索引（llms.txt spec）
- web/skills/liki.md — 对外 skill 文件（角色、工作流、API 调用、报告模板）
- web/skills/report-chart.md — 八字报告模板（GenerateFromData 用）
- web/skills/report-bond.md — 合盘报告模板（GenerateFromData 用）
- web/skills/report-naming.md — 起名报告模板（GenerateFromData 用）
- data/prompts/chat.txt — 统一系统 prompt（收集 + 3 产品 teaser + Q&A + 购买引导）
