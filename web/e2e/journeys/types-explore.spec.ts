import { test, expect } from '../fixtures';

test.describe('Types exploration', () => {
  test('types gallery page renders with filters and tiles', async ({ page }) => {
    await page.goto('/en/types.html');
    await expect(page.locator('h1')).toContainText('25 Personality Types', { timeout: 10000 });

    // Filter buttons should be visible (All + 5 elements).
    const filterBtns = page.locator('.flex-wrap.gap-2 button');
    await expect(filterBtns.first()).toBeVisible({ timeout: 5000 });
    const btnCount = await filterBtns.count();
    expect(btnCount).toBeGreaterThanOrEqual(6); // All + Wood/Fire/Earth/Metal/Water

    // Type tiles should be rendered.
    const tiles = page.locator('.grid a.card');
    await expect(tiles.first()).toBeVisible({ timeout: 5000 });
    const tileCount = await tiles.count();
    expect(tileCount).toBe(25); // 5×5 matrix

    // Click a filter (e.g., "Wood").
    const woodBtn = page.locator('.flex-wrap.gap-2 button').filter({ hasText: 'Wood' });
    if (await woodBtn.isVisible()) {
      await woodBtn.click();
      // The gallery updates via Alpine reactivity — wait for grid to re-render.
      await page.waitForTimeout(500);

      // After filtering, fewer tiles should show (only Wood-primary types).
      const filteredTiles = page.locator('.grid a.card');
      const filteredCount = await filteredTiles.count();
      expect(filteredCount).toBeGreaterThan(0);
      expect(filteredCount).toBeLessThanOrEqual(5); // Wx: WW, WF, WE, WM, WR
    }
  });

  test('type detail page renders sections', async ({ page }) => {
    // Visit a specific type detail page (Wood-Fire = WF).
    await page.goto('/en/types/WF/index.html');
    await expect(page.locator('h1')).toBeVisible({ timeout: 10000 });

    // Should have element badges.
    const badges = page.locator('.badge');
    await expect(badges.first()).toBeVisible({ timeout: 5000 });

    // Should have a portrait section or description.
    const content = page.locator('.card-body, .max-w-3xl');
    await expect(content.first()).toBeVisible({ timeout: 5000 });

    // Should have a CTA at the bottom.
    await expect(page.locator('a[href*="assess"]').last()).toBeVisible({ timeout: 5000 });
  });

  test('types gallery zh-CN renders', async ({ page }) => {
    await page.goto('/zh-CN/types.html');
    await expect(page.locator('h1')).toContainText('25', { timeout: 10000 });
    // Filter buttons should include Chinese element names.
    await expect(page.locator('button:has-text("木")')).toBeVisible({ timeout: 5000 });
  });

  test('type detail page zh-CN renders', async ({ page }) => {
    await page.goto('/zh-CN/types/WF/index.html');
    await expect(page.locator('h1')).toBeVisible({ timeout: 10000 });
    // Should have Chinese element badge (first badge is the primary element).
    await expect(page.locator('.badge').first()).toBeVisible({ timeout: 5000 });
  });

  test('navigation from gallery to detail and back', async ({ page }) => {
    await page.goto('/en/types.html');
    await expect(page.locator('h1')).toBeVisible({ timeout: 10000 });

    // Click the first type tile.
    const firstTile = page.locator('.grid a.card').first();
    const tileHref = await firstTile.getAttribute('href');
    await firstTile.click();

    // Should navigate to the detail page.
    await page.waitForURL(/\/en\/types\//, { timeout: 5000 });
    await expect(page.locator('h1')).toBeVisible({ timeout: 5000 });

    // Navigate back to gallery.
    await page.goto('/en/types.html');
    await expect(page.locator('h1')).toBeVisible({ timeout: 5000 });
  });
});
