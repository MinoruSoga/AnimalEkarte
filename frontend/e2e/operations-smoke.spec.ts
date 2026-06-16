import { test, expect } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Smoke coverage for operations / integrations pages.
// Each test: navigate → assert page-specific stable element is visible.
// Seed prerequisites: admin@noavet.jp (system admin) has full access to ResourceHospitalSettings
// and ResourceReservations which guard these routes.

test.describe('運用・外部連携ページ smoke E2E', () => {
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

  test('健診リマインダー抽出 (/lstep/checkup-sync) が表示される', async () => {
    await page.goto('/lstep/checkup-sync', { waitUntil: 'domcontentloaded' });
    // CheckupSyncPage renders its own h1 (not via PageLayout)
    await expect(page.getByRole('heading', { name: '健診リマインダー抽出' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('自動配信トリガー監視 (/lstep/delivery-monitor) が表示される', async () => {
    await page.goto('/lstep/delivery-monitor', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '自動配信トリガー監視' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('Lステップ分析レポート (/lstep/analytics) が表示される', async () => {
    await page.goto('/lstep/analytics', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'Lステップ分析レポート' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('LINE予約 基本設定 (/line-reservation) が表示される', async () => {
    await page.goto('/line-reservation', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '基本設定' })).toBeVisible({ timeout: 15000 });
  });

  test('LINE予約 ページ編集 (/line-reservation/page-editor) が表示される', async () => {
    await page.goto('/line-reservation/page-editor', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'ページ編集' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('LINE予約枠設定 (/line-reservation/slots) が表示される', async () => {
    await page.goto('/line-reservation/slots', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'LINE予約枠' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('マニュアル (/manual) が表示される — サイドバータブが確認できる', async () => {
    await page.goto('/manual', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveURL(/\/manual/, { timeout: 10000 });
    // ManualSidebar always renders a tablist with "画面別" and "業務フロー" tabs
    await expect(page.getByRole('tab', { name: '画面別' })).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('tab', { name: '業務フロー' })).toBeVisible({ timeout: 5000 });
  });
});
