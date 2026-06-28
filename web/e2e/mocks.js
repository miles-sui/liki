// Shared Playwright API mock helpers.
// Use page.route() to intercept and fulfill backend requests.
// JWT is HttpOnly — can't be set from tests, so all APIs are mocked.

export function sseLine(obj) {
  return `data: ${JSON.stringify(obj)}\n\n`;
}

export async function mockLogin(page, orderID = 'e2e-order-123') {
  await page.route('**/api/auth/login', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { order_id: orderID, has_birth_info: false, redirect: '/chat' } }),
    });
  });
}

export async function mockOrderStatus(page, orderID = 'e2e-order-123') {
  await page.route(`**/api/orders/${orderID}/status`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: { order_id: orderID, status: 'paid', product: 'naming', chat_expires_at: '2027-06-26 00:00:00', amount: 2990, currency: 'USD' },
      }),
    });
  });
}

export async function mockNamingSSE(page, events) {
  await page.route('**/api/agent/naming', async (route) => {
    await route.fulfill({
      status: 200,
      headers: { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' },
      body: events,
    });
  });
}
