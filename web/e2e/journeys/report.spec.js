// Report page tests — GET /api/reports/:orderID per report.js.
// orderIDFromURL() regex: /\/report\/([a-f0-9-]+)/i — extracts UUID from path.
// States: loading → error (no ID) | ready (status=paid) | polling (status=pending) | pollTimeout
// Rendered by lit-html ReportApp class — #status-area shows status card, #report-content shows report.
// Response shape: { data: { order_id, product, chart_json, llm_json, status, ... } }

import { test, expect } from '../fixtures.js';

// Shared naming report mock data.
const NAMING_CHART_JSON = JSON.stringify({
  naming: {
    analysis: { surname: '陈', yong_shen: '火', zodiac: '鼠' },
    candidates: [{ name: '陈明远', wuxing: '火木土' }],
  },
});

const NAMING_LLM_JSON = '# 起名报告\n\n## 名字分析\n\n**陈明远** — 得分 95。';

// Mock naming report API.
async function mockNamingReport(page, orderID = 'babe-face') {
  await page.route(`**/api/reports/${orderID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          order_id: orderID,
          product: 'naming',
          chart_json: NAMING_CHART_JSON,
          llm_json: NAMING_LLM_JSON,
          status: 'paid',
          amount: 2990,
          currency: 'CNY',
        },
      }),
    });
  });
}

test.describe('Report page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── error states (existing) ──

  test('missing order ID shows error', async () => {
    await page.goto('/en/report/');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // ReportApp sets phase='error' when orderID is missing — renders .status-card.status-error.
    await expect(page.locator('.status-card.status-error')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.status-card.status-error .status-actions a')).toBeVisible();
  });

  test('invalid order ID shows error', async () => {
    await page.goto('/en/report/nonexistent-id-12345');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // ReportApp calls loadReport → API returns 404 → phase='error'.
    await expect(page.locator('.status-card.status-error')).toBeVisible({ timeout: 10000 });
  });

  // ── save banner ──

  test('save banner visible by default, has copy link and close buttons', async () => {
    await page.goto('/en/report/abba');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    const banner = page.locator('.save-banner');
    await expect(banner).toBeVisible({ timeout: 5000 });
    // Copy button is #banner-copy-btn per report.html.
    await expect(page.locator('#banner-copy-btn')).toBeVisible();

    await page.locator('#banner-close-btn').click();
    await expect(banner).not.toBeVisible({ timeout: 3000 });
  });

  // ── naming report ──

  test('paid naming report renders candidates', async () => {
    await mockNamingReport(page);
    await page.goto('/zh-Hans/report/babe-face');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });

    const interpretation = page.locator('#naming-interpretation');
    await expect(interpretation).toContainText('起名报告');
    await expect(interpretation).toContainText('陈明远');

    await expect(page.locator('#report-header-title')).toContainText('起名报告');
  });

  // ── polling → ready transition ──

  const POLLING_CASES = [
    { name: 'payment polling → ready', orderID: 'abba-123', pendingTries: 2 },
    { name: 'llm_json polling → ready', orderID: 'cafe-456', pendingTries: 3 },
  ];

  for (const { name, orderID, pendingTries } of POLLING_CASES) {
    test(`transitions from ${name}`, async () => {
      let callCount = 0;
      await page.route(`**/api/reports/${orderID}`, async (route) => {
        callCount++;
        if (callCount <= pendingTries) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              data: { order_id: orderID, product: 'naming', status: 'pending' },
            }),
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              data: {
                order_id: orderID,
                product: 'naming',
                chart_json: NAMING_CHART_JSON,
                llm_json: NAMING_LLM_JSON,
                status: 'paid',
              },
            }),
          });
        }
      });

      await page.goto(`/zh-Hans/report/${orderID}`);
      await page.waitForSelector('#report-header-title', { timeout: 10000 });

      await expect(page.locator('.status-card.status-payment')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('#report-content')).toBeVisible({ timeout: 15000 });
      await expect(page.locator('#naming-interpretation')).toBeVisible();
    });
  }

  // ── save banner interactions ──

  test('banner close persists across page reloads via sessionStorage', async () => {
    await mockNamingReport(page, 'face-fade');
    await page.goto('/zh-Hans/report/face-fade');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Banner visible initially.
    await expect(page.locator('.save-banner')).toBeVisible({ timeout: 5000 });

    // Close it.
    await page.locator('#banner-close-btn').click();
    await expect(page.locator('.save-banner')).not.toBeVisible({ timeout: 3000 });

    // Reload — banner stays hidden.
    await page.reload();
    await page.waitForSelector('#report-header-title', { timeout: 10000 });
    await page.waitForTimeout(500);
    await expect(page.locator('.save-banner')).not.toBeVisible();
  });

  // ── copy link ──

  test('copy link button updates text on click', async () => {
    await mockNamingReport(page, 'cafe-feed');
    await page.goto('/zh-Hans/report/cafe-feed');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });

    // Click copy link in the save banner.
    const copyBtn = page.locator('#banner-copy-btn');
    await expect(copyBtn).toBeVisible();

    // Clipboard permission is auto-granted in Playwright test context.
    await copyBtn.click();
    // Button text changes to '已复制' briefly, then back.
    await expect(copyBtn).toContainText('已复制', { timeout: 3000 });
  });
});
