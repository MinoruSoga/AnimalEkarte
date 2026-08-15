import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { DEMO_ACCOUNTING_KANA_PET, DEMO_ACCOUNTING_OFFPAGE_PET } from './helpers/demo-seed';
import { AccountingPage } from './pages/accounting-page';

// Seed prerequisites:
//   - E2E_LOGIN_* account has accounting view permission
//   - List text search is server-side (?search=) with kana symmetry
//   - DEMO_ACCOUNTING_OFFPAGE_PET is not on page 1 without search
//   - Iris has no billing rows in 003_demo
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

  test('会計一覧: ひらがな検索でページ外のカタカナペット名がヒットする (サーバサイド・かな非区別)', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });

      // Off-page pet must not be visible before search (proves not client-only page filter)
      await expect(
        page.locator('tbody').getByText(DEMO_ACCOUNTING_OFFPAGE_PET.displayName, { exact: true }),
      ).toHaveCount(0);

      const searchToggle = accounting.searchToggle();
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();

      const searchInput = accounting.searchInput();
      await expect(searchInput).toBeVisible({ timeout: 5000 });

      // べるす → ベルス (server NormalizeKana + translate)
      await searchInput.fill(DEMO_ACCOUNTING_OFFPAGE_PET.hiraganaSearch);
      await expect(
        page.locator('tbody').getByText(DEMO_ACCOUNTING_OFFPAGE_PET.displayName, { exact: true }).first(),
      ).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('会計一覧: カタカナ検索でも同一ペット名が表示される (ひらがな・カタカナ統一検索)', async () => {
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

      await searchInput.fill(DEMO_ACCOUNTING_OFFPAGE_PET.katakanaSearch);
      await expect(
        page.locator('tbody').getByText(DEMO_ACCOUNTING_OFFPAGE_PET.displayName, { exact: true }).first(),
      ).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('会計一覧: ページ1のペットもサーバ検索でヒットする', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await accounting.searchToggle().click();
      await accounting.searchInput().fill(DEMO_ACCOUNTING_KANA_PET.hiraganaSearch);
      await expect(accounting.kanaPetCell()).toBeVisible({ timeout: 15000 });
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
