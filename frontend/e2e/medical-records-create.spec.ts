import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

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
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/medical-records/new?petId=1 — カルテ入力フォームが表示される (seed: pet=Iris)', async () => {
    const page = await context.newPage();
    try {
      // Navigate directly to new record form with pet_id=1 (Iris from seed)
      await page.goto('/medical-records/new?petId=1', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'カルテ入力' })).toBeVisible({
        timeout: 15000,
      });

      await expect(page.getByText('Iris').first()).toBeVisible({ timeout: 10000 });
      const saveButton = page.getByRole('button', { name: '保存' });
      await expect(saveButton).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
