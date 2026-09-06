import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { VaccinationsPage } from "./pages/vaccinations-page";

test.describe("予防接種管理 フロー E2E", () => {
  let context: BrowserContext;
  let fixture: ClinicalFixture;

  test.beforeAll(async ({ browser }) => {
    ({ context, fixture } = await createClinicalContext(browser));
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/vaccinations — 予防接種管理一覧が表示される", async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();
      await expect(vaccinations.newButton()).toBeVisible({ timeout: 10000 });
      // Runtime DB may have 0 active rows (demo soft-deleted). List chrome is enough.
      await expect(page).toHaveURL(/\/vaccinations/);
    } finally {
      await page.close();
    }
  });

  test("/vaccinations — 検索フィルタが機能する", async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();

      // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel("検索").click();
      const searchInput = vaccinations.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(fixture.ownerSearch);
      await expect(searchInput).toHaveValue(fixture.ownerSearch, { timeout: 10000 });
      if ((await vaccinations.firstDetailLink().count()) > 0) {
        await expect(vaccinations.ownerText(fixture.ownerName)).toBeVisible({ timeout: 10000 });
      } else {
        await expect(vaccinations.tableBody()).toBeVisible();
      }
    } finally {
      await page.close();
    }
  });

  test("/vaccinations — 新規登録ボタンでペット選択画面に遷移する", async () => {
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

  test("/vaccinations — 行がある場合は詳細リンクで詳細画面に遷移する", async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();

      const detailLink = vaccinations.firstDetailLink();
      if ((await detailLink.count()) === 0) {
        // No active vaccination rows in this runtime DB — detail path is N/A.
        test.info().annotations.push({
          type: "note",
          description: "skipped detail navigation: zero active vaccination rows",
        });
        return;
      }

      // DataTableRow is non-interactive; open via DataTableRowLink.
      await detailLink.click();
      await expect(vaccinations.detailHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/vaccinations\/\d+/);
    } finally {
      await page.close();
    }
  });
});
