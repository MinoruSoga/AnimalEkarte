import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import { MedicalRecordsPage } from "./pages/medical-records-page";

test.describe("カルテ一覧 ページネーション/ソート/フィルタ E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    ({ context } = await createClinicalContext(browser));
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/medical-records — ページ2ボタンをクリックすると ?page=2 に遷移する", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });

      const page2 = medical.pageButton(2);
      await expect(page2).toBeVisible({ timeout: 15000 });
      await page2.click();

      await expect(page).toHaveURL(/[?&]page=2\b/);
    } finally {
      await page.close();
    }
  });

  test("/medical-records — 列ヘッダクリックで sort/order が server 側に反映される（URL 状態）", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });

      // 1回目クリック: 既定 desc でソート開始
      await medical.sortHeader("飼主名").click();
      await expect(page).toHaveURL(/sort=owner_name/);
      await expect(page).toHaveURL(/order=desc/);
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });

      // 2回目クリック（同一列）: asc へトグル
      await medical.sortHeader("飼主名").click();
      await expect(page).toHaveURL(/order=asc/);
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/medical-records — ステータスフィルタ（確定済）を適用すると一覧が絞り込まれる", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });

      await medical.filterAddButton().click();
      await page.getByRole("option", { name: "ステータス" }).click();
      await page.getByRole("button", { name: "次と一致" }).click();
      await page.getByRole("option", { name: "確定済", exact: true }).click();

      // フィルタ適用後、削除ピル（アクティブフィルタの視覚的証跡）が表示される
      await expect(medical.filterRemoveButton("ステータス")).toBeVisible({ timeout: 15000 });
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
