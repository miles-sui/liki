import { test, expect } from '../fixtures';
import { register, req } from '../helpers/api';

test.describe('Password reset flow', () => {
  const TS = Date.now();
  const NAME = `reset-test-${TS}`;
  const PASSWORD = 'test-pass-123';
  const NEW_PASSWORD = 'new-pass-456';
  const EMAIL = `reset-test-${TS}@example.com`;

  test('forgot password page renders and submits with real email', async ({ page }) => {
    // Create a user and set their email.
    const { token } = await register(NAME, PASSWORD);
    await req('PATCH', '/api/users/me', { email: EMAIL }, token);

    await page.goto('/en/forgot-password');
    await expect(page.locator('h1')).toContainText('Reset Password');

    // Fill with the user's actual email.
    await page.fill('input[type="email"]', EMAIL);
    await page.locator('button[type="submit"]').click();

    // Per API contract, endpoint always returns 200 (anti-enumeration).
    await expect(page.locator('text=If that email is registered')).toBeVisible({ timeout: 10000 });
  });

  test('reset password page shows error on invalid token', async ({ page }) => {
    await page.goto('/en/reset-password');
    await expect(page.locator('h1')).toContainText('Set New Password');

    await page.fill('input[type="password"]', NEW_PASSWORD);
    await page.locator('button[type="submit"]').click();

    // With no token in URL, the API returns an error.
    await expect(page.locator('.text-error').first()).toBeVisible({ timeout: 10000 });
  });
});
