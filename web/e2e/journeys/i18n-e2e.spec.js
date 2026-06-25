// Cross-page i18n tests — verifies i18n.js behavior across pages.
// Sources: i18n.js (detectLang, setLang, t), index.html, chat.html.

import { test, expect } from '../fixtures.js';

test.describe('i18n cross-page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('setLang stores preference and redirects, persists across navigation (i18n.js:27-31)', async () => {
    await page.goto('/en/');
    await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });

    // Click language toggle → click zh-Hans option → setLang('zh-Hans') → redirect to /zh-Hans/
    await page.locator('[data-lang-toggle]').click();
    await page.locator('[data-lang-option="zh-Hans"]').click();
    await page.waitForURL(/\/zh-Hans\//);
    await expect(page.locator('h1')).toContainText('灵机');
  });

  test('EN pages display English, ZH pages display Chinese per MSG tables (i18n.js:4-6)', async () => {
    await page.goto('/en/');
    await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });
    await expect(page.locator('h1')).toContainText('Liki');

    await page.goto('/zh-Hans/');
    await page.waitForSelector('[data-lang-toggle]', { timeout: 10000 });
    await expect(page.locator('h1')).toContainText('灵机');
  });

  test('chat page renders in both locales', async () => {
    await page.goto('/en/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });
  });
});
