import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { MedicalRecordsPage } from './pages/medical-records-page';

// E2E test for medical record creation form initialization.
// Seed data: pet 1 "Iris(イリス)" exists in seed.
// admin@noavet.jp is system_admin with full access.
//
// Design: form load assertions only (no save or write).
// Tests the direct URL navigation to new record form and form field visibility.
// Pet selection button and full navigation flow are covered by clinical-flows.spec.ts.
// Actual save is scoped out to avoid DB pollution.

test.describe('カルテ新規作成フォーム E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/medical-records/new?petId=1 — カルテ入力フォームが表示される (seed: pet=Iris)', async () => {
    const page = await context.newPage();
    const medical = new MedicalRecordsPage(page);
    try {
      // Navigate directly to new record form with pet_id=1 (Iris from seed)
      await medical.gotoNew('?petId=1');
      // Wait for network idle to ensure all lazy-loaded content is ready
      await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => null);

      // Verify the form loaded by checking for critical elements
      // Use more resilient selectors that wait longer for the async component to render
      await expect(medical.heading(undefined, 1)).toBeVisible({
        timeout: 15000,
      });

      await expect(medical.irisText()).toBeVisible({ timeout: 10000 });
      await expect(medical.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
