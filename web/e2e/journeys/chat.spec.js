// Chat page tests — JWT login + POST /api/agent/naming SSE streaming.
// Mocks all backend APIs since JWT is HttpOnly and can't be set from tests.

import { test, expect } from '../fixtures.js';
import { sseLine, mockLogin, mockOrderStatus, mockNamingSSE } from '../mocks.js';

// ── chat-specific mock helpers ──

async function mockLoginError(page) {
  await page.route('**/api/auth/login', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({ error: { message: 'No valid order found for this email' } }),
    });
  });
}

async function mockLoginMulti(page) {
  await page.route('**/api/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          orders: [
            { order_id: 'o-001', summary: '起名服务 #1', expires_at: '2027-01-01T00:00:00Z', has_birth_info: false },
            { order_id: 'o-002', summary: '起名服务 #2', expires_at: '2027-02-01T00:00:00Z', has_birth_info: true },
          ],
        },
      }),
    });
  });
}

async function mockOrderStatusExpired(page, orderID = 'e2e-order-123') {
  await page.route(`**/api/orders/${orderID}/status`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: { order_id: orderID, status: 'paid', product: 'naming', chat_expires_at: '2025-01-01 00:00:00', amount: 2990, currency: 'USD' },
      }),
    });
  });
}

async function mockOrderSelect(page, orderID = 'o-001') {
  await page.route('**/api/orders/select', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { order_id: orderID } }),
    });
  });
}

// ── shared flows ──

async function gotoChat(page, lang = 'zh-Hans') {
  await page.goto(`/${lang}/chat.html`);
  await page.waitForSelector('.chat-shell');
  await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
}

async function enterChatViaLogin(page) {
  await page.locator('.login-form input[type="email"]').fill('user@example.com');
  await page.locator('.login-form button').click();
  await page.waitForSelector('.chat-messages', { timeout: 10000 });
}

// ── SSE flows ──

const namingSSEFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在分析八字…' }),
  sseLine({ type: 'text-delta', content: '好的，我来帮您分析起名方案。' }),
  sseLine({ type: 'text-delta', content: '\n\n## 八字分析\n\n' }),
  sseLine({ type: 'text-delta', content: '您的日主为**甲木**，生于寅月。' }),
].join('');

const errorSSEFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'error', content: '服务暂时不可用，请稍后重试' }),
].join('');

const phaseOnlyFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在查询地理信息…' }),
  sseLine({ type: 'phase', content: '正在计算八字…' }),
  sseLine({ type: 'text-delta', content: '好的。' }),
].join('');

// ── data-driven error test cases ──

const SSE_ERROR_CASES = [
  {
    name: 'SSE error event shows toast',
    setup: async (page) => { await mockNamingSSE(page, errorSSEFlow); },
    assert: async (page) => {
      await expect(page.locator('.error-toast')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('.error-toast')).toContainText('服务暂时不可用');
    },
  },
  {
    name: 'HTTP 500 shows in toast',
    setup: async (page) => {
      await page.route('**/api/agent/naming', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: { message: '服务器内部错误' } }),
        });
      });
    },
    assert: async (page) => {
      await expect(page.locator('.error-toast')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('.error-toast')).toContainText('服务器内部错误');
    },
  },
  {
    name: '401 during SSE returns to login state',
    setup: async (page) => {
      await page.route('**/api/agent/naming', async (route) => {
        await route.fulfill({ status: 401, contentType: 'application/json', body: '{}' });
      });
    },
    assert: async (page) => {
      await expect(page.locator('.login-form')).toBeVisible({ timeout: 5000 });
    },
  },
  {
    name: 'SSE abort mid-stream preserves partial content',
    setup: async (page) => {
      await page.route('**/api/agent/naming', async (route) => {
        await route.fulfill({
          status: 200,
          headers: { 'Content-Type': 'text/event-stream' },
          body: [
            sseLine({ type: 'thinking' }),
            sseLine({ type: 'text-delta', content: '好的，我来帮您' }),
          ].join(''),
        });
      });
    },
    assert: async (page) => {
      await expect(page.locator('.msg-asst').last()).toContainText('好的，我来帮您', { timeout: 10000 });
      await expect(page.locator('.btn-send')).toBeVisible({ timeout: 5000 });
    },
  },
  {
    name: 'SSE partial content preserved on stream end',
    setup: async (page) => {
      await page.route('**/api/agent/naming', async (route) => {
        await route.fulfill({
          status: 200,
          headers: { 'Content-Type': 'text/event-stream' },
          body: [
            sseLine({ type: 'thinking' }),
            sseLine({ type: 'text-delta', content: '您的日主为乙木' }),
          ].join(''),
        });
      });
    },
    assert: async (page) => {
      await expect(page.locator('.msg-asst').last()).toContainText('您的日主为乙木', { timeout: 10000 });
      await expect(page.locator('.btn-send')).toBeVisible({ timeout: 5000 });
    },
  },
];

// ============================================================================

test.describe('Chat page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── unauthenticated ──

  test('loads in login state by default', async () => {
    await gotoChat(page);
    await expect(page.locator('.login-form')).toBeVisible();
    await expect(page.locator('.login-form button')).toContainText('继续');
    await expect(page.locator('.chat-shell')).toBeVisible();
  });

  test('login button is disabled when email is empty', async () => {
    await gotoChat(page);
    await expect(page.locator('.login-form button')).toBeDisabled();
  });

  // ── login flow ──

  test('successful login with single order transitions to chat', async () => {
    await mockLogin(page);
    await mockOrderStatus(page);
    await gotoChat(page);
    await enterChatViaLogin(page);

    await expect(page.locator('.login-form')).not.toBeVisible();
    await expect(page.locator('.msg-asst')).toBeVisible();
  });

  test('login error shows inline error message', async () => {
    await mockLoginError(page);
    await gotoChat(page);

    await page.locator('.login-form input[type="email"]').fill('nobody@example.com');
    await page.locator('.login-form button').click();

    await expect(page.locator('.login-form p.text-red-500')).toBeVisible({ timeout: 5000 });
  });

  test('multiple orders show order selection state', async () => {
    await mockLoginMulti(page);
    await gotoChat(page);

    await page.locator('.login-form input[type="email"]').fill('user@example.com');
    await page.locator('.login-form button').click();

    await expect(page.locator('.order-list')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.order-card')).toHaveCount(2);
    await expect(page.locator('.order-card').first()).toContainText('起名服务 #1');
  });

  test('clicking order card transitions to chat', async () => {
    await mockLoginMulti(page);
    await mockOrderSelect(page);
    await mockOrderStatus(page, 'o-001');
    await gotoChat(page);

    await page.locator('.login-form input[type="email"]').fill('user@example.com');
    await page.locator('.login-form button').click();
    await expect(page.locator('.order-list')).toBeVisible({ timeout: 5000 });

    await page.locator('.order-card').first().click();
    await expect(page.locator('.chat-messages')).toBeVisible({ timeout: 5000 });
  });

  // ── authenticated ──

  test.describe('authenticated', () => {
    test.beforeEach(async () => {
      await mockLogin(page);
      await mockOrderStatus(page);
      await gotoChat(page);
      await enterChatViaLogin(page);
    });

    test('typing and sending adds user bubble', async () => {
      await mockNamingSSE(page, namingSSEFlow);

      await page.locator('#chat-input').fill('我的姓氏是陈');
      await page.locator('.btn-send').click();

      await expect(page.locator('.msg-user')).toContainText('我的姓氏是陈');
    });

    test('empty input does not send', async () => {
      await expect(page.locator('.btn-send')).toBeDisabled();
    });

    test('text-delta events render markdown as HTML', async () => {
      await mockNamingSSE(page, namingSSEFlow);

      await page.locator('#chat-input').fill('帮我起名');
      await page.locator('.btn-send').click();

      const lastAsst = page.locator('.msg-asst').last();
      await expect(lastAsst).toBeVisible({ timeout: 10000 });
      const html = await lastAsst.innerHTML();
      expect(html).toContain('八字分析');
      expect(html).toContain('甲木');
    });

    test('phase events show progress text', async () => {
      await mockNamingSSE(page, phaseOnlyFlow);

      await page.locator('#chat-input').fill('帮我起名');
      await page.locator('.btn-send').click();

      await expect(page.locator('.msg-asst').last()).toContainText('好的', { timeout: 10000 });
    });

    test('stop button appears during streaming and aborts', async () => {
      await page.route('**/api/agent/naming', () => new Promise(() => {}));

      await page.locator('#chat-input').fill('hi');
      await page.locator('.btn-send').click();

      await expect(page.locator('.btn-stop')).toBeVisible({ timeout: 5000 });
      await page.locator('.btn-stop').click();
      await expect(page.locator('.btn-send')).toBeVisible({ timeout: 5000 });
    });

    test('new chat button clears messages', async () => {
      await mockNamingSSE(page, namingSSEFlow);

      await page.locator('#chat-input').fill('帮我起名');
      await page.locator('.btn-send').click();
      await expect(page.locator('.msg-user')).toBeVisible({ timeout: 10000 });

      await page.locator('.btn-newchat').click();
      await expect(page.locator('.msg-user')).not.toBeVisible();
    });

    test('Enter key sends message', async () => {
      await mockNamingSSE(page, namingSSEFlow);

      await page.locator('#chat-input').fill('我的姓氏是陈');
      await page.locator('#chat-input').press('Enter');

      await expect(page.locator('.msg-user')).toContainText('我的姓氏是陈');
    });

    test('close button dismisses error toast', async () => {
      await mockNamingSSE(page, errorSSEFlow);

      await page.locator('#chat-input').fill('帮我起名');
      await page.locator('.btn-send').click();
      await expect(page.locator('.error-toast')).toBeVisible({ timeout: 5000 });

      await page.locator('.error-toast button').click();
      await expect(page.locator('.error-toast')).not.toBeVisible();
    });

    test('chat input receives focus on click', async () => {
      await page.locator('#chat-input').click();
      await expect(page.locator('#chat-input')).toBeFocused();
    });

    // Data-driven SSE error handling tests.
    for (const { name, setup, assert } of SSE_ERROR_CASES) {
      test(name, async () => {
        await setup(page);
        await page.locator('#chat-input').fill('帮我起名');
        await page.locator('.btn-send').click();
        await assert(page);
      });
    }
  });

  // ── mobile ──

  test.describe('mobile viewport', () => {
    test.beforeEach(async () => {
      await page.setViewportSize({ width: 375, height: 812 });
      await mockLogin(page);
      await mockOrderStatus(page);
      await gotoChat(page);
      await enterChatViaLogin(page);
    });

    test('shows input toolbar and sends message', async () => {
      await mockNamingSSE(page, namingSSEFlow);

      await page.locator('#chat-input').fill('你好');
      await page.locator('.btn-send').click();

      await expect(page.locator('.msg-user')).toContainText('你好');
    });

    test('shows header in compact layout', async ({ context }) => {
      // Need fresh page — only check layout before entering chat.
      await page.close();
      page = await context.newPage();
      await page.setViewportSize({ width: 375, height: 812 });
      await mockLogin(page);
      await mockOrderStatus(page);
      await gotoChat(page);
      await expect(page.locator('header')).toBeVisible();
    });
  });

  // ── expired order ──

  test('expired order shows expired banner and hides input', async () => {
    await mockLogin(page);
    await mockOrderStatusExpired(page);
    await gotoChat(page);

    await page.locator('.login-form input[type="email"]').fill('user@example.com');
    await page.locator('.login-form button').click();

    await expect(page.locator('.expired-banner')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#chat-input')).not.toBeVisible();
  });
});
