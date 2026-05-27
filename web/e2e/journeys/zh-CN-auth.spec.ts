import { test, expect } from '../fixtures';
import { register, loginViaToken } from '../helpers/api';

test.describe('zh-CN authentication', () => {
  const USER = 'e2e-zh-auth-' + Date.now();
  const PASSWORD = 'test12345678';

  test('zh-CN register page renders and submits', async ({ page }) => {
    await page.goto('/zh-CN/register');
    await expect(page.locator('h1')).toContainText('创建你的账号', { timeout: 10000 });

    // Verify key elements render in Chinese.
    await expect(page.getByText('用户名')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('至少 8 个字符')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button[type="submit"]')).toContainText('创建账号');

    // Fill and submit registration.
    await page.locator('input[autocomplete="username"]').fill(USER);
    await page.locator('input[autocomplete="new-password"]').fill(PASSWORD);
    await page.locator('input[autocomplete="email"]').fill('suiqiang+e2e-' + USER + '@foxmail.com');
    await page.locator('button[type="submit"]').click();

    // Should redirect to profile or assess page with user menu visible.
    await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 15000 });
  });

  test('zh-CN login page renders in Chinese', async ({ page }) => {
    await page.goto('/zh-CN/login');
    await expect(page.locator('h1')).toContainText('欢迎回来', { timeout: 10000 });

    // Verify key elements in Chinese.
    await expect(page.getByText('用户名')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button[type="submit"]')).toContainText('登录');
    await expect(page.getByText('忘记密码？')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('还没有账号？')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('创建一个')).toBeVisible({ timeout: 5000 });
  });

  test('zh-CN navbar shows correct locale links', async ({ page }) => {
    await page.goto('/zh-CN/');
    await expect(page.locator('body')).toBeVisible({ timeout: 10000 });

    // The navbar should show Chinese navigation links for anonymous users.
    await expect(page.locator('a[href*="register"], a[href*="login"], button:has-text("登录")').first()).toBeVisible({ timeout: 10000 });
  });

  test('zh-CN user menu shows Chinese labels', async ({ page }) => {
    // Register and login via API, then verify Chinese menu labels.
    const { token } = await register(USER + '3', PASSWORD);
    await loginViaToken(page, token, '/zh-CN/profile/' + encodeURIComponent(USER + '3'));

    await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 15000 });
    await page.locator('button[aria-label="User menu"]').click();

    // Should show Chinese menu items in the dropdown.
    await expect(page.getByText('退出登录').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('设置').first()).toBeVisible({ timeout: 3000 });
  });
});
