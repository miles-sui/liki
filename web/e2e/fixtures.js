import { test as base, expect, chromium } from '@playwright/test';

const test = base.extend({
  browser: async ({}, use) => {
    const browser = await chromium.launch({
      executablePath: process.env.CHROME_BIN || '/usr/bin/google-chrome-stable',
      headless: true,
      args: ['--no-sandbox', '--disable-setuid-sandbox'],
    });
    await use(browser);
    await browser.close();
  },

  context: async ({ browser, baseURL }, use, testInfo) => {
    const projectUse = testInfo.project.use;
    const context = await browser.newContext({
      baseURL,
      viewport: projectUse.viewport || { width: 1280, height: 720 },
      ignoreHTTPSErrors: true,
    });
    await use(context);
    await context.close();
  },
});

export { test, expect };
