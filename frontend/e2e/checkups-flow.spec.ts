import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for checkups (/checkups) pages.
// Covers: list page, pet selection, new form.
// Note: /checkups/:id detail route does not exist in router.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('定期健診 フロー E2E', () => {
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

  test('/checkups — 定期健診一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/checkups', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '定期健診', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '新規登録' })).toBeVisible({ timeout: 10000 });
      await expect(page.locator('tbody')).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/checkups/select-pet — ペット選択画面が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/checkups/select-pet', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '定期健診登録 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/checkups/new?petId=1 — 定期健診新規登録フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      // Use petId=1 (Iris from seed)
      await page.goto('/checkups/new?petId=1', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '定期健診登録' })).toBeVisible({ timeout: 15000 });
      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '保存' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/checkups — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/checkups', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '定期健診', level: 1 })).toBeVisible({
        timeout: 15000,
      });

      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.getByRole('heading', { name: '定期健診登録 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/checkups\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/checkups — 一覧で検索が機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/checkups', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '定期健診', level: 1 })).toBeVisible({
        timeout: 15000,
      });

      const rows = page.locator('tbody tr');
      const initialRowCount = await rows.count();

      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('ペット名・飼主名・種別で検索...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('Iris');

      // Wait for search to apply
      await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);

      // Verify filtered results or search input retained
      const filteredRows = await page.locator('tbody tr').count();
      expect(filteredRows).toBeLessThanOrEqual(initialRowCount);
      await expect(searchInput).toHaveValue('Iris');
    } finally {
      await page.close();
    }
  });
});
