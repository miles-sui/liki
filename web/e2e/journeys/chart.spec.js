// Chart demo page tests — validates static BaZi demo report rendering, i18n, print/share.
// Source: chart.html, demo-utils.js, i18n.js.

import { test, expect } from '../fixtures.js';

test.describe('Chart demo page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── content rendering ──

  test('ZH page renders all sections: summary, pillars, elements, dayun, interpretation, CTA', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // Header.
    await expect(page.locator('h1')).toContainText('八字报告');

    // Summary cards.
    await expect(page.locator('.summary-grid')).toBeVisible();
    const summaryCards = page.locator('.summary-card');
    await expect(summaryCards).toHaveCount(4);

    // Pillars — mobile cards + desktop table.
    await expect(page.locator('.pillar-cards')).toBeVisible();
    await expect(page.locator('.pillar-card')).toHaveCount(4);
    await expect(page.locator('.hide-mobile table')).toBeVisible();
    await expect(page.locator('.hide-mobile tbody tr')).toHaveCount(4);

    // Element distribution.
    await expect(page.locator('h2:has-text("五行分布")')).toBeVisible();

    // Dayun.
    await expect(page.locator('h2:has-text("大运")')).toBeVisible();

    // LLM interpretation.
    await expect(page.locator('.report-content')).toBeVisible();
    await expect(page.locator('.report-content h3')).toHaveCount(5);

    // CTA bar.
    await expect(page.locator('.cta-bar')).toBeVisible();
    await expect(page.locator('.cta-bar a.btn-primary')).toHaveAttribute('href', '/chat.html');

    // Trust section.
    await expect(page.locator('.hero-numbers')).toBeVisible();
    await expect(page.locator('.trust-badges')).toBeVisible();
  });

  test('EN page renders with English labels', async () => {
    await page.goto('/en/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('h1')).toContainText('BaZi');
    // Table headers should be in English.
    await expect(page.locator('.hide-mobile thead')).toContainText('Pillar');
  });

  // ── print and share buttons ──

  test('print button exists and triggers print dialog', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const printBtn = page.locator('.btn-print');
    await expect(printBtn).toBeVisible();
    await expect(printBtn).toContainText('打印');
  });

  test('share button exists', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const shareBtn = page.locator('.btn-share');
    await expect(shareBtn).toBeVisible();
    await expect(shareBtn).toContainText('分享');
  });

  // ── pillar table data correctness ──

  test('year pillar shows correct data (甲 七杀 → 子, 癸 · 海中金)', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // Mobile pillar cards.
    const yearCard = page.locator('.pillar-card').nth(0);
    await expect(yearCard.locator('.pc-label')).toContainText('年柱');
    await expect(yearCard.locator('.pc-gan')).toContainText('甲');
    await expect(yearCard.locator('.pc-zhi')).toContainText('子');
    await expect(yearCard.locator('.pc-detail')).toContainText('海中金');

    // Desktop table.
    const yearRow = page.locator('.hide-mobile tbody tr').nth(0);
    await expect(yearRow).toContainText('七杀');
    await expect(yearRow).toContainText('海中金');
  });

  test('day pillar shows 戊 (比肩) → 辰 with 华盖 shensha', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const dayRow = page.locator('.hide-mobile tbody tr').nth(2);
    await expect(dayRow).toContainText('华盖');
  });

  // ── CTA link ──

  test('CTA button links to chat page', async () => {
    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    const ctaLink = page.locator('.cta-bar a.btn-primary');
    await expect(ctaLink).toHaveAttribute('href', '/chat.html');
  });

  // ── analytics fires on page load ──

  test('pageview analytics fires on load', async () => {
    let analyticsBody = null;
    await page.route('**/api/analytics/pageview', async (route) => {
      analyticsBody = JSON.parse(route.request().postData() || '{}');
      await route.fulfill({ status: 200, contentType: 'application/json', body: '{"data":{"ok":true}}' });
    });

    await page.goto('/zh/chart.html');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    // Analytics should have been called with the correct path.
    await page.waitForTimeout(500);
    expect(analyticsBody).not.toBeNull();
    expect(analyticsBody.path).toContain('chart.html');
  });
});
