# LLM Integration — 实现规格

> 引擎做计算，LLM 做解读。引擎产出结构化数据（干支、十神、五行、冲合），LLM 扮演"命理师"角色，把多维信息融合成自然语言解读。LLM 只做最后一层，不碰计算。

---

## 1. 角色定位

```
引擎（确定性计算）                          LLM（综合解读）
─────────────────                          ──────────────
八字排盘：四柱、十神、纳音、五行分布
大运：起运年龄、十步大运干支              →  命格解读
流年：当年干支、十神、冲合                →  流年解读
合婚：天干合、地支冲合、十神互见、        →  合婚解读
      十二长生、藏干、神煞
Bond 动力学：δ 向量、生克矩阵             →  关系解读
起名引擎：Top 100、三才五格、音韵         →  起名方案解读
择日日历：建除、十神、冲合、神煞          →  择日解读
─────────────────                          ──────────────
纯函数，相同输入永远相同输出              有创造力，每次可能不同措辞
单元测试                                     人工审阅 + prompt eval
```

**LLM 不做的：**
- 不排盘、不算十神、不算笔画——引擎已算完
- 不排序候选——排序是引擎的确定性逻辑
- 不替用户做决策——只呈现，不说"你应该选这个"

**LLM 做的：**
- 多维度数据融合成流畅解读
- 用"命理师"的语调和隐喻说话（"你的甲木像一棵大树..."）
- 识别维度之间的冲突或呼应（"日柱合土，年柱却子午冲——根基不稳但核心吸引强"）
- 给出有温度的建议

---

## 2. 调用场景

| 场景 | 触发 | LLM 需要融合的维度 | 流式 |
|------|------|-------------------|------|
| `bazi_mingge` | 命格解读 | 日主 + 月令 + 强弱 + 用神 + 喜神 + 五行分布 + 十神 + 纳音 | 是 |
| `bazi_dayun` | 大运解读 | 当前大运干支 + 十神 + 日主生克 + 前后大运衔接 | 是 |
| `bazi_liunian` | 流年解读 | 流年干支 + 十神 + 大运关系 + 四柱冲合 + 神煞 | 是 |
| `hehun` | 合婚解读 | 天干合/克 + 四柱地支冲合刑害 + 十神互见 + 十二长生 + 藏干 + 神煞 + 日主生克 | 是 |
| `naming` | 起名方案 | Top 20 候选 + 三才五格 + 音韵 + 字义 + 五行补益方向 | 是 |
| `huangli` | 择日方案 | 选中日 + 建除 + 十神 + 冲合 + 神煞 + 事件五行 | 是 |
| `relationship` | 关系解读 | Bond δ + profile + 元素生克方向 | 是 |
| `career` | 职业建议 | profile + types.yaml 映射 | 是 |
| `qa` | 自由问答 | 用户问题 + 相关引擎数据（八字/profile/bond 等） | 是 |

**起名终筛仍然是纯模板。** 字义冲突和谐音黑白名单存静态表，每次请求实时过滤。

---

## 3. 提示词模板

模板存 `configs/llm/{scene}_{lang}.yaml`，构建时 `//go:embed` 嵌入。

### 3.1 模板结构

```yaml
# configs/llm/bazi_mingge_zh-CN.yaml

role: |
  你是一位资深命理师，精通子平八字。你的解读风格：
  - 用自然隐喻（甲木=大树，丙火=太阳，戊土=大地）
  - 不说教，不吓人，不给绝对判断
  - 关注用户能做什么，而非注定什么
  - 语言流畅有温度，像朋友聊天

input_template: |
  请解读以下命盘：

  【基本盘】
  日主：{day_master}（{day_master_element}）
  月令：{month_branch}（{month_element}）
  身强/身弱：{strength}
  用神：{yong_shen}
  喜神：{xi_shen}

  【四柱八字】
  {pillars_table}

  【十神分布】
  {ten_gods_table}

  【五行分布】
  木 {wood_count}  火 {fire_count}  土 {earth_count}  金 {metal_count}  水 {water_count}
  
  【日柱纳音】
  {na_yin}

output_guide: |
  请按以下结构输出：
  1. 日主性格——用隐喻，1-2 句
  2. 命格特点——身强/弱意味着什么
  3. 五行平衡——哪个元素在帮你，哪个在消耗你
  4. 用神方向——你需要什么样的能量来平衡
  5. 一句话总结
```

### 3.2 合婚模板特点

合婚维度最多，模板最关键：

```yaml
# configs/llm/hehun_zh-CN.yaml

input_template: |
  请解读这对伴侣的合盘：

  【甲方】{name_a}
  日主：{day_master_a}（{element_a}）
  四柱：{pillars_a}
  
  【乙方】{name_b}
  日主：{day_master_b}（{element_b}）
  四柱：{pillars_b}

  【天干关系】
  {gan_relations}

  【地支关系】
  年柱：{year_zhi_relation}
  月柱：{month_zhi_relation}
  日柱：{day_zhi_relation}
  时柱：{hour_zhi_relation}

  【十神互见】
  A 看 B：{tengods_a_on_b}
  B 看 A：{tengods_b_on_a}

  【十二长生】
  A 日主在 B 日支：{zhang_sheng_a}
  B 日主在 A 日支：{zhang_sheng_b}

  【日主生克】
  {day_master_relation}

  【纳音】
  {na_yin_relation}

  【神煞交集】
  {shen_sha_intersection}

output_guide: |
  请按以下结构输出：
  1. 核心吸引——天干+日柱组合，最关键的连接是什么
  2. 互补与摩擦——哪里生，哪里克，不是好事或坏事，是动力
  3. 十二长生匹配度——两个人当前的生命阶段是否同步
  4. 注意事项——冲/刑/害在哪个宫位，代表什么层面的分歧
  5. 相处建议——基于生克方向的具体行动建议
```

### 3.3 关键规则

- `role` 不写"你是 AI 助手"，直接定义角色和语调
- `input_template` 用 `{key}` 占位符，引擎填入结构化数据
- `output_guide` 引导结构但不强制——LLM 按需要调整
- 模板与代码解耦，改文案不改代码
- 每个 scene + locale 一个文件

---

## 4. SSE 流式契约

所有解读场景使用 SSE 流式输出，通过 Reports 系统统一入口。

### 4.1 端点

```
POST /api/reports
  body: { "scene": "bazi_mingge", "locale": "zh-CN", "context": { ... } }
  → SSE stream
```

`context` 包含该场景所需的全部引擎数据，由引擎计算完后构造。

### 4.2 SSE 事件

```
event: chunk
data: {"text": "你是甲木日主，生于寅月。"}

event: chunk
data: {"text": "甲木像一棵挺拔的大树..."}

event: done
data: {"report_id": 42}
```

### 4.3 错误

```
event: error
data: {"code": "internal", "message": "An unexpected error occurred"}
```

发送 error 后立即关闭连接。

### 4.4 客户端

```js
const es = new EventSource('/api/reports');
es.addEventListener('chunk', (e) => appendText(JSON.parse(e.data).text));
es.addEventListener('done', (e) => saveReport(JSON.parse(e.data).report_id));
es.addEventListener('error', (e) => showError(JSON.parse(e.data)));
```

---

## 5. 非流式调用

### 5.1 自由问答

`/app#ask` 入口，用户输入自由文本问题。

```
POST /api/qa
  body: { "question": "我适合五月结婚吗？", "locale": "zh-CN" }
  → SSE stream（同上格式）
```

后端先做意图识别（八字/择日/关系/职业），拉取相关引擎数据，拼入 prompt context。

### 5.2 意图识别

```go
type Intent string

const (
    IntentBazi     Intent = "mingli"      // 八字相关 → 拉 chart
    IntentHuangli  Intent = "huangli"   // 择日相关 → 拉 chart + calendar
    IntentRelation Intent = "relation"  // 关系相关 → 拉 bond + profile
    IntentCareer   Intent = "career"    // 职业相关 → 拉 profile + types
    IntentGeneral  Intent = "general"   // 通用 → 仅当前上下文
)
```

意图识别用关键词匹配：问"结婚""日子"→ huangli，"合不合""配不配"→ relation，"适合什么工作"→ career，其余走 general。

---

## 6. 模型与配置

```yaml
# configs/llm/models.yaml
provider: anthropic
model: claude-haiku-4-5     # 解读不需要复杂推理，快速够用
max_tokens: 2048
timeout: 15s
```

---

## 7. 错误与降级

| 情况 | 行为 |
|------|------|
| LLM 超时（15s） | 发送 error 事件，关闭 SSE。前端展示引擎数据 + "解读生成失败，请重试" |
| Provider 不可用 | 同上 |
| Token 超限 | 截断 context（优先保留日柱和关键维度），记录 warning 日志 |
| 返回内容异常（空/乱码） | 同上 |

**降级原则：引擎数据永不可被 LLM 阻断。** LLM 失败，前端展示引擎计算的结构化数据，只是缺少自然语言解读。

---

## 8. 可观测性

```go
log.Printf("llm call scene=%s locale=%s model=%s tokens_in=%d tokens_out=%d latency_ms=%d err=%v",
    scene, locale, model, inputTokens, outputTokens, latencyMs, err)
```

---

## 9. 客户端封装

```go
// internal/app/infra/llm/client.go

type Client interface {
    // Stream 流式调用，通过 channel 返回 chunk
    Stream(ctx context.Context, scene string, locale string, context map[string]any) (<-chan Chunk, error)
    
    // Complete 一次性调用（备用）
    Complete(ctx context.Context, scene string, locale string, context map[string]any) (string, error)
}

type Chunk struct {
    Text  string
    Error error
    Done  bool
}
```

`infra/llm/` 下实现 provider 适配，不与 engine 耦合。

---

## 10. 提示词版本管理

- 模板文件在 `configs/llm/` 下，Git 版本管理
- 修改提示词不走代码发布——改 YAML → 重启服务生效
- 每个模板有独立版本注释：`# version: 1`，升级时递增
