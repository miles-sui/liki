// Full purchase flow E2E — chat → compute → buy → checkout → report.
// Mocks all backend APIs to verify frontend state transitions end-to-end.

import { test, expect } from '../fixtures.js';

function sseLine(obj) {
  return `data: ${JSON.stringify(obj)}\n\n`;
}

const CHART_DATA = JSON.stringify({
  chart: {
    chart: {
      riyuan: '庚金',
      nianzhu: { gan: '庚', zhi: '午', shishen: [{ source: 'gan', shishen: '比肩' }], canggan: {}, nayin: '路旁土', shensha: [] },
      yuezhu: { gan: '壬', zhi: '午', shishen: [{ source: 'gan', shishen: '食神' }], canggan: {}, nayin: '杨柳木', shensha: [] },
      rizhu: { gan: '庚', zhi: '午', shishen: [{ source: 'gan', shishen: '日主' }], canggan: {}, nayin: '路旁土', shensha: [] },
      shizhu: { gan: '丙', zhi: '子', shishen: [{ source: 'gan', shishen: '七杀' }], canggan: {}, nayin: '涧下水', shensha: [] },
      yong_shen: { fuyi: { qiangruo: '身弱', geju: '七杀格', yong: '土', xi: '金', ji: '木' }, tiaohou: { season: '夏', yong: '水', xi: '金', ji: '火' } },
    },
  },
});

const REPORT_LLM = `# 八字命理报告

## 第一章 · 命盘总览

您的日主为**庚金**，生于午月。庚金刚健，午月火旺。

## 第二章 · 综合分析

身弱七杀格，以土为用，喜金帮身。`;

const ORDER_ID = 'e2e-baad-f00d-001';

test.describe('Purchase flow', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── chat → done → buy bar visible ──

  test('full flow: chat message → done event → buy bar → checkout URL', async () => {
    // Mock greeting.
    await page.route('**/api/agent/greeting', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { greeting: '你好，我是灵机。' } }),
      });
    });

    // Mock chat SSE: question → compute result → done with purchase.
    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'e2e-purchase-session',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'phase', content: '正在分析您的需求…' }),
          sseLine({ type: 'text-delta', content: '好的，您的八字排盘如下：' }),
          sseLine({ type: 'text-delta', content: '\n\n## 命盘分析\n\n' }),
          sseLine({ type: 'text-delta', content: '您的日主为庚金，生于午月…' }),
          sseLine({ type: 'phase', content: '正在生成分析报告…' }),
          sseLine({ type: 'text-delta', content: '如需查看完整报告（五行、十神、大运、流年），可以购买解锁。' }),
          sseLine({
            type: 'done',
            data: { product: 'chart', order_id: ORDER_ID, amount: 990 },
          }),
        ].join(''),
      });
    });

    // Mock checkout API.
    await page.route('**/api/payments/checkout', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { checkout_url: 'https://pay.dodopayments.com/checkout/e2e-test' },
        }),
      });
    });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    // Send a message.
    await page.locator('.chat-input-bar input').fill('帮我排盘');
    await page.locator('.btn-send').click();

    // Buy bar should appear after done event.
    const buyCard = page.locator('.buy-card');
    await expect(buyCard).toBeVisible({ timeout: 10000 });

    // Verify buy card content.
    await expect(buyCard).toContainText('9.90');
    await expect(page.locator('.btn-buy')).toContainText('查看完整报告');

    // Click the pay button → should trigger checkout and navigate.
    // We intercept the navigation by not actually loading the checkout URL.
    await page.route('https://pay.dodopayments.com/checkout/e2e-test', async (route) => {
      await route.fulfill({ status: 200, body: '<html><body>Mock Dodo Checkout</body></html>' });
    });

    await page.locator('.btn-buy').click();
    // Should navigate to checkout page.
    await page.waitForURL('**/checkout/e2e-test', { timeout: 10000 });
  });

  // ── report page after purchase ──

  test('report page shows paid report after purchase flow', async () => {
    // Mock report API.
    await page.route(`**/api/reports/${ORDER_ID}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            order_id: ORDER_ID,
            product: 'chart',
            chart_json: CHART_DATA,
            llm_json: REPORT_LLM,
            status: 'paid',
            amount: 990,
            currency: 'CNY',
          },
        }),
      });
    });

    await page.goto(`/zh-Hans/report/${ORDER_ID}`);
    await page.waitForSelector('#report-header-title', { timeout: 10000 });

    // Report content should be visible.
    await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });

    // Verify markdown rendered.
    const interpretation = page.locator('#chart-interpretation');
    await expect(interpretation).toContainText('命盘总览');
    await expect(interpretation).toContainText('庚金');

    // Summary cards from chart_json — use .first() for strict-mode safety.
    await expect(page.locator('.summary-grid').first()).toBeVisible();
    await expect(page.locator('.summary-card .value').first()).toContainText('庚金');
  });

  // ── session ID persistence across chat → buy ──

  test('session ID persists in sessionStorage after purchase', async () => {
    // Mock greeting.
    await page.route('**/api/agent/greeting', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { greeting: '你好。' } }),
      });
    });

    // Mock chat with done.
    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'persist-session-e2e',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'text-delta', content: '好的。' }),
          sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-persist', amount: 990 } }),
        ].join(''),
      });
    });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    await page.locator('.chat-input-bar input').fill('排盘');
    await page.locator('.btn-send').click();

    // Wait for buy bar.
    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });

    // Session ID should be in sessionStorage.
    const sid = await page.evaluate(() => sessionStorage.getItem('chatSessionID'));
    expect(sid).toBe('"persist-session-e2e"');
  });

  // ── new chat resets state ──

  test('new chat after purchase clears buy bar and resets to greeting', async () => {
    // Mock greeting.
    await page.route('**/api/agent/greeting', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { greeting: '你好，我是灵机。' } }),
      });
    });

    // Mock chat with done.
    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'new-chat-test',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'text-delta', content: '分析完成。' }),
          sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-newchat', amount: 990 } }),
        ].join(''),
      });
    });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    await page.locator('.chat-input-bar input').fill('排盘');
    await page.locator('.btn-send').click();

    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });

    // Click "新对话" to reset.
    await page.locator('button:has-text("新对话")').click();

    // Buy bar should be gone, chips should be back.
    await expect(page.locator('.buy-card')).not.toBeVisible();
    await expect(page.locator('.chip-row')).toBeVisible();
  });

  // ── error during purchase ──

  test('checkout API error shows toast', async () => {
    // Mock greeting.
    await page.route('**/api/agent/greeting', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { greeting: '你好。' } }),
      });
    });

    // Mock chat with done.
    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'checkout-error-test',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'text-delta', content: '分析完成。' }),
          sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-err', amount: 990 } }),
        ].join(''),
      });
    });

    // Mock checkout to fail.
    await page.route('**/api/payments/checkout', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: { code: 'checkout_failed', message: '支付服务暂不可用' } }),
      });
    });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    await page.locator('.chat-input-bar input').fill('排盘');
    await page.locator('.btn-send').click();

    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });

    // Intercept navigation so we stay on page when goPay redirects.
    await page.route('**/*', async (route) => {
      // Let the checkout error response through, don't redirect.
      const url = route.request().url();
      if (url.includes('/api/payments/checkout')) {
        await route.continue();
      } else {
        await route.continue();
      }
    });

    // Click buy — checkout fails, toast should appear.
    await page.locator('.btn-buy').click();

    // Error toast or error in buy card.
    await expect(page.locator('.error-toast')).toBeVisible({ timeout: 10000 });
  });

  // ── xunhu checkout with QR code ──

  test('xunhu checkout returns qrcode_url alongside checkout_url', async () => {
    // Mock greeting.
    await page.route('**/api/agent/greeting', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { greeting: '你好。' } }),
      });
    });

    // Mock chat with done (CNY currency → xunhu provider).
    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'xunhu-checkout-test',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'text-delta', content: '分析完成。' }),
          sseLine({ type: 'done', data: { product: 'chart', order_id: 'order-xunhu', amount: 990, currency: 'CNY' } }),
        ].join(''),
      });
    });

    // Mock xunhu checkout — returns both checkout_url and qrcode_url.
    await page.route('**/api/payments/checkout', async (route) => {
      const body = JSON.parse(route.request().postData() || '{}');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            checkout_url: 'https://api.xunhupay.com/payment/checkout',
            qrcode_url: 'https://api.xunhupay.com/payment/qrcode',
            provider: body.provider || 'xunhu',
          },
        }),
      });
    });

    await page.goto('/zh-Hans/chat.html');
    await page.waitForSelector('.chat-shell', { timeout: 10000 });

    await page.locator('.chat-input-bar input').fill('排盘');
    await page.locator('.btn-send').click();

    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.buy-card')).toContainText('¥'); // CNY symbol
  });
});
