import { test, expect } from '../fixtures';

test('register, logout, login, and delete account lifecycle', async ({ page }) => {
  const USER = 'e2e-acct-' + Date.now();
  const PASSWORD = 'test12345678';

  // 1. Register
  await page.goto('/en/register');
  await expect(page.locator('h1')).toContainText('Create your account');

  await page.locator('input[autocomplete="username"]').fill(USER);
  await page.locator('input[autocomplete="new-password"]').fill(PASSWORD);
  await page.locator('input[autocomplete="email"]').fill('suiqiang+e2e-' + USER + '@foxmail.com');
  await page.locator('button[type="submit"]').click();

  // Registration redirects to the profile page.
  await page.waitForURL(new RegExp('/en/profile/' + USER.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')), { timeout: 10000 });

  // 2. Logout — user is now authenticated.
  const width = page.viewportSize()?.width ?? 1280;
  if (width < 1024) {
    // Mobile: use hamburger menu.
    await page.locator('button[aria-label="Toggle menu"]').click();
    await page.locator('button:has-text("Logout")').last().click();
  } else {
    // Desktop: use user menu dropdown.
    await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 5000 });
    await page.locator('button[aria-label="User menu"]').click();
    await page.locator('button:has-text("Logout")').first().click();
  }

  // After logout, redirect to /en/.
  await page.waitForURL(/\/en/, { timeout: 10000 });
  await page.evaluate(() => localStorage.clear());

  // 3. Login
  await page.goto('/en/login');
  await expect(page.locator('h1')).toContainText('Welcome back');
  await page.locator('input[autocomplete="username"]').fill(USER);
  await page.locator('input[autocomplete="current-password"]').fill(PASSWORD);
  await page.locator('button[type="submit"]').click();

  // After login, redirect to profile.
  await page.waitForURL(new RegExp('/en/profile/' + USER.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')), { timeout: 10000 });

  // 4. Delete account — via Settings page.
  if (width < 1024) {
    // Navigate directly — mobile drawer transitions make menu clicks brittle.
    await page.goto('/en/settings');
    await page.waitForURL(/\/en\/settings/, { timeout: 10000 });
    // Open the delete modal.
    await page.locator('#section-danger button:has-text("Delete Account")').click();
    // Confirm in the modal (Alpine custom modal, not browser confirm()).
    await page.locator('button:has-text("Yes, delete my account")').click();
  } else {
    await expect(page.locator('button[aria-label="User menu"]')).toBeVisible({ timeout: 5000 });
    await page.locator('button[aria-label="User menu"]').click();
    await page.locator('a:has-text("Settings")').first().click({ force: true });
    await page.waitForURL(/\/en\/settings/, { timeout: 10000 });
    // Open the delete modal.
    await page.locator('#section-danger button:has-text("Delete Account")').click();
    // Confirm in the modal (Alpine custom modal, not browser confirm()).
    await page.locator('button:has-text("Yes, delete my account")').click();
  }

  // After deletion, redirect to login page.
  await page.waitForURL(/\/en\/login/, { timeout: 10000 });
  await expect(page.getByText('Welcome back')).toBeVisible({ timeout: 5000 });
});
