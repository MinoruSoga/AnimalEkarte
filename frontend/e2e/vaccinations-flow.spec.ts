import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for vaccinations (/vaccinations) pages.
// Seed data: 12+ vaccination records at clinic_id=1 (vaccinations).
// admin@noavet.jp is system_admin with full access.
//
// Design: fresh page per test within shared context.

test.describe('予防接種管理 フロー E2E', () => {
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

  test('/vaccinations — 予防接種管理一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/vaccinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '予防接種管理' })).toBeVisible();
      // seed has 12+ records; at least one row must be visible
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 検索フィルタが機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/vaccinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '予防接種管理' })).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、予防接種名...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(page.getByText('林 文明').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/vaccinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '予防接種管理' })).toBeVisible();

      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.getByRole('heading', { name: 'ワクチン接種 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/vaccinations\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 行クリックで予防接種詳細画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/vaccinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '予防接種管理' })).toBeVisible();
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();
      await expect(page.getByRole('heading', { name: '予防接種詳細・編集' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/vaccinations\/\d+/);
    } finally {
      await page.close();
    }
  });
});
