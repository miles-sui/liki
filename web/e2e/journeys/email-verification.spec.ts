import { test, expect } from '../fixtures';
import { req, register, getEmailVerToken, loginViaToken } from '../helpers/api';

test.describe('Email verification', () => {
  test('verify-email page shows error with no token', async ({ page }) => {
    await page.goto('/en/verify-email');
    await expect(page.locator('h1')).toContainText('Email Verification', { timeout: 10000 });

    // No token param — API call fails → error message shown.
    await expect(page.locator('.text-error').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.text-error').first()).toContainText('Invalid or expired token');
  });

  test('verify-email page shows error with invalid token', async ({ page }) => {
    await page.goto('/en/verify-email?token=invalid-token-12345');
    await expect(page.locator('h1')).toContainText('Email Verification', { timeout: 10000 });
    await expect(page.locator('.text-error').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.text-error').first()).toContainText('Invalid or expired token');
  });

  test('full email verification flow with real email', async ({ page }) => {
    const name = 'e2e-email-' + Date.now();
    const password = 'test12345678';
    const email = 'e2e-verify-' + Math.random().toString(36).slice(2) + '@foxmail.com';

    // 1. Register without email (registration doesn't accept email).
    const { token, userId } = await register(name, password);

    // 2. Update profile to set email — this triggers the verification email.
    await req('PATCH', '/api/users/me', { email }, token);

    // 3. Retrieve the verification token from the database by user ID.
    const verToken = getEmailVerToken(userId);
    test.skip(!verToken, 'Email verification token not found — RESEND_API_KEY may not be configured');
    if (!verToken) return;

    // 4. Visit the verify-email page with the token.
    await page.goto('/en/verify-email?token=' + encodeURIComponent(verToken!));
    await expect(page.locator('h1')).toContainText('Email Verification', { timeout: 10000 });

    // 5. Should show success.
    await expect(page.locator('text=Email verified successfully')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('text=Go to Settings')).toBeVisible({ timeout: 5000 });

    // 6. Login and check email_verified status via API.
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));
    await expect(page.locator('body')).toBeVisible({ timeout: 5000 });

    const verified = await page.evaluate(async (t) => {
      const res = await fetch('/api/users/me', {
        headers: { Authorization: 'Bearer ' + t },
      });
      const json = await res.json();
      return json.data?.email_verified === true;
    }, token);
    expect(verified).toBe(true);
  });
});
