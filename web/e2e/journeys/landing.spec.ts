import { test, expect } from '../fixtures';
import { register, loginViaToken } from '../helpers/api';

test.describe('Landing page', () => {
  test('anonymous visitor sees hero, name input, CTA, and examples', async ({ page }) => {
    await page.goto('/en/');
    await expect(page.locator('h1')).toBeVisible({ timeout: 10000 });

    // Hero text should be visible.
    await expect(page.getByText('Enter your name')).toBeVisible({ timeout: 5000 });

    // Name input field.
    const nameInput = page.locator('input[type="text"]').first();
    await expect(nameInput).toBeVisible({ timeout: 5000 });

    // CTA button should be disabled until name is entered.
    const startBtn = page.locator('button[type="submit"]').first();
    await expect(startBtn).toBeDisabled({ timeout: 5000 });

    // Typing a name enables the button.
    await nameInput.fill('Test');
    await expect(startBtn).toBeEnabled({ timeout: 5000 });

    // Example profiles section.
    await expect(page.getByText('See what you\'ll discover')).toBeVisible({ timeout: 5000 });

    // Example profile cards (3 cards).
    const exampleCards = page.locator('#example-profiles .card-fade-up');
    await expect(exampleCards.first()).toBeVisible({ timeout: 5000 });
    expect(await exampleCards.count()).toBe(3);

    // Bottom CTA.
    await expect(page.locator('a[href*="assess"]').last()).toBeVisible({ timeout: 5000 });

    // Stats badge.
    await expect(page.locator('#home-stats')).toBeVisible({ timeout: 5000 });
  });

  test('logged-in user sees welcome back and profile link', async ({ page }) => {
    const name = 'e2e-landing-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await loginViaToken(page, token, '/en/');

    // Hero section should show "Welcome back" with the user's name.
    await expect(page.getByText('Welcome back')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(name).first()).toBeVisible({ timeout: 5000 });

    // Should have a "View Your Profile" link.
    await expect(page.locator('a:has-text("View Your Profile")')).toBeVisible({ timeout: 5000 });

    // Should also have a "Re-assess" link.
    await expect(page.locator('a:has-text("Re-assess")')).toBeVisible({ timeout: 5000 });
  });

  test('zh-CN landing page renders with Chinese title', async ({ page }) => {
    await page.goto('/zh-CN/');
    await expect(page.locator('h1')).toBeVisible({ timeout: 10000 });

    // Title must include the Chinese brand name.
    await expect(page).toHaveTitle(/真象/);

    // Name input field with Chinese placeholder text.
    const nameInput = page.locator('input[type="text"]').first();
    await expect(nameInput).toBeVisible({ timeout: 5000 });
  });
});
