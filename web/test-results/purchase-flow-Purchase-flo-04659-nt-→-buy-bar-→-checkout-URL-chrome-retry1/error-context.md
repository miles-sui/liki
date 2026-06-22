# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: purchase-flow.spec.js >> Purchase flow >> full flow: chat message → done event → buy bar → checkout URL
- Location: e2e/journeys/purchase-flow.spec.js:48:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
Call log:
  - navigating to "http://localhost:8080/zh/chat.html", waiting until "load"

```

# Test source

```ts
  1   | // Full purchase flow E2E — chat → compute → buy → checkout → report.
  2   | // Mocks all backend APIs to verify frontend state transitions end-to-end.
  3   | 
  4   | import { test, expect } from '../fixtures.js';
  5   | 
  6   | function sseLine(obj) {
  7   |   return `data: ${JSON.stringify(obj)}\n\n`;
  8   | }
  9   | 
  10  | const CHART_DATA = JSON.stringify({
  11  |   chart: {
  12  |     chart: {
  13  |       riyuan: '庚金',
  14  |       nianzhu: { gan: '庚', zhi: '午', shishen: [{ source: 'gan', shishen: '比肩' }], canggan: {}, nayin: '路旁土', shensha: [] },
  15  |       yuezhu: { gan: '壬', zhi: '午', shishen: [{ source: 'gan', shishen: '食神' }], canggan: {}, nayin: '杨柳木', shensha: [] },
  16  |       rizhu: { gan: '庚', zhi: '午', shishen: [{ source: 'gan', shishen: '日主' }], canggan: {}, nayin: '路旁土', shensha: [] },
  17  |       shizhu: { gan: '丙', zhi: '子', shishen: [{ source: 'gan', shishen: '七杀' }], canggan: {}, nayin: '涧下水', shensha: [] },
  18  |       yong_shen: { fuyi: { qiangruo: '身弱', geju: '七杀格', yong: '土', xi: '金', ji: '木' }, tiaohou: { season: '夏', yong: '水', xi: '金', ji: '火' } },
  19  |     },
  20  |   },
  21  | });
  22  | 
  23  | const REPORT_LLM = `# 八字命理报告
  24  | 
  25  | ## 第一章 · 命盘总览
  26  | 
  27  | 您的日主为**庚金**，生于午月。庚金刚健，午月火旺。
  28  | 
  29  | ## 第二章 · 综合分析
  30  | 
  31  | 身弱七杀格，以土为用，喜金帮身。`;
  32  | 
  33  | const ORDER_ID = 'e2e-baad-f00d-001';
  34  | 
  35  | test.describe('Purchase flow', () => {
  36  |   let page;
  37  | 
  38  |   test.beforeEach(async ({ context }) => {
  39  |     page = await context.newPage();
  40  |   });
  41  | 
  42  |   test.afterEach(async () => {
  43  |     await page.close();
  44  |   });
  45  | 
  46  |   // ── chat → done → buy bar visible ──
  47  | 
  48  |   test('full flow: chat message → done event → buy bar → checkout URL', async () => {
  49  |     // Mock greeting.
  50  |     await page.route('**/api/agent/greeting', async (route) => {
  51  |       await route.fulfill({
  52  |         status: 200,
  53  |         contentType: 'application/json',
  54  |         body: JSON.stringify({ data: { greeting: '你好，我是灵机。' } }),
  55  |       });
  56  |     });
  57  | 
  58  |     // Mock chat SSE: question → compute result → done with purchase.
  59  |     await page.route('**/api/agent/chat', async (route) => {
  60  |       await route.fulfill({
  61  |         status: 200,
  62  |         headers: {
  63  |           'Content-Type': 'text/event-stream',
  64  |           'Cache-Control': 'no-cache',
  65  |           'X-Session-ID': 'e2e-purchase-session',
  66  |         },
  67  |         body: [
  68  |           sseLine({ type: 'thinking' }),
  69  |           sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  70  |           sseLine({ type: 'text-delta', content: '好的，您的八字排盘如下：' }),
  71  |           sseLine({ type: 'text-delta', content: '\n\n## 命盘分析\n\n' }),
  72  |           sseLine({ type: 'text-delta', content: '您的日主为庚金，生于午月…' }),
  73  |           sseLine({ type: 'phase', content: '正在生成分析报告…' }),
  74  |           sseLine({ type: 'text-delta', content: '如需查看完整报告（五行、十神、大运、流年），可以购买解锁。' }),
  75  |           sseLine({
  76  |             type: 'done',
  77  |             data: { product: 'chart', order_id: ORDER_ID, amount: 990 },
  78  |           }),
  79  |         ].join(''),
  80  |       });
  81  |     });
  82  | 
  83  |     // Mock checkout API.
  84  |     await page.route('**/api/payments/checkout', async (route) => {
  85  |       await route.fulfill({
  86  |         status: 200,
  87  |         contentType: 'application/json',
  88  |         body: JSON.stringify({
  89  |           data: { checkout_url: 'https://pay.dodopayments.com/checkout/e2e-test' },
  90  |         }),
  91  |       });
  92  |     });
  93  | 
> 94  |     await page.goto('/zh/chat.html');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
  95  |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  96  | 
  97  |     // Send a message.
  98  |     await page.locator('.chat-input-bar input').fill('帮我排盘');
  99  |     await page.locator('.btn-send').click();
  100 | 
  101 |     // Buy bar should appear after done event.
  102 |     const buyCard = page.locator('.buy-card');
  103 |     await expect(buyCard).toBeVisible({ timeout: 10000 });
  104 | 
  105 |     // Verify buy card content.
  106 |     await expect(buyCard).toContainText('69.00');
  107 |     await expect(page.locator('.btn-buy')).toContainText('查看完整报告');
  108 | 
  109 |     // Click the pay button → should trigger checkout and navigate.
  110 |     // We intercept the navigation by not actually loading the checkout URL.
  111 |     await page.route('https://pay.dodopayments.com/checkout/e2e-test', async (route) => {
  112 |       await route.fulfill({ status: 200, body: '<html><body>Mock Dodo Checkout</body></html>' });
  113 |     });
  114 | 
  115 |     await page.locator('.btn-buy').click();
  116 |     // Should navigate to Dodo checkout.
  117 |     await page.waitForURL('**/checkout/e2e-test', { timeout: 10000 });
  118 |   });
  119 | 
  120 |   // ── report page after purchase ──
  121 | 
  122 |   test('report page shows paid report after purchase flow', async () => {
  123 |     // Mock report API.
  124 |     await page.route(`**/api/reports/${ORDER_ID}`, async (route) => {
  125 |       await route.fulfill({
  126 |         status: 200,
  127 |         contentType: 'application/json',
  128 |         body: JSON.stringify({
  129 |           data: {
  130 |             order_id: ORDER_ID,
  131 |             product: 'chart',
  132 |             chart_json: CHART_DATA,
  133 |             llm_json: REPORT_LLM,
  134 |             status: 'paid',
  135 |             amount: 990,
  136 |             currency: 'CNY',
  137 |           },
  138 |         }),
  139 |       });
  140 |     });
  141 | 
  142 |     await page.goto(`/zh/report/${ORDER_ID}`);
  143 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  144 | 
  145 |     // Report content should be visible.
  146 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  147 | 
  148 |     // Verify markdown rendered.
  149 |     const interpretation = page.locator('#chart-interpretation');
  150 |     await expect(interpretation).toContainText('命盘总览');
  151 |     await expect(interpretation).toContainText('庚金');
  152 | 
  153 |     // Summary cards from chart_json — use .first() for strict-mode safety.
  154 |     await expect(page.locator('.summary-grid').first()).toBeVisible();
  155 |     await expect(page.locator('.summary-card .value').first()).toContainText('庚金');
  156 |   });
  157 | 
  158 |   // ── session ID persistence across chat → buy ──
  159 | 
  160 |   test('session ID persists in sessionStorage after purchase', async () => {
  161 |     // Mock greeting.
  162 |     await page.route('**/api/agent/greeting', async (route) => {
  163 |       await route.fulfill({
  164 |         status: 200,
  165 |         contentType: 'application/json',
  166 |         body: JSON.stringify({ data: { greeting: '你好。' } }),
  167 |       });
  168 |     });
  169 | 
  170 |     // Mock chat with done.
  171 |     await page.route('**/api/agent/chat', async (route) => {
  172 |       await route.fulfill({
  173 |         status: 200,
  174 |         headers: {
  175 |           'Content-Type': 'text/event-stream',
  176 |           'Cache-Control': 'no-cache',
  177 |           'X-Session-ID': 'persist-session-e2e',
  178 |         },
  179 |         body: [
  180 |           sseLine({ type: 'thinking' }),
  181 |           sseLine({ type: 'text-delta', content: '好的。' }),
  182 |           sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-persist', amount: 990 } }),
  183 |         ].join(''),
  184 |       });
  185 |     });
  186 | 
  187 |     await page.goto('/zh/chat.html');
  188 |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  189 | 
  190 |     await page.locator('.chat-input-bar input').fill('排盘');
  191 |     await page.locator('.btn-send').click();
  192 | 
  193 |     // Wait for buy bar.
  194 |     await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });
```