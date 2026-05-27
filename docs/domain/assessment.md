# Assessment 聚合 — 实现规格

> 评估提交、身份分类、他评邀请、自评 vs 他评对比、匿名认领。Flow 时间投影引用共享内核。

---

## 1. 聚合边界

```
User ──owns──> Assessment(s) ──produces──> PersonalityProfile (d,p,identity)
  │
  ├──owns──> ReviewLink(s) ──receives──> Assessment(s) [type=peer]
```

| 实体 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| Assessment | 实体 | 提交创建，不可变 | 测量事件——核心数据创建后不可修改（触发器强制） |
| ReviewLink | 聚合根 | 用户创建，可软删除 | token 唯一，30 天有效。软删除/续期操作由 ReviewLinkRepository 实现 |
| PersonalityProfile | 值对象 | 从 Assessment 派生 | 用户当前人格状态——d 向量 + p 向量 + identity。概念独立于测量事件 |

PersonalityProfile 是领域一等公民，不同于 Assessment（测量事件）。Profile 使用 `engine.Deviation`（d 向量，Σ=0）和 `engine.Proportion`（p 向量，Σ=1）类型——两者都是强类型值对象，有完整的方法集（Relu/Add/Sub/Dot/ToProportion）。Identity 仅存 ID（如 "WF"），label 按查看者 locale 从 types.yaml 查找。`domain.NewProfile(d, p, id)` 构造——不再裸露 `[5]float64`。

**JSON 序列化下沉至 sqlite 层**：application 层传递 `domain.PersonalityProfile` 和 `[]engine.Answer`（已解析类型），JSON 序列化/反序列化在 `sqlite.AssessmentRepo` 内部完成。`marshalJSON()` 辅助函数替代 `mustMarshalJSON()`（panic → 空字符串降级）。

---

## 2. 行为规则

### 2.1 提交评估

三种提交路径：

**路径 A — 已登录 (HTTP 201)**：评估保存到 `assessments` 表，关联 `user_id`。返回 `id` + `profile` + `identity` + `data_complete`。

**路径 B — 匿名无 token (HTTP 200)**：返回计算结果但不保存。客户端生成新 UUID 存储到 localStorage。

**路径 C — 匿名带 token (HTTP 200)**：计算结果保存到该 token (`legacy_user_token`)，响应含 `anonymous_token`。后续注册可认领。

**选题算法 — 三级优先级（固定题库，按轮推送）**：

题库 30 题固定不变（不轮换、不随机抽）。每题 3 选项（三元追选，选 2 = 排除 1），6 轮 × 5 题，轮内元素出场平衡（每元素每轮恰好 3 次）。按轮推送，每次返回完整一轮的 5 道题：

1. **补缺**：推送用户未答过的轮次。已登录按 `user_id`，匿名按 `anonymous_token`。满 6 轮后 `data_complete = true`
2. **覆盖**：在未答轮次中，优先总体回答数低的轮次
3. **争议**：在未答轮次中，优先已有回答中分歧大的轮次

**答题交互**：
- 每轮 5 道三元追选题，选最符合的两个（排除 1 个），必须选两个不同选项
- 进度条："已完成 X/6 轮"
- 5 题全部作答后"提交"按钮激活
- 提交成功 → 结果区替换。底部"再做 5 题"按钮调用 `GET /api/assessments/next-round`
- `data_complete = true` 时不再展示"再做一轮"

**异常处理**：
- answers 为空 → 客户端校验拦截，按钮保持禁用
- QID 不在题库 (400) → toast "无效题目，请刷新重试"
- 已登录超 10 次/天 (429) → toast + 按钮禁用
- 匿名 IP 超 30 次/15min (429) → toast
- 网络异常或 500 → toast + 按钮恢复，已选选项保留

### 2.2 ComputeD 算法

```
输入: Answer[] (三元追选：每题 3 选项，选 2 个 = 排除 1 个，各计 1pt)

raw[e]     = Σ (e 被选中 ? 1 : 0)
d[e]       = 5 × (raw[e] / Σraw − 0.2)                   // Σ=0
p[e]       = raw[e] / Σraw                                // Σ=1, 仅用于展示
```

**边界**：0 答案 → d=零向量, p=均匀 0.2×5。全部选 W → d 反映极端偏好。某元素 raw=0 → p=0%, d=-1。

**反函数**：`DToP(d)` → `p[i] = 0.2 + d[i] / 5`。当从 DB 读取 profile_json 无原始 answers 时使用。

### 2.3 身份分类

```
identity_id = argmax  cos(d, d_protoₖ) = argmax  d · d_protoₖ / (‖d‖ · ‖d_protoₖ‖)
// d_protoₖ 为 25 个单纯形原型（Σd=0），d 是 profile 的形状向量
```

平局按生克关系优先级裁定。Identity 仅存 ID（如 "WF"），`label` 按查看者 locale 从 `types.yaml` 查找，`category` 由 ID 推导（单字符 → pure，双字符 → composite）。

### 2.4 结果展示

评估结果页展示：
- **雷达图**：d 向量入场动画 600ms
- **身份标签**：label + id + category（如 "Pioneer-Luminary · WF · composite"）
- **距离面板**：最近 5 个类型 + 距离值，最近与次近 < 0.05 时突出展示 + 引导语
- **角色匹配**：从 `characters.json` 查找
- **免责声明**：「这个结果基于你当下的选择，反映的是倾向而非永久标签」

**相邻类型引导**：当最近与次近类型距离 < 0.05 时，引导语如"你在 WF 和 FW 之间，两种描述都部分适合你"——防止微小波动导致标签切换困惑。

**匿名结果页额外展示**：banner "临时结果，刷新即丢失" + 注册 CTA 卡片 + 共享设备警告（小字）。

**已登录结果页导航出口**：探索其他类型、点击身份标签查看类型详情、查看历次变迁、再做一轮、Bond 试玩引导。

### 2.5 类型距离面板

客户端计算：`d → ReLU → d⁺ → 与 25 个原型逐一点积 → 降序排列`。展示最近 5 个类型及距离值。最近类型 = identity（一致性验证）。

### 2.6 评估历史

登录用户查看自己的评估历史，按时间倒序，分页（默认 20 条/页，最大 100 条）。

### 2.7 历次评估变迁

客户端可视化：历次 d 向量按时间轴连线（折线图或河流图），x 轴时间，y 轴 d 值。每次标注 identity 标签和日期。仅一次评估时不展示。

### 2.8 两次评估对比

客户端功能：从历史中任选两次，逐元素 diff。并排柱状图或叠加雷达图，标注变化量和方向（↑↓）。差异最大元素附简短说明。

### 2.9 匿名认领

- 注册时携带 `anonymous_token` 即自动认领——通过窄接口 `user.Claimer` 调用 `ClaimAnonymous()`
- 独立端点 `POST /api/assessments/claim` 保留用于跨设备场景（设备 A 匿名评估，设备 B 登录后认领）
- 认领逻辑：`UPDATE assessments SET user_id = ? WHERE legacy_user_token = ? AND user_id IS NULL`
- Claimer 是窄接口——仅 `ClaimAnonymous(ctx, userID, token) (int64, error)`，不暴露完整的 AssessmentRepository

**接口位置**：`user.Claimer`（user 上下文使用）——在消费方定义。

### 2.10 自评 vs 他评对比

登录用户查看自评与他评的聚合对比。结果包含三组 Profile + Identity：

- **self**：用户最新自评
- **peers_aggregated**：所有有效链接的全部 peer 答案池化 → ComputeD
- **combined**：自评 + 总体他评答案池化 → ComputeD

通过 `?link_id=` 查看单轮结果。

**池化逻辑**：每个他评链接对应一个独立评价轮次。池内同一评审者对同一题重复提交取最后一次。被评者看不到单个评审者原始数据，仅看聚合。

**深入分析门槛**：单链接评审者 ≥ 3 人（每人 ≥ 7 题，总数据点 ≥ 25）时 `deep_analysis_available = true`。不满足时不返回深入分析字段——小样本聚合可反推个体答案。免费和付费同样适用。

**元素分布对比图**（客户端）：三组 d 叠加在同一雷达图——自评、他评（peers_aggregated 的 d）、当前季节（`GET /api/flow` 的 generates/restrains 方向）。

---

## 3. ReviewLink — 他评邀请链接

### 3.1 生命周期

用户创建 → 生成唯一 token（`crypto/rand` 16 字节 hex，32 字符）→ 30 天有效 → 到期自动失效或手动续期（延长 30 天）→ 可软删除（`deleted_at` 时间戳）。

已删除链接不出现在列表中，已提交的他评数据保留。

### 3.2 评审者着陆页（三态）

**态 A — 链接有效**：展示被评者姓名 + 答题表单。`recommended_qids` 由选题算法生成（补缺 > 覆盖 > 争议），已排除当前评审者已答题。登录评审者 `reviewer_name` 自动填充只读，匿名评审者需填写。

**态 B — 链接已过期**：展示失效提示。未登录用户仅展示"已失效"。已登录 + 链接创建者额外展示"续期 30 天"按钮。

**态 C — 链接不存在或已删除**：展示 "此链接不存在或已被删除" + 返回首页链接。

### 3.3 提交他评


- `reviewer_name` 匿名必填，登录自动填充只读
- 作答推荐题目（≥1 题即可提交），3 选 2 同自评交互
- 提交成功 (201) → 表单替换为确认区：展示被评者 identity + 数据充足度提示
- 同一评审者重复提交同一题：最后一次覆盖，无额外提示
- "继续评价"按钮 → 选题返回新一批推荐题目，新答案追加不覆盖

**评审者识别**：去重依据 `user_id`（登录）或 `anonymous_token`（匿名）。`reviewer_name` 仅展示不参与去重。

**终态 CTA**：
- 匿名评审者：底部评估 CTA——"想知道你的类型吗？来做你自己的评估。"
- 登录评审者：底部"查看我的评价记录" → `/me/reviews-given`

**异常处理**：
- answers 为空 → 客户端拦截（按钮保持禁用直到至少选 1 题）
- `reviewer_name` 缺失（匿名）→ 输入框红框
- token 过期 → 页面刷新为过期态
- QID 不在题库 (400) → toast
- 网络异常或 500 → toast + 按钮恢复

### 3.4 查看评价记录

登录用户查看自己（作为评审者）提交过的所有他评记录。匿名评审者凭 `anonymous_token` 查看。按提交时间倒序，每项含被评者姓名、评价时间、已答题目数。

---

## 4. Flow 时间投影

Flow 有独立的 application 包 `internal/app/application/flow/`，包含 `GetFlow`、`GetFlowYearly`、`GetSolarTerms` 三个 use case。`ProfileLoader` 接口（`LoadProfile(ctx, userID) (*PersonalityProfile, error)`）由 consumer 定义。

Flow 不建表——每次请求实时计算。取用户最新 d，调用共享内核的 `engine.ComputeFlow()`，根据当前节气月判定生克方向：

```
generates  = S · t_month   → 被生元素索引
restrains = C · t_month   → 被克元素索引
```

不做数值偏移（不再计算 `d_eff`）。季节效应为纯方向指示——用户的 d 剖面全年不变，各月仅生克方向轮转。

`GetSolarTerms()` 使用 `sync.Once` 缓存当年节气日数据（避免每请求重复计算黄经迭代），跨年自动刷新。节气月判定使用太阳黄经迭代算法。

12 月年历 `/api/flow/yearly` 返回逐月生克方向。节气月表参见 INDEX §共享内核。

---

## 5. 不可变性

`assessments` 行一旦创建即为不可变事实。`BEFORE UPDATE` 触发器拦截任何对 `identity_id`、`answers_json`、`profile_json`、`assessment_type`、`review_link_id` 的修改。应用层不提供 UPDATE——仅 INSERT 和 SELECT。

```
assessments 表:
  id, user_id, assessment_type (self|peer), identity_id,
  answers_json, profile_json, created_at,
  review_link_id (peer 时非空), reviewer_name (peer 时非空),
  legacy_user_token (匿名认领凭据)

review_links 表:
  id, subject_user_id, token (UNIQUE), expires_at, created_at, deleted_at
```

---

## 6. 功能边界

| 运算 | 位置 | 端点 |
|------|------|------|
| d/p 计算 | 服务端 | POST /assessments |
| 身份分类 | 服务端 | POST /assessments（含 identity） |
| 他评聚合 | 服务端 | GET /assessments/peers（池化 → ComputeD） |
| Flow 时间 | 服务端 | GET /flow（d + 节气月判定 + solar month table） |
| Flow 年度 | 服务端 | GET /flow/yearly（12 月逐月 generates / restrains 方向） |
| 节气历 | 客户端 | 纯展示，数据来自 GET /api/solar-terms |
| 月度简报 | 客户端 | 数据来自 GET /api/flow/yearly |
| 元素提示 | 客户端 | 数据来自 GET /api/flow/yearly，通行证解锁 |

---

## 参见

- [INDEX](../INDEX.md) — S/C 矩阵、25 原型向量、Solar Month Table、envelope 格式
- [API](../API.md) — HTTP 契约
- [user](user.md) — User 聚合
- [match](match.md) — Bond 计算
- [commerce](commerce.md) — 通行证
- [theory/THEORY](../theory/THEORY.md) — 理论推导
