# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: report.spec.js >> Report page >> invalid order ID shows error
- Location: e2e/journeys/report.spec.js:142:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/en/report/nonexistent-id-12345
Call log:
  - navigating to "http://localhost:8080/en/report/nonexistent-id-12345", waiting until "load"

```

# Test source

```ts
  43  | 
  44  | 宜从事木火行业，向东方发展。`;
  45  | 
  46  | // Mock chart report API — returns paid report with chart_json + llm_json.
  47  | async function mockChartReport(page, orderID = 'dead-beef') {
  48  |   await page.route(`**/api/reports/${orderID}`, async (route) => {
  49  |     await route.fulfill({
  50  |       status: 200,
  51  |       contentType: 'application/json',
  52  |       body: JSON.stringify({
  53  |         data: {
  54  |           order_id: orderID,
  55  |           product: 'chart',
  56  |           chart_json: CHART_JSON,
  57  |           llm_json: LLM_JSON,
  58  |           status: 'paid',
  59  |           amount: 990,
  60  |           currency: 'CNY',
  61  |         },
  62  |       }),
  63  |     });
  64  |   });
  65  | }
  66  | 
  67  | // Mock bond report API.
  68  | async function mockBondReport(page, orderID = 'cafe-fade') {
  69  |   const bondChartJSON = JSON.stringify({
  70  |     chart_a: { chart: { riyuan: '庚金', nianzhu: { gan: '庚', zhi: '午' }, yuezhu: { gan: '壬', zhi: '午' }, rizhu: { gan: '庚', zhi: '午' }, shizhu: { gan: '丙', zhi: '子' } } },
  71  |     chart_b: { chart: { riyuan: '甲木', nianzhu: { gan: '甲', zhi: '子' }, yuezhu: { gan: '丙', zhi: '寅' }, rizhu: { gan: '甲', zhi: '寅' }, shizhu: { gan: '戊', zhi: '辰' } } },
  72  |     bond: { gan_rel: [{ a: '庚', b: '甲', rel: '冲' }], zhi_rel: [], key_hints: ['金木相冲'] },
  73  |   });
  74  |   await page.route(`**/api/reports/${orderID}`, async (route) => {
  75  |     await route.fulfill({
  76  |       status: 200,
  77  |       contentType: 'application/json',
  78  |       body: JSON.stringify({
  79  |         data: {
  80  |           order_id: orderID,
  81  |           product: 'bond',
  82  |           chart_json: bondChartJSON,
  83  |           llm_json: '# 合盘报告\n\n## 缘分分析\n\n金木相冲，需调和。',
  84  |           status: 'paid',
  85  |           amount: 1990,
  86  |           currency: 'CNY',
  87  |         },
  88  |       }),
  89  |     });
  90  |   });
  91  | }
  92  | 
  93  | // Mock naming report API.
  94  | async function mockNamingReport(page, orderID = 'babe-face') {
  95  |   const namingChartJSON = JSON.stringify({
  96  |     naming: {
  97  |       analysis: { surname: '陈', yong_shen: '火', zodiac: '鼠' },
  98  |       candidates: [{ name: '陈明远', score: 95, wuxing: '火木土' }],
  99  |     },
  100 |   });
  101 |   await page.route(`**/api/reports/${orderID}`, async (route) => {
  102 |     await route.fulfill({
  103 |       status: 200,
  104 |       contentType: 'application/json',
  105 |       body: JSON.stringify({
  106 |         data: {
  107 |           order_id: orderID,
  108 |           product: 'naming',
  109 |           chart_json: namingChartJSON,
  110 |           llm_json: '# 起名报告\n\n## 名字分析\n\n**陈明远** — 得分 95。',
  111 |           status: 'paid',
  112 |           amount: 2990,
  113 |           currency: 'CNY',
  114 |         },
  115 |       }),
  116 |     });
  117 |   });
  118 | }
  119 | 
  120 | test.describe('Report page', () => {
  121 |   let page;
  122 | 
  123 |   test.beforeEach(async ({ context }) => {
  124 |     page = await context.newPage();
  125 |   });
  126 | 
  127 |   test.afterEach(async () => {
  128 |     await page.close();
  129 |   });
  130 | 
  131 |   // ── error states (existing) ──
  132 | 
  133 |   test('missing order ID shows error', async () => {
  134 |     await page.goto('/en/report/');
  135 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  136 | 
  137 |     // ReportApp sets phase='error' when orderID is missing — renders .status-card.status-error.
  138 |     await expect(page.locator('.status-card.status-error')).toBeVisible({ timeout: 10000 });
  139 |     await expect(page.locator('.status-card.status-error .status-actions a')).toBeVisible();
  140 |   });
  141 | 
  142 |   test('invalid order ID shows error', async () => {
> 143 |     await page.goto('/en/report/nonexistent-id-12345');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:8080/en/report/nonexistent-id-12345
  144 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  145 | 
  146 |     // ReportApp calls loadReport → API returns 404 → phase='error'.
  147 |     await expect(page.locator('.status-card.status-error')).toBeVisible({ timeout: 10000 });
  148 |   });
  149 | 
  150 |   // ── save banner ──
  151 | 
  152 |   test('save banner visible by default, has copy link and close buttons', async () => {
  153 |     await page.goto('/en/report/abba');
  154 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  155 | 
  156 |     const banner = page.locator('.save-banner');
  157 |     await expect(banner).toBeVisible({ timeout: 5000 });
  158 |     // Copy button is #banner-copy-btn per report.html.
  159 |     await expect(page.locator('#banner-copy-btn')).toBeVisible();
  160 | 
  161 |     await page.locator('#banner-close-btn').click();
  162 |     await expect(banner).not.toBeVisible({ timeout: 3000 });
  163 |   });
  164 | 
  165 |   // ── successful chart report rendering ──
  166 | 
  167 |   test('paid chart report renders markdown and summary cards', async () => {
  168 |     await mockChartReport(page);
  169 |     await page.goto('/zh/report/dead-beef');
  170 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  171 | 
  172 |     // Phase transitions to ready — report content becomes visible.
  173 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  174 | 
  175 |     // Summary cards rendered from chart_json — use .first() to avoid strict-mode
  176 |     // conflict between chart summary and tiaohou summary sections.
  177 |     await expect(page.locator('.summary-grid').first()).toBeVisible();
  178 |     await expect(page.locator('.summary-card .value').first()).toContainText('戊土');
  179 | 
  180 |     // Markdown → HTML rendered in #chart-interpretation (product=chart).
  181 |     const interpretation = page.locator('#chart-interpretation');
  182 |     await expect(interpretation).toContainText('命盘总览');
  183 |     await expect(interpretation).toContainText('戊土');
  184 |     await expect(interpretation).toContainText('五行分析');
  185 |     await expect(interpretation).toContainText('十神解读');
  186 |     await expect(interpretation).toContainText('大运流年');
  187 | 
  188 |     // Title reflects product — use #report-header-title to avoid strict-mode conflict.
  189 |     await expect(page.locator('#report-header-title')).toContainText('八字报告');
  190 |   });
  191 | 
  192 |   test('paid chart report in EN locale', async () => {
  193 |     await mockChartReport(page, 'beef-dead');
  194 |     await page.goto('/en/report/beef-dead');
  195 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  196 | 
  197 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  198 | 
  199 |     // EN title
  200 |     await expect(page.locator('#report-header-title')).toContainText('BaZi');
  201 |   });
  202 | 
  203 |   // ── bond report ──
  204 | 
  205 |   test('paid bond report renders bond-specific data', async () => {
  206 |     await mockBondReport(page);
  207 |     await page.goto('/zh/report/cafe-fade');
  208 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  209 | 
  210 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  211 | 
  212 |     // Bond report markdown rendered in .report-content.
  213 |     const interpretation = page.locator('#bond-interpretation');
  214 |     await expect(interpretation).toContainText('缘分分析');
  215 |     await expect(interpretation).toContainText('金木相冲');
  216 | 
  217 |     // Title reflects bond product.
  218 |     await expect(page.locator('#report-header-title')).toContainText('合盘报告');
  219 |   });
  220 | 
  221 |   // ── naming report ──
  222 | 
  223 |   test('paid naming report renders candidates', async () => {
  224 |     await mockNamingReport(page);
  225 |     await page.goto('/zh/report/babe-face');
  226 |     await page.waitForSelector('#report-header-title', { timeout: 10000 });
  227 | 
  228 |     await expect(page.locator('#report-content')).toBeVisible({ timeout: 10000 });
  229 | 
  230 |     const interpretation = page.locator('#naming-interpretation');
  231 |     await expect(interpretation).toContainText('起名报告');
  232 |     await expect(interpretation).toContainText('陈明远');
  233 | 
  234 |     await expect(page.locator('#report-header-title')).toContainText('起名报告');
  235 |   });
  236 | 
  237 |   // ── polling → ready transition ──
  238 | 
  239 |   test('transitions from payment polling to ready when order is paid', async () => {
  240 |     let callCount = 0;
  241 |     await page.route('**/api/reports/abba-123', async (route) => {
  242 |       callCount++;
  243 |       if (callCount <= 2) {
```