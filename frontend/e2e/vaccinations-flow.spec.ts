import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { VaccinationsPage } from './pages/vaccinations-page';

// E2E flow tests for vaccinations (/vaccinations) pages.
// Seed data: 12+ vaccination records at clinic_id=1 (vaccinations).
// admin@noavet.jp is system_admin with full access.
//
// Design: fresh page per test within shared context.

test.describe('予防接種管理 フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/vaccinations — 予防接種管理一覧が表示される', async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();
      // seed has 12+ records; at least one row must be visible
      await expect(vaccinations.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 検索フィルタが機能する', async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = vaccinations.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill('林');
      await expect(vaccinations.hayashiText()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();

      await vaccinations.newButton().click();
      await expect(vaccinations.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/vaccinations\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/vaccinations — 行クリックで予防接種詳細画面に遷移する', async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();
      await expect(vaccinations.firstRow()).toBeVisible({ timeout: 15000 });

      await vaccinations.firstRow().click();
      await expect(vaccinations.detailHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/vaccinations\/\d+/);
    } finally {
      await page.close();
    }
  });
});
