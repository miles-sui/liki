// Landing page tests — validates static page rendering, navigation, and i18n.
// Source: index.html, i18n.js (detectLang, setLang), docs/llms.txt health check.

import { test, expect } from '../fixtures.js';

test.describe('Landing page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  test('EN page renders site name, 3 product cards, steps, footer', async () => {
    await page.goto('/en/');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // From i18n.js MSG.en: site.name = "Liki"
    await expect(page.locator('h1')).toHaveText('Liki');
    // index.chart.title, index.bond.title, index.naming.title
    const cards = page.locator('a.card');
    await expect(cards).toHaveCount(3);
    await expect(cards.nth(0)).toContainText('BaZi');
    await expect(cards.nth(1)).toContainText('Bond');
    await expect(cards.nth(2)).toContainText('Naming');
    // index.steps.title
    await expect(page.getByText('How It Works')).toBeVisible();
    // site.footer
    await expect(page.locator('footer')).toBeVisible();
  });

  test('ZH page renders Chinese content from MSG.zh', async () => {
    await page.goto('/zh/');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('h1')).toContainText('灵机');
    const cards = page.locator('a.card');
    await expect(cards.nth(0)).toContainText('八字');
    await expect(cards.nth(1)).toContainText('合盘');
    await expect(cards.nth(2)).toContainText('起名');
  });

  test('product cards link to correct pages (href from index.html)', async () => {
    await page.goto('/en/');

    const cards = page.locator('a.card');
    await expect(cards.nth(0)).toHaveAttribute('href', 'chat.html');
    await expect(cards.nth(1)).toHaveAttribute('href', 'chat.html');
    await expect(cards.nth(2)).toHaveAttribute('href', 'chat.html');
  });

  test('language switch calls setLang(), redirects to other locale', async () => {
    await page.goto('/en/');

    // setLang() in i18n.js:32 — stores to localStorage and redirects
    await page.locator('header a[href="#"]').click();
    await page.waitForURL(/\/zh\//);
    await expect(page.locator('h1')).toContainText('灵机');
  });

  test('pending order banner visible when sessionStorage has orderID (index.html:13-20)', async () => {
    await page.goto('/en/');
    await page.evaluate(() => sessionStorage.setItem('orderID', 'test-order-123'));
    await page.reload();
    await page.waitForLoadState('networkidle');

    await expect(page.locator('#pending-banner')).toBeVisible({ timeout: 5000 });
  });
});
