import { test, expect } from '../fixtures';
import { register, submitAssessment, loginViaToken, req } from '../helpers/api';

test.describe('Anonymous assessment claiming', () => {
  test('complete anonymous assessment, register, and claim it', async ({ page }) => {
    const ts = Date.now();
    const anonToken = 'e2e-claim-' + ts;

    // 1. Submit assessment anonymously (no auth token).
    await submitAssessment(undefined, anonToken);

    // 2. Register a new user with the anonymous token — this claims the assessment.
    const name = 'e2e-claim-' + ts;
    const { token } = await register(name, 'test12345678', anonToken);

    // 3. Login and visit profile — claiming means the assessment profile is now linked.
    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    // Claimed assessment MUST show the profile radar (not "No assessment yet").
    await expect(page.locator('#profile-radar')).toBeVisible({ timeout: 10000 });
  });

  test('register without anonymous token has no profile', async ({ page }) => {
    const name = 'e2e-noclaim-' + Date.now();
    const { token } = await register(name, 'test12345678');

    await loginViaToken(page, token, '/en/profile/' + encodeURIComponent(name));
    await expect(page.locator('[x-data="profilePage()"]')).toBeVisible({ timeout: 10000 });

    // Without claiming, the user has no assessment — profile page shows empty state.
    await expect(page.getByText("This user hasn't completed an assessment yet.")).toBeVisible({ timeout: 10000 });
  });

  test('anonymous assessment is linkable to match link flow', async () => {
    const ts = Date.now();
    const anonToken = 'e2e-linkclaim-' + ts;

    await submitAssessment(undefined, anonToken);

    const name = 'e2e-linkclaim-' + ts;
    const { token } = await register(name, 'test12345678', anonToken);

    // Verify via the profile API that the assessment was claimed.
    const { data } = await req('GET', '/api/profiles/' + encodeURIComponent(name), undefined, token);
    // Claimed assessment means profile data exists and is_owner is true.
    expect(data.profile).toBeTruthy();
    expect(data.is_owner).toBe(true);
  });
});
