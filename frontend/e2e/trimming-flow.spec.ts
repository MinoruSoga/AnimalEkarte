import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { DEMO_IRIS_PET } from './helpers/demo-seed';
import { TrimmingPage } from './pages/trimming-page';

// E2E flow tests for trimming (/trimming) pages.
// Demo seed: Iris pet id=1000099 (not petId=1).
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

  test(`/trimming/new?petId=${DEMO_IRIS_PET.id} — トリミング登録フォームが表示される`, async () => {
    const page = await context.newPage();
    const trimming = new TrimmingPage(page);
    try {
      await trimming.gotoNew(`?petId=${DEMO_IRIS_PET.id}`);
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
