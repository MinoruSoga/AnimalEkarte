import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E for PatientSelectionTable kana-normalised search inside ReservationFormModal.
//
// Seed data (003_seed_demo.sql, /v1/pets page 1, limit=20):
//   pet id=4  name='タロウ'      name_kana='たろう'   owner_id=3 (鈴木 一郎)
//   pet id=5  name='ジロウ'      name_kana='じろう'   owner_id=3 (鈴木 一郎)
//   pet id=1  name='Iris(イリス)' name_kana='いりす'  owner_id=1 (林 文明)
//
// NOTE: /v1/pets returns the first 20 pets (paginated). pet id=80 (ピーター)
// is beyond page 1 and is intentionally excluded from these tests.
//
// Issue #161: カナ混同検索を未対応画面へ適用
// Surface: PatientSelectionTable (petName field) inside 新規予約作成 modal.
//
// Execution: ./scripts/run-e2e.sh e2e/reservation-patient-search.spec.ts
//
// Design: open reservation modal via /reservations → "新規予約登録" button.
// PatientSelectionTable renders by default (ownerMode=existing).
// All locators are scoped to the dialog to avoid calendar-DOM collision.
// Search triggers only after "検索" button click (hasSearched state gate).
// Results are filtered client-side via normalizedIncludes (katakana→hiragana).

test.describe('予約新規作成 患者選択 かな混同検索', () => {
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

  async function openPatientSelectionTable(context: BrowserContext) {
    const page = await context.newPage();
    await page.goto('/reservations', { waitUntil: 'domcontentloaded' });
    await expect(page.getByRole('button', { name: '今日' })).toBeVisible({ timeout: 15000 });

    await page.getByRole('button', { name: '新規予約登録' }).click();

    // Wait for the modal dialog — DialogTitle "新規予約作成"
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 10000 });
    await expect(dialog.getByText('新規予約作成')).toBeVisible({ timeout: 5000 });

    // PatientSelectionTable renders by default (ownerMode=existing, lg desktop shows search panel)
    await expect(dialog.getByText('患者検索')).toBeVisible({ timeout: 5000 });

    return page;
  }

  test('ひらがな「たろう」でペット名検索するとカタカナ「タロウ」が表示される', async () => {
    const page = await openPatientSelectionTable(context);
    try {
      const dialog = page.getByRole('dialog');

      // petName field — scoped to dialog to avoid any page-level ambiguity
      await dialog.locator('#petName').fill('たろう');

      // Trigger search via Enter key (onKeyDown handler in PatientSelectionTable)
      await dialog.locator('#petName').press('Enter');

      // normalizedIncludes('タロウ', 'たろう') → normalizeKana('タロウ')='たろう' includes 'たろう' → true
      await expect(dialog.getByText('タロウ')).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('カタカナ「タロウ」でペット名検索しても「タロウ」が表示される (かな統一検索)', async () => {
    const page = await openPatientSelectionTable(context);
    try {
      const dialog = page.getByRole('dialog');

      await dialog.locator('#petName').fill('タロウ');

      await dialog.getByRole('button', { name: '検索' }).click();

      // normalizedIncludes('タロウ', 'タロウ') → same after normalization → true
      await expect(dialog.getByText('タロウ')).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('ひらがな「いりす」でペット名検索するとカタカナ混在「Iris(イリス)」が表示される', async () => {
    const page = await openPatientSelectionTable(context);
    try {
      const dialog = page.getByRole('dialog');

      await dialog.locator('#petName').fill('いりす');

      await dialog.getByRole('button', { name: '検索' }).click();

      // normalizedIncludes('Iris(イリス)', 'いりす') → 'iris(いりす)' includes 'いりす' → true
      await expect(dialog.getByText('Iris(イリス)')).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
