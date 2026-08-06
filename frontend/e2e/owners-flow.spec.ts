import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { OwnersPage } from './pages/owners-page';

// E2E flow tests for owners (/owners) pages.
// Seed data: owner id=1 "林 文明", pet id=1 "Iris(イリス)" at clinic_id=1.
// admin@noavet.jp is system_admin with full access.

test.describe('飼主フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/owners — 飼主一覧が表示される', async () => {
    const page = await context.newPage();
    const owners = new OwnersPage(page);
    try {
      await owners.gotoList();
      await expect(owners.listHeading()).toBeVisible();
      // テーブル行が1件以上存在する
      await expect(owners.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/owners — 検索で「林」を入力すると林 文明が表示される', async () => {
    const page = await context.newPage();
    const owners = new OwnersPage(page);
    try {
      await owners.gotoList();
      await expect(owners.listHeading()).toBeVisible();

      // PropertyFilter: 検索トグルをクリックして入力欄を表示
      await owners.searchToggle().click();
      const searchInput = owners.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(owners.hayashiText()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/owners — 行クリックで飼主詳細(編集)画面に遷移する', async () => {
    const page = await context.newPage();
    const owners = new OwnersPage(page);
    try {
      await owners.gotoList();
      await expect(owners.listHeading()).toBeVisible();

      await owners.searchToggle().click();
      const searchInput = owners.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(owners.hayashiText()).toBeVisible({ timeout: 15000 });

      // 行クリック（first() で複数行対応）
      await owners.hayashiText().click();
      // 飼主編集画面に遷移
      await expect(owners.editHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/owners\/\d+/);
    } finally {
      await page.close();
    }
  });

  test('/owners/new — 飼主新規登録フォームが表示される', async () => {
    const page = await context.newPage();
    const owners = new OwnersPage(page);
    try {
      await owners.gotoList();
      await expect(owners.listHeading()).toBeVisible();

      await owners.newButton().click();
      await expect(owners.registerHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/owners\/new/);
    } finally {
      await page.close();
    }
  });

  test('/owners/1 — 飼主編集フォームに飼主名「林 文明」が表示される', async () => {
    const page = await context.newPage();
    const owners = new OwnersPage(page);
    try {
      await owners.gotoDetail(1);
      await expect(owners.editHeading()).toBeVisible({ timeout: 15000 });
      // 飼主名フィールドに「林 文明」が入力されている
      // Note: getByDisplayValue は Testing Library の API。Playwright では locator + toHaveValue を使う
      await expect(owners.ownerNameInput()).toHaveValue('林 文明', { timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
