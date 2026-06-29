// Landing page + purchase flow tests.
// Covers: content rendering, order creation, checkout, payment return, full flow.
// Source: index.html, pricing.js, chat.js, report.js.

import { test, expect } from '../fixtures.js';
import { sseLine, mockLogin, mockOrderStatus, mockNamingSSE } from '../mocks.js';

const ORDER_ID = 'e2e-pf-001';

const namingSSEFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在分析八字…' }),
  sseLine({ type: 'text-delta', content: '好的，我来帮您分析起名方案。' }),
  sseLine({ type: 'text-delta', content: '\n\n## 起名分析\n\n' }),
  sseLine({ type: 'text-delta', content: '根据您的八字，推荐名字**陈明远**。' }),
].join('');

// landing-specific mock helpers

async function mockCreateOrder(page, orderID = ORDER_ID) {
  await page.route('**/api/orders', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { order_id: orderID } }),
    });
  });
}

async function mockCheckout(page) {
  await page.route('**/api/payments/checkout', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { checkout_url: 'https://pay.example.com/checkout/pf-test' } }),
    });
  });
}

async function mockCheckoutQR(page) {
  await page.route('**/api/payments/checkout', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { checkout_url: 'https://pay.example.com/checkout/pf-test', qrcode_url: 'https://qr.example.com/qr.png' } }),
    });
  });
}

async function mockCheckoutError(page) {
  await page.route('**/api/payments/checkout', async (route) => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ error: { message: 'Payment service unavailable' } }),
    });
  });
}

test.describe('Landing page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── content rendering ──

  test('EN page renders branding, tagline, pay form, and pricing', async () => {
    await page.goto('/en/');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('h1').first()).toHaveText('Liki');
    await expect(page.locator('main h2:first-of-type')).toContainText('Find Your Name');
    await expect(page.locator('#purchase-form')).toBeVisible();
    await expect(page.locator('#purchase-email')).toBeVisible();
    await expect(page.locator('#purchase-form button[type="submit"]')).toContainText('Purchase');
    await expect(page.locator('footer')).toBeVisible();
  });

  test('ZH-Hans page renders Chinese content', async () => {
    await page.goto('/zh-Hans/');
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });

    await expect(page.locator('h1').first()).toContainText('Liki');
    await expect(page.locator('main h2:first-of-type')).toContainText('找到你的名字');
    await expect(page.locator('#purchase-form')).toBeVisible();
    await expect(page.locator('.name-scroll')).toBeVisible();
  });

  // ── order + checkout ──

  test('order creation and checkout redirect', async () => {
    await mockCreateOrder(page);
    await mockCheckout(page);
    await page.route('**/api/location', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { currency: 'USD' } }),
      });
    });
    await page.route('https://pay.example.com/checkout/pf-test', async (route) => {
      await route.fulfill({ status: 200, body: '<html><body>Mock Checkout</body></html>' });
    });

    await page.goto('/en/');

    let ordersBody = null;
    await page.unroute('**/api/orders');
    await page.route('**/api/orders', async (route) => {
      ordersBody = JSON.parse(route.request().postData() || '{}');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { order_id: ORDER_ID } }),
      });
    });

    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();

    await page.waitForURL('**/checkout/pf-test', { timeout: 10000 });
    expect(ordersBody).toMatchObject({ email: 'buyer@example.com', product: 'naming', currency: 'USD' });
  });

  test('order creation server error shows inline error', async () => {
    await page.route('**/api/orders', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: { message: 'Server error' } }),
      });
    });

    await page.goto('/en/');
    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();

    await expect(page.locator('#purchase-error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#purchase-error')).not.toBeEmpty();
  });

  test('order creation network error shows inline error', async () => {
    await page.goto('/en/');

    await page.route('**/api/orders', async (route) => {
      await route.abort('failed');
    });

    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();

    await expect(page.locator('#purchase-error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#purchase-form button[type="submit"]')).toBeEnabled({ timeout: 5000 });
  });

  test('checkout API error is handled gracefully', async () => {
    await mockCreateOrder(page);
    await mockCheckoutError(page);

    await page.goto('/en/');
    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();

    await expect(page.locator('#purchase-error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#purchase-form button[type="submit"]')).toBeEnabled({ timeout: 5000 });
  });

  // ── QR code ──

  test('desktop shows QR modal when qrcode_url is present', async () => {
    await mockCreateOrder(page);
    await mockCheckoutQR(page);
    await page.goto('/zh-Hans/');
    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();

    await expect(page.locator('.qr-modal-overlay')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.qr-modal')).toBeVisible();
    await expect(page.locator('.qr-modal-img')).toHaveAttribute('src', 'https://qr.example.com/qr.png');
    // URL should NOT have changed (no redirect when QR is shown)
    expect(page.url()).not.toContain('checkout');
  });

  // ── language switch ──

  test('language switch redirects to correct locale URL', async () => {
    await page.goto('/en/');

    await page.locator('[data-lang-toggle]').click();
    await page.locator('[data-lang-option="zh-Hans"]').click();

    await page.waitForURL(/\/zh-Hans\//);
    expect(page.url()).toContain('/zh-Hans/');
  });

  // ── payment return → chat ──

  test('payment return: chat.html?order_id=xxx → login → enter chat', async () => {
    await mockOrderStatus(page, ORDER_ID);
    await mockLogin(page, ORDER_ID);

    let statusCallCount = 0;
    await page.route(`**/api/orders/${ORDER_ID}/status`, async (route) => {
      statusCallCount++;
      if (statusCallCount === 1) {
        await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: { order_id: ORDER_ID, status: 'paid', product: 'naming', chat_expires_at: '2027-06-26 00:00:00', amount: 2990, currency: 'USD' },
          }),
        });
      }
    });

    await page.goto(`/zh-Hans/chat.html?order_id=${ORDER_ID}`);

    await expect(page.locator('.login-overlay')).toBeVisible({ timeout: 10000 });

    await page.locator('.login-card input[type="email"]').fill('buyer@example.com');
    await page.locator('.login-card button').click();

    await expect(page.locator('.chat-messages')).toBeVisible({ timeout: 10000 });
  });

  // ── full flow: landing → checkout → login → SSE ──

  test('full flow: landing order → checkout → login → naming SSE', async () => {
    await mockCreateOrder(page);
    await mockCheckout(page);
    await page.route('https://pay.example.com/checkout/pf-test', async (route) => {
      await route.fulfill({ status: 200, body: '<html><body>Checkout</body></html>' });
    });

    await page.goto('/zh-Hans/');

    await page.locator('#purchase-email').fill('buyer@example.com');
    await page.locator('#purchase-form button[type="submit"]').click();
    await page.waitForURL('**/checkout/pf-test', { timeout: 10000 });

    // Simulate payment return.
    await mockOrderStatus(page, ORDER_ID);
    await mockLogin(page, ORDER_ID);

    let callCount = 0;
    await page.route(`**/api/orders/${ORDER_ID}/status`, async (route) => {
      callCount++;
      if (callCount === 1) {
        await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            data: { order_id: ORDER_ID, status: 'paid', product: 'naming', chat_expires_at: '2027-06-26 00:00:00', amount: 2990, currency: 'USD' },
          }),
        });
      }
    });

    await page.goto(`/zh-Hans/chat.html?order_id=${ORDER_ID}`);
    await expect(page.locator('.login-overlay')).toBeVisible({ timeout: 10000 });

    await page.locator('.login-card input[type="email"]').fill('buyer@example.com');
    await page.locator('.login-card button').click();
    await expect(page.locator('.chat-messages')).toBeVisible({ timeout: 10000 });

    // Send message and verify SSE response.
    await mockNamingSSE(page, namingSSEFlow);

    await page.locator('#chat-input').fill('帮我起名');
    await page.locator('.btn-send').click();

    const lastAsst = page.locator('.msg-asst').last();
    await expect(lastAsst).toBeVisible({ timeout: 10000 });
    const html = await lastAsst.innerHTML();
    expect(html).toContain('起名分析');
    expect(html).toContain('陈明远');
  });
});
