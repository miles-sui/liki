# Commerce 聚合 — 实现规格

> 自愿捐赠。Dodo Payments 作为支付网关。三轮：文档 → 测试 → 实现。

---

## 1. 领域模型

所有功能免费开放。收入来自自愿捐赠。支付渠道（Dodo Payments）通过官方 Go SDK (`github.com/dodopayments/dodopayments-go`) 集成。SDK 封装 checkout session 创建和 webhook 验签（Standard Webhooks HMAC-SHA256）。应用层通过 `application/commerce/` 中的窄接口消费，不直接依赖 SDK 类型。

`donations` 表记录捐赠流水：`id`、`user_id`、`amount`（美分，>0）、`created_at`。

User 通过 LEFT JOIN 读取 `MIN(donations.created_at)`，映射为 `Supporter` 标记：有捐赠记录时为支持者。`supporter_since` 取首次捐赠时间。

| 概念 | 存储 / 位置 | 说明 |
|------|------|------|
| 捐赠记录 | `donations` 表 | `user_id`、`amount`、`created_at` |
| 支持者标记 | User 查询时 LEFT JOIN 读取 | `supporter: true/false`，`supporter_since: 时间或 null` |
| 支付会话 | Dodo Payments | 创建 checkout → 用户跳转 → webhook 回调 |

---

## 2. 功能开放

全部功能免费：

- 评估 + identity
- 全部类型工具
- 他评创建及提交
- 自评 vs 他评对比
- Bond（不限次数）
- 全年 Flow 预报 + 元素提示
- 历次变迁 + 两次对比
- 他评深度分析
- Bond 历史
- 完整数据导出
- 原始评估记录 JSON（GDPR 数据携带权）

---

## 3. 捐赠档位

三档固定金额，无自定义：

| 档位 | 展示 | 美分值 |
|------|------|--------|
| 基础 | $9.90 | `990` |
| 支持 | $19.90 | `1990` |
| 慷慨 | $29.90 | `2990` |

前端展示三档选择器，后端 `POST /api/payments/checkout` 接收美分值并校验合法性。

---

## 4. 捐赠 checkout

创建捐赠会话，返回支付页面 URL。请求体含 `amount`（美分，三档之一）。用户跳转至 Dodo checkout 页面完成付款。

### 4.1 Webhook 回调

Dodo webhook 回调流程：
1. 验签——Standard Webhooks 规范 HMAC-SHA256，验签失败返回 401
2. 解析事件类型——仅处理 `payment.succeeded`
3. 从 event data 提取 `user_id`（metadata）、`amount`、`email`
4. 写入 `donations`：`user_id`、`amount`
5. 同一用户可多次捐赠，每次追加一条记录

### 4.2 支持者标记

`GET /api/users/me` 返回 `supporter` 和 `supporter_since` 字段。纯粹的身份标识，不给功能特权。

---

## 5. 捐赠感谢邮件

捐赠成功后 best-effort 发送感谢邮件。`infra/resend.Client` 和 `infra/tencent.Client` 均需实现 `SendThankYouEmail(ctx, to, locale)` 方法。按 locale 选择 EN/ZH 模板。未配置 API key 时静默跳过。邮件失败不回滚捐赠记录。

---

## 参见

- [INDEX](../INDEX.md) — 全局约定
- [API](../API.md) — HTTP 契约
- [user](user.md) — User 聚合
