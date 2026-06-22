# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: chart.spec.js >> Chart demo page >> day zhu shows 戊 (比肩) → 辰 with 华盖 shensha
- Location: e2e/journeys/chart.spec.js:103:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chart.html
Call log:
  - navigating to "http://localhost:8080/zh/chart.html", waiting until "load"

```

# Test source

```ts
  4   | import { test, expect } from '../fixtures.js';
  5   | 
  6   | test.describe('Chart demo page', () => {
  7   |   let page;
  8   | 
  9   |   test.beforeEach(async ({ context }) => {
  10  |     page = await context.newPage();
  11  |   });
  12  | 
  13  |   test.afterEach(async () => {
  14  |     await page.close();
  15  |   });
  16  | 
  17  |   // ── content rendering ──
  18  | 
  19  |   test('ZH page renders all sections: summary, zhu, elements, dayun, interpretation, CTA', async () => {
  20  |     await page.goto('/zh/chart.html');
  21  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  22  | 
  23  |     // Header — use h1[data-i18n] to avoid strict-mode conflict with <liki-header> brand h1.
  24  |     await expect(page.locator('h1[data-i18n]')).toContainText('八字报告');
  25  | 
  26  |     // Summary cards.
  27  |     await expect(page.locator('.summary-grid')).toBeVisible();
  28  |     const summaryCards = page.locator('.summary-card');
  29  |     await expect(summaryCards).toHaveCount(4);
  30  | 
  31  |     // Pillars — mobile cards exist in DOM + desktop table is visible.
  32  |     await expect(page.locator('.zhu-cards')).toBeAttached();
  33  |     await expect(page.locator('.zhu-card')).toHaveCount(4);
  34  |     await expect(page.locator('.hide-mobile table')).toBeVisible();
  35  |     await expect(page.locator('.hide-mobile tbody tr')).toHaveCount(4);
  36  | 
  37  |     // Element distribution.
  38  |     await expect(page.locator('h2:has-text("五行分布")')).toBeVisible();
  39  | 
  40  |     // Dayun.
  41  |     await expect(page.locator('h2:has-text("大运")')).toBeVisible();
  42  | 
  43  |     // LLM interpretation.
  44  |     await expect(page.locator('.report-content')).toBeVisible();
  45  |     await expect(page.locator('.report-content h3')).toHaveCount(5);
  46  | 
  47  |     // CTA bar.
  48  |     await expect(page.locator('.cta-bar')).toBeVisible();
  49  |     await expect(page.locator('.cta-bar a.btn-primary')).toHaveAttribute('href', '/chat.html');
  50  | 
  51  |     // Trust section — .hero-numbers may not be present on all page variants.
  52  |     await expect(page.locator('.trust-badges')).toBeVisible();
  53  |   });
  54  | 
  55  |   test('EN page renders with English labels', async () => {
  56  |     await page.goto('/en/chart.html');
  57  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  58  | 
  59  |     await expect(page.locator('h1[data-i18n]')).toContainText('BaZi');
  60  |     // Table headers should be in English.
  61  |     await expect(page.locator('.hide-mobile thead')).toContainText('Pillar');
  62  |   });
  63  | 
  64  |   // ── print and share buttons ──
  65  | 
  66  |   test('print button exists and triggers print dialog', async () => {
  67  |     await page.goto('/zh/chart.html');
  68  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  69  | 
  70  |     const printBtn = page.locator('.btn-print');
  71  |     await expect(printBtn).toBeVisible();
  72  |     await expect(printBtn).toContainText('打印');
  73  |   });
  74  | 
  75  |   test('share button exists', async () => {
  76  |     await page.goto('/zh/chart.html');
  77  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  78  | 
  79  |     const shareBtn = page.locator('.btn-share');
  80  |     await expect(shareBtn).toBeVisible();
  81  |     await expect(shareBtn).toContainText('分享');
  82  |   });
  83  | 
  84  |   // ── zhu table data correctness ──
  85  | 
  86  |   test('year zhu shows correct data (甲 七杀 → 子, 癸 · 海中金)', async () => {
  87  |     await page.goto('/zh/chart.html');
  88  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  89  | 
  90  |     // Mobile zhu cards.
  91  |     const yearCard = page.locator('.zhu-card').nth(0);
  92  |     await expect(yearCard.locator('.zc-label')).toContainText('年柱');
  93  |     await expect(yearCard.locator('.zc-gan')).toContainText('甲');
  94  |     await expect(yearCard.locator('.zc-zhi')).toContainText('子');
  95  |     await expect(yearCard.locator('.zc-detail')).toContainText('海中金');
  96  | 
  97  |     // Desktop table.
  98  |     const yearRow = page.locator('.hide-mobile tbody tr').nth(0);
  99  |     await expect(yearRow).toContainText('七杀');
  100 |     await expect(yearRow).toContainText('海中金');
  101 |   });
  102 | 
  103 |   test('day zhu shows 戊 (比肩) → 辰 with 华盖 shensha', async () => {
> 104 |     await page.goto('/zh/chart.html');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/chart.html
  105 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  106 | 
  107 |     const dayRow = page.locator('.hide-mobile tbody tr').nth(2);
  108 |     await expect(dayRow).toContainText('华盖');
  109 |   });
  110 | 
  111 |   // ── CTA link ──
  112 | 
  113 |   test('CTA button links to chat page', async () => {
  114 |     await page.goto('/zh/chart.html');
  115 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  116 | 
  117 |     const ctaLink = page.locator('.cta-bar a.btn-primary');
  118 |     await expect(ctaLink).toHaveAttribute('href', '/chat.html');
  119 |   });
  120 | 
  121 |   // ── analytics fires on page load ──
  122 | 
  123 |   test('pageview analytics fires on load', async () => {
  124 |     let analyticsBody = null;
  125 |     await page.route('**/api/analytics/pageview', async (route) => {
  126 |       analyticsBody = JSON.parse(route.request().postData() || '{}');
  127 |       await route.fulfill({ status: 200, contentType: 'application/json', body: '{"data":{"ok":true}}' });
  128 |     });
  129 | 
  130 |     await page.goto('/zh/chart.html');
  131 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  132 | 
  133 |     // Analytics should have been called with the correct path.
  134 |     await page.waitForTimeout(500);
  135 |     expect(analyticsBody).not.toBeNull();
  136 |     expect(analyticsBody.path).toContain('chart.html');
  137 |   });
  138 | });
  139 | 
```