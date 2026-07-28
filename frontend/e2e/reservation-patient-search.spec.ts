import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import {
  OUTSIDE_FIRST_PAGE_PET,
  readRuntimePetReferences,
} from './helpers/pet-search-regression';
import { ReservationFormPage } from './pages/reservation-form-page';

// PatientSelectionTable delegates its unified `#search` field to GET /v1/pets.
// Runtime pet 1003298 (SPANKY) is intentionally outside the unfiltered first
// page. Its pet-name search term is unchanged by kana normalization, so
// finding the exact pet pair proves this is server-side search rather than
// filtering the 20 rows already loaded in the browser.
//
// Execution: ./scripts/run-e2e.sh e2e/reservation-patient-search.spec.ts
//
// Design: open reservation modal via /reservations → "新規予約登録" button.
// PatientSelectionTable renders by default (ownerMode=existing).
// All locators are scoped to the dialog to avoid calendar-DOM collision.
// Input is debounced and searches automatically; there is no manual search button.

test.describe('予約新規作成 患者選択 サーバー検索', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  async function openPatientSelectionTable(context: BrowserContext) {
    const page = await context.newPage();
    const reservation = new ReservationFormPage(page);
    await reservation.gotoReservations();
    await expect(reservation.todayButton()).toBeVisible({ timeout: 15000 });

    await reservation.newReservationButton().click();

    // Wait for the modal dialog — DialogTitle "新規予約作成"
    await expect(reservation.dialog()).toBeVisible({ timeout: 10000 });
    await expect(reservation.dialogTitle()).toBeVisible({ timeout: 5000 });

    // PatientSelectionTable renders by default (ownerMode=existing, lg desktop shows search panel)
    await expect(reservation.patientSearchLabel()).toBeVisible({ timeout: 5000 });

    return { page, reservation };
  }

  test('先頭20件にいない患者1003298を自動検索しPatientSelectionTableで選択できる', async () => {
    const { page, reservation } = await openPatientSelectionTable(context);
    try {
      const firstPageResponse = await page.request.get('/api/v1/pets?page=1&limit=20', {
        headers: {
          Accept: 'application/json',
          'X-Requested-With': 'XMLHttpRequest',
        },
      });
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
          url.searchParams.get('search') === OUTSIDE_FIRST_PAGE_PET.name &&
          !url.searchParams.has('include_deceased')
        );
      });

      await reservation.patientSearchInput().fill(OUTSIDE_FIRST_PAGE_PET.name);

      const searchResponse = await searchResponsePromise;
      expect(searchResponse.status()).toBe(200);
      const searchPayload: unknown = await searchResponse.json();
      expect(readRuntimePetReferences(searchPayload)).toContainEqual(
        OUTSIDE_FIRST_PAGE_PET,
      );

      await expect(
        reservation.patientRow(OUTSIDE_FIRST_PAGE_PET.name),
      ).toBeVisible();
      const selectButton = reservation.selectPatientButton(
        OUTSIDE_FIRST_PAGE_PET.id,
        OUTSIDE_FIRST_PAGE_PET.name,
      );
      await expect(selectButton).toBeVisible();
      await expect(selectButton).toBeEnabled();

      await selectButton.click();
      await expect(
        reservation.selectedPatientButton(
          OUTSIDE_FIRST_PAGE_PET.id,
          OUTSIDE_FIRST_PAGE_PET.name,
        ),
      ).toBeVisible();
    } finally {
      await page.close();
    }
  });
});
