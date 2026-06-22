# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: i18n-e2e.spec.js >> i18n cross-page >> EN pages display English, ZH pages display Chinese per MSG tables (i18n.js:4-6)
- Location: e2e/journeys/i18n-e2e.spec.js:28:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/en/
Call log:
  - navigating to "http://localhost:8080/en/", waiting until "load"

```

# Test source

```ts
  1  | // Cross-page i18n tests — verifies i18n.js behavior across pages.
  2  | // Sources: i18n.js (detectLang, setLang, t), index.html, chat.html.
  3  | 
  4  | import { test, expect } from '../fixtures.js';
  5  | 
  6  | test.describe('i18n cross-page', () => {
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
  17 |   test('setLang stores preference and redirects, persists across navigation (i18n.js:27-31)', async () => {
  18 |     await page.goto('/en/');
  19 |     await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });
  20 | 
  21 |     // Click language toggle → click zh option → setLang('zh') → redirect to /zh/
  22 |     await page.locator('[data-lang-toggle]').click();
  23 |     await page.locator('[data-lang-option="zh"]').click();
  24 |     await page.waitForURL(/\/zh\//);
  25 |     await expect(page.locator('h1')).toContainText('灵机');
  26 |   });
  27 | 
  28 |   test('EN pages display English, ZH pages display Chinese per MSG tables (i18n.js:4-6)', async () => {
> 29 |     await page.goto('/en/');
     |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/en/
  30 |     await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });
  31 |     await expect(page.locator('h1')).toContainText('Liki');
  32 | 
  33 |     await page.goto('/zh/');
  34 |     await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });
  35 |     await expect(page.locator('h1')).toContainText('灵机');
  36 |   });
  37 | 
  38 |   test('chat page renders in both locales', async () => {
  39 |     await page.goto('/en/chat.html');
  40 |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  41 | 
  42 |     await page.goto('/zh/chat.html');
  43 |     await page.waitForSelector('.chat-shell', { timeout: 10000 });
  44 |   });
  45 | });
  46 | 
```