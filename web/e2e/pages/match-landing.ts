import { expect, type Page, type Locator } from '@playwright/test';

/**
 * Page Object for the match link landing page (match_landing.html).
 * Encapsulates name input, assessment submission, use-existing flow, and bond result.
 */
export class MatchLandingPage {
  constructor(private page: Page) {}

  async goto(token: string, locale = 'en'): Promise<void> {
    await this.page.goto(`/${locale}/m/${token}`);
  }

  /** Wait for the page to load (Alpine component). */
  async waitForLoad(): Promise<void> {
    await this.page.locator('[x-data="matchLandingPage()"]').waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Fill in the name field for anonymous submission. */
  async fillName(name: string): Promise<void> {
    const input = this.page.locator('input[type="text"]').first();
    await input.fill(name);
  }

  /** Complete all assessment questions (one at a time with nav arrows). */
  async completeAssessment(): Promise<void> {
    for (let q = 0; q < 30; q++) {
      // Wait for the question card to be visible.
      const card = this.page.locator('.card.bg-base-200').first();
      await card.waitFor({ state: 'visible', timeout: 10000 });

      // Select two options. The option buttons are the only <button> elements inside the card.
      const btns = card.locator('button');
      await btns.nth(0).click();
      await btns.nth(2).click();
      await expect(btns.nth(0)).toHaveClass(/btn-primary/, { timeout: 3000 });
      await expect(btns.nth(2)).toHaveClass(/btn-primary/, { timeout: 3000 });

      if (q < 29) {
        const nextBtn = this.page.locator('button:has-text("❯")');
        await nextBtn.waitFor({ state: 'visible', timeout: 5000 });
        await nextBtn.click();
        await expect(card).not.toBeVisible({ timeout: 5000 }).catch(() => {});
      } else {
        // Last question — submit via in-card "Submit Round" or outer "See Results".
        const submitBtn = this.page.locator('button:has-text("Submit Round"), button:has-text("See Results")').first();
        await submitBtn.waitFor({ state: 'visible', timeout: 10000 });
        await submitBtn.click();
      }
    }
  }

  /** Click "Compute Bond" / "Use Existing" button. */
  async clickUseExisting(): Promise<void> {
    await this.page.locator('button:has-text("Compute Bond")').click();
  }

  /** Click "Or complete a new assessment instead" link. */
  async clickOrAssess(): Promise<void> {
    await this.page.locator('a:has-text("Or complete")').click();
  }

  /** Wait for bond result cards to appear (attached to DOM + ECharts ready). */
  async waitForBondResult(): Promise<void> {
    await this.page.locator('#bond-influence-self').waitFor({ state: 'attached', timeout: 15000 });
    await this.page.waitForFunction(() => {
      const el = document.getElementById('bond-influence-self');
      return !!(el && (window as any).echarts?.getInstanceByDom(el));
    }, { timeout: 15000 });
  }

  /** Get bond result card elements. */
  getBondCards(): { self: Locator; other: Locator } {
    return {
      self: this.page.locator('#bond-influence-self'),
      other: this.page.locator('#bond-influence-other'),
    };
  }

  /** Check if the "Compute Bond" (use existing) button is visible. */
  async hasUseExistingButton(): Promise<boolean> {
    return this.page.locator('button:has-text("Compute Bond")').isVisible({ timeout: 3000 }).catch(() => false);
  }

  /** Check if the page shows an error. */
  async hasError(): Promise<boolean> {
    return this.page.locator('[x-if="!loading && error"]').isVisible({ timeout: 3000 }).catch(() => false);
  }

  /** Get the error message text. */
  async getErrorText(): Promise<string> {
    return this.page.locator('.text-error').textContent() || '';
  }

  /** Read radar chart series data via ECharts instance. Returns array of [name, value[]] pairs. */
  async getRadarSeriesData(chartId: string): Promise<[string, number[]][]> {
    return this.page.evaluate((id) => {
      const el = document.getElementById(id);
      if (!el) return [];
      const inst = (window as any).echarts?.getInstanceByDom(el);
      if (!inst) return [];
      const opt = inst.getOption();
      if (!opt.series) return [];
      return opt.series.map((s: any) => [s.name || '', s.data?.[0]?.value || []] as [string, number[]]);
    }, chartId);
  }
}
