import { expect, test, type BrowserContext } from '@playwright/test';

import { createAuthedContext } from './helpers/context';
import {
  OUTSIDE_FIRST_PAGE_PET,
  readRuntimePetReferences,
} from './helpers/pet-search-regression';
import { MedicalRecordsPage } from './pages/medical-records-page';

// Representative E2E for the shared usePetSelectionPage flow.
// The medical-record selector includes deceased pets in both its initial page
// and debounced server-side search. Selection is asserted but not activated:
// /medical-records/new performs auto-create work when mounted.
//
// Execution: ./scripts/run-e2e.sh e2e/medical-records-patient-search.spec.ts

test.describe('カルテ作成 ペット選択 サーバー検索', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('include_deceased=trueの先頭20件にいない患者1003298を検索し選択可能と表示する', async () => {
    const page = await context.newPage();
    const medicalRecords = new MedicalRecordsPage(page);

    try {
      const firstPageResponsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === 'GET' &&
          url.pathname.endsWith('/api/v1/pets') &&
          url.searchParams.get('page') === '1' &&
          url.searchParams.get('limit') === '20' &&
          url.searchParams.get('include_deceased') === 'true' &&
          !url.searchParams.has('search')
        );
      });

      await medicalRecords.gotoSelectPet();
      await expect(medicalRecords.selectPetHeading()).toBeVisible();

      const firstPageResponse = await firstPageResponsePromise;
      expect(firstPageResponse.status()).toBe(200);
      const firstPagePayload: unknown = await firstPageResponse.json();
      const firstPagePets = readRuntimePetReferences(firstPagePayload);
      expect(firstPagePets).toHaveLength(20);
      expect(firstPagePets).not.toContainEqual(OUTSIDE_FIRST_PAGE_PET);

      const searchResponsePromise = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === 'GET' &&
          url.pathname.endsWith('/api/v1/pets') &&
          url.searchParams.get('page') === '1' &&
          url.searchParams.get('limit') === '20' &&
          url.searchParams.get('include_deceased') === 'true' &&
          url.searchParams.get('search') === OUTSIDE_FIRST_PAGE_PET.name
        );
      });

      await medicalRecords.patientSearchInput().fill(
        OUTSIDE_FIRST_PAGE_PET.name,
      );

      const searchResponse = await searchResponsePromise;
      expect(searchResponse.status()).toBe(200);
      const searchPayload: unknown = await searchResponse.json();
      expect(readRuntimePetReferences(searchPayload)).toContainEqual(
        OUTSIDE_FIRST_PAGE_PET,
      );

      await expect(
        medicalRecords.patientRow(OUTSIDE_FIRST_PAGE_PET.name),
      ).toBeVisible();
      const selectButton = medicalRecords.selectPatientButton(
        OUTSIDE_FIRST_PAGE_PET.id,
        OUTSIDE_FIRST_PAGE_PET.name,
      );
      await expect(selectButton).toBeVisible();
      await expect(selectButton).toBeEnabled();
    } finally {
      await page.close();
    }
  });
});
