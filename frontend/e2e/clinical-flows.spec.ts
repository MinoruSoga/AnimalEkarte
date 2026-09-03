import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { MedicalRecordsPage } from "./pages/medical-records-page";

// E2E flow tests for clinical (medical records) pages.
// Demo seed: owner 「林\u3000文明」(ideographic space U+3000), pet Iris id=1000099.
// admin@noavet.jp is system_admin with full access.

test.describe("カルテ フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/medical-records — カルテ管理一覧が表示される", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();
      // テーブル行が1件以上存在する
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/medical-records — 検索で「林」を入力すると林 文明のカルテが表示される", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();

      // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel("検索").click();
      const searchInput = medical.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill("林");
      await expect(medical.hayashiText()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/medical-records — 行クリックでカルテ編集画面に遷移する", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });

      // DataTableRow is non-interactive; navigation is via DataTableRowLink only.
      await medical.firstDetailLink().click();
      await expect(medical.editHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/medical-records\/\d+/);
    } finally {
      await page.close();
    }
  });

  test("/medical-records/select-pet — ペット選択画面が表示される", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoSelectPet();
      await expect(medical.selectPetHeading()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/medical-records — 新規カルテ登録ボタンでペット選択画面に遷移する", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();

      await medical.newButton().click();
      await expect(medical.selectPetHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/medical-records\/select-pet/);
    } finally {
      await page.close();
    }
  });
});
