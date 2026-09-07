import { expect, test, type BrowserContext } from "@playwright/test";

import { createClinicalContext } from "./helpers/clinical-suite";
import type { ClinicalFixture } from "./helpers/clinical-fixture";
import { readRuntimePetReferences } from "./helpers/pet-search-regression";
import { MedicalRecordsPage } from "./pages/medical-records-page";

test.describe("カルテ作成 ペット選択 サーバー検索", () => {
  let context: BrowserContext;
  let fixture: ClinicalFixture;

  test.beforeAll(async ({ browser }) => {
    ({ context, fixture } = await createClinicalContext(browser));
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("先頭20件にいない合成ペットを検索し選択可能と表示する", async () => {
    const page = await context.newPage();
    const medicalRecords = new MedicalRecordsPage(page);
    const target = fixture.outsideFirstPagePet;

    try {
      const firstPageResponsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === "GET" &&
          url.pathname.endsWith("/api/v1/pets") &&
          url.searchParams.get("page") === "1" &&
          url.searchParams.get("limit") === "20" &&
          url.searchParams.get("include_deceased") === "true" &&
          !url.searchParams.has("search")
        );
      });

      await medicalRecords.gotoSelectPet();
      await expect(medicalRecords.selectPetHeading()).toBeVisible();

      const firstPageResponse = await firstPageResponsePromise;
      expect(firstPageResponse.status()).toBe(200);
      const firstPagePayload: unknown = await firstPageResponse.json();
      const firstPagePets = readRuntimePetReferences(firstPagePayload);
      expect(firstPagePets).toHaveLength(20);
      expect(firstPagePets).not.toContainEqual(target);

      const searchResponsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === "GET" &&
          url.pathname.endsWith("/api/v1/pets") &&
          url.searchParams.get("page") === "1" &&
          url.searchParams.get("limit") === "20" &&
          url.searchParams.get("include_deceased") === "true" &&
          url.searchParams.get("search") === target.name
        );
      });

      await medicalRecords.patientSearchInput().fill(target.name);

      const searchResponse = await searchResponsePromise;
      expect(searchResponse.status()).toBe(200);
      const searchPayload: unknown = await searchResponse.json();
      expect(readRuntimePetReferences(searchPayload)).toContainEqual(target);

      await expect(medicalRecords.patientRow(target.name)).toBeVisible();
      const selectButton = medicalRecords.selectPatientButton(target.id, target.name);
      await expect(selectButton).toBeVisible();
      await expect(selectButton).toBeEnabled();
    } finally {
      await page.close();
    }
  });
});
