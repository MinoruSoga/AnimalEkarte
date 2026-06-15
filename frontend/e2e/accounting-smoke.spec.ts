import { test, expect } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Seed prerequisites:
//   - admin@noavet.jp (is_system_admin=true) has full accounting permission
//   - owner 1 (林 文明) has completed billings for pet 1 (Iris(イリス), name_kana=いりす)
//   - /v1/accountings returns these records with no date filter set

test.describe('会計 smoke E2E', () => {
  let context: BrowserContext;
  let page: Page;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    page = await context.newPage();
    await loginAsDemoAdmin(page);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('会計一覧 (/accounting) が表示される', async () => {
    // domcontentloaded returns before lazy route chunks load; toBeVisible polls
    // until AccountingList mounts and UnifiedTabs renders (30 s covers Docker lag).
    await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });
    await expect(page.getByRole('tab', { name: '未納者一覧' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('tab', { name: '当日会計' })).toBeVisible({ timeout: 5000 });
  });

  test('会計一覧: ひらがな「いりす」でカタカナ「Iris(イリス)」が表示される (かな非区別検索)', async () => {
    await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });

    // Search input is hidden behind a toggle — click the search button first.
    // The button renders once AccountingListTable mounts (after API resolves).
    const searchToggle = page.getByRole('button', { name: '検索' });
    await expect(searchToggle).toBeVisible({ timeout: 15000 });
    await searchToggle.click();

    const searchInput = page.getByPlaceholder('飼主名、ペット名...');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    // normalizeKana('Iris(イリス)') → 'iris(いりす)'; includes('いりす') → true
    await searchInput.fill('いりす');
    await expect(page.locator('tbody').getByText('Iris', { exact: false })).toBeVisible({
      timeout: 5000,
    });
  });

  test('会計一覧: カタカナ「イリス」でも「Iris(イリス)」が表示される (ひらがな・カタカナ統一検索)', async () => {
    await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });

    const searchToggle = page.getByRole('button', { name: '検索' });
    await expect(searchToggle).toBeVisible({ timeout: 15000 });
    await searchToggle.click();

    const searchInput = page.getByPlaceholder('飼主名、ペット名...');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    // normalizeKana('イリス') → 'いりす', same match as hiragana above
    await searchInput.fill('イリス');
    await expect(page.locator('tbody').getByText('Iris', { exact: false })).toBeVisible({
      timeout: 5000,
    });
  });

  test('未納者一覧タブ (/accounting?tab=unpaid) がアクティブになる', async () => {
    await page.goto('/accounting?tab=unpaid', { waitUntil: 'domcontentloaded' });
    const unpaidTab = page.getByRole('tab', { name: '未納者一覧' });
    await expect(unpaidTab).toBeVisible({ timeout: 30000 });
    await expect(unpaidTab).toHaveAttribute('data-state', 'active');
  });

  test('月次集計レポート (/accounting/reports) が表示される', async () => {
    await page.goto('/accounting/reports');
    await expect(page).toHaveURL(/\/accounting\/reports/, { timeout: 10000 });
    await expect(page.getByRole('heading', { name: '月次集計レポート' })).toBeVisible({
      timeout: 10000,
    });
  });
});
