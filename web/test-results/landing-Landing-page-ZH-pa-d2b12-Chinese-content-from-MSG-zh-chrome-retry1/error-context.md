# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: landing.spec.js >> Landing page >> ZH page renders Chinese content from MSG.zh
- Location: e2e/journeys/landing.spec.js:35:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/
Call log:
  - navigating to "http://localhost:8080/zh/", waiting until "load"

```

# Test source

```ts
  1  | // Landing page tests — validates static page rendering, navigation, and i18n.
  2  | // Source: index.html, i18n.js (detectLang, setLang), docs/llms.txt health check.
  3  | 
  4  | import { test, expect } from '../fixtures.js';
  5  | 
  6  | test.describe('Landing page', () => {
  7  |   let page;
  8  | 
  9  |   test.beforeEach(async ({ context }) => {
  10 |     page = await context.newPage();
  11 |   });
  12 | 
  13 |   test.afterEach(async () => {
  14 |     await page.close();
  15 |   });
  16 | 
  17 |   test('EN page renders site name, 3 product cards, steps, footer', async () => {
  18 |     await page.goto('/en/');
  19 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  20 | 
  21 |     // From i18n.js MSG.en: site.name = "Liki"
  22 |     await expect(page.locator('h1')).toHaveText('Liki');
  23 |     // index.chart.title, index.bond.title, index.naming.title
  24 |     const cards = page.locator('a.product-card');
  25 |     await expect(cards).toHaveCount(3);
  26 |     await expect(cards.nth(0)).toContainText('BaZi');
  27 |     await expect(cards.nth(1)).toContainText('Compatibility');
  28 |     await expect(cards.nth(2)).toContainText('Naming');
  29 |     // index.steps.title
  30 |     await expect(page.getByText('Want to Try First?')).toBeVisible();
  31 |     // site.footer
  32 |     await expect(page.locator('footer')).toBeVisible();
  33 |   });
  34 | 
  35 |   test('ZH page renders Chinese content from MSG.zh', async () => {
> 36 |     await page.goto('/zh/');
     |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/
  37 |     await page.waitForSelector('[data-i18n]', { timeout: 10000 });
  38 | 
  39 |     await expect(page.locator('h1')).toContainText('灵机');
  40 |     const cards = page.locator('a.product-card');
  41 |     await expect(cards.nth(0)).toContainText('八字');
  42 |     await expect(cards.nth(1)).toContainText('合盘');
  43 |     await expect(cards.nth(2)).toContainText('起名');
  44 |   });
  45 | 
  46 |   test('product cards link to correct pages (href from index.html)', async () => {
  47 |     await page.goto('/en/');
  48 | 
  49 |     const cards = page.locator('a.product-card');
  50 |     await expect(cards.nth(0)).toHaveAttribute('href', 'chat.html?product=chart');
  51 |     await expect(cards.nth(1)).toHaveAttribute('href', 'chat.html?product=bond');
  52 |     await expect(cards.nth(2)).toHaveAttribute('href', 'chat.html?product=naming');
  53 |   });
  54 | 
  55 |   test('language switch calls setLang(), redirects to other locale', async () => {
  56 |     await page.goto('/en/');
  57 | 
  58 |     // Open lang dropdown and click zh option
  59 |     await page.locator('[data-lang-toggle]').click();
  60 |     await page.locator('[data-lang-option="zh"]').click();
  61 |     await page.waitForURL(/\/zh\//);
  62 |     await expect(page.locator('h1')).toContainText('灵机');
  63 |   });
  64 | 
  65 |   // Removed: pending order banner feature was removed in refactoring.
  66 | });
  67 | 
```