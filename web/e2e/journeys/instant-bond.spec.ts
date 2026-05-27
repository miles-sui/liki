import { test, expect } from '../fixtures';
import {
  register, submitAssessment, setPublic, loginViaToken, req,
} from '../helpers/api';
import { ProfilePage } from '../pages/profile';

test.describe('Instant Bond', () => {
  test('should compare with another user from their profile', async ({ page }) => {
    const nameA = 'e2e-instant-a-' + Date.now();
    const nameB = 'e2e-instant-b-' + Date.now();
    const a = await register(nameA, 'test12345678');
    const b = await register(nameB, 'test12345678');
    await submitAssessment(a.token);
    await submitAssessment(b.token);
    await setPublic(a.token);
    await setPublic(b.token);

    // A visits B's public profile.
    await loginViaToken(page, a.token, '/en/profile/' + encodeURIComponent(nameB));
    const profilePage = new ProfilePage(page);
    await profilePage.waitForProfile();

    // Click "Compare with Me" to compute instant bond.
    await profilePage.clickCompareWithMe();
    await profilePage.waitForBondResult();

    // Verify both bond cards are attached and ECharts rendered.
    const cards = profilePage.getBondCards();
    await expect(cards.self).toBeAttached();
    await expect(cards.other).toBeAttached();

    // Verify radar charts have 5-element data series.
    const selfData = await profilePage.getRadarSeriesData('bond-influence-self');
    const otherData = await profilePage.getRadarSeriesData('bond-influence-other');
    expect(selfData.length).toBeGreaterThan(0);
    expect(selfData[0][1].length).toBe(5);
    expect(otherData.length).toBeGreaterThan(0);
    expect(otherData[0][1].length).toBe(5);
  });

  test('should not allow comparing with a private user', async ({ page }) => {
    const nameA = 'e2e-private-a-' + Date.now();
    const nameB = 'e2e-private-b-' + Date.now();
    const a = await register(nameA, 'test12345678');
    const b = await register(nameB, 'test12345678');
    await submitAssessment(a.token);
    await submitAssessment(b.token);

    // B is private by default (register sets is_public=false).
    // Explicitly ensure private via PATCH.
    await req('PATCH', '/api/users/me', { is_public: false }, b.token);

    // A visits B's profile (private, not owner → 404).
    await loginViaToken(page, a.token, '/en/profile/' + encodeURIComponent(nameB));
    const profilePage = new ProfilePage(page);
    await profilePage.waitForProfile();

    // Should show "not found" since profile is private.
    await expect(page.getByRole('heading', { name: 'Profile not found' })).toBeVisible({ timeout: 5000 });
  });
});
