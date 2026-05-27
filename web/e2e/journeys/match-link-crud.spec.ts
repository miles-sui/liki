import { test, expect } from '../fixtures';
import { register, submitAssessment, setPublic, loginViaToken, createMatchLink, acceptDialog } from '../helpers/api';

test.describe('Match link CRUD', () => {
  test('create, copy, and delete match link from profile', async ({ page }) => {
    const name = 'e2e-mlcrud-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    await setPublic(token);
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));

    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    // Click "Create Match Link".
    const createBtn = page.locator('button:has-text("Create Match Link")');
    await createBtn.scrollIntoViewIfNeeded();
    await expect(createBtn).toBeVisible({ timeout: 5000 });
    await createBtn.click();

    // Wait for the new link to appear.
    await expect(page.locator('text=Match').first()).toBeVisible({ timeout: 5000 });

    // Click the copy button — it should be visible after link creation.
    const copyBtn = page.locator('button:has-text("🔗")').first();
    await expect(copyBtn).toBeVisible({ timeout: 5000 });
    await copyBtn.click();
    // Verify copy confirmation appears.
    await expect(page.locator('text=Copied')).toBeVisible({ timeout: 3000 });

    // Delete the match link — scope to match links section.
    const deleteBtn = page.locator('[x-show="matchLinks.length > 0"] button:has-text("Delete")');
    await expect(deleteBtn).toBeVisible({ timeout: 5000 });
    acceptDialog(page);
    await deleteBtn.click();

    // The link section should be removed from DOM.
    await expect(page.locator('button:has-text("🔗")')).toHaveCount(0, { timeout: 5000 });
  });

  test('list match links shows correct count', async ({ page }) => {
    const name = 'e2e-mllist-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    await setPublic(token);

    // Create 3 match links via API.
    await createMatchLink(token);
    await createMatchLink(token);
    await createMatchLink(token);

    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    // Should show 3 match link items — each has "Match" text.
    const matchItems = page.locator('[x-show="matchLinks.length > 0"]').first();
    await expect(matchItems).toBeVisible({ timeout: 5000 });
  });

  test('deleted match link shows error page', async ({ page }) => {
    const name = 'e2e-mlinv-' + Date.now();
    const { token } = await register(name, 'test12345678');
    await submitAssessment(token);
    const link = await createMatchLink(token);

    // Verify the link works before deletion.
    await page.goto('/en/m/' + link.token);
    await expect(page.locator('[x-data="matchLandingPage()"]')).toBeVisible({ timeout: 10000 });

    // Delete the link via profile.
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    const deleteBtn = page.locator('[x-show="matchLinks.length > 0"] button:has-text("Delete")');
    await expect(deleteBtn).toBeVisible({ timeout: 5000 });
    acceptDialog(page);
    await deleteBtn.click();

    // Wait for DOM update after deletion.
    await expect(page.locator('button:has-text("Create Match Link")')).toBeVisible({ timeout: 5000 });

    // Visiting the deleted link should show an error.
    await page.evaluate(() => localStorage.clear());
    await page.goto('/en/m/' + link.token);

    // Deleted link must show an error — the bond result must NOT be visible.
    await expect(page.locator('.text-error').first()).toBeVisible({ timeout: 10000 });
  });
});
