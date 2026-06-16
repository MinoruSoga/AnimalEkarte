import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

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
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/accounting — 会計一覧の行クリックで会計精算画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
      // AccountingList mounts after API resolves (may take 30 s with Docker lag)
      await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();
      // AccountingDetail renders PageLayout with title="会計精算" only after API resolves
      await expect(page.getByRole('heading', { name: '会計精算' })).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/accounting — 会計一覧の「Iris」行クリックで会計精算画面に遷移する (seed: 林 文明 / Iris)', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });

      // Search for Iris to get a stable row from seed data
      const searchToggle = page.getByRole('button', { name: '検索' });
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名...');
      await expect(searchInput).toBeVisible({ timeout: 5000 });
      await searchInput.fill('Iris');

      // Click the first Iris row
      await expect(page.locator('tbody').getByText('Iris', { exact: false }).first()).toBeVisible({
        timeout: 5000,
      });
      await page.locator('tbody tr').filter({ hasText: 'Iris' }).first().click();

      // Detail page renders after API resolves
      await expect(page.getByRole('heading', { name: '会計精算' })).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/accounting/reports — 月次集計レポートにセレクタが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/accounting/reports', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '月次集計レポート' })).toBeVisible({
        timeout: 15000,
      });
      // Reports page always renders year/month selectors
      await expect(page.getByRole('combobox').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/accounting/:id — 会計精算フォームの確定ボタンが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/accounting', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('tab', { name: '会計一覧' })).toBeVisible({ timeout: 30000 });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      // Click first row to navigate to detail
      await page.locator('tbody tr').first().click();
      await expect(page.getByRole('heading', { name: '会計精算' })).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);

      const confirmButton = page.getByRole('button', { name: /会計を確定する|修正を保存する/ });
      await expect(confirmButton).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
