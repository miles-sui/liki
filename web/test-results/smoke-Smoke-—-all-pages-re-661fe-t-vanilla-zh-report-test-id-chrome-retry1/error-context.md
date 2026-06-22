# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: smoke.spec.js >> Smoke — all pages render without errors or warnings >> report (vanilla): /zh/report/test-id
- Location: e2e/journeys/smoke.spec.js:72:5

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/report/test-id
Call log:
  - navigating to "http://localhost:8080/zh/report/test-id", waiting until "load"

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e6]:
    - heading "This site can’t be reached" [level=1] [ref=e7]
    - paragraph [ref=e8]:
      - strong [ref=e9]: localhost
      - text: refused to connect.
    - generic [ref=e10]:
      - paragraph [ref=e11]: "Try:"
      - list [ref=e12]:
        - listitem [ref=e13]: Checking the connection
        - listitem [ref=e14]:
          - link "Checking the proxy and the firewall" [ref=e15] [cursor=pointer]:
            - /url: "#buttons"
    - generic [ref=e16]: ERR_CONNECTION_REFUSED
  - generic [ref=e17]:
    - button "Reload" [ref=e19] [cursor=pointer]
    - button "Details" [ref=e20] [cursor=pointer]
```

# Test source

```ts
  1   | // Smoke test — visits every page and asserts no console errors/warnings + framework renders.
  2   | // Catches: Vue compiler errors, source map 404s, i18n load failures,
  3   | //          unresolved i18n keys (raw key text leaked to page).
  4   | // Render markers: vanilla pages use [data-i18n] (Web Components), chat uses .chat-shell (Vue).
  5   | 
  6   | import { test, expect } from '../fixtures.js';
  7   | 
  8   | // Pages to test with their framework-specific render marker and i18n content checks.
  9   | const PAGES = [
  10  |   { path: '/zh/',              marker: '[data-i18n]',      name: 'index (vanilla)',
  11  |     checks: [
  12  |       { selector: 'h1.font-brand', text: '灵机 Liki' },
  13  |       { selector: 'main h2:first-of-type', text: 'AI命理助手' },
  14  |       { selector: 'section.text-center p', text: '八字分析' },
  15  |     ] },
  16  |   { path: '/zh/chat.html',     marker: '.chat-shell',     name: 'chat (Vue)',
  17  |     checks: [
  18  |       { selector: '.brand', text: '灵机对话' },
  19  |     ] },
  20  |   { path: '/zh/chart.html',    marker: '[data-i18n]',      name: 'chart (vanilla)',
  21  |     checks: [
  22  |       { selector: 'h1', text: '八字报告' },
  23  |     ] },
  24  |   { path: '/zh/naming.html',   marker: '[data-i18n]',      name: 'naming (vanilla)',
  25  |     checks: [
  26  |       { selector: 'h1', text: '起名报告' },
  27  |     ] },
  28  |   { path: '/zh/disclaimer.html', marker: '[data-i18n]',    name: 'disclaimer (vanilla)',
  29  |     checks: [
  30  |       { selector: 'h1', text: '免责声明' },
  31  |     ] },
  32  |   { path: '/zh/compatibility.html', marker: '[data-i18n]', name: 'compatibility (vanilla)',
  33  |     checks: [
  34  |       { selector: 'h1', text: '合盘报告' },
  35  |     ] },
  36  |   { path: '/zh/report/test-id', marker: '#report-header-title', name: 'report (vanilla)',
  37  |     checks: [
  38  |       { selector: 'h1', text: '命理报告' },
  39  |     ] },
  40  |   { path: '/en/',              marker: '[data-i18n]',      name: 'index EN',
  41  |     checks: [
  42  |       { selector: 'h1.font-brand', text: 'Liki' },
  43  |       { selector: 'main h2:first-of-type', text: 'AI Chinese Metaphysics Assistant' },
  44  |       { selector: 'section.text-center p', text: 'BaZi analysis' },
  45  |     ] },
  46  |   { path: '/en/chat.html',     marker: '.chat-shell',     name: 'chat EN',
  47  |     checks: [
  48  |       { selector: '.brand', text: 'Liki Chat' },
  49  |     ] },
  50  |   { path: '/en/report/test-id', marker: '#report-header-title', name: 'report EN',
  51  |     checks: [
  52  |       { selector: 'h1', text: 'Report' },
  53  |     ] },
  54  |   { path: '/hk/',              marker: '[data-i18n]',      name: 'index HK',
  55  |     checks: [
  56  |       { selector: 'h1.font-brand', text: '靈機 Liki' },
  57  |       { selector: 'main h2:first-of-type', text: 'AI命理助手' },
  58  |       { selector: 'section.text-center p', text: '八字分析' },
  59  |     ] },
  60  |   { path: '/hk/chat.html',     marker: '.chat-shell',     name: 'chat HK',
  61  |     checks: [
  62  |       { selector: '.brand', text: '靈機對話' },
  63  |     ] },
  64  | ];
  65  | 
  66  | // i18n key pattern: if a raw key like "site.name" or "index.hero.subtitle" appears
  67  | // as visible text, the i18n lookup failed and leaked the key.
  68  | const I18N_KEY_RE = /\b[a-z]{2,}\.[a-z]{2,}\.[a-z]+\b|\b[a-z]{2,}\.[a-z]{2,}\b/;
  69  | 
  70  | test.describe('Smoke — all pages render without errors or warnings', () => {
  71  |   for (const { path, marker, name, checks } of PAGES) {
  72  |     test(`${name}: ${path}`, async ({ page }) => {
  73  |       const consoleProblems = [];
  74  |       page.on('console', msg => {
  75  |         if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
  76  |           consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
  77  |         }
  78  |       });
  79  |       page.on('pageerror', err => {
  80  |         consoleProblems.push(`[uncaught] ${err.message}`);
  81  |       });
  82  | 
> 83  |       await page.goto(path);
      |                  ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/report/test-id
  84  | 
  85  |       // Framework must initialize: Web Components [data-i18n] or Vue .chat-shell.
  86  |       await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });
  87  | 
  88  |       // Wait for i18n to finish (FOUC guard removes html visibility:hidden).
  89  |       await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
  90  | 
  91  |       // i18n content must resolve — raw key like "site.name" means fetch failed.
  92  |       for (const { selector, text } of checks) {
  93  |         await expect(page.locator(selector).first()).toContainText(text, { timeout: 5000 });
  94  |       }
  95  | 
  96  |       // No unresolved i18n keys leaked to the page body.
  97  |       const bodyText = await page.locator('body').innerText();
  98  |       const leakedKeys = bodyText.split('\n').filter(line => I18N_KEY_RE.test(line.trim()));
  99  |       if (leakedKeys.length > 0) {
  100 |         throw new Error(`Unresolved i18n keys on ${path}:\n  ${leakedKeys.join('\n  ')}`);
  101 |       }
  102 | 
  103 |       if (consoleProblems.length > 0) {
  104 |         throw new Error(`Console errors on ${path}:\n  ${consoleProblems.join('\n  ')}`);
  105 |       }
  106 |     });
  107 |   }
  108 | });
  109 | 
```