import type { Page, Locator } from '@playwright/test';

/**
 * Page Object for the profile page (profile.html).
 * Encapsulates bond-related actions: Compare with Me, bond cards, match links.
 */
export class ProfilePage {
  constructor(private page: Page) {}

  async goto(username: string, locale = 'en'): Promise<void> {
    await this.page.goto(`/${locale}/profile/${encodeURIComponent(username)}`);
  }

  /** Wait for the profile to load (Alpine x-data="profilePage()"). */
  async waitForProfile(): Promise<void> {
    await this.page.locator('[x-data="profilePage()"]').waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Click "Compare with Me" to compute an instant bond. */
  async clickCompareWithMe(): Promise<void> {
    await this.page.locator('button:has-text("Compare with Me")').click();
  }

  /** Wait for bond result cards to appear (attached to DOM + ECharts ready). */
  async waitForBondResult(): Promise<void> {
    await this.page.locator('#bond-influence-self').waitFor({ state: 'attached', timeout: 15000 });
    await this.page.waitForFunction(() => {
      const el = document.getElementById('bond-influence-self');
      return !!(el && (window as any).echarts?.getInstanceByDom(el));
    }, { timeout: 15000 });
  }

  /** Get bond radar chart elements. */
  getBondCards(): { self: Locator; other: Locator } {
    return {
      self: this.page.locator('#bond-influence-self'),
      other: this.page.locator('#bond-influence-other'),
    };
  }

  /** Get the identity badge element. */
  getIdentityBadge(): Locator {
    return this.page.locator('[x-text="identity.label"], .badge-outline').first();
  }

  /** Click "Create Match Link" button. */
  async clickCreateMatchLink(): Promise<void> {
    await this.page.locator('button:has-text("Create Match Link")').click();
  }

  /** Check if "View All Bonds" link is visible. */
  async hasViewAllBonds(): Promise<boolean> {
    return this.page.locator('a:has-text("View All Bonds")').isVisible({ timeout: 3000 }).catch(() => false);
  }

  /** Click "View All Bonds" link. */
  async clickViewAllBonds(): Promise<void> {
    await this.page.locator('a:has-text("View All Bonds")').click();
  }

  /** Check if the Compare button is visible for a visitor. */
  async hasCompareButton(): Promise<boolean> {
    return this.page.locator('button:has-text("Compare with Me")').isVisible({ timeout: 2000 }).catch(() => false);
  }

  /** Check if profile shows "not found" message. */
  async isNotFound(): Promise<boolean> {
    return this.page.getByText('Profile not found').isVisible({ timeout: 3000 }).catch(() => false);
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
