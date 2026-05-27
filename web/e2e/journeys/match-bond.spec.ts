import { test, expect } from '../fixtures';
import {
  register, submitAssessment, setPublic, createMatchLink,
  loginViaToken, submitMatchLink, FULL_ANSWERS,
} from '../helpers/api';
import { MatchLandingPage } from '../pages/match-landing';

test.describe('Match Link', () => {
  test.describe('anonymous flow', () => {
    test('should see bond result on match landing after anonymous submission', async ({ page }) => {
      const creatorName = 'e2e-claim-creator-' + Date.now();
      const creator = await register(creatorName, 'test12345678');
      await submitAssessment(creator.token);
      await setPublic(creator.token);
      const link = await createMatchLink(creator.token);

      // Submit anonymously via API.
      await submitMatchLink(link.token, FULL_ANSWERS);

      // B (another registered user) visits the link to see if bond exists.
      const nameB = 'e2e-claim-visitor-' + Date.now();
      const b = await register(nameB, 'test12345678');
      await submitAssessment(b.token);

      // Visit as logged-in user to see use_existing option with bond result.
      await loginViaToken(page, b.token, '/en/m/' + link.token);
      const mlPage = new MatchLandingPage(page);
      await mlPage.waitForLoad();

      // Should see "Compute Bond" button (use existing).
      await expect(page.locator('button:has-text("Compute Bond")')).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('authenticated flow', () => {
    test('should use existing profile to create instant bond', async ({ page }) => {
      const creatorName = 'e2e-useexist-creator-' + Date.now();
      const creator = await register(creatorName, 'test12345678');
      await submitAssessment(creator.token);
      const link = await createMatchLink(creator.token);

      const otherName = 'e2e-useexist-other-' + Date.now();
      const other = await register(otherName, 'test12345678');
      await submitAssessment(other.token);

      // B opens match link while logged in.
      await loginViaToken(page, other.token, '/en/m/' + link.token);
      const mlPage = new MatchLandingPage(page);
      await mlPage.waitForLoad();

      await expect(page.locator('button:has-text("Compute Bond")')).toBeVisible({ timeout: 5000 });
      await mlPage.clickUseExisting();
      await mlPage.waitForBondResult();

      const cards = mlPage.getBondCards();
      await expect(cards.self).toBeAttached();
      await expect(cards.other).toBeAttached();
    });
  });
});
