# Reports（报告）— 实现规格

> 解读统一出口。引擎产出结构化数据，Reports 将其转为自然语言解读。一问一报告，SSE 流式生成，可回看可分享。

---

## 1. 聚合边界

```
引擎计算 ──→ engine_data (JSON) ──→ Reports ──→ LLM 解读 ──→ content (Markdown)
                                         │
                                         ├── 流式输出（SSE chunk）
                                         ├── 持久化（可回看历史）
                                         ├── 分享链接（公开访问）
                                         └── 软删除（用户管理）
```

| 概念 | 类型 | 生命周期 | 关键约束 |
|------|------|---------|---------|
| Report | 聚合根 | 用户创建，软删除 | engine_data 为不可变快照，content 生成后不可编辑 |
| Scene | 枚举 | 静态 | 对齐 API 域前缀：mingli/naming/dates/relationship/career/general |
| SubScene | 枚举 | 静态 | 细分场景：mingge/dayun/liunian/liuyue/hehun/parenting 等 |
| ReportShare | 值对象 | 按需生成 | 公开 token，可撤销 |

---

## 2. 领域类型

### 2.1 报告

```go
type Report struct {
    ID          int64           `json:"id"`
    UserID      int64           `json:"user_id"`
    Scene       Scene           `json:"scene"`
    SubScene    string          `json:"sub_scene,omitempty"`
    Question    string          `json:"question,omitempty"`    // QA 场景才有
    EngineData  json.RawMessage `json:"engine_data"`           // 引擎计算的结构化数据快照
    Content     string          `json:"content"`               // LLM 生成的 Markdown 解读
    Locale      string          `json:"locale"`
    CreatedAt   time.Time       `json:"created_at"`
}
```

### 2.2 Scene 枚举

```go
type Scene string

const (
    SceneMingli         Scene = "mingli"         // 八字解读（mingge/dayun/liunian/liuyue/hehun）
    SceneNaming       Scene = "naming"       // 起名方案
    SceneAlmanac      Scene = "almanac"        // 择日方案
    SceneRelationship Scene = "relationship" // 25types 关系解读
    SceneCareer       Scene = "career"       // 职业建议
    SceneGeneral      Scene = "general"      // 自由问答
)
```

SubScene 按 Scene 细分：

| Scene | SubScene | 说明 |
|-------|----------|------|
| mingli | mingge | 命格解读 |
| mingli | dayun | 大运解读 |
| mingli | liunian | 流年解读 |
| mingli | liuyue | 流月解读 |
| mingli | hehun | 合婚解读 |
| naming | — | 起名方案（无子场景） |
| dates | wedding/engage/open/sign/move/... | 按事件类型 |
| relationship | bond | Bond 关系 |
| relationship | parenting | 亲子关系 |
| relationship | partner | 合伙人 |
| career | — | 职业建议 |
| general | — | 自由问答 |

### 2.3 创建请求

```go
type CreateReportRequest struct {
    Scene      Scene           `json:"scene"`
    SubScene   string          `json:"sub_scene,omitempty"`
    Question   string          `json:"question,omitempty"`
    EngineData json.RawMessage `json:"engine_data"`
    Locale     string          `json:"locale"`
}
```

### 2.4 列表

```go
type ReportItem struct {
    ID        int64     `json:"id"`
    Scene     Scene     `json:"scene"`
    SubScene  string    `json:"sub_scene,omitempty"`
    Question  string    `json:"question,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

type ReportList struct {
    Items []ReportItem `json:"items"`
    Total int          `json:"total"`
}
```

### 2.5 分享

```go
type ReportShare struct {
    Token     string    `json:"token"`
    ReportID  int64     `json:"report_id"`
    ExpiresAt time.Time `json:"expires_at,omitempty"` // NULL = 永久
}
```

---

## 3. 行为规则

### 3.1 创建与流式生成

```
POST /api/reports
  → 1. 校验 scene / engine_data
  → 2. 按 scene+locale 选提示词模板（configs/llm/）
  → 3. 拼 prompt，调 LLM
  → 4. SSE 流式输出 chunk
  → 5. 流结束，拼接 content，写入 reports 表
  → 6. 发送 done 事件含 report_id
```

SSE 事件流：

```
event: chunk
data: {"text": "你是甲木日主，生于寅月。"}

event: chunk
data: {"text": "甲木像一棵挺拔的大树..."}

event: done
data: {"report_id": 42}
```

### 3.2 不可变性

- `engine_data` 是生成时的快照——之后引擎算法升级不影响已有报告
- `content` 生成后不可编辑——报告是当时的解读，改了就失去"快照"意义
- 用户只能删，不能改

### 3.3 历史列表

```
GET /api/reports?scene=mingli&limit=20&offset=0
  → 按 created_at 倒序
  → scene 可选过滤
  → 不返回 content（列表只展示标题+时间，省带宽）
```

### 3.4 详情

```
GET /api/reports/{id}
  → 返回完整 Report（含 engine_data + content）
  → 仅报告所有者可访问
```

### 3.5 软删除

```
DELETE /api/reports/{id}
  → 设置 deleted_at，不物理删除
  → 已删除的报告不出现在列表中
  → 分享链接立即失效
```

### 3.6 分享

```
POST /api/reports/{id}/share
  → 生成 share token（UUID）
  → 返回公开链接

GET /api/reports/shared/{token}
  → 公开访问，无需登录
  → 返回 content（不返回 engine_data）
  → 已删除的报告 → 404
```

分享是可选的——用户主动操作才生成。默认不公开。

---

## 4. 仓库接口

```go
type ReportRepository interface {
    Create(ctx context.Context, report *Report) (int64, error)
    FindByID(ctx context.Context, id int64, userID int64) (*Report, error)
    ListByUser(ctx context.Context, userID int64, scene string, limit, offset int) ([]ReportItem, int, error)
    SoftDelete(ctx context.Context, id int64, userID int64) (bool, error)
    
    // 分享
    CreateShare(ctx context.Context, reportID int64, token string) error
    FindShareByToken(ctx context.Context, token string) (*ReportShare, *Report, error)
    RevokeShare(ctx context.Context, reportID int64) error
}
```

---

## 5. 与 LLM 层的关系

Reports 不直接调 LLM，通过 `llm.Client` 接口：

```go
// internal/app/application/reports/service.go

func (s *Service) Generate(ctx context.Context, req CreateReportRequest, userID int64) (<-chan llm.Chunk, error) {
    // 1. 选模板
    tmpl := s.templates.Get(string(req.Scene), req.SubScene, req.Locale)
    
    // 2. 拼 prompt
    prompt := tmpl.Render(req.EngineData, req.Question)
    
    // 3. 调 LLM
    chunks, err := s.llmClient.Stream(ctx, prompt)
    
    // 4. 收集 chunk，流结束后存库
    go s.collectAndSave(ctx, chunks, req, userID)
    
    return chunks, err
}
```

Reports 的职责是编排——模板选择、prompt 拼装、存储、分享。LLM 调用本身由 `infra/llm/` 封装。

---

## 6. 错误

| Code | HTTP | 含义 |
|------|------|------|
| `invalid_request` | 400 | scene 缺失或非法；engine_data 缺失 |
| `not_found` | 404 | 报告不存在或不属于当前用户 |
| `not_found` | 404 | 分享链接不存在、报告已删除、或 token 无效 |
| `payment_required` | 402 | 免费用户尝试生成付费 scene 的报告 |

### 付费 Scene

| Scene | SubScene | 付费？ |
|-------|----------|--------|
| mingli | mingge | **免费** — 不走 Reports，直接返回 |
| mingli | dayun | 付费 |
| mingli | liunian | 付费 |
| mingli | liuyue | 免费（留存工具） |
| mingli | hehun | 付费 |
| naming | — | 付费 |
| dates | — | 付费 |
| relationship | — | 付费 |
| career | — | 付费 |
| general | — | 免费额度内免费 |

---

## 7. 与现有系统的关系

| | BaZi / Naming / Dates | Reports |
|------|------|------|
| 产出 | 结构化 engine_data | 自然语言 content |
| 确定性 | 相同输入永远相同输出 | 非确定性（LLM） |
| 持久化 | 不持久化（值对象） | 持久化（报告历史） |
| 认证 | mingge 免费匿名 | 全部需登录 |
| 分享 | — | 公开 token 链接 |
