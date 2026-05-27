import { test, expect } from '../fixtures';
import { register, login, req } from '../helpers/api';

test.describe('Account deletion', () => {
  test('soft-delete account and reactivate by logging in within 7 days', async ({ page }) => {
    const name = 'e2e-delete-' + Date.now();
    const password = 'test12345678';
    const { token } = await register(name, password);

    // Set token in localStorage and visit profile.
    await page.goto('/en/profile/' + encodeURIComponent(name));
    await page.evaluate((t) => { localStorage.setItem('token', t); }, token);
    await page.reload();
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    const width = page.viewportSize()?.width ?? 1280;

    // Navigate to settings page, then delete account.
    if (width < 1024) {
      await page.goto('/en/settings');
      await page.waitForURL(/\/en\/settings/, { timeout: 10000 });
      await page.locator('#section-danger button:has-text("Delete Account")').click();
      await page.locator('button:has-text("Yes, delete my account")').click();
    } else {
      await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 10000 });
      await page.locator('button[aria-label="User menu"]').click();
      await page.locator('a:has-text("Settings")').first().click({ force: true });
      await page.waitForURL(/\/en\/settings/, { timeout: 10000 });
      await page.locator('#section-danger button:has-text("Delete Account")').click();
      await page.locator('button:has-text("Yes, delete my account")').click();
    }

    // After deletion, redirect to login page.
    await expect(page.locator('h1')).toContainText('Welcome back', { timeout: 10000 });

    // Login via API — must reactivate the account (within 7-day grace period).
    // Per API contract §POST /api/auth/login: login within 7 days clears deactivated_at.
    const loginResult = await login(name, password);
    expect(loginResult.token).toBeTruthy();
    expect(loginResult.userId).toBeGreaterThan(0);

    // Verify the token is usable by fetching the user profile via API.
    const { status, data } = await req('GET', '/api/users/me', undefined, loginResult.token);
    expect(status).toBe(200);
    expect(data.name).toBe(name);

    // Clean up: delete the reactivated account via API.
    await req('DELETE', '/api/users/me', undefined, loginResult.token);
  });
});
