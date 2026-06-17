// Report page tests — GET /api/reports/:orderID per report.js.
// orderIDFromURL() regex: /\/report\/([a-f0-9-]+)/i — extracts UUID from path.
// States: loading → error (no ID) | ready (status=paid) | polling (status=pending) | pollTimeout
// Response shape from report.js:145-173: loadReport → apiGet('/reports/'+orderID)

import { test, expect } from '../fixtures.js';

// Realistic mock data for a paid chart report.
const CHART_JSON = JSON.stringify({
  chart: {
    chart: {
      riyuan: '戊土',
      nianzhu: { gan: '甲', zhi: '子', shishen: [{ source: 'gan', shishen: '七杀' }], canggan: { 子: '癸' }, nayin: '海中金', shensha: [{ name: '天乙贵人' }] },
      yuezhu: { gan: '丙', zhi: '寅', shishen: [{ source: 'gan', shishen: '偏印' }], canggan: { 寅: '甲丙戊' }, nayin: '炉中火', shensha: [] },
      rizhu: { gan: '戊', zhi: '辰', shishen: [{ source: 'gan', shishen: '日主' }], canggan: { 辰: '戊乙癸' }, nayin: '大林木', shensha: [{ name: '华盖' }] },
      shizhu: { gan: '辛', zhi: '酉', shishen: [{ source: 'gan', shishen: '伤官' }], canggan: { 酉: '辛' }, nayin: '石榴木', shensha: [] },
      yong_shen: { fuyi: { qiangruo: '身强', geju: '正印格', yong: '木', xi: '水', ji: '土' }, tiaohou: { season: '冬', yong: '火', xi: '木', ji: '木' } },
      dayun: [],
    },
  },
});

const LLM_JSON = `# 八字命理报告

## 第一章 · 命盘总览

您的日主为**戊土**，生于寅月。戊土厚重，寅月木旺。

## 第二章 · 五行分析

命局五行分布：木 3、火 1、土 2、金 1、水 1。

## 第三章 · 十神解读

年柱甲子，甲为七杀，子为正财。月柱丙寅，丙为偏印，寅为七杀。

## 第四章 · 大运流年

当前大运戊辰，比肩帮身，财运亨通。

## 第五章 · 综合建议

宜从事木火行业，向东方发展。`;

// Mock chart report API — returns paid report with chart_json + llm_json.
async function mockChartReport(page, orderID = 'test-chart-id') {
  await page.route(`**/api/reports/${orderID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          order_id: orderID,
          product: 'chart',
          chart_json: CHART_JSON,
          llm_json: LLM_JSON,
          status: 'paid',
          amount: 990,
          currency: 'CNY',
        },
      }),
    });
  });
}

// Mock bond report API.
async function mockBondReport(page, orderID = 'test-bond-id') {
  const bondChartJSON = JSON.stringify({
    chart_a: { chart: { riyuan: '庚金', nianzhu: { gan: '庚', zhi: '午' }, yuezhu: { gan: '壬', zhi: '午' }, rizhu: { gan: '庚', zhi: '午' }, shizhu: { gan: '丙', zhi: '子' } } },
    chart_b: { chart: { riyuan: '甲木', nianzhu: { gan: '甲', zhi: '子' }, yuezhu: { gan: '丙', zhi: '寅' }, rizhu: { gan: '甲', zhi: '寅' }, shizhu: { gan: '戊', zhi: '辰' } } },
    bond: { gan_rel: [{ a: '庚', b: '甲', rel: '冲' }], zhi_rel: [], key_hints: ['金木相冲'] },
  });
  await page.route(`**/api/reports/${orderID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          order_id: orderID,
          product: 'bond',
          chart_json: bondChartJSON,
          llm_json: '# 合盘报告\n\n## 缘分分析\n\n金木相冲，需调和。',
          status: 'paid',
          amount: 1990,
          currency: 'CNY',
        },
      }),
    });
  });
}

// Mock naming report API.
async function mockNamingReport(page, orderID = 'test-naming-id') {
  const namingChartJSON = JSON.stringify({
    naming: {
      analysis: { surname: '陈', yong_shen: '火', zodiac: '鼠' },
      candidates: [{ name: '陈明远', score: 95, wuxing: '火木土' }],
    },
  });
  await page.route(`**/api/reports/${orderID}`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          order_id: orderID,
          product: 'naming',
          chart_json: namingChartJSON,
          llm_json: '# 起名报告\n\n## 名字分析\n\n**陈明远** — 得分 95。',
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

    await page.locator('[x-show="error"]').waitFor({ state: 'visible', timeout: 10000 });
    await expect(page.locator('[x-show="error"] p')).toBeVisible();
    await expect(page.locator('[x-show="error"] a')).toBeVisible();
  });

  test('invalid order ID shows error', async () => {
    await page.goto('/en/report/nonexistent-id-12345');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await page.locator('[x-show="error"]').waitFor({ state: 'visible', timeout: 10000 });
    await expect(page.locator('[x-show="error"] p')).toBeVisible();
  });

  // ── save banner ──

  test('save banner visible by default, has copy link and close buttons', async () => {
    await page.goto('/en/report/test-id');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    const banner = page.locator('.save-banner');
    await expect(banner).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button[x-text="copyBtnText"]')).toBeVisible();

    await page.locator('.save-banner').getByText('×').click();
    await expect(banner).not.toBeVisible({ timeout: 3000 });
  });

  // ── successful chart report rendering ──

  test('paid chart report renders markdown and summary cards', async () => {
    await mockChartReport(page);
    await page.goto('/zh/report/test-chart-id');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Phase transitions to ready, report content visible.
    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });

    // Summary cards rendered from chart_json.
    await expect(page.locator('.summary-grid')).toBeVisible();
    await expect(page.locator('.summary-card .value').first()).toContainText('戊土');

    // Markdown → HTML rendered.
    const content = await page.locator('[x-show="ready"]').innerHTML();
    expect(content).toContain('命盘总览');
    expect(content).toContain('戊土');
    expect(content).toContain('五行分析');
    expect(content).toContain('十神解读');
    expect(content).toContain('大运流年');

    // Title reflects product.
    await expect(page.locator('h1')).toContainText('八字报告');
  });

  test('paid chart report in EN locale', async () => {
    await mockChartReport(page, 'en-chart-id');
    await page.goto('/en/report/en-chart-id');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });

    // EN title
    await expect(page.locator('h1')).toContainText('BaZi');
  });

  // ── bond report ──

  test('paid bond report renders bond-specific data', async () => {
    await mockBondReport(page);
    await page.goto('/zh/report/test-bond-id');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });

    // Bond report markdown.
    const content = await page.locator('[x-show="ready"]').innerHTML();
    expect(content).toContain('缘分分析');
    expect(content).toContain('金木相冲');

    // Title reflects bond product.
    await expect(page.locator('h1')).toContainText('合盘报告');
  });

  // ── naming report ──

  test('paid naming report renders candidates', async () => {
    await mockNamingReport(page);
    await page.goto('/zh/report/test-naming-id');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });

    const content = await page.locator('[x-show="ready"]').innerHTML();
    expect(content).toContain('起名报告');
    expect(content).toContain('陈明远');

    await expect(page.locator('h1')).toContainText('起名报告');
  });

  // ── polling → ready transition ──

  test('transitions from payment polling to ready when order is paid', async () => {
    let callCount = 0;
    await page.route('**/api/reports/poll-123', async (route) => {
      callCount++;
      if (callCount === 1) {
        // First poll: still pending.
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: { order_id: 'poll-123', product: 'chart', status: 'pending' },
          }),
        });
      } else {
        // Subsequent polls: paid with report data.
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              order_id: 'poll-123',
              product: 'chart',
              chart_json: CHART_JSON,
              llm_json: LLM_JSON,
              status: 'paid',
            },
          }),
        });
      }
    });

    await page.goto('/zh/report/poll-123');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Starts in payment polling phase.
    await expect(page.locator('.status-card.status-payment')).toBeVisible({ timeout: 5000 });

    // Eventually transitions to ready.
    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 15000 });
    await expect(page.locator('.summary-grid')).toBeVisible();
  });

  // ── generating → ready transition ──

  test('transitions from generating polling to ready when llm_json arrives', async () => {
    let callCount = 0;
    await page.route('**/api/reports/gen-456', async (route) => {
      callCount++;
      if (callCount <= 2) {
        // Paid but no llm_json yet — generating status.
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              order_id: 'gen-456',
              product: 'chart',
              chart_json: CHART_JSON,
              status: 'paid',
            },
          }),
        });
      } else {
        // llm_json ready.
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: {
              order_id: 'gen-456',
              product: 'chart',
              chart_json: CHART_JSON,
              llm_json: LLM_JSON,
              status: 'paid',
            },
          }),
        });
      }
    });

    await page.goto('/zh/report/gen-456');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Starts in generating phase.
    await expect(page.locator('.status-card.status-generating')).toBeVisible({ timeout: 5000 });

    // Eventually transitions to ready with transition animation.
    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 15000 });
    await expect(page.locator('.summary-grid')).toBeVisible();
  });

  // ── save banner interactions ──

  test('banner close persists across page reloads via sessionStorage', async () => {
    await mockChartReport(page, 'banner-test');
    await page.goto('/zh/report/banner-test');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Banner visible initially.
    await expect(page.locator('.save-banner')).toBeVisible({ timeout: 5000 });

    // Close it.
    await page.locator('.save-banner').getByText('×').click();
    await expect(page.locator('.save-banner')).not.toBeVisible({ timeout: 3000 });

    // Reload — banner stays hidden.
    await page.reload();
    await page.waitForSelector('#report-header-title', { timeout: 10000 });
    await page.waitForTimeout(500);
    await expect(page.locator('.save-banner')).not.toBeVisible();
  });

  // ── copy link ──

  test('copy link button updates text on click', async () => {
    await mockChartReport(page, 'copy-test');
    await page.goto('/zh/report/copy-test');
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    await page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });

    // Click copy link in the save banner.
    const copyBtn = page.locator('.save-banner button[x-text="copyBtnText"]');
    await expect(copyBtn).toBeVisible();

    // Clipboard permission is auto-granted in Playwright test context.
    await copyBtn.click();
    // Button text changes to '已复制' briefly, then back.
    await expect(copyBtn).toContainText('已复制', { timeout: 3000 });
  });
});
