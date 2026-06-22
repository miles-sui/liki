// Smoke test — visits every page and asserts no console errors/warnings + framework renders.
// Catches: Vue compiler errors, source map 404s, i18n load failures,
//          unresolved i18n keys (raw key text leaked to page).
// Render markers: vanilla pages use [data-i18n] (Web Components), chat uses .chat-shell (Vue).

import { test, expect } from '../fixtures.js';

// Pages to test with their framework-specific render marker and i18n content checks.
const PAGES = [
  { path: '/zh/',              marker: '[data-i18n]',      name: 'index (vanilla)',
    checks: [
      { selector: 'h1.font-brand', text: '灵机 Liki' },
      { selector: 'main h2:first-of-type', text: 'AI命理助手' },
      { selector: 'section.text-center p', text: '八字分析' },
    ] },
  { path: '/zh/chat.html',     marker: '.chat-shell',     name: 'chat (Vue)',
    checks: [
      { selector: '.brand', text: '灵机对话' },
    ] },
  { path: '/zh/chart.html',    marker: '[data-i18n]',      name: 'chart (vanilla)',
    checks: [
      { selector: 'h1', text: '八字报告' },
    ] },
  { path: '/zh/naming.html',   marker: '[data-i18n]',      name: 'naming (vanilla)',
    checks: [
      { selector: 'h1', text: '起名报告' },
    ] },
  { path: '/zh/disclaimer.html', marker: '[data-i18n]',    name: 'disclaimer (vanilla)',
    checks: [
      { selector: 'h1', text: '免责声明' },
    ] },
  { path: '/zh/compatibility.html', marker: '[data-i18n]', name: 'compatibility (vanilla)',
    checks: [
      { selector: 'h1', text: '合盘报告' },
    ] },
  { path: '/zh/report/test-id', marker: '#report-header-title', name: 'report (vanilla)',
    checks: [
      { selector: 'h1', text: '命理报告' },
    ] },
  { path: '/en/',              marker: '[data-i18n]',      name: 'index EN',
    checks: [
      { selector: 'h1.font-brand', text: 'Liki' },
      { selector: 'main h2:first-of-type', text: 'AI Chinese Metaphysics Assistant' },
      { selector: 'section.text-center p', text: 'BaZi analysis' },
    ] },
  { path: '/en/chat.html',     marker: '.chat-shell',     name: 'chat EN',
    checks: [
      { selector: '.brand', text: 'Liki Chat' },
    ] },
  { path: '/en/report/test-id', marker: '#report-header-title', name: 'report EN',
    checks: [
      { selector: 'h1', text: 'Report' },
    ] },
  { path: '/hk/',              marker: '[data-i18n]',      name: 'index HK',
    checks: [
      { selector: 'h1.font-brand', text: '靈機 Liki' },
      { selector: 'main h2:first-of-type', text: 'AI命理助手' },
      { selector: 'section.text-center p', text: '八字分析' },
    ] },
  { path: '/hk/chat.html',     marker: '.chat-shell',     name: 'chat HK',
    checks: [
      { selector: '.brand', text: '靈機對話' },
    ] },
  // Legal pages — marker + console check only (Web Component shadow DOM may contain i18n keys)
  { path: '/zh/about.html',    marker: '[data-i18n]',    name: 'about (vanilla)' },
  { path: '/zh/contact.html',  marker: '[data-i18n]',    name: 'contact (vanilla)' },
  { path: '/zh/privacy.html',  marker: '[data-i18n]',    name: 'privacy (vanilla)' },
  { path: '/zh/terms.html',    marker: '[data-i18n]',    name: 'terms (vanilla)' },
  // Static resources
  { path: '/skills/liki.md',   marker: null,             name: 'skills',   resource: true },
  { path: '/llms.txt',         marker: null,             name: 'llms.txt', resource: true },
  { path: '/api/openapi.json',     marker: null,             name: 'openapi',  resource: true },
];

// i18n key pattern: if a raw key like "site.name" or "index.hero.subtitle" appears
// as visible text, the i18n lookup failed and leaked the key.
const I18N_KEY_RE = /\b[a-z]{2,}\.[a-z]{2,}\.[a-z]+\b|\b[a-z]{2,}\.[a-z]{2,}\b/;

test.describe('Smoke — all pages render without errors or warnings', () => {
  for (const { path, marker, name, checks, resource } of PAGES) {
    test(`${name}: ${path}`, async ({ page }) => {
      const consoleProblems = [];
      page.on('console', msg => {
        if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
          consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
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
      const bodyText = await page.locator('body').innerText();
      const leakedKeys = bodyText.split('\n').filter(line => I18N_KEY_RE.test(line.trim()));
      if (leakedKeys.length > 0) {
        throw new Error(`Unresolved i18n keys on ${path}:\n  ${leakedKeys.join('\n  ')}`);
      }

      if (consoleProblems.length > 0) {
        throw new Error(`Console errors on ${path}:\n  ${consoleProblems.join('\n  ')}`);
      }
    });
  }
});
