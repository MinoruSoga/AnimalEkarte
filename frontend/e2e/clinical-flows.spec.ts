import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for clinical (medical records) pages.
// Seed data: owner id=1 "林 文明", pet id=1 "Iris(イリス)", medical record id=91.
// admin@noavet.jp is system_admin with full access.

test.describe('カルテ フロー E2E', () => {
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

  test('/medical-records — カルテ管理一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/medical-records', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ管理' })).toBeVisible();
      // テーブル行が1件以上存在する
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/medical-records — 検索で「林」を入力すると林 文明のカルテが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/medical-records', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ管理' })).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、カルテNo、主訴で検索...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(page.getByText('林 文明').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/medical-records — 行クリックでカルテ編集画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/medical-records', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ管理' })).toBeVisible();
      await expect(page.locator('tbody tr').first()).toBeVisible({ timeout: 15000 });

      await page.locator('tbody tr').first().click();
      await expect(page.getByRole('heading', { name: 'カルテ編集' })).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/medical-records\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/medical-records/select-pet — ペット選択画面が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/medical-records/select-pet', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ作成 - ペット選択' })).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/medical-records — 新規カルテ登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/medical-records', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ管理' })).toBeVisible();

      await page.getByRole('button', { name: '新規カルテ登録' }).click();
      await expect(page.getByRole('heading', { name: 'カルテ作成 - ペット選択' })).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/medical-records\/select-pet/);
    } finally {
      await page.close();
    }
  });
});
