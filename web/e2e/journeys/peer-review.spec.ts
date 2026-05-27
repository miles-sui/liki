import { test, expect } from '../fixtures';
import { register, submitAssessment, createReviewLink, renewReviewLink, loginViaToken, acceptDialog } from '../helpers/api';

test('create link, peer review, view profile, delete link', async ({ page, context }) => {
  const name = 'e2e-peer-a-' + Date.now();
  const a = await register(name, 'test12345678');
  await submitAssessment(a.token);
  const link = await createReviewLink(a.token);

  // Review links are managed on the profile page
  await loginViaToken(page, a.token, '/en/profile/' + encodeURIComponent(name));
  await expect(page.locator('body')).toBeVisible({ timeout: 5000 });

  // Peer submits review
  const peerPage = await context.newPage();
  await peerPage.goto('/en/r/' + link.token);

  const nameInput = peerPage.locator('input[type="text"]').first();
  await expect(nameInput).toBeVisible({ timeout: 5000 });
  await nameInput.fill('Reviewer Pete');

  const cards = peerPage.locator('.card.bg-base-200');
  const cardCount = await cards.count();
  for (let i = 0; i < cardCount; i++) {
    const card = cards.nth(i);
    const buttons = card.locator('button.btn-sm');
    const btnCount = await buttons.count();
    if (btnCount >= 2) {
      await buttons.nth(0).click();
      await buttons.nth(Math.min(2, btnCount - 1)).click();
    }
  }

  const submitBtn = peerPage.locator('button[type="submit"]');
  if (await submitBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
    await submitBtn.click();
  }
  await expect(peerPage.getByText('Thank You!')).toBeVisible({ timeout: 8000 });

  // Peers radar should be visible on profile page
  await page.goto('/en/profile/' + encodeURIComponent(name));
  await expect(page.locator('#peers-radar')).toBeVisible({ timeout: 10000 });

  // Delete the review link from the profile page
  const deleteBtn = page.locator('[x-show="reviewLinks.length > 0"] button:has-text("Delete")').first();
  await expect(deleteBtn).toBeVisible({ timeout: 5000 });
  acceptDialog(page);
  await deleteBtn.click();
  await expect(deleteBtn).not.toBeAttached({ timeout: 5000 });

  // Verify link is no longer available
  const reviewUrl = '/en/r/' + link.token;
  await peerPage.goto(reviewUrl, { waitUntil: 'networkidle' });
  await expect(peerPage.getByText('Link not available')).toBeVisible({ timeout: 5000 });

  await peerPage.close();
});

test('renew review link', async ({ page }) => {
  const name = 'e2e-review-renew-' + Date.now();
  const { token } = await register(name, 'test12345678');
  await submitAssessment(token);

  // Create a review link.
  const link = await createReviewLink(token);

  // Renew it — extends expiry, token stays the same.
  const renewed = await renewReviewLink(token, link.id);
  expect(renewed.token).toBeTruthy();
  // Renew extends expiry but keeps the same token.
  expect(renewed.token).toBe(link.token);

  // The (same) token still works after renewal.
  await page.goto('/en/r/' + renewed.token);
  await expect(page.locator('input[type="text"]').first()).toBeVisible({ timeout: 10000 });
});
