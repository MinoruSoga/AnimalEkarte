import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for examinations (/examinations) pages.
// Covers: list page, pet selection, new form, detail form.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('検査管理 フロー E2E', () => {
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

  test('/examinations — 検査管理一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/examinations', { waitUntil: 'domcontentloaded' });
      // Use level:1 to avoid strict locator violation with multiple headings
      await expect(page.getByRole('heading', { name: /検査管理/, level: 1 })).toBeVisible({ timeout: 15000 });
      // Verify page loaded even if list is empty
      await expect(page).toHaveURL(/\/examinations/);
    } finally {
      await page.close();
    }
  });

  test('/examinations/select-pet — ペット選択画面が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/examinations/select-pet', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '検査登録 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/examinations/new?petId=1 — 検査登録フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      // Use petId=1 (Iris from seed)
      await page.goto('/examinations/new?petId=1', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '新規検査登録' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '保存' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/examinations/:id — 検査詳細フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      // Navigate to list first to find an examination ID
      await page.goto('/examinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '検査管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();

      await expect(page.getByRole('heading', { name: '検査詳細・編集', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/examinations\/\d+/);
      await expect(page.getByRole('button', { name: '保存' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/examinations — 検査一覧で検索が機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/examinations', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '検査管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
      const firstTestType = (await page.locator('tbody tr').first().locator('td').nth(3).textContent())?.trim();
      expect(firstTestType).toBeTruthy();

      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、検査種別...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill(firstTestType ?? '');
      await expect(page.locator('tbody tr').first()).toContainText(firstTestType ?? '', { timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
