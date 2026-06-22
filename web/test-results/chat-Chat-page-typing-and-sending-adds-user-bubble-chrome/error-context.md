# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: chat.spec.js >> Chat page >> typing and sending adds user bubble
- Location: e2e/journeys/chat.spec.js:133:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
Call log:
  - navigating to "http://localhost:8080/zh/chat.html", waiting until "load"

```

# Test source

```ts
  1   | // Chat page tests — POST /api/agent/chat with SSE streaming.
  2   | // Mocks the backend to test all SSE event types and UI states.
  3   | 
  4   | import { test, expect } from '../fixtures.js';
  5   | 
  6   | // SSE event builders
  7   | function sseLine(obj) {
  8   |   return `data: ${JSON.stringify(obj)}\n\n`;
  9   | }
  10  | 
  11  | const mockChartFlow = [
  12  |   sseLine({ type: 'thinking' }),
  13  |   sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  14  |   sseLine({ type: 'text-delta', content: '好的，我来帮您排盘。' }),
  15  |   sseLine({ type: 'phase', content: '正在确认服务类型…' }),
  16  |   sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  17  |   sseLine({ type: 'text-delta', content: '请提供出生信息。' }),
  18  |   sseLine({ type: 'phase', content: '正在查询地理信息…' }),
  19  |   sseLine({ type: 'phase', content: '正在整理出生数据…' }),
  20  |   sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  21  |   sseLine({ type: 'text-delta', content: '信息已收集完成。' }),
  22  |   sseLine({ type: 'phase', content: '正在计算命理数据…' }),
  23  |   sseLine({ type: 'phase', content: '正在生成分析报告…' }),
  24  |   sseLine({ type: 'text-delta', content: '## 八字命盘分析\n\n' }),
  25  |   sseLine({ type: 'text-delta', content: '您的日主为**甲木**，生于寅月。' }),
  26  |   sseLine({ type: 'text-delta', content: '格局为**建禄格**，用神为**火土**。' }),
  27  |   sseLine({ type: 'text-delta', content: '\n\n## 大运分析\n\n' }),
  28  |   sseLine({ type: 'text-delta', content: '当前大运为**戊辰**，财运亨通。' }),
  29  |   sseLine({
  30  |     type: 'done',
  31  |     data: {
  32  |       product: 'chart',
  33  |       order_id: 'test-order-123',
  34  |       amount: 990,
  35  |     },
  36  |   }),
  37  | ].join('');
  38  | 
  39  | const mockErrorFlow = [
  40  |   sseLine({ type: 'thinking' }),
  41  |   sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  42  |   sseLine({ type: 'error', content: '服务暂时不可用，请稍后重试' }),
  43  | ].join('');
  44  | 
  45  | const mockQuestionFlow = [
  46  |   sseLine({ type: 'thinking' }),
  47  |   sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  48  |   sseLine({ type: 'text-delta', content: '请问您的出生年份和性别是什么？' }),
  49  | ].join('');
  50  | 
  51  | async function mockGreeting(page, text = '你好，我是灵机 Liki，精通八字命理与姓名学。有什么可以帮你的？') {
  52  |   await page.route('**/api/agent/greeting', async (route) => {
  53  |     await route.fulfill({
  54  |       status: 200,
  55  |       contentType: 'application/json',
  56  |       body: JSON.stringify({ data: { greeting: text } }),
  57  |     });
  58  |   });
  59  | }
  60  | 
  61  | async function gotoChat(page, lang = 'zh') {
> 62  |   await page.goto(`/${lang}/chat.html`);
      |              ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
  63  |   await page.waitForSelector('.chat-shell');
  64  | }
  65  | 
  66  | async function mockAgentChat(page, body, status = 200) {
  67  |   await page.route('**/api/agent/chat', async (route) => {
  68  |     await route.fulfill({
  69  |       status,
  70  |       headers: {
  71  |         'Content-Type': 'text/event-stream',
  72  |         'Cache-Control': 'no-cache',
  73  |         'X-Session-ID': 'e2e-session-id',
  74  |       },
  75  |       body,
  76  |     });
  77  |   });
  78  | }
  79  | 
  80  | test.describe('Chat page', () => {
  81  |   let page;
  82  | 
  83  |   test.beforeEach(async ({ context }) => {
  84  |     page = await context.newPage();
  85  |   });
  86  | 
  87  |   test.afterEach(async () => {
  88  |     await page.close();
  89  |   });
  90  | 
  91  |   // ── welcome state ──
  92  | 
  93  |   test('greeting bubble and chips render on page load', async () => {
  94  |     await mockGreeting(page);
  95  |     await gotoChat(page);
  96  | 
  97  |     // Greeting is shown as first assistant bubble.
  98  |     await expect(page.locator('.msg-asst')).toBeVisible();
  99  |     await expect(page.locator('.chip')).toHaveCount(3);
  100 |     await expect(page.locator('.chip').nth(0)).toHaveText('🔮 帮我算八字');
  101 |     await expect(page.locator('.chip').nth(1)).toHaveText('💑 看看我和 TA');
  102 |     await expect(page.locator('.chip').nth(2)).toHaveText('📛 帮宝宝起名');
  103 |   });
  104 | 
  105 |   test('chips never show raw i18n keys', async () => {
  106 |     await mockGreeting(page);
  107 |     await gotoChat(page);
  108 | 
  109 |     await expect(page.locator('.chip')).toHaveCount(3);
  110 |     // Verify no raw i18n key is rendered — covers race condition where
  111 |     // i18next.t(key) returns the key itself before resources are loaded.
  112 |     const chipTexts = await page.locator('.chip').allTextContents();
  113 |     for (const text of chipTexts) {
  114 |       expect(text).not.toMatch(/^[a-z]+\.[a-z]+/); // i18n key pattern
  115 |     }
  116 |   });
  117 | 
  118 |   test('chips hide when user sends a message', async () => {
  119 |     await gotoChat(page);
  120 |     await mockAgentChat(page, mockChartFlow);
  121 | 
  122 |     await page.locator('.chat-input-bar input').fill('想看八字');
  123 |     await page.locator('.btn-send').click();
  124 | 
  125 |     // Chips should disappear after user sends a message.
  126 |     await expect(page.locator('.chip-row')).not.toBeVisible();
  127 |     // Messages area should be visible.
  128 |     await expect(page.locator('.chat-messages')).toBeVisible();
  129 |   });
  130 | 
  131 |   // ── message sending ──
  132 | 
  133 |   test('typing and sending adds user bubble', async () => {
  134 |     await gotoChat(page);
  135 |     await mockAgentChat(page, mockChartFlow);
  136 | 
  137 |     await page.locator('.chat-input-bar input').fill('你好');
  138 |     await page.locator('.btn-send').click();
  139 | 
  140 |     await expect(page.locator('.msg-user')).toHaveText('你好');
  141 |   });
  142 | 
  143 |   test('empty input does not send', async () => {
  144 |     await gotoChat(page);
  145 |     await mockAgentChat(page, mockChartFlow);
  146 | 
  147 |     await expect(page.locator('.btn-send')).toBeDisabled();
  148 |   });
  149 | 
  150 |   test('clicking a chip sends its message', async () => {
  151 |     await gotoChat(page);
  152 |     await mockAgentChat(page, mockChartFlow);
  153 | 
  154 |     await page.locator('.chip').nth(0).click();
  155 | 
  156 |     await expect(page.locator('.msg-user')).toHaveText('我想排八字命盘');
  157 |   });
  158 | 
  159 |   // ── SSE streaming ──
  160 | 
  161 |   test('text-delta events render markdown content incrementally', async () => {
  162 |     await mockGreeting(page);
```