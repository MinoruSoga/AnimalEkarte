import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { MedicalRecordsPage } from "./pages/medical-records-page";
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

  // EMR-60 受入4: fill 直後の DOM・page error・console・trace を採取する。
  // trace: "on" はこの describe だけに限定（config の on-first-retry を上書き）。
  test.describe("採証: 検索フィルタ", () => {
    test.use({ trace: "on" });

    test("/vaccinations — 検索フィルタが機能する", async () => {
      const page = await context.newPage();
      const vaccinations = new VaccinationsPage(page);
      const consoleEntries: string[] = [];
      const pageErrors: string[] = [];
      page.on("console", (message) => consoleEntries.push(`${message.type()}: ${message.text()}`));
      page.on("pageerror", (error) => pageErrors.push(error.message));
      try {
        await vaccinations.gotoList();
        await expect(vaccinations.listHeading()).toBeVisible();

        // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
        await page.getByLabel("検索").click();
        const searchInput = vaccinations.searchInput();
        await expect(searchInput).toBeVisible();

        // search は deferred 経由で GET /v1/vaccinations?search=… を発行する。
        // fill より先に waitForResponse を登録して取りこぼさない。
        const filteredResponsePromise = page.waitForResponse((response) => {
          const url = new URL(response.url());
          return (
            response.request().method() === "GET" &&
            url.pathname.endsWith("/api/v1/vaccinations") &&
            url.searchParams.get("search") === fixture.ownerSearch
          );
        });
        await searchInput.fill(fixture.ownerSearch);
        await expect(searchInput).toHaveValue(fixture.ownerSearch, { timeout: 10000 });
        // fill 直後（filtered response 到達前）の一覧 DOM を証跡として採取する。
        // eslint-disable-next-line no-restricted-properties -- Playwright の innerHTML() は読み取り専用の証跡採取（代入ではない）
        const domAfterFill = await vaccinations
          .tableBody()
          .innerHTML()
          .catch(() => "<tbody unavailable>");
        await test.info().attach("vaccination-filter-dom-after-fill.html", {
          body: domAfterFill,
          contentType: "text/html",
        });

        const filteredResponse = await filteredResponsePromise;
        expect(filteredResponse.status()).toBe(200);
        // サーバー検索結果が描画されるまで待ってから行を検証する（count() スナップショットは禁止）。
        await expect(vaccinations.ownerText(fixture.ownerName)).toBeVisible({ timeout: 15000 });
        await expect(vaccinations.detailLinkForPet(fixture.petName)).toBeVisible();
        await expect(vaccinations.detailLinkForPet(fixture.outsideFirstPagePet.name)).toBeVisible();
        expect(pageErrors).toEqual([]);
      } finally {
        await test.info().attach("vaccination-filter-console.txt", {
          body: consoleEntries.join("\n") || "(no console entries)",
          contentType: "text/plain",
        });
        await test.info().attach("vaccination-filter-pageerrors.txt", {
          body: pageErrors.join("\n") || "(no page errors)",
          contentType: "text/plain",
        });
        await page.close();
      }
    });
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

  test("/vaccinations — 詳細リンクがmedicalRecordId紐付けに応じてカルテタブ/standalone詳細へ遷移する", async () => {
    const page = await context.newPage();
    const vaccinations = new VaccinationsPage(page);
    const medicalRecords = new MedicalRecordsPage(page);
    try {
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible();
      await expect(vaccinations.firstRow()).toBeVisible({ timeout: 15000 });

      // medicalRecordId 紐付き行 → カルテの予防接種タブへ遷移する。
      const chartLink = vaccinations.detailLinkForPet(fixture.petName);
      await expect(chartLink).toBeVisible({ timeout: 15000 });
      await chartLink.click();
      await expect(page).toHaveURL(/\/medical-records\/\d+\?/);
      const chartUrl = new URL(page.url());
      expect(chartUrl.searchParams.get("tab")).toBe("予防接種");
      expect(chartUrl.searchParams.get("vaccinationId")).toMatch(/^\d+$/);
      await expect(medicalRecords.editHeading()).toBeVisible({ timeout: 15000 });
      await expect(vaccinations.tab("予防接種")).toHaveAttribute("aria-selected", "true");

      // medicalRecordId なし行 → standalone 詳細フォームへ遷移する。
      await vaccinations.gotoList();
      await expect(vaccinations.listHeading()).toBeVisible({ timeout: 15000 });
      const standaloneLink = vaccinations.detailLinkForPet(fixture.outsideFirstPagePet.name);
      await expect(standaloneLink).toBeVisible({ timeout: 15000 });
      await standaloneLink.click();
      await expect(vaccinations.detailHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/vaccinations\/\d+/);
    } finally {
      await page.close();
    }
  });
});
