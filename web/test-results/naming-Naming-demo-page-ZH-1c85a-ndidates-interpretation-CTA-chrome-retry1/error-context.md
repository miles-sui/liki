# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: naming.spec.js >> Naming demo page >> ZH page renders all sections: summary, candidates, interpretation, CTA
- Location: e2e/journeys/naming.spec.js:19:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/naming.html
Call log:
  - navigating to "http://localhost:8080/zh/naming.html", waiting until "load"

```

# Test source

```ts
  1   | // Naming demo page tests — validates static naming demo report rendering, i18n, interactions.
  2   | // Source: naming.html, demo-utils.js, i18n.js.
  3   | 
  4   | import { test, expect } from '../fixtures.js';
  5   | 
  6   | test.describe('Naming demo page', () => {
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
  19  |   test('ZH page renders all sections: summary, candidates, interpretation, CTA', async () => {
> 20  |     await page.goto('/zh/naming.html');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/naming.html
  21  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  22  | 
  23  |     // Header — use h1[data-i18n] to avoid strict-mode conflict with <liki-header> brand h1.
  24  |     await expect(page.locator('h1[data-i18n]')).toContainText('起名报告');
  25  | 
  26  |     // Summary cards.
  27  |     await expect(page.locator('.summary-grid')).toBeVisible();
  28  | 
  29  |     // Name candidate cards.
  30  |     const nameCards = page.locator('.name-card');
  31  |     await expect(nameCards).toHaveCount(3);
  32  | 
  33  |     // LLM interpretation.
  34  |     await expect(page.locator('.report-content')).toBeVisible();
  35  |   });
  36  | 
  37  |   test('EN page renders with English labels', async () => {
  38  |     await page.goto('/en/naming.html');
  39  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  40  | 
  41  |     await expect(page.locator('h1[data-i18n]')).toContainText('Naming');
  42  |   });
  43  | 
  44  |   // ── name candidate cards ──
  45  | 
  46  |   test('first name card shows 陈煜霖 with correct details', async () => {
  47  |     await page.goto('/zh/naming.html');
  48  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  49  | 
  50  |     const firstCard = page.locator('.name-card').nth(0);
  51  |     await expect(firstCard.locator('.name-title')).toContainText('陈煜霖');
  52  |     await expect(firstCard.locator('.tag-green')).toContainText('大吉');
  53  |     // Character details.
  54  |     await expect(firstCard.locator('.name-card-chars')).toContainText('煜 (13画)');
  55  |     await expect(firstCard.locator('.name-card-chars')).toContainText('霖 (16画)');
  56  |     // Wu Ge.
  57  |     await expect(firstCard.locator('.name-card-wuge')).toBeVisible();
  58  |   });
  59  | 
  60  |   test('all three candidates have different names', async () => {
  61  |     await page.goto('/zh/naming.html');
  62  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  63  | 
  64  |     const titles = page.locator('.name-title');
  65  |     await expect(titles).toHaveCount(3);
  66  | 
  67  |     const names = await titles.allTextContents();
  68  |     expect(names[0]).not.toBe(names[1]);
  69  |     expect(names[1]).not.toBe(names[2]);
  70  |     expect(names[0]).not.toBe(names[2]);
  71  |   });
  72  | 
  73  |   // ── interpretation content ──
  74  | 
  75  |   test('interpretation has three sections: 命理基础, 候选分析, 选择建议', async () => {
  76  |     await page.goto('/zh/naming.html');
  77  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  78  | 
  79  |     const h3s = page.locator('.report-content h3');
  80  |     await expect(h3s).toHaveCount(3);
  81  |     await expect(h3s.nth(0)).toContainText('命理基础');
  82  |     await expect(h3s.nth(1)).toContainText('候选名字分析');
  83  |     await expect(h3s.nth(2)).toContainText('选择建议');
  84  |   });
  85  | 
  86  |   // ── print and share ──
  87  | 
  88  |   test('print and share buttons exist', async () => {
  89  |     await page.goto('/zh/naming.html');
  90  |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  91  | 
  92  |     await expect(page.locator('.btn-print')).toBeVisible();
  93  |     await expect(page.locator('.btn-share')).toBeVisible();
  94  |   });
  95  | 
  96  |   // ── CTA ──
  97  | 
  98  |   test('CTA links to chat page', async () => {
  99  |     await page.goto('/zh/naming.html');
  100 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  101 | 
  102 |     const ctaLink = page.locator('.cta-bar a.btn-primary');
  103 |     await expect(ctaLink).toHaveAttribute('href', '/chat.html');
  104 |   });
  105 | 
  106 |   // ── sample note ──
  107 | 
  108 |   test('sample note is visible', async () => {
  109 |     await page.goto('/zh/naming.html');
  110 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  111 | 
  112 |     // "此为示例报告" text near top.
  113 |     await expect(page.getByText('示例报告')).toBeVisible();
  114 |   });
  115 | });
  116 | 
```