const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './journeys',
  fullyParallel: false,
  workers: 1,
  retries: 2,
  timeout: 60000,
  expect: { timeout: 10000 },
  use: {
    baseURL: 'http://localhost:8080',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chrome',
      use: {
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: 'mobile',
      testMatch: /assess|account/,
      use: {
        ...devices['iPhone 14'],
      },
    },
  ],
});
