import { test, expect } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Smoke coverage for business / accounting-extras pages.
// Each test: navigate → assert page-specific h1 heading is visible.
// Seed prerequisites: admin@noavet.jp has full access to all business resources.

test.describe('業務ページ smoke E2E', () => {
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

  test('在庫管理 (/inventory) が表示される', async () => {
    await page.goto('/inventory', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '在庫管理' })).toBeVisible({ timeout: 15000 });
  });

  test('見積書管理 (/estimates) が表示される', async () => {
    await page.goto('/estimates', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '見積書管理' })).toBeVisible({ timeout: 15000 });
  });

  test('シフト管理 (/shifts) が表示される', async () => {
    await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'シフト管理' })).toBeVisible({ timeout: 15000 });
  });

  test('レジ締め (/accounting/close) が表示される', async () => {
    await page.goto('/accounting/close', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'レジ締め' })).toBeVisible({ timeout: 15000 });
  });

  test('締め履歴 (/accounting/close/history) が表示される', async () => {
    await page.goto('/accounting/close/history', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '締め履歴' })).toBeVisible({ timeout: 15000 });
  });
});
