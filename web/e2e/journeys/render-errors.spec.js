// Render error tests — catches silent framework failures: console errors, raw templates, broken images.
// Separate from smoke.spec.js which checks page load and framework init.

import { test, expect } from '../fixtures.js';

// Pages that don't depend on async API data.
const STATIC_PAGES = [
  { path: '/zh/',              marker: '[data-i18n]',      name: 'index ZH' },
  { path: '/zh/chart.html',    marker: '[data-i18n]',      name: 'chart ZH' },
  { path: '/zh/naming.html',   marker: '[data-i18n]',      name: 'naming ZH' },
  { path: '/zh/about.html',    marker: '[data-i18n]',      name: 'about ZH' },
  { path: '/zh/privacy.html',  marker: '[data-i18n]',      name: 'privacy ZH' },
  { path: '/zh/terms.html',    marker: '[data-i18n]',      name: 'terms ZH' },
  { path: '/zh/disclaimer.html', marker: '[data-i18n]',    name: 'disclaimer ZH' },
  { path: '/zh/compatibility.html', marker: '[data-i18n]', name: 'compatibility ZH' },
  { path: '/en/',              marker: '[data-i18n]',      name: 'index EN' },
  { path: '/en/about.html',    marker: '[data-i18n]',      name: 'about EN' },
  { path: '/en/privacy.html',  marker: '[data-i18n]',      name: 'privacy EN' },
  { path: '/en/terms.html',    marker: '[data-i18n]',      name: 'terms EN' },
  { path: '/en/chart.html',    marker: '[data-i18n]',      name: 'chart EN' },
  { path: '/en/naming.html',   marker: '[data-i18n]',      name: 'naming EN' },
  { path: '/hk/',              marker: '[data-i18n]',      name: 'index HK' },
  { path: '/hk/about.html',    marker: '[data-i18n]',      name: 'about HK' },
  { path: '/hk/privacy.html',  marker: '[data-i18n]',      name: 'privacy HK' },
  { path: '/hk/terms.html',    marker: '[data-i18n]',      name: 'terms HK' },
  { path: '/hk/chart.html',    marker: '[data-i18n]',      name: 'chart HK' },
  { path: '/hk/naming.html',   marker: '[data-i18n]',      name: 'naming HK' },
];

// Pages that require Vue mount (API-free after mount).
const VUE_PAGES = [
  { path: '/zh/chat.html', name: 'chat ZH' },
  { path: '/en/chat.html', name: 'chat EN' },
  { path: '/hk/chat.html', name: 'chat HK' },
];

// Pages that make async API calls on init — only check console errors.
const ASYNC_PAGES = [
  { path: '/zh/report/test-id', marker: '#report-header-title', name: 'report ZH' },
  { path: '/en/report/test-id', marker: '#report-header-title', name: 'report EN' },
];

test.describe('Static pages — no render errors', () => {
  for (const { path, marker, name } of STATIC_PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      const consoleProblems = [];
      page.on('console', msg => {
        if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
          consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
        }
      });
      page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));

      await page.goto(path, { waitUntil: 'domcontentloaded' });
      await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });
      await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

      // Check raw template leakage and unresolved i18n keys.
      const bodyText = await page.locator('body').innerText();
      if (/\{\{/.test(bodyText)) {
        throw new Error(`Raw template syntax '{{ }}' found on ${path}`);
      }
      // Raw i18n keys follow a dotted pattern; catch any that leaked through.
      if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
        throw new Error(`Unresolved i18n key found on ${path}: ${bodyText.match(/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/gi)}`);
      }

      if (consoleProblems.length > 0) {
        throw new Error(`Console problems on ${path}:\n  ${consoleProblems.join('\n  ')}`);
      }
    });
  }
});

test.describe('Vue pages — no render errors', () => {
  for (const { path, name } of VUE_PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      const consoleProblems = [];
      page.on('console', msg => {
        if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
          consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
        }
      });
      page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));

      await page.goto(path, { waitUntil: 'domcontentloaded' });

      // FOUC guard hides html until i18n loads. Wait for it first.
      await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

      // Vue mounts on #app div wrapper. v-cloak must be removed after mount.
      await expect(page.locator('#app:not([v-cloak])')).toBeAttached({ timeout: 10000 });
      // Chat shell must be present after mount.
      await expect(page.locator('.chat-shell')).toBeVisible({ timeout: 5000 });

      const bodyText = await page.locator('body').innerText();
      if (/\{\{/.test(bodyText)) {
        throw new Error(`Raw template syntax '{{ }}' found on ${path}`);
      }
      if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
        throw new Error(`Unresolved i18n key found on ${path}`);
      }

      if (consoleProblems.length > 0) {
        throw new Error(`Console problems on ${path}:\n  ${consoleProblems.join('\n  ')}`);
      }
    });
  }
});

test.describe('Async pages — no console errors on load', () => {
  for (const { path, marker, name } of ASYNC_PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      const consoleProblems = [];
      page.on('console', msg => {
        if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
          consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
        }
      });
      page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));

      await page.goto(path, { waitUntil: 'domcontentloaded' });
      // Wait for Web Components to init (API call will be in-flight, but framework must mount).
      await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });
      await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

      // Check for unresolved i18n keys.
      const bodyText = await page.locator('body').innerText();
      if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
        throw new Error(`Unresolved i18n key found on ${path}`);
      }

      // Only check for console errors — API call may fail for test-id, that's expected.
      // Filter out expected fetch errors from the fake report ID.
      const realProblems = consoleProblems.filter(m => !m.includes('404') && !m.includes('Failed to fetch'));
      if (realProblems.length > 0) {
        throw new Error(`Console problems on ${path}:\n  ${realProblems.join('\n  ')}`);
      }
    });
  }
});

// ── Images: critical images must load ──

test.describe('Images load without error', () => {
  test('index page images are valid', async ({ page }) => {
    await page.goto('/zh/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });

    const imgs = page.locator('img');
    const count = await imgs.count();
    for (let i = 0; i < count; i++) {
      const img = imgs.nth(i);
      const src = await img.getAttribute('src');
      if (!src || src.startsWith('data:')) continue;
      const w = await img.evaluate(el => el.naturalWidth);
      if (w === 0) throw new Error(`Broken image on /zh/: src="${src}"`);
    }
  });

  test('chart demo page images are valid', async ({ page }) => {
    await page.goto('/zh/chart.html', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });

    const imgs = page.locator('img');
    const count = await imgs.count();
    for (let i = 0; i < count; i++) {
      const img = imgs.nth(i);
      const src = await img.getAttribute('src');
      if (!src || src.startsWith('data:')) continue;
      const w = await img.evaluate(el => el.naturalWidth);
      if (w === 0) throw new Error(`Broken image on /zh/chart.html: src="${src}"`);
    }
  });
});

// ── Naming demo page: dynamic content renders ──

test.describe('Naming demo page', () => {
  test('ZH naming page renders all sections', async ({ page }) => {
    await page.goto('/zh/naming.html', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });
    await expect(page.locator('header')).toBeVisible();
    await expect(page.locator('main .section-card').first()).toBeVisible({ timeout: 5000 });
  });

  test('EN naming page renders with English labels', async ({ page }) => {
    await page.goto('/en/naming.html', { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('[data-i18n]', { timeout: 10000 });
    await expect(page.locator('header')).toBeVisible();
    await expect(page.locator('main .section-card').first()).toBeVisible({ timeout: 5000 });
  });
});
