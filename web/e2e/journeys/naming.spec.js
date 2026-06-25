// Naming demo page tests — validates static naming demo report rendering, i18n, interactions.
// Source: naming.html, demo-utils.js, i18n.js.

import { test, expect } from '../fixtures.js';

test.describe('Naming demo page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── content rendering ──

  test('ZH page renders all sections: summary, candidates, interpretation, CTA', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // Header — use h1[data-i18n] to avoid strict-mode conflict with <liki-header> brand h1.
    await expect(page.locator('h1[data-i18n]')).toContainText('起名报告');

    // Summary cards.
    await expect(page.locator('.summary-grid')).toBeVisible();

    // Name candidate cards.
    const nameCards = page.locator('.name-card');
    await expect(nameCards).toHaveCount(3);

    // LLM interpretation.
    await expect(page.locator('.report-content')).toBeVisible();
  });

  test('EN page renders with English labels', async () => {
    await page.goto('/en/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('h1[data-i18n]')).toContainText('Naming');
  });

  // ── name candidate cards ──

  test('first name card shows 陳煜霖 with correct details', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const firstCard = page.locator('.name-card').nth(0);
    await expect(firstCard.locator('.name-title')).toContainText('陳煜霖');
    await expect(firstCard.locator('.tag-green')).toContainText('大吉');
    // Character details.
    await expect(firstCard.locator('.name-card-chars')).toContainText('煜 (13畫)');
    await expect(firstCard.locator('.name-card-chars')).toContainText('霖 (16畫)');
    // Wu Ge.
    await expect(firstCard.locator('.name-card-wuge')).toBeVisible();
  });

  test('all three candidates have different names', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const titles = page.locator('.name-title');
    await expect(titles).toHaveCount(3);

    const names = await titles.allTextContents();
    expect(names[0]).not.toBe(names[1]);
    expect(names[1]).not.toBe(names[2]);
    expect(names[0]).not.toBe(names[2]);
  });

  // ── interpretation content ──

  test('interpretation has three sections: 命理基礎, 候選分析, 選擇建議', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const h3s = page.locator('.report-content h3');
    await expect(h3s).toHaveCount(3);
    await expect(h3s.nth(0)).toContainText('命理基礎');
    await expect(h3s.nth(1)).toContainText('候選名字分析');
    await expect(h3s.nth(2)).toContainText('選擇建議');
  });

  // ── print and share ──

  test('print and share buttons exist', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('.btn-print')).toBeVisible();
    await expect(page.locator('.btn-share')).toBeVisible();
  });

  // ── CTA ──

  test('CTA links to chat page', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const ctaLink = page.locator('.cta-bar a.btn-primary');
    await expect(ctaLink).toHaveAttribute('href', '/chat.html');
  });

  // ── sample note ──

  test('sample note is visible', async () => {
    await page.goto('/zh-Hans/naming.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // "此为示例报告" text near top.
    await expect(page.getByText('示例报告')).toBeVisible();
  });
});
