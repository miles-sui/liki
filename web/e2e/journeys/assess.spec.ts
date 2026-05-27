import { test, expect } from '../fixtures';
import { AssessPage } from '../pages/assess';

// The standalone /assess page navigates through all 30 questions with ❯ arrows,
// then submits at the end via "Submit Round". Uses the Page Object to reduce
// per-question overhead and avoid flaky visibility checks.
test('complete full assessment and reach result page', async ({ page }) => {
  const assess = new AssessPage(page);
  await assess.goto('en');
  await assess.completeAllQuestions();
  await assess.waitForResult();
  await expect(page.locator('#result-radar')).toBeVisible({ timeout: 10000 });
});

test('complete full assessment in zh-CN locale', async ({ page }) => {
  const assess = new AssessPage(page);
  await assess.goto('zh-CN');
  await expect(page.locator('text=选择最符合你的选项（可多选或不选）')).toBeVisible({ timeout: 5000 });
  await assess.completeAllQuestions('zh-CN');
  await page.waitForURL(/\/zh-CN\/result/, { timeout: 30000 });
  await expect(page.locator('#result-radar')).toBeVisible({ timeout: 10000 });
});

test('zero selections submit shows toast and does not navigate', async ({ page }) => {
  const assess = new AssessPage(page);
  await assess.goto('zh-CN');
  await assess.waitForQuestions();
  // Navigate to Q30 (last question) where submit button appears.
  await assess.goNextTimes(29);
  // Submit without selecting any options.
  await assess.submit('zh-CN');
  // Toast with the empty-submit message should appear.
  const toastText = await assess.getToastText();
  expect(toastText).toContain('请至少选择一些选项');
  // Should still be on the assess page (no redirect).
  expect(page.url()).toContain('/assess');
});
