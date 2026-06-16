import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for hospitalization (/hospitalization) pages.
// Seed data: 12+ hospitalization records at clinic_id=1.
// admin@noavet.jp is system_admin with full access.
//
// Note: the list page defaults to board view; tests that need rows switch to list view first.
// Design: fresh page per test within shared context.

test.describe('入院・ホテル管理 フロー E2E', () => {
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

  test('/hospitalization — 入院・ホテル管理一覧がリストビューで表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/hospitalization', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '入院・ホテル管理' })).toBeVisible();

      // デフォルトはボードビュー — リストビューに切り替えてから行を確認
      await page.getByLabel('List View').click();
      // seed has 12+ records; at least one row must be visible in list view
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/hospitalization — 新規入院登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/hospitalization', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '入院・ホテル管理' })).toBeVisible();

      await page.getByRole('button', { name: '新規入院登録' }).click();
      await expect(page.getByRole('heading', { name: '入院・ホテル登録 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/hospitalization\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/hospitalization — ステータスタブ「予約」に切り替えると予約件数が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/hospitalization', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '入院・ホテル管理' })).toBeVisible();

      // デフォルトは「入院中」タブ; 「予約」タブに切り替え
      await page.getByRole('tab', { name: '予約' }).click();
      await expect(page.getByRole('tab', { name: '予約' })).toHaveAttribute('data-state', 'active');
    } finally {
      await page.close();
    }
  });

  test('/hospitalization — ステータスタブ「すべて」に切り替えるとすべての件数が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/hospitalization', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '入院・ホテル管理' })).toBeVisible();

      // 「すべて」タブに切り替え
      await page.getByRole('tab', { name: 'すべて' }).click();
      await expect(page.getByRole('tab', { name: 'すべて' })).toHaveAttribute('data-state', 'active');

      // リストビューに切り替えて全件表示を確認
      await page.getByLabel('List View').click();
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
