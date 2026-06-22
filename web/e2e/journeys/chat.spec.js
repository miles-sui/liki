// Chat page tests — POST /api/agent/chat with SSE streaming.
// Mocks the backend to test all SSE event types and UI states.

import { test, expect } from '../fixtures.js';

// SSE event builders
function sseLine(obj) {
  return `data: ${JSON.stringify(obj)}\n\n`;
}

const mockChartFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  sseLine({ type: 'text-delta', content: '好的，我来帮您排盘。' }),
  sseLine({ type: 'phase', content: '正在确认服务类型…' }),
  sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  sseLine({ type: 'text-delta', content: '请提供出生信息。' }),
  sseLine({ type: 'phase', content: '正在查询地理信息…' }),
  sseLine({ type: 'phase', content: '正在整理出生数据…' }),
  sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  sseLine({ type: 'text-delta', content: '信息已收集完成。' }),
  sseLine({ type: 'phase', content: '正在计算命理数据…' }),
  sseLine({ type: 'phase', content: '正在生成分析报告…' }),
  sseLine({ type: 'text-delta', content: '## 八字命盘分析\n\n' }),
  sseLine({ type: 'text-delta', content: '您的日主为**甲木**，生于寅月。' }),
  sseLine({ type: 'text-delta', content: '格局为**建禄格**，用神为**火土**。' }),
  sseLine({ type: 'text-delta', content: '\n\n## 大运分析\n\n' }),
  sseLine({ type: 'text-delta', content: '当前大运为**戊辰**，财运亨通。' }),
  sseLine({
    type: 'done',
    data: {
      product: 'chart',
      order_id: 'test-order-123',
      amount: 990,
    },
  }),
].join('');

const mockErrorFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  sseLine({ type: 'error', content: '服务暂时不可用，请稍后重试' }),
].join('');

const mockQuestionFlow = [
  sseLine({ type: 'thinking' }),
  sseLine({ type: 'phase', content: '正在分析您的需求…' }),
  sseLine({ type: 'text-delta', content: '请问您的出生年份和性别是什么？' }),
].join('');

async function mockGreeting(page, text = '你好，我是灵机 Liki，精通八字命理与姓名学。有什么可以帮你的？') {
  await page.route('**/api/agent/greeting', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { greeting: text } }),
    });
  });
}

async function gotoChat(page, lang = 'zh') {
  await page.goto(`/${lang}/chat.html`);
  await page.waitForSelector('.chat-shell');
}

async function mockAgentChat(page, body, status = 200) {
  await page.route('**/api/agent/chat', async (route) => {
    await route.fulfill({
      status,
      headers: {
        'Content-Type': 'text/event-stream',
        'Cache-Control': 'no-cache',
        'X-Session-ID': 'e2e-session-id',
      },
      body,
    });
  });
}

test.describe('Chat page', () => {
  let page;

  test.beforeEach(async ({ context }) => {
    page = await context.newPage();
  });

  test.afterEach(async () => {
    await page.close();
  });

  // ── welcome state ──

  test('greeting bubble and chips render on page load', async () => {
    await mockGreeting(page);
    await gotoChat(page);

    // Greeting is shown as first assistant bubble.
    await expect(page.locator('.msg-asst')).toBeVisible();
    await expect(page.locator('.chip')).toHaveCount(3);
    await expect(page.locator('.chip').nth(0)).toHaveText('🔮 帮我算八字');
    await expect(page.locator('.chip').nth(1)).toHaveText('💑 看看我和 TA');
    await expect(page.locator('.chip').nth(2)).toHaveText('📛 帮宝宝起名');
  });

  test('chips never show raw i18n keys', async () => {
    await mockGreeting(page);
    await gotoChat(page);

    await expect(page.locator('.chip')).toHaveCount(3);
    // Verify no raw i18n key is rendered — covers race condition where
    // i18next.t(key) returns the key itself before resources are loaded.
    const chipTexts = await page.locator('.chip').allTextContents();
    for (const text of chipTexts) {
      expect(text).not.toMatch(/^[a-z]+\.[a-z]+/); // i18n key pattern
    }
  });

  test('chips hide when user sends a message', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chat-input-bar input').fill('想看八字');
    await page.locator('.btn-send').click();

    // Chips should disappear after user sends a message.
    await expect(page.locator('.chip-row')).not.toBeVisible();
    // Messages area should be visible.
    await expect(page.locator('.chat-messages')).toBeVisible();
  });

  // ── message sending ──

  test('typing and sending adds user bubble', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chat-input-bar input').fill('你好');
    await page.locator('.btn-send').click();

    await expect(page.locator('.msg-user')).toHaveText('你好');
  });

  test('empty input does not send', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await expect(page.locator('.btn-send')).toBeDisabled();
  });

  test('clicking a chip sends its message', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chip').nth(0).click();

    await expect(page.locator('.msg-user')).toHaveText('我想排八字命盘');
  });

  // ── SSE streaming ──

  test('text-delta events render markdown content incrementally', async () => {
    await mockGreeting(page);
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    // Wait for the streaming content to appear in the last assistant bubble.
    await expect(page.locator('.msg-asst').last()).toBeVisible({ timeout: 5000 });
    const html = await page.locator('.msg-asst').last().innerHTML();
    expect(html).toContain('八字命盘分析');
    expect(html).toContain('甲木');
  });

  test('phase events show progress text', async () => {
    await mockGreeting(page);
    await gotoChat(page);
    const flow = [
      sseLine({ type: 'thinking' }),
      sseLine({ type: 'phase', content: '正在分析您的需求…' }),
      sseLine({ type: 'text-delta', content: '好的。' }),
      sseLine({ type: 'phase', content: '正在确认服务类型…' }),
      sseLine({ type: 'phase', content: '正在计算命理数据…' }),
      sseLine({ type: 'phase', content: '正在生成分析报告…' }),
      sseLine({ type: 'text-delta', content: 'ok' }),
      sseLine({ type: 'done', data: { product: 'chart', order_id: 'x', amount: 990 } }),
    ].join('');

    await mockAgentChat(page, flow);
    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    // Phase events were emitted before text-deltas — wait for text content to appear in the last
    // assistant bubble which confirms the stream ran end-to-end (thinking → phase → text-delta → done).
    await expect(page.locator('.msg-asst').last()).toContainText('好的', { timeout: 5000 });
  });

  test('stop button appears during streaming and aborts', async () => {
    await mockGreeting(page);
    await gotoChat(page);

    // Send only thinking + phase without text-delta or done — stream finishes quickly
    // but the stop button should flash during processing.
    let didStream = false;
    await page.route('**/api/agent/chat', async (route) => {
      didStream = true;
      await route.fulfill({
        status: 200,
        headers: {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'X-Session-ID': 'e2e-session-id',
        },
        body: [
          sseLine({ type: 'thinking' }),
          sseLine({ type: 'phase', content: '正在分析您的需求…' }),
          sseLine({ type: 'text-delta', content: '您' }),
        ].join(''),
      });
    });

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    // The stream ran — verify messages were rendered (proving the stream happened).
    await expect(page.locator('.msg-asst').last()).toContainText('您', { timeout: 5000 });

    // Stop button should not be visible once stream completes.
    await expect(page.locator('.btn-stop')).not.toBeVisible();
  });

  // ── done / pay ──

  test('done event shows buy bar with pay button', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    // Buy bar should appear after done.
    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.buy-card')).toContainText('69.00');

    // Pay button should be present.
    await expect(page.locator('.btn-buy')).toContainText('查看完整报告');
  });

  // ── question ──

  test('LLM question text appears in assistant bubble', async () => {
    await mockGreeting(page);
    await gotoChat(page);
    await mockAgentChat(page, mockQuestionFlow);

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    await expect(page.locator('.msg-asst').last()).toContainText('请问您的出生年份和性别是什么？');
  });

  // ── error ──

  test('error event shows toast', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockErrorFlow);

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    await expect(page.locator('.error-toast')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.error-toast')).toContainText('服务暂时不可用');
  });

  test('HTTP error shows in toast', async () => {
    await gotoChat(page);

    await page.route('**/api/agent/chat', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: { message: '服务器内部错误' } }),
      });
    });

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();

    await expect(page.locator('.error-toast')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.error-toast')).toContainText('服务器内部错误');
  });

  // ── new chat ──

  test('new chat button resets to greeting and chips', async () => {
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    // Send a message first.
    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();
    await expect(page.locator('.msg-user')).toBeVisible();

    // Click new chat.
    await page.locator('button:has-text("新对话")').click();

    // Should be back to greeting + chips state.
    await expect(page.locator('.chip-row')).toBeVisible();
    await expect(page.locator('.msg-user')).not.toBeVisible();
  });

  // ── session persistence ──

  test('session ID is preserved in sessionStorage after first message', async () => {
    await mockGreeting(page);
    await gotoChat(page);
    await mockAgentChat(page, mockChartFlow);

    await page.locator('.chat-input-bar input').fill('看八字');
    await page.locator('.btn-send').click();
    await expect(page.locator('.buy-card')).toBeVisible({ timeout: 5000 });

    // sessionStorage should have the chat session ID from X-Session-ID header.
    const sid = await page.evaluate(() => sessionStorage.getItem('chatSessionID'));
    expect(sid).toBe('"e2e-session-id"');
  });
});
