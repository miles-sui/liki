# 25types — 架构

> **六个聚合，一张图。共享内核。ADR。文档入口。**

---

## 1. 领域概览

| 聚合 | 类型 | 说明 |
|------|------|------|
| **User** | 聚合根 | 注册用户。owns Assessment, ReviewLink；participates MatchRequest。有 IsDeactivated/CanReactivate 行为 |
| **Assessment** | 实体 | 一次测量事件。不可变。包含 identity_id 和原始 answers/profile JSON |
| **PersonalityProfile** | 值对象 | 用户当前人格状态（d 向量 + p 向量 + identity）。从 Assessment 派生但概念独立 |
| **ReviewLink** | 聚合根 | 他评邀请链接。可软删除。token 唯一。30 天有效 |
| **MatchRequest** | 聚合根 | 匹配请求。状态机：pending → accepted/declined。有 IsTerminal 行为 |
| **Bond** | 值对象 | 两人之间的关系动力学（delta_a / delta_b 双向）。engine 计算，domain 建模 |
| **Flow** | 值对象 | 用户人格的季节性时间投影。按月为单位 |
| **MingliChart** | 值对象 | 八字四柱排盘（年/月/日/时柱 + 藏干 + 十神 + 纳音 + 十二长生 + 大运）。纯计算，不建表 |
| **MingliMatch** | 值对象 | 合八字匹配（七维评分 + 加权总分 + 等级判定）。纯计算，不建表 |
| **Content** | 静态数据 | 元素色值、翻译、类型文案在前端 JSON/YAML。原型向量、节气月向量由后端 engine 持有。评估题库由后端 API serve。八字解读模板在前端 YAML。 |

### 值对象

| 值对象 | 位置 | 说明 |
|--------|------|------|
| **Deviation** | engine | d 向量，Σ=0。类型安全，有 Relu/Add/Sub/Dot/ToProportion 方法 |
| **Proportion** | engine | p 向量，Σ=1。仅用于展示。有 ToDeviation 方法 |
| **AssessmentType** | domain | "self" / "peer" 枚举 |
| **MatchStatus** | domain | "pending" / "accepted" / "declined" 状态机。有 IsTerminal() |
| **Stem / Branch / Element** | engine | 天干(1-10)、地支(1-12)、五行(1-5) 枚举。八字基础类型 |
| **Pillar** | engine | 一柱 = Stem + Branch。不可变 |
| **BirthInfo** | domain | 出生时间 + 地点输入。八字排盘与合八字的入参 |

### 上下文关系

```
Auth ──protects──> Users, Assessment, Reviews, Matches, Flow
Assessment ──produces──> PersonalityProfile (d,p,identity)
PersonalityProfile ──drives──> Bond (between two users), Flow (seasonal)
Reviews ──references──> Assessment (peer aggregation)
Matches ──depends──> PersonalityProfile (both users' d for Bond)
Flow ──depends──> PersonalityProfile (user's d)
BaZi ──depends──> BirthInfo (input), Solar Terms (节气判定) — 纯计算，不依赖任何聚合
BaZi Match ──depends──> two BaZiCharts — 纯计算，不依赖用户状态
Payments ──depends──> Users (user_subscriptions: passport, plan, bond_count)
PassportChecker ──gate──> Bond (first free, subsequent paid), Flow (current vs yearly)
Content ──no dependencies── (static, Caddy-served)
```

### 依赖方向

```
Content     ← 零依赖，Caddy 直出
  ↓
Auth        ← 依赖 User (name/password/token_version)
  ↓
Users       ← 依赖 Assessment (profile)
  ↓
Assessment  ← 自带题库 (assessment.yaml)，由后端 API serve
  ↓
Reviews     ← 依赖 Assessment (peer 答案池化), 依赖 User (subject)
  ↓
Matches     ← 依赖 Assessment (双方 d), 依赖 User (匹配关系)
  ↓
Flow        ← 依赖 Assessment (用户 d), 依赖 Solar Terms (Go API 实时计算)
  ↓
Payments    ← 依赖 User (user_subscriptions: passport, plan, bond_count)
Bond/Flow   ← 依赖 PassportChecker (通行证门控)
```

---

## 2. 架构决策记录 (ADR)

### ADR-001: 速率限制在 Caddy 代理层

**决策**: 速率限制在 Caddy 代理层实施，Go 应用层不设限流。

**理由**: Caddy 通过 `caddy-ratelimit` 模块，以 `{remote_host}` 为 key 获取真实客户端 IP。若交给 Go 处理限流，到达 Go 的 IP 是 Docker 内网地址（`172.x.x.x`），无法区分真实客户端。

| 限流 zone | 速率 | 覆盖路由 |
|-----------|------|----------|
| `auth` | 20 req/min | `route /api/auth/*` |
| `assess` | 30 req/min | `route /api/assessments` |

超限返回 429，含 `Retry-After` header。

### ADR-002: 前端技术选型

**决策**: MPA 纯静态 HTML + Alpine.js + Tailwind CSS + daisyUI + ECharts。Eleventy 构建时两次构建分别产出 EN/ZH 页面，Caddy file_server 直出。

**理由**:
- MPA 架构每个页面独立 HTML，无客户端路由。SEO 友好
- Alpine.js (~15KB) 处理组件内状态（滑块、dropdown、toggle），纯声明式
- Eleventy 构建时注入单语言数据，零运行时翻译查找开销
- esbuild 打包 ECharts 定制包 + JS，Tailwind CLI 构建 CSS
- 总 JS < 150KB gzip，总 CSS < 10KB gzip

### ADR-003: DNS-01 挑战代替 HTTP-01

**决策**: Caddy 使用 DNS-01 挑战获取 Let's Encrypt 证书，而非 HTTP-01。

**理由**: Cloudflare 橙云（代理模式）下，CF 边缘节点拦截所有 HTTP 流量并代行 TLS 终结。传统 HTTP-01 挑战的验证请求同样被拦截，Caddy 无法完成验证。DNS-01 通过 Cloudflare API 在 DNS 区域自动创建 `_acme-challenge` TXT 记录来完成域名验证，不依赖 HTTP 通路。

Caddy 需通过 `xcaddy build --with github.com/caddy-dns/cloudflare` 构建自定义镜像。

### ADR-004: 无 `/api/v1/` 前缀

**决策**: API 路由不含版本前缀。当前为 `/api/*`，未来若需版本并存，通过 Caddy 路由按路径分发（如 `/api/v2/*` → `backend_v2:8080`），Go 端始终不感知版本前缀。

**理由**: 单版本期间 `/api/v1/` 是纯噪音。真正需要版本并存时，Caddy 是天然的 API 网关——按路径前缀路由到不同容器，Go 端代码不变。

---

## 3. 领域划分决策记录 (Domain Partitioning Decisions)

以下记录影响架构完整性的关键领域划分决策，每项决策含依据、权衡和演进方向，为后续扩展提供约束。

---

### DPD-001: MatchRequest 独立为聚合根

**决策**: MatchRequest 是独立聚合根，非 Profile 或 User 的子实体。

**理由**:
- MatchRequest 有自己的生命周期（pending → accepted/declined 状态机），不依赖 Profile 的创建或更新
- MatchRequest 涉及的并发安全（同向唯一、交叉冲突）是独立的一致性边界——不应耦合到 Profile 的更新事务中
- 一方注销时 MatchRequest 的清理策略与 Assessment/ReviewLink 不同：直接 DELETE 而非软删除或 NULL 脱钩

**演进方向**: 若引入群组匹配或多层嵌套关系，MatchRequest 聚合可能分裂为 ConnectionRequest（一对一）和 GroupInvite（一对多），但独立聚合根的地位不变。

---

### DPD-002: MatchRepository 拆分为请求仓库与 Bond 仓库

**决策**: 将原 `MatchRepository` 拆分为 `MatchRequestRepository`（授权生命周期）和 `BondRepository`（计算查询）。

**理由**:
- 接口隔离原则 (ISP)：请求生命周期操作（Send/Respond/ListInbox）与 Bond 计算查询（LoadProfile/ListMatches）服务于不同 use case，不应强制同一实现
- 调用者语义清晰：`SendMatchRequestUseCase` 仅依赖 `MatchRequestRepository`，`BondUseCase` 仅依赖 `BondRepository`——避免了"需要 MatchRepo 只是为了调 LoadProfile"的困惑
- 两个接口可以指向同一 sqlite 实现（当前），也可以在未来指向不同数据源（如 Bond 计算迁移到独立服务）

**代码位置**: `internal/app/application/match/ports.go`

---

### DPD-003: PersonalityProfile 作为值对象，独立于 Assessment

**决策**: `PersonalityProfile`（d 向量 + p 向量 + Identity）是值对象，概念独立于 `Assessment`（测量事件）。

**理由**:
- Assessment 是"测量事件"——不可变事实，创建后永不修改
- PersonalityProfile 是"测量结果"——用户当前人格状态，可以来自最新自评、他评聚合或外部导入
- 若不分离：删除 Assessment 时 Profile 丢失；多次测量合并时模型难以演化；Peer 聚合的 Profile 无处归属
- 当前表结构暂未物理分离（profile 仍存于 assessments 表），但代码层面 `domain.PersonalityProfile` 已是独立值对象类型——通过 `FindLatestProfile()` 查询而不要求调用方知晓存储细节

**演进方向**: 可考虑 `user_profiles` 独立表——新 Assessment 提交时 UPSERT，消除 `ORDER BY id DESC LIMIT 1` 查询模式。

---

### DPD-004: engine 为共享内核

**决策**: Shared kernel 包是全部有界上下文的 Shared Kernel，包含 Deviation、Proportion、Identity、S/C 矩阵、原型向量、ComputeBond、ComputeFlow，不依赖任何其他层。

**边界约定**:
- engine 类型不可变——所有操作返回新值
- engine 函数纯函数——无 I/O，无副作用，确定性
- engine 不定义接口——接口由消费方（application 层）定义
- engine 变更需跨上下文评审——所有聚合都受影响

**反模式识别**:
- `engine` 包不应包含任何 SQL/HTTP 相关代码
- 不应在 engine 中定义 Repository 或 Service 接口
- JSON 序列化/反序列化不进 engine——那是 infrastructure 层职责

**代码位置**: `internal/ganzhi/`, `internal/tianwen/`, `internal/25types/`

---

### DPD-005: Deviation 和 Proportion 是一等值对象

**决策**: d 向量（Deviation）和 p 向量（Proportion）使用强类型 `[5]float64` 别名，自带不变式保护和方法集，不裸露为裸数组。

**理由**:
- 类型安全：`Deviation`（Σ=0）和 `Proportion`（Σ=1）不能混用——编译器阻止错误赋值
- 不变式集中：sum≈0 的验证在构造器中，而非散布在各处
- 行为内聚：Relu、Add、Sub、Dot、ToProportion 作为方法，调用代码自文档化
- 消除 JSON 散落：Profile 直接用 Deviation/Proportion 类型，JSON 序列化在 repository 实现层消化

**Pre-Refactor 反模式**:
```go
// 旧：裸数组，无类型安全，匿名 [5]float64 遍布各处
var dPos [5]float64
for i, v := range d { if v > 0 { dPos[i] = v } }
score := dot(dPos, proto)
```

**Post-Refactor**:
```go
dPos := d.Relu()
score := dPos.Dot(proto)
```

**代码位置**: `internal/25types/deviation.go`, `internal/25types/proportion.go`

---

### DPD-006: Claimer 接口窄化 + 注册即认领

**决策**: 匿名评估认领通过窄接口 `Claimer` 完成，注册时若携带 `anonymous_token` 则自动认领——不使用粗粒度的 `AssessmentRepository`。

**理由**:
- ISP：`RegisterUseCase` 只需要 `ClaimAnonymous(ctx, userID, token)` ——不需要完整的 AssessmentRepository（20 个方法）
- `Claimer` 接口在 user 包中定义（消费者定义接口）——由 user 上下文消费
- 注册时自动认领减少了客户端交互步骤——无需先注册再调 `/api/assessments/claim`

**接口定义**:
```go
// internal/app/application/user/ports.go
type Claimer interface {
    ClaimAnonymous(ctx context.Context, userID int64, token string) (int64, error)
}
```

**独立 `/claim` 端点保留**: 用于跨设备场景——用户在设备 A 匿名评估，在设备 B 登录后认领。

---

### DPD-007: ForgotPassword 完整流程归入 Use Case

**决策**: ForgotPassword 的完整流程（生成 token → 存储 DB → 日志记录）在 use case 层完成，handler 仅负责解析请求和委托。

**Pre-Refactor 反模式**: Use case 只生成 token 然后 `log.Printf`（不调 repo），handler 独自完成全流程（生成 token + 调 repo + log）。两套重复逻辑，use case 是死代码。

**Post-Refactor**: `ForgotPasswordUseCase` 调用 `repo.SetPasswordResetToken` 且包含 log，handler 仅一行委托。

**代码位置**: `internal/application/user/service.go:194-208`

---

### DPD-008: stringly-typed 错误 → sentinel error

**决策**: 所有领域错误使用 `errors.Is()` 可比较的 sentinel error，禁止 `err.Error() == "..."` 字符串比较。

**理由**:
- 字符串比较脆弱——错误消息微调即破坏 handler 的匹配逻辑
- sentinel error 支持 `errors.Is` 链式展开（`fmt.Errorf("...: %w", ErrXxx)`）
- 让每个错误有明确的聚合归属——`domain/errors.go` 按 User / Assessment / ReviewLink / Match / Commerce 分组

**Pre-Refactor 反模式**:
```go
switch err.Error() {
case "to_user_id is required": ...
case "cannot send match request to yourself": ...
```

**Post-Refactor**:
```go
switch {
case errors.Is(err, domain.ErrToUserIDRequired): ...
case errors.Is(err, domain.ErrCannotMatchSelf): ...
```

**新增 sentinel**: `ErrNameEmpty`, `ErrNoFields` — 替换了 `UpdateMeUseCase` 中的 `errors.New()` 字面量。

---

### DPD-009: JSON 序列化下沉至 sqlite 层

**决策**: `application/` 层不使用 `json.Marshal`/`json.Unmarshal`。序列化/反序列化是 infrastructure 层职责。

**理由**:
- application 层传递的是已解析的领域类型（`domain.PersonalityProfile`, `[]engine.Answer`）
- repository 接口的输入/输出是领域类型，不出现 `json.RawMessage` 或 `string`（JSON）
- 切换存储后端（SQLite → PostgreSQL）时，只需修改 sqlite 层的序列化逻辑

**辅助函数**: `marshalJSON(v) string` 替代 `mustMarshalJSON(v) string`——后者在 JSON 序列化失败时 panic（crash-risk），前者返回空字符串（优雅降级）。

---

### DPD-010: Bond 和 Flow 是值对象，不建表

**决策**: Bond（关系动力学）和 Flow（季节性投影）是纯计算值对象——每次请求实时计算，不持久化。

**理由**:
- 它们从 PersonalityProfile 确定性派生——存储是冗余的
- 节气月判定随时间变化——旧计算值会因闰年偏移而不一致
- 不建表避免了"存什么时刻的计算结果"的语义难题

**例外**: Bond 历史（通行证功能）和 Flow 记录若需持久化，应在独立表中存储快照（含时间戳和计算参数），而非作为 Bond/Flow 值的持久化形式。

---

### DPD-011: admin 端点需认证

**决策**: `/api/admin/*` 路由套 `auth` middleware，所有登录用户可访问。

**理由**: 生成激活码属敏感操作。后续可加 `RequireAdmin` middleware（检查 `role` claim）收窄为仅管理员。

**代码位置**: `internal/app/http/server.go:108-110`

---

### DPD-013: BaZi 八字 — 纯计算值对象，不建表

**决策**: MingliChart 和 MingliMatch 是纯计算值对象，实时计算不持久化，API 匿名可用。

**理由**:
- 八字排盘从出生时间 + 地点确定性派生——存储是冗余的
- 合八字是两份排盘的即时比对——不需要历史记录
- 匿名可用降低采纳门槛——"先试用再注册"的经典转化路径
- Profile 中嵌入 birth_info 后自动计算并返回 bazi_chart——计算即保存

**演进方向**: 若需"保存排盘"功能（如用户想保存多个出生时间的排盘），可在 profile 层扩展 `birth_infos[]`，而非为 BaZi 单独建表。

**代码位置**: `internal/mingli/bazi/`, `internal/app/domain/mingli_match_link.go`, `internal/mingli/http/bazi.go`

---

### DPD-012: Resend 作为邮件基础设施

**决策**: 使用 Resend API 发送 transactional email，以 `nil`-safe 接口注入到 use case。

**架构决策**:
- `user.EmailSender` 和 `commerce.WelcomeSender` 是两个窄接口——各自在消费方（user/commerce 包）定义
- 实现方 `infra/resend.Client` 同时实现两个接口——一个具体类型满足多个窄接口
- Use case 在 `sender != nil` 时才调用——未配置 `RESEND_API_KEY` 环境变量时优雅降级为 `log.Printf`，不阻塞业务
- `ForgotPasswordUseCase` 和 `UpdateMeUseCase` 中的邮件发送是**同步 best-effort**——邮件失败仅日志记录，不影响 API 响应
- 邮件内容在 `resend` 包内按 locale（en/zh-CN）选择模板，use case 和 handler 不感知邮件格式

**未发送邮件时的 fallback**:
- `ForgotPasswordUseCase`: 输出 `log.Printf("[email] password reset for %s: /reset-password?token=%s ..."` — 管理员可通过日志手动发送
- `UpdateMeUseCase`: 生成 token 存入 DB 但不发送——管理员可通过 `/verify-email?token=` 手动验证

**代码位置**: `internal/app/infra/resend/client.go`, `internal/application/user/service.go:195-210`, `cmd/server/main.go:38-48`

---

## 4. 共享内核 (Shared Kernel)

Shared kernel 包是整个系统的共享内核。所有有界上下文依赖它，它不依赖任何其他层。

**约定**:
- engine 类型是不可变的——所有操作返回新值
- engine 函数是纯函数——无 I/O，无副作用，确定性
- Deviation（d 向量，Σ=0）和 Proportion（p 向量，Σ=1）是一等值对象
- S、C 矩阵、原型向量、α 定位参数是定义性常量，变更需跨团队评审
- 新的数学概念优先放入 engine

### 4.1 元素编码

| 元素 | 编码 | 角色 (EN) | 角色 (ZH) | 本质（洪范） | 一气周流 |
|------|------|-----------|-----------|------|---------|
| Wood (木) | W | Pioneer | 开创者 | 曲直 — 方向性生长 | 左升 — 从下往上推动 |
| Fire (火) | F | Luminary | 发光者 | 炎上 — 从中心向四周辐射 | 上散 — 到顶，向四周打开 |
| Earth (土) | E | Cultivator | 培育者 | 稼穡 — 耕作培育，播种收获 | 中 — 斡旋四维 |
| Metal (金) | M | Refiner | 规范者 | 从革 — 循标准又打破 | 右降 — 从上往下收敛 |
| Water (水) | R | Reservoir | 蕴蓄者 | 润下 — 向下渗透 | 下藏 — 沉淀保存 |

木火为发，金水为收，土居中斡旋。经典依据见 [THEORY](theory/THEORY.md)。

S/C 矩阵索引（0-indexed）：Wood=0, Fire=1, Earth=2, Metal=3, Water=4。S[j][k] / C[j][k] 用数字下标直接访问。

### 4.1b 天干地支编码

八字引擎的干支枚举，与五元素编码独立：

| 天干 | 值 | 阴阳 | 五行 | 地支 | 值 | 五行 |
|------|---|------|------|------|---|------|
| 甲 | 1 | 阳 | 木 | 子 | 1 | 水 |
| 乙 | 2 | 阴 | 木 | 丑 | 2 | 土 |
| 丙 | 3 | 阳 | 火 | 寅 | 3 | 木 |
| 丁 | 4 | 阴 | 火 | 卯 | 4 | 木 |
| 戊 | 5 | 阳 | 土 | 辰 | 5 | 土 |
| 己 | 6 | 阴 | 土 | 巳 | 6 | 火 |
| 庚 | 7 | 阳 | 金 | 午 | 7 | 火 |
| 辛 | 8 | 阴 | 金 | 未 | 8 | 土 |
| 壬 | 9 | 阳 | 水 | 申 | 9 | 金 |
| 癸 | 10 | 阴 | 水 | 酉 | 10 | 金 |
| | | | | 戌 | 11 | 土 |
| | | | | 亥 | 12 | 水 |

六十甲子序数：`SixtyCycleName(stem, branch)`，范围 0-59。公式：`(stem*6 - branch*5 - 1) % 60`。

理论依据：[theory/bazi-theory.md](theory/bazi-theory.md)。

### 4.2 d 向量与 p 向量

- `Deviation` (d vector): `[5]float64` 的类型别名，Σ=0。`d[i] > 0` 为偏盛，`d[i] < 0` 为偏弱。有 `Relu()`, `Add()`, `Sub()`, `Dot()`, `ToProportion()`, `Sum()`, `MarshalJSON()` 方法。JSON 序列化为命名元素对象 `{"wood":..., "fire":..., ...}`
- `Proportion` (p vector): `[5]float64` 的类型别名，Σ=1，各分量 ∈ [0,1]。仅用于展示，不参与计算。有 `ToDeviation()`, `MarshalJSON()` 方法。JSON 序列化为命名元素对象 `{"wood":..., "fire":..., ...}`
- `d⁺ = d.Relu()`: 只有偏盛元素参与生克力输出

### 4.3 生矩阵 S (5×5)

S[j][k] = 1 当且仅当元素 j 生元素 k。`Wood→Fire→Earth→Metal→Water→Wood`

|   | W | F | E | M | R |
|---|---|---|---|---|---|
| W | 0 | 1 | 0 | 0 | 0 |
| F | 0 | 0 | 1 | 0 | 0 |
| E | 0 | 0 | 0 | 1 | 0 |
| M | 0 | 0 | 0 | 0 | 1 |
| R | 1 | 0 | 0 | 0 | 0 |

### 4.4 克矩阵 C (5×5)

C[j][k] = 1 当且仅当元素 j 克元素 k。`Wood→Earth→Water→Fire→Metal→Wood`

|   | W | F | E | M | R |
|---|---|---|---|---|---|
| W | 0 | 0 | 1 | 0 | 0 |
| F | 0 | 0 | 0 | 1 | 0 |
| E | 0 | 0 | 0 | 0 | 1 |
| M | 1 | 0 | 0 | 0 | 0 |
| R | 0 | 1 | 0 | 0 | 0 |

Sᵀ 和 Cᵀ 通过矩阵转置运算：
- `Sᵀd⁺` — "生我者对我的推动"：生我者的 d⁺ 分量加到被生元素上
- `Cᵀd⁺` — "克我者对我的约束"：克我者的 d⁺ 分量从被克元素上减去

### 4.5 25 个原型向量

定义在单纯形 p-空间（Σp=1，各分量 ∈ [0,1]），内部自动转为 d-空间（d[i] = p[i] - 0.2，Σd=0）参与余弦分类 `cos(d, d_proto)`。纯型为顶点。20 个复合型由统一参数 α=0.52 定位：主=α=0.52，次=1-α=0.48。

**Pure (×5):**

| ID | Wood | Fire | Earth | Metal | Water |
|----|------|------|-------|-------|-------|
| W  | 1.00 | 0.00 | 0.00 | 0.00 | 0.00 |
| F  | 0.00 | 1.00 | 0.00 | 0.00 | 0.00 |
| E  | 0.00 | 0.00 | 1.00 | 0.00 | 0.00 |
| M  | 0.00 | 0.00 | 0.00 | 1.00 | 0.00 |
| R  | 0.00 | 0.00 | 0.00 | 0.00 | 1.00 |

**Composite — 生序 (×10):**

| ID | Wood | Fire | Earth | Metal | Water |
|----|------|------|-------|-------|-------|
| WF | 0.52 | 0.48 | 0 | 0 | 0 |
| FW | 0.48 | 0.52 | 0 | 0 | 0 |
| FE | 0 | 0.52 | 0.48 | 0 | 0 |
| EF | 0 | 0.48 | 0.52 | 0 | 0 |
| EM | 0 | 0 | 0.52 | 0.48 | 0 |
| ME | 0 | 0 | 0.48 | 0.52 | 0 |
| MR | 0 | 0 | 0 | 0.52 | 0.48 |
| RM | 0 | 0 | 0 | 0.48 | 0.52 |
| RW | 0.48 | 0 | 0 | 0 | 0.52 |
| WR | 0.52 | 0 | 0 | 0 | 0.48 |

**Composite — 克序 (×10):**

| ID | Wood | Fire | Earth | Metal | Water |
|----|------|------|-------|-------|-------|
| WE | 0.52 | 0 | 0.48 | 0 | 0 |
| EW | 0.48 | 0 | 0.52 | 0 | 0 |
| FM | 0 | 0.52 | 0 | 0.48 | 0 |
| MF | 0 | 0.48 | 0 | 0.52 | 0 |
| ER | 0 | 0 | 0.52 | 0 | 0.48 |
| RE | 0 | 0 | 0.48 | 0 | 0.52 |
| MW | 0.48 | 0 | 0 | 0.52 | 0 |
| WM | 0.52 | 0 | 0 | 0.48 | 0 |
| RF | 0 | 0.48 | 0 | 0 | 0.52 |
| FR | 0 | 0.52 | 0 | 0 | 0.48 |

### 4.6 Solar Month Table

节气月向量由 Go API `/api/flow` 和 `/api/solar-terms` 实时计算返回，不依赖静态文件。

| 月 | EN | Wood | Fire | Earth | Metal | Water |
|----|-----|------|------|-------|-------|-------|
| 寅月 | Early Spring  | 1 | 0 | 0 | 0 | 0 |
| 卯月 | Mid Spring    | 1 | 0 | 0 | 0 | 0 |
| 辰月 | Late Spring   | 0 | 0 | 1 | 0 | 0 |
| 巳月 | Early Summer  | 0 | 1 | 0 | 0 | 0 |
| 午月 | Mid Summer    | 0 | 1 | 0 | 0 | 0 |
| 未月 | Late Summer   | 0 | 0 | 1 | 0 | 0 |
| 申月 | Early Autumn  | 0 | 0 | 0 | 1 | 0 |
| 酉月 | Mid Autumn    | 0 | 0 | 0 | 1 | 0 |
| 戌月 | Late Autumn   | 0 | 0 | 1 | 0 | 0 |
| 亥月 | Early Winter  | 0 | 0 | 0 | 0 | 1 |
| 子月 | Mid Winter    | 0 | 0 | 0 | 0 | 1 |
| 丑月 | Late Winter   | 0 | 0 | 1 | 0 | 0 |

节气月判定使用太阳黄经迭代算法。服务启动时预计算当年 12 个节气日边界，请求时 O(log 12) 查区间。跨年自动刷新。Flow 计算：`generates = S · t_month`, `restrains = C · t_month`（纯方向，无数值偏移）。

太阳黄经边界（12 节气起始度数）：立春 315°, 惊蛰 345°, 清明 15°, 立夏 45°, 芒种 75°, 小暑 105°, 立秋 135°, 白露 165°, 寒露 195°, 立冬 225°, 大雪 255°, 小寒 285°。

### 4.7 Envelope 格式

统一响应格式——Go 业务 API 和 Content JSON 共用：

**成功 — 单对象**: `{ "data": { ... } }`
**成功 — 列表**: `{ "data": { "items": [...], "total": 42 } }`
**错误**: `{ "error": { "code": "invalid_request", "message": "human-readable detail" } }`

---

## 5. 全局约定

### 5.1 JWT 策略

- **签名算法**: HMAC-SHA256（HS256）。单服务架构下 HMAC 足够
- **有效期**: 30 天
- **密钥**: 环境变量 `JWT_SECRET`，32+ 字节 base64。不提供默认值
- **失效机制**: `token_version` 计数器。密码变更或账户注销时 +1，旧 token 立即失效
- **密钥轮换**: 旧密钥验证新请求 30 天（覆盖最长 JWT 有效期），之后仅新密钥验证
- **存储**: 客户端 `localStorage`，JS fetch 以 `Authorization: Bearer <token>` header 发送

### 5.2 Auth 标记

| 标记 | 含义 |
|------|------|
| 🔒 | RequireAuth — 需要认证。未认证返回 `401 unauthorized` |
| 🔓 | OptionalAuth — 可选认证。已登录则自动关联用户，未登录以匿名模式运行 |

### 5.3 Base URL

```
/api           — Go 业务 API (Caddy 反向代理 → :8080)
/content       — 静态内容 (Caddy file_server 直接 serve)
```

### 5.4 多语言

前端页面采用 URL-based locale 路由：`/en/about`、`/zh-CN/about`。Caddy 在 `/` 通过 `Accept-Language` 嗅探 302 重定向到对应语言首页。共享资源（JS/CSS/fonts/img）无 locale 前缀。

业务 API 通过 Header: `X-Locale: zh-CN` 或 `Accept-Language`。客户端从 URL 路径提取当前 locale，通过 `api()` fetch wrapper 自动注入 `X-Locale` header。

默认 locale: `en`。优先级：URL 路径 > `X-Locale` > `Accept-Language` > `en` fallback。

### 5.5 已知权衡与限制

| 项目 | 决策 | 理由 |
|------|------|------|
| **CSP `unsafe-inline`** | 允许 `script-src 'unsafe-inline'` | Alpine.js 依赖内联 `x-data`/`@click` 属性展开，无法避免 |
| **账户锁定** | 无登录失败锁定/递增延迟 | 当前仅 Caddy 层 `/api/auth/*` 20req/min 限流。暴力破解风险通过 argon2id 哈希成本 + 限流缓解 |
| **GDPR 硬删除** | 仅软删除（`deactivated_at`），7 天后禁止 reactivate | 数据保留用于审计和研究；用户数据不会物理清除 |
| **JWT 密钥轮换** | 单活跃 key，无双 key 验证窗口 | 密码变更通过 `token_version` 递增立即失效，降低轮换紧迫性 |

---


### 5.6 数据库时间格式

SQLite TEXT 列存储时间戳必须统一使用 RFC3339 格式。Go 代码使用 `sql.NullString` 扫描，通过共享 helper `parseNullTime` 解析为 `*time.Time`（`modernc.org/sqlite` 不支持直接扫描到 `sql.NullTime`）。

**DDL 约定**：所有 `_at` 列的 DEFAULT 必须使用 `strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`，严禁 `datetime('now')`（产物为 `"2006-01-02 15:04:05"`，与 RFC3339 不兼容）。

**Go 扫描约定**：
```go
var createdAt sql.NullString
row.Scan(&createdAt)
t := parseNullTime(createdAt) // → *time.Time, nil if invalid/empty
```

`parseNullTime` 定义在 `internal/app/sqlite/user_repo.go`，是 sqlite 包内共享的时间解析函数。

**历史问题**：migrations 010、011、013、017 曾使用 `datetime('now')`，已通过 014（match_links/bond_events）和 019（donations）的 UPDATE 修正。新增 migration 时务必检查 DEFAULT 格式。

## 6. 文档索引

### 产品与架构

| 文档 | 内容 |
|------|------|
| [PRODUCT.md](PRODUCT.md) | 产品定位、双引擎架构、获客架构、54 用例全景、变现模型 |
| [ENGINE-LLM.md](ENGINE-LLM.md) | 引擎 × LLM 功能分解：哪些纯计算、哪些 AI 表达、数据流与协作 |
| [API.md](API.md) | HTTP 契约唯一权威源 |
| [OPERATIONS.md](OPERATIONS.md) | 部署、备份、CI/CD、可观测性 |
| [TESTING.md](TESTING.md) | 测试策略、金字塔、用例清单 |
| [FRONTEND.md](FRONTEND.md) | 前端技术栈、页面清单、组件树、交互模式 |

### 实现规格

| 文档 | 内容 |
|------|------|
| [domain/user.md](domain/user.md) | User 聚合：注册、登录、密码管理、资料编辑、注销、GDPR |
| [domain/assessment.md](domain/assessment.md) | Assessment 聚合：评估提交、身份分类、他评、Peers 聚合、Flow 投影 |
| [domain/match.md](domain/match.md) | Match 聚合：匹配请求、Bond 计算、并发安全 |
| [domain/bazi.md](domain/bazi.md) | BaZi 八字：排盘、合八字、真太阳时、城市经纬度 |
| [domain/commerce.md](domain/commerce.md) | 支付：订阅、激活码、webhook |
| [domain/naming.md](domain/naming.md) | 起名：喜用神、生肖喜忌、三才五格、汉字数据库 |
| [domain/huangli.md](domain/huangli.md) | 择日：日柱对敲、天干十神、地支冲合、事件五行、建除十二神 |
| [domain/reports.md](domain/reports.md) | Reports：解读统一出口、SSE 流式生成、scene/sub_scene 模型、分享 |
| [domain/llm.md](domain/llm.md) | LLM 集成：提示词模板、SSE 流式契约、降级策略 |

### 理论基础

| 文档 | 内容 |
|------|------|
| [theory/THEORY.md](theory/THEORY.md) | 五行生克代数推导 |
| [theory/SPEC.md](theory/SPEC.md) | 编码规范与参数选择 |
| [theory/QUESTIONNAIRE.md](theory/QUESTIONNAIRE.md) | 问卷设计方法论 |
| [theory/ASSESSMENT.md](theory/ASSESSMENT.md) | 评估逻辑 |
| [theory/CULTURE.md](theory/CULTURE.md) | 文化语境 |
| [theory/bazi-theory.md](theory/bazi-theory.md) | 八字命理学：天干地支、四柱推算、合婚规则、真太阳时 |
| [theory/NOVELTY.md](theory/NOVELTY.md) | 创新声明 |

### 附录

| 文档 | 内容 |
|------|------|
| [appendix/errors.md](appendix/errors.md) | 完整错误码参照 |
