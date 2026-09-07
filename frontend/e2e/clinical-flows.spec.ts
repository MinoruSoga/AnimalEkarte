import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { MedicalRecordsPage } from "./pages/medical-records-page";

test.describe("カルテ フロー E2E", () => {
  let context: BrowserContext;
  let fixture: ClinicalFixture;

  test.beforeAll(async ({ browser }) => {
    ({ context, fixture } = await createClinicalContext(browser));
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
      await expect(medical.firstRow()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/medical-records — 検索で合成飼主のカルテが表示される", async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      await medical.gotoList();
      await expect(medical.listHeading()).toBeVisible();

      await page.getByLabel("検索").click();
      const searchInput = medical.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(fixture.ownerSearch);
      await expect(medical.ownerText(fixture.ownerName)).toBeVisible({ timeout: 10000 });
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
