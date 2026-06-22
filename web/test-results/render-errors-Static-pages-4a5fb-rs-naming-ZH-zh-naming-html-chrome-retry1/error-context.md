# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: render-errors.spec.js >> Static pages — no render errors >> naming ZH: /zh/naming.html
- Location: e2e/journeys/render-errors.spec.js:45:5

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/naming.html
Call log:
  - navigating to "http://localhost:8080/zh/naming.html", waiting until "domcontentloaded"

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
  1   | // Render error tests — catches silent framework failures: console errors, raw templates, broken images.
  2   | // Separate from smoke.spec.js which checks page load and framework init.
  3   | 
  4   | import { test, expect } from '../fixtures.js';
  5   | 
  6   | // Pages that don't depend on async API data.
  7   | const STATIC_PAGES = [
  8   |   { path: '/zh/',              marker: '[data-i18n]',      name: 'index ZH' },
  9   |   { path: '/zh/chart.html',    marker: '[data-i18n]',      name: 'chart ZH' },
  10  |   { path: '/zh/naming.html',   marker: '[data-i18n]',      name: 'naming ZH' },
  11  |   { path: '/zh/about.html',    marker: '[data-i18n]',      name: 'about ZH' },
  12  |   { path: '/zh/privacy.html',  marker: '[data-i18n]',      name: 'privacy ZH' },
  13  |   { path: '/zh/terms.html',    marker: '[data-i18n]',      name: 'terms ZH' },
  14  |   { path: '/zh/disclaimer.html', marker: '[data-i18n]',    name: 'disclaimer ZH' },
  15  |   { path: '/zh/compatibility.html', marker: '[data-i18n]', name: 'compatibility ZH' },
  16  |   { path: '/en/',              marker: '[data-i18n]',      name: 'index EN' },
  17  |   { path: '/en/about.html',    marker: '[data-i18n]',      name: 'about EN' },
  18  |   { path: '/en/privacy.html',  marker: '[data-i18n]',      name: 'privacy EN' },
  19  |   { path: '/en/terms.html',    marker: '[data-i18n]',      name: 'terms EN' },
  20  |   { path: '/en/chart.html',    marker: '[data-i18n]',      name: 'chart EN' },
  21  |   { path: '/en/naming.html',   marker: '[data-i18n]',      name: 'naming EN' },
  22  |   { path: '/hk/',              marker: '[data-i18n]',      name: 'index HK' },
  23  |   { path: '/hk/about.html',    marker: '[data-i18n]',      name: 'about HK' },
  24  |   { path: '/hk/privacy.html',  marker: '[data-i18n]',      name: 'privacy HK' },
  25  |   { path: '/hk/terms.html',    marker: '[data-i18n]',      name: 'terms HK' },
  26  |   { path: '/hk/chart.html',    marker: '[data-i18n]',      name: 'chart HK' },
  27  |   { path: '/hk/naming.html',   marker: '[data-i18n]',      name: 'naming HK' },
  28  | ];
  29  | 
  30  | // Pages that require Vue mount (API-free after mount).
  31  | const VUE_PAGES = [
  32  |   { path: '/zh/chat.html', name: 'chat ZH' },
  33  |   { path: '/en/chat.html', name: 'chat EN' },
  34  |   { path: '/hk/chat.html', name: 'chat HK' },
  35  | ];
  36  | 
  37  | // Pages that make async API calls on init — only check console errors.
  38  | const ASYNC_PAGES = [
  39  |   { path: '/zh/report/test-id', marker: '#report-header-title', name: 'report ZH' },
  40  |   { path: '/en/report/test-id', marker: '#report-header-title', name: 'report EN' },
  41  | ];
  42  | 
  43  | test.describe('Static pages — no render errors', () => {
  44  |   for (const { path, marker, name } of STATIC_PAGES) {
  45  |     test(`${name}: ${path}`, async ({ page }) => {
  46  |       const consoleProblems = [];
  47  |       page.on('console', msg => {
  48  |         if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
  49  |           consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
  50  |         }
  51  |       });
  52  |       page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));
  53  | 
> 54  |       await page.goto(path, { waitUntil: 'domcontentloaded' });
      |                  ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/zh/naming.html
  55  |       await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });
  56  |       await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
  57  | 
  58  |       // Check raw template leakage and unresolved i18n keys.
  59  |       const bodyText = await page.locator('body').innerText();
  60  |       if (/\{\{/.test(bodyText)) {
  61  |         throw new Error(`Raw template syntax '{{ }}' found on ${path}`);
  62  |       }
  63  |       // Raw i18n keys follow a dotted pattern; catch any that leaked through.
  64  |       if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
  65  |         throw new Error(`Unresolved i18n key found on ${path}: ${bodyText.match(/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/gi)}`);
  66  |       }
  67  | 
  68  |       if (consoleProblems.length > 0) {
  69  |         throw new Error(`Console problems on ${path}:\n  ${consoleProblems.join('\n  ')}`);
  70  |       }
  71  |     });
  72  |   }
  73  | });
  74  | 
  75  | test.describe('Vue pages — no render errors', () => {
  76  |   for (const { path, name } of VUE_PAGES) {
  77  |     test(`${name}: ${path}`, async ({ page }) => {
  78  |       const consoleProblems = [];
  79  |       page.on('console', msg => {
  80  |         if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
  81  |           consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
  82  |         }
  83  |       });
  84  |       page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));
  85  | 
  86  |       await page.goto(path, { waitUntil: 'domcontentloaded' });
  87  | 
  88  |       // FOUC guard hides html until i18n loads. Wait for it first.
  89  |       await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
  90  | 
  91  |       // Vue mounts on #app div wrapper. v-cloak must be removed after mount.
  92  |       await expect(page.locator('#app:not([v-cloak])')).toBeAttached({ timeout: 10000 });
  93  |       // Chat shell must be present after mount.
  94  |       await expect(page.locator('.chat-shell')).toBeVisible({ timeout: 5000 });
  95  | 
  96  |       const bodyText = await page.locator('body').innerText();
  97  |       if (/\{\{/.test(bodyText)) {
  98  |         throw new Error(`Raw template syntax '{{ }}' found on ${path}`);
  99  |       }
  100 |       if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
  101 |         throw new Error(`Unresolved i18n key found on ${path}`);
  102 |       }
  103 | 
  104 |       if (consoleProblems.length > 0) {
  105 |         throw new Error(`Console problems on ${path}:\n  ${consoleProblems.join('\n  ')}`);
  106 |       }
  107 |     });
  108 |   }
  109 | });
  110 | 
  111 | test.describe('Async pages — no console errors on load', () => {
  112 |   for (const { path, marker, name } of ASYNC_PAGES) {
  113 |     test(`${name}: ${path}`, async ({ page }) => {
  114 |       const consoleProblems = [];
  115 |       page.on('console', msg => {
  116 |         if ((msg.type() === 'error' || msg.type() === 'warning') && !(msg.location().url || '').includes('favicon')) {
  117 |           consoleProblems.push(`[${msg.type()}] ${msg.text()}`);
  118 |         }
  119 |       });
  120 |       page.on('pageerror', err => consoleProblems.push(`[uncaught] ${err.message}`));
  121 | 
  122 |       await page.goto(path, { waitUntil: 'domcontentloaded' });
  123 |       // Wait for Web Components to init (API call will be in-flight, but framework must mount).
  124 |       await expect(page.locator(marker).first()).toBeVisible({ timeout: 10000 });
  125 |       await expect(page.locator('html')).not.toHaveCSS('visibility', 'hidden', { timeout: 10000 });
  126 | 
  127 |       // Check for unresolved i18n keys.
  128 |       const bodyText = await page.locator('body').innerText();
  129 |       if (/\babout\.\w+|privacy\.\w+|terms\.\w+|site\.\w+|index\.\w+|chart\.\w+|report\.\w+|chat\.\w+|naming\.\w+|bond\.\w+|zhu\.\w+|footer\.\w+|form\.\w+|disclaimer\.\w+|compatibility\.\w+|nav\.\w+\b/i.test(bodyText)) {
  130 |         throw new Error(`Unresolved i18n key found on ${path}`);
  131 |       }
  132 | 
  133 |       // Only check for console errors — API call may fail for test-id, that's expected.
  134 |       // Filter out expected fetch errors from the fake report ID.
  135 |       const realProblems = consoleProblems.filter(m => !m.includes('404') && !m.includes('Failed to fetch'));
  136 |       if (realProblems.length > 0) {
  137 |         throw new Error(`Console problems on ${path}:\n  ${realProblems.join('\n  ')}`);
  138 |       }
  139 |     });
  140 |   }
  141 | });
  142 | 
  143 | // ── Images: critical images must load ──
  144 | 
  145 | test.describe('Images load without error', () => {
  146 |   test('index page images are valid', async ({ page }) => {
  147 |     await page.goto('/zh/', { waitUntil: 'domcontentloaded' });
  148 |     await expect(page.locator('[data-i18n]').first()).toBeVisible({ timeout: 10000 });
  149 | 
  150 |     const imgs = page.locator('img');
  151 |     const count = await imgs.count();
  152 |     for (let i = 0; i < count; i++) {
  153 |       const img = imgs.nth(i);
  154 |       const src = await img.getAttribute('src');
```