import os from 'node:os';
import { test, expect } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Note: arm64/aarch64 runtime (e.g., Docker on Apple Silicon) will skip these tests
// because Playwright Chromium does not run reliably in that environment.
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

const isArm64Runtime = process.platform === 'linux' && os.arch() === 'arm64';
const TREATMENT_ITEMS_PATH = '/settings/treatment-items';
const parentCategoryCombobox = (page: Page) =>
  page.getByRole('combobox').filter({ hasText: 'なし（ルート）' });

async function gotoTreatmentItems(page: Page, tab: string) {
  await page.goto(`${TREATMENT_ITEMS_PATH}?tab=${tab}`, { waitUntil: 'domcontentloaded' });
  await expect(page).toHaveURL(new RegExp(`/settings/treatment-items\\?tab=${tab}`));
  // Lazy chunk for settings may take 15-20 s through Docker — rely on expect.timeout (15 s).
  await expect(page.getByRole('heading', { name: '治療プランマスタ' })).toBeVisible();
}

test.describe('Master CRUD E2E Tests', () => {
  let context: BrowserContext;

  // arm64 環境では describe 全体をスキップする
  if (isArm64Runtime) {
    test.describe.configure({ skip: true });
  }

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('A: Chief Complaint Navigation', async () => {
    const page = await context.newPage();
    try {
      // Navigate to settings page
      await page.goto('/settings', { waitUntil: 'domcontentloaded' });

      // Check if Chief Complaint card is visible
      const chiefComplaintCard = page.locator('text=主訴カテゴリ').first();
      await expect(chiefComplaintCard).toBeVisible();

      // Click and navigate to chief complaint page
      await chiefComplaintCard.click();
      await page.waitForURL('**/chief-complaint');
      await expect(page).toHaveURL(/\/settings\/interview\/chief-complaint/);

      // Verify page content loaded
      await expect(page.locator('body')).toContainText('主訴カテゴリ');
    } finally {
      await page.close();
    }
  });

  test('B: Treatment Plan - Procedure Tab and Parent Selector', async () => {
    const page = await context.newPage();
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(page, 'procedure');

      // Click on 処置 (Procedure) tab
      const procedureTab = page.getByRole('tab', { name: '処置' });
      await expect(procedureTab).toBeVisible();
      await expect(procedureTab).toHaveAttribute('data-state', 'active');

      // Look for parent category selector in the form/panel
      // The selector should be visible when new item creation panel opens
      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.getByText('親カテゴリ')).toBeVisible();

      // Parent selector should be visible or selectable
      await expect(parentCategoryCombobox(page)).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('C: Treatment Plan - Root with Children Cannot Change Parent', async () => {
    const page = await context.newPage();
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(page, 'procedure');

      // Click on 処置 (Procedure) tab
      const procedureTab = page.getByRole('tab', { name: '処置' });
      await expect(procedureTab).toHaveAttribute('data-state', 'active');

      // Seed data has "注射" as a root category with children.
      const rootItem = page.locator('tbody').getByText('注射', { exact: true }).first();
      await expect(rootItem).toBeVisible();
      await rootItem.click();

      // Check for the "cannot change parent" message
      await expect(page.getByText('子項目があるため変更できません')).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('D: All 5 Tabs Exist and Show Parent Selector', async () => {
    const page = await context.newPage();
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(page, 'consultation');

      const tabs = ['診察', '検査', '予防接種', '処置', '定期健診'];

      for (const tabName of tabs) {
        const tab = page.getByRole('tab', { name: tabName });
        await expect(tab).toBeVisible(`Tab ${tabName} should be visible`);

        await tab.click();
        await expect(tab).toHaveAttribute('data-state', 'active');

        // For each tab, the parent selector should be accessible in creation panel
        await page.getByRole('button', { name: '新規登録' }).click();
        await expect(page.getByText('親カテゴリ')).toBeVisible();
        await expect(parentCategoryCombobox(page)).toBeVisible();
        await page.getByRole('button', { name: '閉じる' }).click();
      }
    } finally {
      await page.close();
    }
  });
});
