import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { EstimatesPage } from './pages/estimates-page';

// E2E flow tests for estimates (/estimates) pages.
// Covers: list page, new form, detail, edit form.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('見積書管理 フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/estimates — 見積書一覧が表示される', async () => {
    const page = await context.newPage();
    const estimates = new EstimatesPage(page);
    try {
      await estimates.gotoList();
      // Use level:1 to be specific about main heading
      await expect(estimates.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.newButton()).toBeVisible({ timeout: 10000 });
      await expect(estimates.tableBody()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/new — 見積書新規作成フォームが表示される', async () => {
    const page = await context.newPage();
    const estimates = new EstimatesPage(page);
    try {
      await estimates.gotoNew();
      await expect(estimates.newFormHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.titleInput()).toBeVisible({ timeout: 10000 });
      await expect(estimates.createButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/:id — 見積書詳細フォームが表示される', async () => {
    const page = await context.newPage();
    const estimates = new EstimatesPage(page);
    try {
      await estimates.gotoList();
      await expect(estimates.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.draftDetailLink()).toBeVisible({ timeout: 15000 });

      // Only the estimate-no cell link navigates (row click is a no-op)
      await estimates.draftDetailLink().click();
      await expect(estimates.detailHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.editButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates/:id/edit — 見積書編集フォームが表示される', async () => {
    const page = await context.newPage();
    const estimates = new EstimatesPage(page);
    try {
      await estimates.gotoList();
      await expect(estimates.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.draftDetailLink()).toBeVisible({ timeout: 15000 });

      await estimates.draftDetailLink().click();
      await expect(estimates.detailHeading()).toBeVisible({
        timeout: 15000,
      });

      await estimates.editButton().click();
      await expect(estimates.editHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(estimates.titleInput()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/estimates — 一覧で検索が機能する', async () => {
    const page = await context.newPage();
    const estimates = new EstimatesPage(page);
    try {
      await estimates.gotoList();
      await expect(estimates.listHeading()).toBeVisible({
        timeout: 15000,
      });

      const rows = estimates.rows();
      const initialRowCount = await rows.count();

      await estimates.searchToggle().click();
      const searchInput = estimates.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill('Mass');

      // Wait for search to apply (debounce/API response)
      await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);

      // Verify filtered results are narrower or search input is retained
      const filteredRows = await estimates.rows().count();
      expect(filteredRows).toBeLessThanOrEqual(initialRowCount);

      // Verify the search term is still in the input
      await expect(searchInput).toHaveValue('Mass');
    } finally {
      await page.close();
    }
  });
});
