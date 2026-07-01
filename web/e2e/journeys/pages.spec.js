// Smoke test — visits every page and asserts no console errors/warnings + framework renders.
// Catches: Vue compiler errors, source map 404s, i18n load failures,
//          unresolved i18n keys (raw key text leaked to page).
// Render markers: vanilla pages use [data-i18n] (Web Components), chat uses .chat-shell (Vue).

import { test, expect } from '../fixtures.js';

// Pages to test with their framework-specific render marker and i18n content checks.
const PAGES = [
  { path: '/zh-Hans/',              marker: '[data-i18n]',      name: 'index (vanilla)',
    checks: [
      { selector: 'span.font-brand', text: 'Liki' },
      { selector: 'main h2:first-of-type', text: '找到你的名字' },
    ] },
  { path: '/zh-Hans/chat.html',     marker: '.chat-shell',     name: 'chat (Vue)',
    checks: [
      { selector: '.brand', text: '灵机对话' },
    ] },
  { path: '/zh-Hans/disclaimer.html', marker: '[data-i18n]',    name: 'disclaimer (vanilla)',
    checks: [
      { selector: 'h1', text: '免责声明' },
    ] },
  { path: '/zh-Hans/report/test-id', marker: '#report-header-title', name: 'report (vanilla)',
    checks: [
      { selector: 'h1', text: '起名报告' },
    ] },
  { path: '/en/',              marker: '[data-i18n]',      name: 'index EN',
    checks: [
      { selector: 'span.font-brand', text: 'Liki' },
      { selector: 'main h2:first-of-type', text: 'Find Your Name' },
    ] },
  { path: '/en/chat.html',     marker: '.chat-shell',     name: 'chat EN',
    checks: [
      { selector: '.brand', text: 'Liki Chat' },
    ] },
  { path: '/en/report/test-id', marker: '#report-header-title', name: 'report EN',
    checks: [
      { selector: 'h1', text: 'Report' },
    ] },
  { path: '/zh-Hant/',              marker: '[data-i18n]',      name: 'index ZH-Hant',
    checks: [
      { selector: 'span.font-brand', text: 'Liki' },
      { selector: 'main h2:first-of-type', text: '找到你的名字' },
    ] },
  { path: '/zh-Hant/chat.html',     marker: '.chat-shell',     name: 'chat ZH-Hant',
    checks: [
      { selector: '.brand', text: '靈機對話' },
    ] },
  // Legal pages — marker renders, i18n resolves, keys don't leak
  { path: '/zh-Hans/about.html',    marker: '[data-i18n]',    name: 'about (vanilla)', skipI18nCheck: true },
  { path: '/zh-Hans/contact.html',  marker: '[data-i18n]',    name: 'contact (vanilla)', skipI18nCheck: true },
  { path: '/zh-Hans/privacy.html',  marker: '[data-i18n]',    name: 'privacy (vanilla)', skipI18nCheck: true },
  { path: '/zh-Hans/terms.html',    marker: '[data-i18n]',    name: 'terms (vanilla)', skipI18nCheck: true },
  // Static resources
  { path: '/skills/liki.md',   marker: null,             name: 'skills',   resource: true },
  { path: '/llms.txt',         marker: null,             name: 'llms.txt', resource: true },
];

// i18n key pattern: if a raw key like "site.name" or "index.hero.subtitle" appears
// as visible text, the i18n lookup failed and leaked the key.
const I18N_KEY_RE = /\b[a-z]{2,}\.[a-z]{2,}\.[a-z]+\b|\b[a-z]{2,}\.[a-z]{2,}\b/;

test.describe('Smoke — all pages render without errors or warnings', () => {
  for (const { path, marker, name, checks, resource, skipI18nCheck } of PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      const consoleProblems = [];
      page.on('console', msg => {
        if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
          consoleProblems.push(`[${msg.type()}] ${msg.text()} (${msg.location().url || 'no-url'})`);
        }
      });
      page.on('requestfailed', req => {
        // Ignore aborted requests (page close during navigation)
        if (req.failure()?.errorText !== 'net::ERR_ABORTED') {
          consoleProblems.push(`[requestfailed] ${req.url()} — ${req.failure()?.errorText || 'unknown'}`);
        }
      });
      page.on('pageerror', err => {
        consoleProblems.push(`[uncaught] ${err.message}`);
      });

      const resp = await page.goto(path);

      // Static resources: just verify HTTP 200.
      if (resource) {
        expect(resp.status()).toBe(200);
        return;
      }

      // Framework must initialize: Web Components [data-i18n] or Vue .chat-shell.
      await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });

      // Wait for i18n to finish (FOUC guard removes html visibility:hidden).
      await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });

      // i18n content must resolve — raw key like "site.name" means fetch failed.
      if (checks) for (const { selector, text } of checks) {
        await expect(page.locator(selector).first()).toContainText(text, { timeout: 5000 });
      }

      // No unresolved i18n keys leaked to the page body.
      if (!skipI18nCheck) {
        const bodyText = await page.locator('body').innerText();
        const leakedKeys = bodyText.split('\n').filter(line => I18N_KEY_RE.test(line.trim()));
        if (leakedKeys.length > 0) {
          throw new Error(`Unresolved i18n keys on ${path}:\n  ${leakedKeys.join('\n  ')}`);
        }
        // Raw template syntax must not leak.
        if (/\{\{/.test(bodyText)) {
          throw new Error(`Raw template syntax '{{ }}' found on ${path}`);
        }
      }

      // Vue pages: v-cloak must be removed after mount.
      if (marker === '.chat-shell') {
        await expect(page.locator('#app:not([v-cloak])')).toBeAttached({ timeout: 10000 });
      }

      if (consoleProblems.length > 0) {
        throw new Error(`Console errors on ${path}:\n  ${consoleProblems.join('\n  ')}`);
      }
    });
  }
});

// ── OG meta tags: social sharing cards ──

test.describe('OG / Twitter meta tags', () => {
  test('index page has correct OG tags', async ({ page }) => {
    await page.goto('/zh-Hans/');
    await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });

    const tags = {
      'og:title':       await page.locator('meta[property="og:title"]').getAttribute('content'),
      'og:description': await page.locator('meta[property="og:description"]').getAttribute('content'),
      'og:image':       await page.locator('meta[property="og:image"]').getAttribute('content'),
      'og:url':         await page.locator('meta[property="og:url"]').getAttribute('content'),
      'og:type':        await page.locator('meta[property="og:type"]').getAttribute('content'),
      'twitter:card':   await page.locator('meta[name="twitter:card"]').getAttribute('content'),
    };

    expect(tags['og:title']).toBeTruthy();
    expect(tags['og:description']).toBeTruthy();
    expect(tags['og:image']).toMatch(/^https:\/\/liki\.hk\/img\/og-image\.png$/);
    expect(tags['og:url']).toBe('https://liki.hk');
    expect(tags['og:type']).toBe('website');
    expect(tags['twitter:card']).toBe('summary_large_image');
  });

  test('OG image is accessible', async ({ page }) => {
    const resp = await page.goto('/img/og-image.png');
    expect(resp.status()).toBe(200);
    const ct = resp.headers()['content-type'];
    expect(ct).toMatch(/^image\/png/);
  });
});

test.describe('Images load without error', () => {
  test('index page images are valid', async ({ page }) => {
    await page.goto('/zh-Hans/');
    await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });

    const imgs = page.locator('img');
    const count = await imgs.count();
    for (let i = 0; i < count; i++) {
      const img = imgs.nth(i);
      const src = await img.getAttribute('src');
      if (!src || src.startsWith('data:')) continue;
      const w = await img.evaluate(el => el.naturalWidth);
      if (w === 0) throw new Error(`Broken image on /zh-Hans/: src="${src}"`);
    }
  });
});

// ── Mobile viewport: critical pages render on mobile ──

test.describe('Mobile viewport', () => {
  test('index renders at 375x812', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto('/zh-Hans/');
    await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#purchase-form')).toBeVisible();
  });

  test('report renders at 375x812', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto('/zh-Hans/report/test-id');
    await expect(page.locator('#report-header-title')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
  });
});
