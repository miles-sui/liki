// Page object for report.html (report viewer page)

export class ReportPage {
  constructor(page) {
    this.page = page;
  }

  async goto(orderID, locale = 'en') {
    await this.page.goto(`/${locale}/report/${orderID || ''}`);
  }

  async waitForError() {
    await this.page.locator('[x-show="error"]').waitFor({ state: 'visible', timeout: 10000 });
  }

  async waitForPolling() {
    await this.page.locator('.poll-hint').waitFor({ state: 'visible', timeout: 10000 });
  }

  async waitForReady() {
    await this.page.locator('[x-show="ready"]').waitFor({ state: 'visible', timeout: 10000 });
  }

  get errorText() {
    return this.page.locator('[x-show="error"] p');
  }

  get homeButton() {
    return this.page.locator('[x-show="error"] a');
  }

  get saveBanner() {
    return this.page.locator('.save-banner');
  }

  get reportContent() {
    return this.page.locator('[x-show="ready"]');
  }
}
