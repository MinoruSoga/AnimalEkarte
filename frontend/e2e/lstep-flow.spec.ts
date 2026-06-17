import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for L-step integration pages:
// /lstep/checkup-sync, /lstep/delivery-monitor, /lstep/analytics
// Covers: page load, basic navigation, interaction with filters/selectors.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('Lステップ連携 フロー E2E', () => {
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

  test('/lstep/checkup-sync — 健診リマインダー抽出ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/checkup-sync', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '健診リマインダー抽出', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '対象者を検索' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/lstep/checkup-sync — ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/checkup-sync', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/lstep\/checkup-sync/);
      await expect(page.locator('select[name="checkup_type"]')).toBeVisible({ timeout: 10000 });
      await expect(page.getByPlaceholder('例: 50000')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/lstep/delivery-monitor — 自動配信トリガー監視ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/delivery-monitor', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '自動配信トリガー監視', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '更新' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByTestId('filter-from')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/lstep/delivery-monitor — ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/delivery-monitor', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/lstep\/delivery-monitor/);
      await expect(page.getByTestId('filter-trigger-type')).toBeVisible({ timeout: 10000 });
      await expect(page.getByTestId('filter-status')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/lstep/analytics — Lステップ分析レポートページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/analytics', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'Lステップ分析レポート', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('月次配信統計')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/lstep/analytics — ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/lstep/analytics', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/lstep\/analytics/);
      await expect(page.getByText('配信後来院率')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('heading', { name: '友だち属性 CSV インポート' })).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});
