# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: report.spec.js >> Report page >> transitions from generating polling to ready when llm_json arrives
- Location: e2e/journeys/report.spec.js:283:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/report/cafe-456
Call log:
  - navigating to "http://localhost:8080/zh/report/cafe-456", waiting until "load"

```

# Test source

```ts
  219 |   });
  220 | 
  221 |   // ── naming report ──
  222 | 
  223 |   test('paid naming report renders candidates', async () => {
  224 |     await mockNamingReport(page);
  225 |     await page.goto('/zh/report/babe-face');
  226 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  227 | 
  228 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  229 | 
  230 |     const interpretation = page.locator('#naming-interpretation');
  231 |     await expect(interpretation).toContainText('起名报告');
  232 |     await expect(interpretation).toContainText('陈明远');
  233 | 
  234 |     await expect(page.locator('#report-header-title')).toContainText('起名报告');
  235 |   });
  236 | 
  237 |   // ── polling → ready transition ──
  238 | 
  239 |   test('transitions from payment polling to ready when order is paid', async () => {
  240 |     let callCount = 0;
  241 |     await page.route('**/api/reports/abba-123', async (route) => {
  242 |       callCount++;
  243 |       if (callCount <= 2) {
  244 |         // First poll: still pending.
  245 |         await route.fulfill({
  246 |           status: 200,
  247 |           contentType: 'application/json',
  248 |           body: JSON.stringify({
  249 |             data: { order_id: 'abba-123', product: 'chart', status: 'pending' },
  250 |           }),
  251 |         });
  252 |       } else {
  253 |         // Subsequent polls: paid with report data.
  254 |         await route.fulfill({
  255 |           status: 200,
  256 |           contentType: 'application/json',
  257 |           body: JSON.stringify({
  258 |             data: {
  259 |               order_id: 'abba-123',
  260 |               product: 'chart',
  261 |               chart_json: CHART_JSON,
  262 |               llm_json: LLM_JSON,
  263 |               status: 'paid',
  264 |             },
  265 |           }),
  266 |         });
  267 |       }
  268 |     });
  269 | 
  270 |     await page.goto('/zh/report/abba-123');
  271 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  272 | 
  273 |     // Starts in payment polling phase.
  274 |     await expect(page.locator('.status-card.status-payment')).toBeVisible({ timeout: 5000 });
  275 | 
  276 |     // Eventually transitions to ready.
  277 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 15000 });
  278 |     await expect(page.locator('.summary-grid').first()).toBeVisible();
  279 |   });
  280 | 
  281 |   // ── generating → ready transition ──
  282 | 
  283 |   test('transitions from generating polling to ready when llm_json arrives', async () => {
  284 |     let callCount = 0;
  285 |     await page.route('**/api/reports/cafe-456', async (route) => {
  286 |       callCount++;
  287 |       if (callCount <= 3) {
  288 |         // Paid but no llm_json yet — generating status.
  289 |         await route.fulfill({
  290 |           status: 200,
  291 |           contentType: 'application/json',
  292 |           body: JSON.stringify({
  293 |             data: {
  294 |               order_id: 'cafe-456',
  295 |               product: 'chart',
  296 |               chart_json: CHART_JSON,
  297 |               status: 'paid',
  298 |             },
  299 |           }),
  300 |         });
  301 |       } else {
  302 |         // llm_json ready.
  303 |         await route.fulfill({
  304 |           status: 200,
  305 |           contentType: 'application/json',
  306 |           body: JSON.stringify({
  307 |             data: {
  308 |               order_id: 'cafe-456',
  309 |               product: 'chart',
  310 |               chart_json: CHART_JSON,
  311 |               llm_json: LLM_JSON,
  312 |               status: 'paid',
  313 |             },
  314 |           }),
  315 |         });
  316 |       }
  317 |     });
  318 | 
> 319 |     await page.goto('/zh/report/cafe-456');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/report/cafe-456
  320 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  321 | 
  322 |     // Starts in generating phase.
  323 |     await expect(page.locator('.status-card.status-generating')).toBeVisible({ timeout: 5000 });
  324 | 
  325 |     // Eventually transitions to ready with transition animation.
  326 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 15000 });
  327 |     await expect(page.locator('.summary-grid').first()).toBeVisible();
  328 |   });
  329 | 
  330 |   // ── save banner interactions ──
  331 | 
  332 |   test('banner close persists across page reloads via sessionStorage', async () => {
  333 |     await mockChartReport(page, 'face-fade');
  334 |     await page.goto('/zh/report/face-fade');
  335 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  336 | 
  337 |     // Banner visible initially.
  338 |     await expect(page.locator('.save-banner')).toBeVisible({ timeout: 5000 });
  339 | 
  340 |     // Close it.
  341 |     await page.locator('#banner-close-btn').click();
  342 |     await expect(page.locator('.save-banner')).not.toBeVisible({ timeout: 3000 });
  343 | 
  344 |     // Reload — banner stays hidden.
  345 |     await page.reload();
  346 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  347 |     await page.waitForTimeout(500);
  348 |     await expect(page.locator('.save-banner')).not.toBeVisible();
  349 |   });
  350 | 
  351 |   // ── copy link ──
  352 | 
  353 |   test('copy link button updates text on click', async () => {
  354 |     await mockChartReport(page, 'cafe-feed');
  355 |     await page.goto('/zh/report/cafe-feed');
  356 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  357 | 
  358 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  359 | 
  360 |     // Click copy link in the save banner.
  361 |     const copyBtn = page.locator('#banner-copy-btn');
  362 |     await expect(copyBtn).toBeVisible();
  363 | 
  364 |     // Clipboard permission is auto-granted in Playwright test context.
  365 |     await copyBtn.click();
  366 |     // Button text changes to '已复制' briefly, then back.
  367 |     await expect(copyBtn).toContainText('已复制', { timeout: 3000 });
  368 |   });
  369 | });
  370 | 
```