import { test, expect } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Smoke coverage for clinical pages.
// Each test: navigate → assert page-specific h1 heading is visible.
// Seed prerequisites: admin@noavet.jp (system admin) has full access to all clinical resources.

test.describe('臨床ページ smoke E2E', () => {
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

  test('受付 (/) — 当日の受付 が表示される', async () => {
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '当日の受付' })).toBeVisible({ timeout: 15000 });
  });

  test('顧客集計 (/aggregation) — 顧客集計ダッシュボード が表示される', async () => {
    await page.goto('/aggregation', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '顧客集計ダッシュボード' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('カルテ管理 (/medical-records) が表示される', async () => {
    await page.goto('/medical-records', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'カルテ管理' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('入院・ホテル管理 (/hospitalization) が表示される', async () => {
    await page.goto('/hospitalization', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '入院・ホテル管理' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('トリミング管理 (/trimming) が表示される', async () => {
    await page.goto('/trimming', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: 'トリミング管理' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('検査管理 (/examinations) が表示される', async () => {
    await page.goto('/examinations', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '検査管理' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('予防接種管理 (/vaccinations) が表示される', async () => {
    await page.goto('/vaccinations', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '予防接種管理' })).toBeVisible({
      timeout: 15000,
    });
  });

  test('定期健診 (/checkups) が表示される', async () => {
    await page.goto('/checkups', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('heading', { name: '定期健診' })).toBeVisible({
      timeout: 15000,
    });
  });
});
