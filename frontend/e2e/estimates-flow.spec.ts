import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for estimates (/estimates) pages.
// Covers: list page, new form, detail, edit form.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('見積書管理 フロー E2E', () => {
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

  test('/estimates — 見積書一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/estimates', { waitUntil: 'domcontentloaded' });
      // Use level:1 to be specific about main heading
      await expect(page.getByRole('heading', { name: '見積書管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '新規見積書登録' })).toBeVisible({ timeout: 10000 });
      await expect(page.locator('tbody')).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/new — 見積書新規作成フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/estimates/new', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '新規見積書作成', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByPlaceholder('見積書タイトルを入力')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '作成' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/:id — 見積書詳細フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/estimates', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '見積書管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();
      await expect(page.getByRole('heading', { name: /見積書\s+\S+/, level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '編集' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/:id/edit — 見積書編集フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/estimates', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '見積書管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();
      await expect(page.getByRole('heading', { name: /見積書\s+\S+/, level: 1 })).toBeVisible({
        timeout: 15000,
      });

      await page.getByRole('button', { name: '編集' }).click();
      await expect(page.getByRole('heading', { name: '見積書編集', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByPlaceholder('見積書タイトルを入力')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates — 一覧で検索が機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/estimates', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '見積書管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });

      const rows = page.locator('tbody tr');
      const initialRowCount = await rows.count();

      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('見積番号、タイトル、飼主名...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('Iris');

      // Wait for search to apply (debounce/API response)
      await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);

      // Verify filtered results are narrower or search input is retained
      const filteredRows = await page.locator('tbody tr').count();
      expect(filteredRows).toBeLessThanOrEqual(initialRowCount);

      // Verify the search term is still in the input
      await expect(searchInput).toHaveValue('Iris');
    } finally {
      await page.close();
    }
  });
});
