/**
 * Custom fixture that uses system Chrome in old headless mode.
 * Playwright 1.59 defaults to chrome-headless-shell which requires bundled Chromium.
 * This forces the old --headless flag which works with any system Chrome.
 */
import { test as base, chromium, type Browser, type BrowserContext } from '@playwright/test';

export const test = base.extend<{}, { baseURL: string }>({
  baseURL: ['http://localhost:8080', { option: true }],

  browser: async ({}, use) => {
    const browser: Browser = await chromium.launch({
      executablePath: '/usr/bin/google-chrome-stable',
      headless: true,
      args: ['--no-sandbox', '--disable-setuid-sandbox'],
    });
    await use(browser);
    await browser.close();
  },

  context: async ({ browser, baseURL }, use, testInfo) => {
    const projectUse = testInfo.project.use;
    const context: BrowserContext = await browser.newContext({
      baseURL,
      viewport: projectUse.viewport || { width: 1280, height: 720 },
      userAgent: projectUse.userAgent,
      screen: (projectUse as any).screen,
      deviceScaleFactor: projectUse.deviceScaleFactor,
      isMobile: projectUse.isMobile,
      hasTouch: projectUse.hasTouch,
      ignoreHTTPSErrors: true,
    });
    await use(context);
    await context.close();
  },
});

export { expect } from '@playwright/test';
