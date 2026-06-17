import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for manual pages:
// /manual (index with sidebar navigation)
// /manual/:category/:slug (dynamic article pages)
// Covers: page load, sidebar category navigation, article display.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('マニュアル フロー E2E', () => {
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

  test('/manual — マニュアルページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/manual\/screens\/00-overview/);
      await expect(page.getByRole('heading', { name: 'Animal Ekarte 取扱説明書', level: 1 })).toBeVisible({
        timeout: 15000,
      });
    } finally {
      await page.close();
    }
  });

  test('/manual — サイドバーが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual', { waitUntil: 'domcontentloaded' });
      const sidebar = page.getByRole('navigation', { name: 'マニュアル目次' });
      await expect(sidebar).toBeVisible({ timeout: 15000 });

      const navLinks = sidebar.getByRole('link');
      const linkCount = await navLinks.count();
      expect(linkCount).toBeGreaterThanOrEqual(1);
    } finally {
      await page.close();
    }
  });

  test('/manual — サイドバーのカテゴリをクリックできる', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual', { waitUntil: 'domcontentloaded' });
      await page.getByRole('tab', { name: '業務フロー' }).click();
      await expect(page.getByRole('tab', { name: '業務フロー' })).toHaveAttribute('aria-selected', 'true');
      const sidebar = page.getByRole('navigation', { name: 'マニュアル目次' });
      await expect(sidebar.getByRole('link', { name: '新規飼主の登録から初診会計まで' })).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });

  test('/manual/:category/:slug — マニュアル記事ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual/screens/03-owners', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/manual\/screens\/03-owners/);
      await expect(page.getByRole('heading', { name: '飼主・ペット管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
    } finally {
      await page.close();
    }
  });

  test('/manual — コンテンツが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/manual\/screens\/00-overview/);
      await expect(page.getByLabel('マニュアル内検索')).toBeVisible({ timeout: 10000 });
      await page.getByLabel('マニュアル内検索').fill('会計');
      await expect(page.getByRole('navigation', { name: 'マニュアル目次' }).getByRole('link').first()).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });

  test('/manual — ナビゲーションリンクが機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/manual', { waitUntil: 'domcontentloaded' });
      const sidebar = page.getByRole('navigation', { name: 'マニュアル目次' });
      await sidebar.getByRole('link', { name: '会計' }).click();
      await expect(page).toHaveURL(/\/manual\/screens\/05-accounting/);
      await expect(page.getByRole('heading', { name: '会計', level: 1 })).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
