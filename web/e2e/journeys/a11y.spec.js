// Accessibility tests — runs axe-core on key pages and reports violations.
// Does NOT fail the build by default; violations are informational.
// Run with: npx playwright test --config e2e/playwright.config.js a11y.spec.js

import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

// Pages to audit with their framework render marker.
const PAGES = [
  { path: '/zh-Hans/',              marker: '[data-i18n]',    name: 'index' },
  { path: '/zh-Hans/chat.html',     marker: '.chat-shell',   name: 'chat (login)' },
  // Legal pages
  { path: '/zh-Hans/about.html',    marker: '[data-i18n]',    name: 'about' },
  { path: '/zh-Hans/disclaimer.html', marker: '[data-i18n]',  name: 'disclaimer' },
  { path: '/zh-Hans/privacy.html',  marker: '[data-i18n]',    name: 'privacy' },
  { path: '/zh-Hans/terms.html',    marker: '[data-i18n]',    name: 'terms' },
];

test.describe('Accessibility audit', () => {
  for (const { path, marker, name } of PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      await page.goto(path);
      await page.waitForSelector(marker, { timeout: 10000 });
      // Ensure no FOUC hiding
      await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .analyze();

      // Group violations by impact for readable output.
      const byImpact = { critical: [], serious: [], moderate: [], minor: [] };
      for (const v of results.violations) {
        byImpact[v.impact || 'minor'].push(`${v.id}: ${v.help} (${v.nodes.length} nodes)`);
      }

      const criticalSerious = [...byImpact.critical, ...byImpact.serious];
      if (criticalSerious.length > 0) {
        console.warn(`\n  A11Y [${name}] critical/serious violations:`);
        for (const v of criticalSerious) {
          console.warn(`    - ${v}`);
        }
      }

      // Only assert no critical/serious violations.
      // Moderate/minor are too noisy for CI and tracked separately.
      expect(criticalSerious, `A11Y violations on ${name}`).toEqual([]);
    });
  }
});

// Dynamic pages: report page needs a specific order ID which can't be tested
// without mocking. We test the static error state instead.
test('report page (error state): no critical a11y violations', async ({ page }) => {
  await page.goto('/zh-Hans/report/test-id');
  await page.waitForSelector('#report-header-title', { timeout: 10000 });
  await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

  // Wait for the error status card to render.
  await page.waitForSelector('.status-card', { timeout: 10000 });

  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa'])
    .analyze();

  const criticalSerious = results.violations.filter(
    v => v.impact === 'critical' || v.impact === 'serious'
  );
  if (criticalSerious.length > 0) {
    console.warn(`\n  A11Y [report error] violations:`);
    for (const v of criticalSerious) {
      console.warn(`    - ${v.id}: ${v.help}`);
    }
  }
  expect(criticalSerious).toEqual([]);
});
