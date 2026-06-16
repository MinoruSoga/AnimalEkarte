import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for trimming (/trimming) pages.
// Seed data: pet 1 "Iris(イリス)" exists in seed.
// admin@noavet.jp is system_admin with full access.
//
// Design: fresh page per test within shared context.

test.describe('トリミング管理 フロー E2E', () => {
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

  test('/trimming — トリミング管理一覧が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/trimming', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'トリミング管理' })).toBeVisible();
      await expect(page.getByRole('button', { name: '新規登録' })).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('/trimming — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/trimming', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'トリミング管理' })).toBeVisible();

      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.getByRole('heading', { name: 'トリミング登録 - ペット選択' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/trimming\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/trimming/new?petId=1 — トリミング登録フォームが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/trimming/new?petId=1', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'トリミング登録' })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '保存' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
