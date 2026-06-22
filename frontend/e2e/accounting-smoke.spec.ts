import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { AccountingPage } from './pages/accounting-page';

// Seed prerequisites:
//   - admin@noavet.jp (is_system_admin=true) has full accounting permission
//   - owner 1 (林 文明) has completed billings for pet 1 (Iris(イリス), name_kana=いりす)
//   - /v1/accountings returns these records with no date filter set
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe('会計 smoke E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('会計一覧 (/accounting) が表示される', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      // domcontentloaded returns before lazy route chunks load; toBeVisible polls
      // until AccountingList mounts and UnifiedTabs renders (30 s covers Docker lag).
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await expect(accounting.unpaidTab()).toBeVisible({ timeout: 5000 });
      await expect(accounting.sameDayTab()).toBeVisible({ timeout: 5000 });
    } finally {
      await page.close();
    }
  });

  test('会計一覧: ひらがな「いりす」でカタカナ「Iris(イリス)」が表示される (かな非区別検索)', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });

      // Search input is hidden behind a toggle — click the search button first.
      // The button renders once AccountingListTable mounts (after API resolves).
      const searchToggle = accounting.searchToggle();
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();

      const searchInput = accounting.searchInput();
      await expect(searchInput).toBeVisible({ timeout: 5000 });

      // normalizeKana('Iris(イリス)') → 'iris(いりす)'; includes('いりす') → true
      // .first() because seed may contain multiple Iris billing rows
      await searchInput.fill('いりす');
      await expect(accounting.irisCell()).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await page.close();
    }
  });

  test('会計一覧: カタカナ「イリス」でも「Iris(イリス)」が表示される (ひらがな・カタカナ統一検索)', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });

      const searchToggle = accounting.searchToggle();
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();

      const searchInput = accounting.searchInput();
      await expect(searchInput).toBeVisible({ timeout: 5000 });

      // normalizeKana('イリス') → 'いりす', same match as hiragana above
      // .first() because seed may contain multiple Iris billing rows
      await searchInput.fill('イリス');
      await expect(accounting.irisCell()).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await page.close();
    }
  });

  test('未納者一覧タブ (/accounting?tab=unpaid) がアクティブになる', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoUnpaid();
      const unpaidTab = accounting.unpaidTab();
      await expect(unpaidTab).toBeVisible({ timeout: 30000 });
      await expect(unpaidTab).toHaveAttribute('data-state', 'active');
    } finally {
      await page.close();
    }
  });

  test('月次集計レポート (/accounting/reports) が表示される', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoReports();
      await expect(page).toHaveURL(/\/accounting\/reports/, { timeout: 10000 });
      await expect(accounting.reportsHeading()).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});
