import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { ExaminationsPage } from "./pages/examinations-page";
import { MedicalRecordsPage } from "./pages/medical-records-page";

test.describe("検査管理 フロー E2E", () => {
  let context: BrowserContext;
  let fixture: ClinicalFixture;

  test.beforeAll(async ({ browser }) => {
    ({ context, fixture } = await createClinicalContext(browser));
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/examinations — 検査管理一覧が表示される", async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoList();
      await expect(examinations.listHeadingPattern()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/examinations/);
    } finally {
      await page.close();
    }
  });

  test("/examinations/select-pet — ペット選択画面が表示される", async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoSelectPet();
      await expect(examinations.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await examinations.patientSearchInput().fill(fixture.petName);
      await expect(examinations.petText(fixture.petName)).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test("/examinations — 詳細リンクがmedicalRecordId紐付けに応じてカルテタブ/standalone詳細へ遷移する", async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    const medicalRecords = new MedicalRecordsPage(page);
    try {
      await examinations.gotoList();
      await expect(examinations.listHeading()).toBeVisible({ timeout: 15000 });
      await expect(examinations.firstRow()).toBeVisible({ timeout: 15000 });

      // 閲覧権限あり + medicalRecordId 紐付き行 → カルテの検査タブへ遷移する。
      const chartLink = examinations.chartTabDetailLink(fixture.petName);
      await expect(chartLink).toBeVisible({ timeout: 15000 });
      await chartLink.click();
      await expect(page).toHaveURL(/\/medical-records\/\d+\?/);
      const chartUrl = new URL(page.url());
      expect(chartUrl.searchParams.get("tab")).toBe("検査");
      expect(chartUrl.searchParams.get("examId")).toMatch(/^\d+$/);
      await expect(medicalRecords.editHeading()).toBeVisible({ timeout: 15000 });
      await expect(examinations.tab("検査")).toHaveAttribute("aria-selected", "true");

      // medicalRecordId なし行 → standalone 詳細フォームへ遷移する。
      await examinations.gotoList();
      await expect(examinations.listHeading()).toBeVisible({ timeout: 15000 });
      const standaloneLink = examinations.standaloneDetailLink(fixture.outsideFirstPagePet.name);
      await expect(standaloneLink).toBeVisible({ timeout: 15000 });
      await standaloneLink.click();
      await expect(examinations.detailHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/examinations\/\d+/);
      await expect(examinations.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/examinations — 検査一覧で検索が機能する", async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoList();
      await expect(examinations.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(examinations.firstRow()).toBeVisible({ timeout: 15000 });
      const firstTestType = (await examinations.firstRowTestTypeCell().textContent())?.trim();
      expect(firstTestType).toBeTruthy();

      await page.getByLabel("検索").click();
      const searchInput = examinations.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(firstTestType ?? "");
      await expect(examinations.firstRow()).toContainText(firstTestType ?? "", { timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
