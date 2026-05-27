import { test, expect } from '../fixtures';
import { register, submitAssessment, loginViaToken } from '../helpers/api';

test.describe('Donate page', () => {
  test('anonymous visitor sees login prompt and no tiers', async ({ page }) => {
    await page.goto('/en/donate');
    await expect(page).toHaveTitle(/Donate/);

    // Login prompt should be visible for anonymous users.
    await expect(page.getByText('Log in to make a donation')).toBeVisible({ timeout: 5000 });

    // Donation tier buttons should NOT be visible.
    await expect(page.locator('button:has-text("$")').first()).not.toBeVisible({ timeout: 3000 });
  });

  test('logged-in user sees donation tiers and can select one', async ({ page }) => {
    const name = 'e2e-donate-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await loginViaToken(page, token, '/en/donate');

    await expect(page).toHaveTitle(/Donate/);

    // Three donation tier cards should be visible.
    const tierCards = page.locator('.grid.grid-cols-1 .card.cursor-pointer');
    await expect(tierCards.first()).toBeVisible({ timeout: 5000 });
    expect(await tierCards.count()).toBe(3);

    // Verify tier labels ($9.90, $19.90, $29.90).
    await expect(page.getByText('$9.90')).toBeVisible({ timeout: 3000 });
    await expect(page.getByText('$19.90')).toBeVisible({ timeout: 3000 });
    await expect(page.getByText('$29.90')).toBeVisible({ timeout: 3000 });

    // Click the middle tier ($19.90).
    await tierCards.nth(1).click();

    // The Donate CTA button should be present and enabled.
    // Use exact match — the user name may contain "donate" (e2e-donate-xxx).
    const cta = page.getByRole('button', { name: 'Donate', exact: true });
    await expect(cta).toBeVisible({ timeout: 3000 });
    await expect(cta).toBeEnabled({ timeout: 3000 });
  });

  test('zh-CN donate page renders with Chinese text', async ({ page }) => {
    await page.goto('/zh-CN/donate');
    // Note: page title comes from frontmatter and is not yet locale-aware.
    await expect(page.locator('h2').first()).toBeVisible({ timeout: 10000 });

    // Chinese login prompt.
    await expect(page.getByText('登录以进行捐赠。')).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Donate links from other pages', () => {
  test('footer has donate link on static pages', async ({ page }) => {
    await page.goto('/en/about');

    // Footer support section with donate link.
    const footerDonate = page.locator('footer a[href*="donate"]');
    await expect(footerDonate.first()).toBeVisible({ timeout: 5000 });
    await expect(footerDonate.first()).toHaveText('Donate');
  });

  test('landing page has donate link below CTA', async ({ page }) => {
    await page.goto('/en/');

    // Subtle donate link below the CTA banner.
    const donateLink = page.locator('a[href*="donate"]').filter({ hasText: 'Support 25types' });
    await expect(donateLink.first()).toBeVisible({ timeout: 5000 });
  });

  test('profile page has donate link when logged in', async ({ page }) => {
    const name = 'e2e-result-donate-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));

    // Donate link in the navbar or page.
    await expect(page.locator('a[href*="donate"]').first()).toBeVisible({ timeout: 5000 });
  });
});
