# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: purchase-flow.spec.js >> Purchase flow >> checkout API error shows toast
- Location: e2e/journeys/purchase-flow.spec.js:248:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
Call log:
  - navigating to "http://localhost:8080/zh/chat.html", waiting until "load"

```

# Test source

```ts
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
  195 | 
  196 |     // Session ID should be in sessionStorage.
  197 |     const sid = await page.evaluate(() => sessionStorage.getItem('chatSessionID'));
  198 |     expect(sid).toBe('"persist-session-e2e"');
  199 |   });
  200 | 
  201 |   // ── new chat resets state ──
  202 | 
  203 |   test('new chat after purchase clears buy bar and resets to greeting', async () => {
  204 |     // Mock greeting.
  205 |     await page.route('**/api/agent/greeting', async (route) => {
  206 |       await route.fulfill({
  207 |         status: 200,
  208 |         contentType: 'application/json',
  209 |         body: JSON.stringify({ data: { greeting: '你好，我是灵机。' } }),
  210 |       });
  211 |     });
  212 | 
  213 |     // Mock chat with done.
  214 |     await page.route('**/api/agent/chat', async (route) => {
  215 |       await route.fulfill({
  216 |         status: 200,
  217 |         headers: {
  218 |           'Content-Type': 'text/event-stream',
  219 |           'Cache-Control': 'no-cache',
  220 |           'X-Session-ID': 'new-chat-test',
  221 |         },
  222 |         body: [
  223 |           sseLine({ type: 'thinking' }),
  224 |           sseLine({ type: 'text-delta', content: '分析完成。' }),
  225 |           sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-newchat', amount: 990 } }),
  226 |         ].join(''),
  227 |       });
  228 |     });
  229 | 
  230 |     await page.goto('/zh/chat.html');
  231 |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  232 | 
  233 |     await page.locator('.chat-input-bar input').fill('排盘');
  234 |     await page.locator('.btn-send').click();
  235 | 
  236 |     await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });
  237 | 
  238 |     // Click "新对话" to reset.
  239 |     await page.locator('button:has-text("新对话")').click();
  240 | 
  241 |     // Buy bar should be gone, chips should be back.
  242 |     await expect(page.locator('.buy-card')).not.toBeVisible();
  243 |     await expect(page.locator('.chip-row')).toBeVisible();
  244 |   });
  245 | 
  246 |   // ── error during purchase ──
  247 | 
  248 |   test('checkout API error shows toast', async () => {
  249 |     // Mock greeting.
  250 |     await page.route('**/api/agent/greeting', async (route) => {
  251 |       await route.fulfill({
  252 |         status: 200,
  253 |         contentType: 'application/json',
  254 |         body: JSON.stringify({ data: { greeting: '你好。' } }),
  255 |       });
  256 |     });
  257 | 
  258 |     // Mock chat with done.
  259 |     await page.route('**/api/agent/chat', async (route) => {
  260 |       await route.fulfill({
  261 |         status: 200,
  262 |         headers: {
  263 |           'Content-Type': 'text/event-stream',
  264 |           'Cache-Control': 'no-cache',
  265 |           'X-Session-ID': 'checkout-error-test',
  266 |         },
  267 |         body: [
  268 |           sseLine({ type: 'thinking' }),
  269 |           sseLine({ type: 'text-delta', content: '分析完成。' }),
  270 |           sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-err', amount: 990 } }),
  271 |         ].join(''),
  272 |       });
  273 |     });
  274 | 
  275 |     // Mock checkout to fail.
  276 |     await page.route('**/api/payments/checkout', async (route) => {
  277 |       await route.fulfill({
  278 |         status: 500,
  279 |         contentType: 'application/json',
  280 |         body: JSON.stringify({ error: { code: 'checkout_failed', message: '支付服务暂不可用' } }),
  281 |       });
  282 |     });
  283 | 
> 284 |     await page.goto('/zh/chat.html');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chat.html
  285 |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  286 | 
  287 |     await page.locator('.chat-input-bar input').fill('排盘');
  288 |     await page.locator('.btn-send').click();
  289 | 
  290 |     await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });
  291 | 
  292 |     // Intercept navigation so we stay on page when goPay redirects.
  293 |     await page.route('**/*', async (route) => {
  294 |       // Let the checkout error response through, don't redirect.
  295 |       const url = route.request().url();
  296 |       if (url.includes('/api/payments/checkout')) {
  297 |         await route.continue();
  298 |       } else {
  299 |         await route.continue();
  300 |       }
  301 |     });
  302 | 
  303 |     // Click buy — checkout fails, toast should appear.
  304 |     await page.locator('.btn-buy').click();
  305 | 
  306 |     // Error toast or error in buy card.
  307 |     await expect(page.locator('.error-toast')).toBeVisible({ timeout: 10000 });
  308 |   });
  309 | });
  310 | 
```