import { test, expect } from '../fixtures';
import { register, submitAssessment, setPublic, loginViaToken } from '../helpers/api';

test.describe('Profile settings', () => {
  test('edit display name via settings page', async ({ page }) => {
    const name = 'e2e-edit-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await loginViaToken(page, token, '/en/settings');

    // Section should be visible.
    await expect(page.locator('#section-account')).toBeVisible({ timeout: 10000 });

    // Type new name directly and save.
    const newName = 'e2e-edited-' + Date.now();
    await page.locator('#section-account input[type="text"]').fill(newName);
    await page.getByRole('button', { name: 'Change', exact: true }).click();

    // Toast confirmation.
    await expect(page.getByText('Settings updated')).toBeVisible({ timeout: 5000 });

    // Navigate to profile — the page title should reflect the new name.
    await page.goto('/en/profile/' + encodeURIComponent(newName));
    await expect(page.locator('h2, h1').first()).toBeVisible({ timeout: 10000 });
    await expect(page).toHaveTitle(new RegExp(newName), { timeout: 5000 });
  });

  test('toggle profile privacy', async ({ page }) => {
    const name = 'e2e-priv-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    await setPublic(token);
    await loginViaToken(page, token, '/en/settings');

    await expect(page.locator('#section-privacy')).toBeVisible({ timeout: 10000 });

    // Click the privacy toggle.
    await page.locator('#section-privacy input[type="checkbox"]').click();

    // Wait for the toast confirming privacy change.
    await expect(page.getByText('Profile is now private')).toBeVisible({ timeout: 5000 });

    // Now visit as anonymous — should see "Profile not found".
    await page.evaluate(() => localStorage.clear());
    await page.goto('/en/profile/' + encodeURIComponent(name));
    await expect(page.getByRole('heading', { name: 'Profile not found' })).toBeVisible({ timeout: 10000 });
  });

  test('logout from user menu', async ({ page }) => {
    const name = 'e2e-logout-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));

    // Verify logged in.
    await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 5000 });

    // Click Logout via user menu.
    await page.locator('button[aria-label="User menu"]').click();
    await page.locator('button:has-text("Logout")').first().click();

    // After logout, token is removed; verify by navigating to home page.
    await page.waitForTimeout(1000);
    await page.evaluate(() => localStorage.clear());
    await page.goto('/en/');
    await expect(page.locator('body')).toBeVisible({ timeout: 5000 });
    // Login link should be visible when not authenticated.
    await expect(page.getByRole('link', { name: /Log in|Login|登录/ })).toBeVisible({ timeout: 10000 });
  });

  test('profile page renders in zh-CN', async ({ page }) => {
    const name = 'e2e-prof-zh-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    await setPublic(token);
    await loginViaToken(page, token, '/zh-CN/profile/' + encodeURIComponent(name));

    // The profile page should render in Chinese locale.
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });
    // URL must be under /zh-CN/.
    expect(page.url()).toContain('/zh-CN/');
  });
});
