import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for shifts (/shifts) calendar page.
// Covers: page load, calendar navigation, basic interaction.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('シフト管理 フロー E2E', () => {
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

  test('/shifts — シフト管理カレンダーが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'シフト管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('button', { name: '前月' })).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '翌月' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/shifts — カレンダーナビゲーションが存在する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'シフト管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText(/\d{4}年\d{1,2}月/)).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/shifts — スタッフセレクタが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'シフト管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByRole('combobox').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/shifts — カレンダーを前月に移動できる', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'シフト管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      const monthDisplay = page.getByText(/\d{4}年\d{1,2}月/);
      const currentText = await monthDisplay.textContent();

      await page.getByRole('button', { name: '前月' }).click();

      await expect(monthDisplay).not.toHaveText(currentText ?? '', { timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/shifts — スタッフフィルタが機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/shifts', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'シフト管理', level: 1 })).toBeVisible({
        timeout: 15000,
      });

      await page.getByRole('combobox').first().click();
      await expect(page.getByRole('option').first()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
