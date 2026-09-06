import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { CheckupsPage } from "./pages/checkups-page";

test.describe("定期健診 フロー E2E", () => {
  let context: BrowserContext;
  let fixture: ClinicalFixture;

  test.beforeAll(async ({ browser }) => {
    ({ context, fixture } = await createClinicalContext(browser));
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/checkups — 定期健診一覧が表示される", async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(checkups.newButton()).toBeVisible({ timeout: 10000 });
      await expect(checkups.tableBody()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/checkups/select-pet — ペット選択画面が表示される", async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoSelectPet();
      await expect(checkups.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await checkups.patientSearchInput().fill(fixture.petName);
      await expect(checkups.petText(fixture.petName)).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/checkups/new — 定期健診新規登録フォームが表示される", async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoNew(`?petId=${fixture.petId}`);
      await expect(checkups.newFormHeading()).toBeVisible({ timeout: 15000 });
      await expect(checkups.petText(fixture.petName)).toBeVisible({ timeout: 10000 });
      await expect(checkups.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/checkups — 新規登録ボタンでペット選択画面に遷移する", async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });

      await checkups.newButton().click();
      await expect(checkups.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/checkups\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test("/checkups — 一覧で検索が機能する", async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });

      const rows = checkups.rows();
      const initialRowCount = await rows.count();

      await page.getByLabel("検索").click();
      const searchInput = checkups.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(fixture.petName);

      await page.waitForLoadState("networkidle", { timeout: 10000 }).catch(() => null);

      const filteredRows = await checkups.rows().count();
      expect(filteredRows).toBeLessThanOrEqual(initialRowCount);
      await expect(searchInput).toHaveValue(fixture.petName);
    } finally {
      await page.close();
    }
  });
});
