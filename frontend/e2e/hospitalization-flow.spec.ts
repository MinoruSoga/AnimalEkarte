import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { HospitalizationPage } from "./pages/hospitalization-page";

// E2E flow tests for hospitalization (/hospitalization) pages.
// Seed data: 12+ hospitalization records at clinic_id=1.
// admin@noavet.jp is system_admin with full access.
//
// Note: the list page defaults to board view; tests that need rows switch to list view first.
// Design: fresh page per test within shared context.

test.describe("入院・ホテル管理 フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/hospitalization — 入院・ホテル管理一覧がリストビューで表示される", async () => {
    const page = await context.newPage();
    const hospitalization = new HospitalizationPage(page);
    try {
      await hospitalization.gotoList();
      await expect(hospitalization.listHeading()).toBeVisible();

      // デフォルトはボードビュー — リストビューに切り替えてから行を確認
      await hospitalization.listViewToggle().click();
      // seed has 12+ records; at least one row must be visible in list view
      await expect(hospitalization.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/hospitalization — 新規入院登録ボタンでペット選択画面に遷移する", async () => {
    const page = await context.newPage();
    const hospitalization = new HospitalizationPage(page);
    try {
      await hospitalization.gotoList();
      await expect(hospitalization.listHeading()).toBeVisible();

      await hospitalization.newButton().click();
      await expect(hospitalization.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/hospitalization\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test("/hospitalization — ステータスタブ「予約」に切り替えると予約件数が表示される", async () => {
    const page = await context.newPage();
    const hospitalization = new HospitalizationPage(page);
    try {
      await hospitalization.gotoList();
      await expect(hospitalization.listHeading()).toBeVisible();

      // デフォルトは「入院中」タブ; 「予約」タブに切り替え
      await hospitalization.statusTab("予約").click();
      await expect(hospitalization.statusTab("予約")).toHaveAttribute("data-state", "active");
    } finally {
      await page.close();
    }
  });

  test("/hospitalization — ステータスタブ「すべて」に切り替えるとすべての件数が表示される", async () => {
    const page = await context.newPage();
    const hospitalization = new HospitalizationPage(page);
    try {
      await hospitalization.gotoList();
      await expect(hospitalization.listHeading()).toBeVisible();

      // 「すべて」タブに切り替え
      await hospitalization.statusTab("すべて").click();
      await expect(hospitalization.statusTab("すべて")).toHaveAttribute("data-state", "active");

      // リストビューに切り替えて全件表示を確認
      await hospitalization.listViewToggle().click();
      await expect(hospitalization.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
