import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { TrimmingPage } from './pages/trimming-page';

// E2E flow tests for trimming (/trimming) pages.
// Seed data: pet 1 "Iris(イリス)" exists in seed.
// admin@noavet.jp is system_admin with full access.
//
// Design: fresh page per test within shared context.

test.describe('トリミング管理 フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/trimming — トリミング管理一覧が表示される', async () => {
    const page = await context.newPage();
    const trimming = new TrimmingPage(page);
    try {
      await trimming.gotoList();
      await expect(trimming.listHeading()).toBeVisible();
      await expect(trimming.newButton()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('/trimming — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    const trimming = new TrimmingPage(page);
    try {
      await trimming.gotoList();
      await expect(trimming.listHeading()).toBeVisible();

      await trimming.newButton().click();
      await expect(trimming.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/trimming\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/trimming/new?petId=1 — トリミング登録フォームが表示される', async () => {
    const page = await context.newPage();
    const trimming = new TrimmingPage(page);
    try {
      await trimming.gotoNew('?petId=1');
      await expect(trimming.newFormHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(trimming.irisText()).toBeVisible({ timeout: 10000 });
      await expect(trimming.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
