# Match Link / Bond — 实现规格

> 匹配链接、Bond 动力计算、bond_events 持久化。Bond 为值对象，bond_json 为历史快照。

---

## 1. 计费模型

```
Link   — 分享通道。创建 + 打开 + 做题都免费。不限次数使用。不消耗。
Bond   — 计费事件。双方 profile 就绪后计算一次 bond，记录到 bond_events。消耗 1 次 quota。
Quota  — 在创建 Bond 时检查（不是创建 Link 时）。NULL=无限，0=耗尽。预留，当前不实现。
```

| 实体 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| MatchLink | 聚合根 | 用户创建，软删除 | token 唯一。免费，可多次使用 |
| Bond | 值对象 | 从两份 PersonalityProfile 实时计算 | 双向 delta (δ_a, δ_b) |
| BondEvent | 事件记录 | 每次 Bond 计算写入一条 | bond_json 快照。同一对人只保留最新一条 |
| Quota | 接口（预留） | commerce 上下文 | 在 Bond 创建前检查，非 Link 创建时 |

**为什么配额挂在 Bond 而非 Link**：Link 只是分享入口，发链接不花钱。真正产生价值的是两个人的匹配计算。用户 A 发链接给 B，B 没有 profile（不做题），= 没产生 bond，不计费。

---

## 2. 仓库拆分

### MatchLinkRepository

```go
Create(ctx context.Context, userID int64, token string) (int64, error)
FindByToken(ctx context.Context, token string) (*domain.MatchLink, error)
ListByUser(ctx context.Context, userID int64) ([]MatchLinkItem, error)
SoftDelete(ctx context.Context, id int64, userID int64) (bool, error)
```

### ProfileRepository

```go
FindByName(ctx context.Context, name string) (*domain.User, error)
LoadProfile(ctx context.Context, userID int64) (*domain.PersonalityProfile, error)
ListPeerAnswersForUser(ctx context.Context, userID int64) ([]engine.Answer, int, error)
FindActiveReviewLink(ctx context.Context, userID int64) (string, bool)
```

### BondStorer

```go
InsertBondEvent(ctx context.Context, linkID *int64, initiatorID, otherID int64, assessmentID *int64, bond *domain.Bond) error
ListBondEvents(ctx context.Context, userID int64) ([]domain.BondEvent, error)
```

### Quota 接口（预留）

```go
type BondQuotaChecker interface {
    CheckBondQuota(ctx context.Context, userID int64) (remaining int, ok bool)
    ConsumeBondQuota(ctx context.Context, userID int64) error
}
```

---

## 3. 行为规则

### 3.1 匹配链接

- 已登录用户创建匹配链接（token 唯一），免费
- 链接可被任何人打开，可多次使用
- 收件人做题或使用已有 profile → 计算 Bond → 写入 bond_events → 消耗 link 创建者 1 次 quota（未来）
- 收件人可通过 `anonymous_token` 注册认领
- 已登录用户有 profile → 直接计算 Bond，跳过做题

### 3.2 即时对比

- POST /api/bonds，link_id=NULL
- 未来也消耗 quota（即时对比也算一次 bond）

### 3.3 Bond 动力计算

```
d^A⁺ = ReLU(d^A)
d^B⁺ = ReLU(d^B)
Δ^A = Sᵀd^B⁺ − Cᵀd^B⁺    // B 对 A 的影响
Δ^B = Sᵀd^A⁺ − Cᵀd^A⁺    // A 对 B 的影响
d^A_eff = d^A + Δ^A
d^B_eff = d^B + Δ^B
```

### 3.4 bond_json 持久化

`bond_events` 存储每次计算的快照：

```
id, link_id, initiator_user_id, other_user_id, other_name,
assessment_id, bond_json, created_at
```

`bond_json = {"self":{...},"other":{...},"delta_a":{...},"delta_b":{...}}`

self/other 方向取决于谁触发的计算（initiator 的 profile = self）。`created_at` 以 RFC3339 格式写入（兼容旧 `datetime('now')` 格式读取）。

### 3.5 同一对人去重

同一对人（A, B）之间可能有多条 bond_event：
- A 的 link → B 使用
- B 的 link → A 使用
- A 对 B 即时对比
- B 对 A 即时对比

**规则：同一对人只保留最新一条。** InsertBondEvent 时，先删除 initiator 与 other 之间已有的所有 bond_event（不考虑方向，即 (A,B) 和 (B,A) 都删除），再插入新的。

这样 profile 页面的 bond 列表每条代表一个唯一的人，语义清晰：你和这个人的最新匹配结果。

### 3.6 展示视角对齐

bond_json 的 self/other 方向取决于谁发起的。但展示时，profile owner 看到的 bond 应该始终把自己放左边（self 位）。

**方案：后端查询时视角对齐。** `GetBonds` 服务层在组装 `BondEventItem` 时，如果 other_user_id = viewer，交换 bond_json 中的 self↔other、delta_a↔delta_b，并修正 `OtherUserID` 指向真正的对方。前端拿到的永远是 viewer=self、对方=other。

---

## 4. Domain 类型

```go
type BondEvent struct {
    ID              int64
    LinkID          *int64
    InitiatorUserID int64
    OtherUserID     int64
    OtherName       string
    AssessmentID    *int64
    BondJSON        string  `json:"-"` // raw JSON snapshot, not serialized
    CreatedAt       time.Time
}

type Bond struct {
    Self   engine.Deviation `json:"self"`
    Other  engine.Deviation `json:"other"`
    DeltaA engine.Deviation `json:"delta_a"`
    DeltaB engine.Deviation `json:"delta_b"`
}
```

---

## 5. API

### POST /api/m/{token}

```
Request:  { answers, anonymous_token, use_existing, other_name }
Response: { profile, bond }
```

无 409。同一 link 可被多人多次使用。
- `use_existing=true`（需登录 + 已有 profile）：直接计算 bond，`otherID` 为当前用户 ID
- `answers` 路径：提交做题答案。登录用户自动关联 `otherID`；匿名用户 `otherID=0`

### POST /api/bonds

即时对比。已登录用户对另一个用户直接计算 bond。

```
Request:  { with_user_id } 或 { with_name }
Response: { bond }
```

### GET /api/profiles/{name}/bonds

```
Response: { items: [{ id, other_user_id, other_name, bond: {self, other, delta_a, delta_b}, source, created_at }], total }
```

每个 item 的 bond 对象已视角对齐（viewer = self）。同一对人只出现一次（最新）。

---

## 6. 错误

| Error | HTTP | 
|-------|------|
| ErrNoProfile | 404 |
| ErrMatchLinkNotFound | 404 |
| ErrQuotaExhausted（预留） | 402 |

---

## 7. 迁移

013_bond_json.sql 已完成：添加 other_name + bond_json 列。

---

## 8. 与旧系统的区别

| | 旧 | 新 |
|------|-----|------|
| Bond 存储 | 仅元数据 | bond_json 快照 |
| Link 使用 | 任意次 | 任意次（不变） |
| Quota 附着点 | 无 | Bond 创建时，非 Link 创建时 |
| 同对去重 | 无 | 只保留最新一条 |
| 展示方向 | 原始方向 | 视角对齐：viewer 始终在 self 位 |
| 匹配通知 | 无 | Link 创建者收到邮件通知 |

---

## 9. Bond 匹配通知邮件

当有人通过 match link 完成 bond 计算时，向 link 创建者发送通知邮件。

### 触发条件

- 仅 `POST /api/m/{token}` 路径，bond 创建成功后触发
- `POST /api/bonds`（即时对比）不触发通知
- 创建者无邮箱时静默跳过

### 接口

`user.EmailSender` 新增：

```go
SendBondNotification(ctx context.Context, to, otherName, creatorName, locale string) error
```

### 邮件内容

**EN:**
- Subject: `Someone matched with you on 25types`
- Body: `{other_name} completed a bond match through your link.\n\nView your bonds: https://25types.com/en/{creator_name}\n\n— 25types`

**ZH-CN:**
- Subject: `有人通过你的链接完成了匹配 — 25types`
- Body: `{other_name} 通过你的链接完成了一次匹配。\n\n查看你的匹配记录：https://25types.com/zh-CN/{creator_name}\n\n— 25types`

`other_name` 取值：已登录用户取用户名，匿名用户取 `other_name` 字段，均无则 "Anonymous" / "匿名用户"。

### 实现

- `infra/resend/client.go` 和 `infra/tencent/client.go` 分别实现
- 同步 best-effort：发送失败仅 `log.Printf`，不阻塞 API 响应
- nil-safe：未配置邮件服务时优雅降级
