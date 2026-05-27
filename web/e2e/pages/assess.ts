import type { Page } from '@playwright/test';

/**
 * Page Object for the assessment flow (assess.html).
 * Encapsulates question selection and submission.
 */
export class AssessPage {
  constructor(private page: Page) {}

  async goto(locale = 'en'): Promise<void> {
    await this.page.goto(`/${locale}/assess`);
  }

  /** Wait for the first question card to appear. */
  async waitForQuestions(): Promise<void> {
    await this.page.locator('.card.bg-base-200').first().waitFor({ state: 'visible', timeout: 15000 });
  }

  /** Select a couple of options on the current card. */
  async selectOptions(): Promise<void> {
    const card = this.page.locator('.card.bg-base-200').first();
    const options = card.locator('button');
    await options.nth(0).click();
    await options.nth(2).click();
    await options.nth(0).waitFor({ state: 'attached', timeout: 3000 });
  }

  /** Click the next arrow button and wait for card transition. */
  async goNext(): Promise<void> {
    await this.page.locator('button:has-text("❯")').click();
    await this.page.waitForTimeout(200);
  }

  /** Click the submit button (locale-aware text). */
  async submit(locale = 'en'): Promise<void> {
    const label = locale === 'zh-CN' ? '提交本轮' : 'Submit Round';
    await this.page.locator(`button:has-text("${label}")`).click();
  }

  /** Complete all 30 questions by selecting options each and advancing. */
  async completeAllQuestions(locale = 'en'): Promise<void> {
    await this.waitForQuestions();
    for (let q = 0; q < 30; q++) {
      await this.selectOptions();
      if (q < 29) {
        await this.goNext();
      } else {
        await this.submit(locale);
      }
    }
  }

  /** Navigate forward n times without selecting options. */
  async goNextTimes(n: number): Promise<void> {
    for (let i = 0; i < n; i++) {
      await this.page.locator('button:has-text("❯")').click();
      await this.page.waitForTimeout(200);
    }
  }

  /** Wait for redirect to result page. */
  async waitForResult(): Promise<void> {
    await this.page.waitForURL(/\/en\/result/, { timeout: 30000 });
  }

  /** Wait for a toast message to appear and return its text. */
  async getToastText(): Promise<string> {
    const toast = this.page.locator('.toast-container .alert');
    await toast.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await toast.first().textContent()) || '';
  }
}
