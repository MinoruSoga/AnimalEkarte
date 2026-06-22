import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { AccountingPage } from './pages/accounting-page';

// E2E flow tests for accounting (/accounting) navigation.
// Smoke coverage (tab visibility, kana search) lives in accounting-smoke.spec.ts.
// This file adds detail navigation: list row → AccountingDetail page.
// Seed data: owner 1 (林 文明) has completed billings for pet 1 (Iris(イリス)).
// admin@noavet.jp is system_admin with full accounting access.
//
// Design: fresh page per test within shared context.

test.describe('会計フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/accounting — 会計一覧の行クリックで会計精算画面に遷移する', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      // AccountingList mounts after API resolves (may take 30 s with Docker lag)
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await expect(accounting.firstRow()).toBeVisible({ timeout: 15000 });

      await accounting.firstRow().click();
      // AccountingDetail renders PageLayout with title="会計精算" only after API resolves
      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/accounting — 会計一覧の「Iris」行クリックで会計精算画面に遷移する (seed: 林 文明 / Iris)', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });

      // Search for Iris to get a stable row from seed data
      const searchToggle = accounting.searchToggle();
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();
      const searchInput = accounting.searchInput();
      await expect(searchInput).toBeVisible({ timeout: 5000 });
      await searchInput.fill('Iris');

      // Click the first Iris row
      await expect(accounting.irisCell()).toBeVisible({ timeout: 5000 });
      await accounting.irisRow().click();

      // Detail page renders after API resolves
      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/accounting/reports — 月次集計レポートにセレクタが表示される', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoReports({ waitUntil: 'domcontentloaded' });
      await expect(accounting.reportsHeading()).toBeVisible({
        timeout: 15000,
      });
      // Reports page always renders year/month selectors
      await expect(accounting.firstCombobox()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/accounting/:id — 会計精算フォームの確定ボタンが表示される', async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await expect(accounting.firstRow()).toBeVisible({ timeout: 15000 });

      // Click first row to navigate to detail
      await accounting.firstRow().click();
      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);

      await expect(accounting.confirmButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
