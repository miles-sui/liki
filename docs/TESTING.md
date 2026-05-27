# Testing — 测试策略与方案

> 测试分层、正交解耦、用例清单。

---

## 1. 测试分层（正交）

```
            无服务器                              有服务器
      ┌──────────────┐              ┌──────────────────┐
      │ Go vet       │              │ smoke.sh (88)    │
      │ Go test unit │              │ E2E Playwright   │
      │ Go test integ│              └──────────────────┘
      │ frontend-src │
      └──────────────┘
```

**分工原则**：

| 层 | 测什么 | 不测什么 |
|----|--------|---------|
| httptest (`server_test.go` + `application/*/service_test.go`) | 端点级：handler 逻辑、错误映射、envelope 格式、middleware。隔离测试，内存 SQLite。 | Caddy 路由、真实文件系统、跨端点有状态流程 |
| smoke.sh | 旅程级：全栈用户旅程（Caddy + Go + SQLite + 文件系统）。跨端点有状态流程。 | 单端点行为（httptest 已覆盖）、单端点错误分支、响应形状细节 |

**原则**：

- 不依赖服务器的测试可独立运行，秒级反馈
- 依赖服务器的测试由 `test-all.sh` 统一编排，不内嵌服务启停
- 每个测试独立运行，不依赖其他测试的副作用
- 一个 test function 只验证一个行为

---

## 2. 分层说明

### 2.1 Go 单元测试

纯函数测试，零依赖。覆盖干支历法、天文计算、人格引擎、命理计算。

**运行**：`go test ./internal/ganzhi/ ./internal/tianwen/ ./internal/25types/ ./internal/httputil/ ./internal/mingli/bazi/ ./internal/mingli/http/`

### 2.2 Go 集成测试

`httptest.NewServer` + 内存 SQLite。覆盖所有 API 端点，含认证、错误码、响应形状。App 层用手写 stub 覆盖业务规则。

**运行**：`go test ./internal/app/sqlite/ ./internal/app/http/ ./internal/app/application/...`

### 2.3 前端源码契约 (test-frontend-src)

静态分析，**不依赖服务器**。覆盖 JS 命名对象处理、HTML 模板模式、dist 构建验证、share card 函数与翻译键。

Phase 1: JS 源码模式 (namedToArray/computeBondShapes)
Phase 2: HTML 模板模式 (x-data 安全、locale 前缀)
Phase 3: dist/ 构建产物验证
Phase 4: share card & radar 函数/翻译键（从 smoke.sh 迁移，无需服务器）

namedToArray/deviationToProportion 数学逻辑由 Go 单元测试以编译期类型安全覆盖，JS 层配合 grep 验证函数存在与调用。

**脚本**：`scripts/test-frontend-src.sh`（39 项检查）
**运行**：`scripts/test-frontend-src.sh`

### 2.4 Smoke — 用户旅程 (HTTP)

curl + jq 旅程级冒烟测试，覆盖全栈用户旅程（Caddy + Go + SQLite + 文件系统）。针对运行中的服务。

端点级行为（handler 逻辑、错误映射、envelope 格式）由 httptest 隔离覆盖，smoke 聚焦跨端点有状态流程：

| Phase | 内容 | 检查数 |
|-------|------|--------|
| 0 | 静态页面 (Caddy 专有) | ~40 |
| 1 | 匿名评估旅程 | ~5 |
| 2 | 完整认证旅程 | ~11 |
| 3 | 他评旅程 | ~12 |
| 4 | match & bond 旅程 | ~10 |
| 5 | 隐私流程 | ~3 |
| 6 | 删号旅程 | ~3 |
| 7 | i18n / 错误页 | ~2 |
| 8 | 限流（仅生产） | ~1 |

**脚本**：`scripts/smoke.sh`（88 项检查）
**运行**：`scripts/smoke.sh [base_url]`

### 2.5 E2E (Playwright)

浏览器测试。仅覆盖关键用户旅程的 happy path。

**文件**：`web/e2e/journeys/`（16 spec，54 用例）
**Page Objects**：`web/e2e/pages/`（assess, bonds, match-landing, profile）
**运行**：`cd web && npm run test:e2e`

---

## 3. 运行方式

### 三种环境

```
# 裸服务（dev-local.sh 或手动）
scripts/dev-local.sh               # 终端1：启动
scripts/test-all.sh                # 终端2：测试

# 本地 Docker
docker compose -f deploy/docker-compose.yml up -d
scripts/test-all.sh
docker compose -f deploy/docker-compose.yml down

# 生产（部署后旅程验证）
scripts/smoke.sh https://25types.com
```

### Makefile 入口

| target | 依赖服务器 | 内容 |
|--------|-----------|------|
| `make check` | 否 | vet + test + test-frontend-src |
| `make test` | 否 | test-unit + test-integration |
| `make test-all` | **是** | test-all.sh（Go + frontend-src + smoke 旅程） |
| `make test-all-e2e` | **是** | test-all.sh --e2e（以上 + Playwright） |

### Shell 命令

```bash
make check                           # CI — 无服务器，快速反馈
make test-all                        # 全量 HTTP（需服务器运行）
make test-all-e2e                    # 全量 + 浏览器（需服务器运行）

# 单脚本
scripts/test-frontend-src.sh         # 前端源码分析
scripts/smoke.sh                     # 本地旅程冒烟（需服务器运行）
scripts/smoke.sh https://25types.com # 部署后旅程冒烟（生产验证）

# E2E 单文件
npx playwright test --grep "match link"
```

---

## 4. cos(d) 单纯形方案

分类算法已从 L2 sphere 迁移至 cos(d) simplex。详情见 engine 包及对应 ADR。

**已实现测试 (59 用例 in engine_test.go + engine_cos_test.go)**

| 类别 | 文件 | 用例数 | 覆盖 |
|------|------|--------|------|
| Deviation / Proportion / Identity | `engine_test.go` | 39 | d 计算、p 计算、cos(d) 分类、ReLU |
| cos(d) 分类 + 原型验证 | `engine_cos_test.go` | 8 + 3 | 25 原型自分类、纯型/复合型方向、Σd=0 |
| Concord | `engine_cos_test.go` | 7 | 生克关系、ε 边界、交换性 |
| Flow | `engine_cos_test.go` | 1 | 生克方向判定 |
| ComputeBond | `engine_test.go` | (部分) | ReLU(d) + Sᵀ/Cᵀ，不受 cos(d) 迁移影响 |

---

## 5. 用户旅程覆盖矩阵

| 旅程 | E2E 覆盖 | 状态 |
|------|----------|------|
| 注册/登录/登出 | account.spec.ts, zh-CN-auth.spec.ts | ✅ |
| 邮箱验证 | email-verification.spec.ts | ✅ |
| 密码找回 | forgot-password.spec.ts | ✅ |
| 注销账户 | account-delete.spec.ts | ⚠️ 间歇失败 |
| 完整评估 (en/zh) | assess.spec.ts | ✅ |
| 匿名评估认领 | anonymous-claim.spec.ts | ✅ |
| 类型画廊/详情 | types-explore.spec.ts | ✅ |
| 首页浏览 | landing.spec.ts | ✅ |
| Profile 设置 | profile-settings.spec.ts | ✅ |
| Peer review | peer-review.spec.ts | ✅ |
| Match link CRUD | match-link-crud.spec.ts | ✅ |
| Match bond 匿名 | match-bond.spec.ts | ✅ |
| Instant bond | instant-bond.spec.ts | ✅ |
| Bonds 画廊 | bonds-gallery.spec.ts | ✅ |
| Donate | donate.spec.ts | ✅ |

---

## 6. 有意识不做的事

| 项目 | 理由 |
|------|------|
| 视觉回归 | Bond 图表是 ECharts canvas，替代：Go 测试验证数值 |
| Accessibility | 无合规要求 |
| 性能/负载 | 单用户平台，SQLite 单连接 |
| E2E 中测邮件 | 邮件是 infra 职责，Go stub 已覆盖 |
| 测试数据工厂 | 当前数据模型简单，API helpers 足够 |

---

## 7. 测试环境

```
                  localhost:8080
                       │
                  ┌────▼────┐
                  │  Caddy  │  TLS + 反向代理 + 静态文件
                  └──┬───┬──┘
             /api/* │   │ /content/* /en/* /zh-CN/*
                    │   │
            ┌───────▼┐ ┌▼──────────┐
            │ Go API │ │ File Server│
            │ :8081  │ │ (Caddy)    │
            └───┬────┘ └───────────┘
                │
        ┌───────▼────────┐
        │  SQLite (data/)│
        └────────────────┘
```

### CI 流程

```
go-test job:
  go vet ./...
  go test ./internal/...
  coverage check

check job (needs go-test):
  make check

e2e job (needs go-test + docker compose up):
  scripts/test-all.sh --e2e
```
