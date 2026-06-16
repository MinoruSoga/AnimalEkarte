import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for owners (/owners) pages.
// Seed data: owner id=1 "林 文明", pet id=1 "Iris(イリス)" at clinic_id=1.
// admin@noavet.jp is system_admin with full access.

test.describe('飼主フロー E2E', () => {
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

  test('/owners — 飼主一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/owners', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '飼主・ペット一覧' })).toBeVisible();
      // テーブル行が1件以上存在する
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/owners — 検索で「林」を入力すると林 文明が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/owners', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '飼主・ペット一覧' })).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、飼主No、種別...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(page.getByText('林 文明').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/owners — 行クリックで飼主詳細(編集)画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/owners', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '飼主・ペット一覧' })).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、飼主No、種別...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(page.getByText('林 文明').first()).toBeVisible({ timeout: 10000 });

      // 行クリック（first() で複数行対応）
      await page.getByText('林 文明').first().click();
      // 飼主編集画面に遷移
      await expect(page.getByRole('heading', { name: '飼主・ペット　編集' })).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/owners\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/owners/new — 飼主新規登録フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/owners', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '飼主・ペット一覧' })).toBeVisible();

      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.getByRole('heading', { name: '飼主・ペット　登録' })).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/owners\/new/);
    } finally {
      await page.close();
    }
  });

  test('/owners/1 — 飼主編集フォームに飼主名「林 文明」が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/owners/1', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '飼主・ペット　編集' })).toBeVisible({ timeout: 15000 });
      // 飼主名フィールドに「林 文明」が入力されている
      // Note: getByDisplayValue は Testing Library の API。Playwright では locator + toHaveValue を使う
      await expect(page.locator('#ownerName')).toHaveValue('林 文明', { timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
