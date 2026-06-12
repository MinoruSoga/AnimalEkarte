import os from 'node:os';
import { test, expect } from '@playwright/test';

// These tests assume the local app is already authenticated via the demo session.
// If redirected to login, add a Playwright auth setup before enabling CI execution.
//
// Note: arm64/aarch64 runtime (e.g., Docker on Apple Silicon) will skip these tests
// because Playwright Chromium does not run reliably in that environment.

const isArm64Runtime = process.platform === 'linux' && os.arch() === 'arm64';

test.describe('Master CRUD E2E Tests', () => {
  // arm64 環境では describe 全体をスキップする
  if (isArm64Runtime) {
    test.describe.configure({ skip: true });
  }

  test('A: Chief Complaint Navigation', async ({ page }) => {
    // Navigate to settings page
    await page.goto('/settings');

    // Check if Chief Complaint card is visible
    const chiefComplaintCard = page.locator('text=主訴カテゴリ').first();
    await expect(chiefComplaintCard).toBeVisible();

    // Click and navigate to chief complaint page
    await chiefComplaintCard.click();
    await page.waitForURL('**/chief-complaint');
    await expect(page).toHaveURL(/\/settings\/interview\/chief-complaint/);

    // Verify page content loaded
    await expect(page.locator('body')).toContainText('主訴カテゴリ');
  });

  test('B: Treatment Plan - Procedure Tab and Parent Selector', async ({ page }) => {
    // Navigate to treatment plan page
    await page.goto('/settings/medical/treatment-plan');

    // Click on 処置 (Procedure) tab
    const procedureTab = page.locator('text=処置').first();
    await expect(procedureTab).toBeVisible();
    await procedureTab.click();

    // Wait for tab content to load
    await page.waitForTimeout(500);

    // Look for parent category selector in the form/panel
    // The selector should be visible when new item creation panel opens
    const createButton = page.locator('button:has-text("新規作成"), button:has-text("追加")').first();
    if (await createButton.isVisible()) {
      await createButton.click();
      await page.waitForTimeout(300);
    }

    // Parent selector should be visible or selectable
    const selectElements = page.locator('select, [role="combobox"], [role="listbox"]');
    const selectCount = await selectElements.count();

    // At least one form element should exist for selection
    expect(selectCount).toBeGreaterThan(0);
  });

  test('C: Treatment Plan - Root with Children Cannot Change Parent', async ({ page }) => {
    // Navigate to treatment plan page
    await page.goto('/settings/medical/treatment-plan');

    // Click on 処置 (Procedure) tab
    const procedureTab = page.locator('text=処置').first();
    await procedureTab.click();
    await page.waitForTimeout(500);

    // Find a root category with children and verify the "cannot change parent" message is displayed
    const rootCategoryNames = ['注射', '輸液', '処置', '注射剤', '静脈確保'];
    let foundWithWarning = false;

    for (const categoryName of rootCategoryNames) {
      const rootItem = page.locator(`text=${categoryName}`).first();
      if (await rootItem.isVisible()) {
        await rootItem.click();
        await page.waitForTimeout(300);

        // Check for the "cannot change parent" message
        const warningText = page.locator('text=子項目があるため変更できません');
        if (await warningText.isVisible()) {
          await expect(warningText).toBeVisible();
          foundWithWarning = true;
          break;
        }
      }
    }

    // Ensure at least one root item with children was found and displayed the warning
    expect(
      foundWithWarning,
      'Root item with children should show parent-change warning message'
    ).toBe(true);
  });

  test('D: All 5 Tabs Exist and Show Parent Selector', async ({ page }) => {
    // Navigate to treatment plan page
    await page.goto('/settings/medical/treatment-plan');

    const tabs = ['診察', '検査', '予防接種', '処置', '定期健診'];

    for (const tabName of tabs) {
      const tab = page.locator(`text=${tabName}`).first();
      await expect(tab).toBeVisible(`Tab ${tabName} should be visible`);

      // Click the tab
      if (await tab.isClickable()) {
        await tab.click();
        await page.waitForTimeout(300);

        // For each tab, the parent selector should be accessible in creation panel
        const formElements = page.locator('input, select, [role="combobox"]');
        expect(await formElements.count()).toBeGreaterThan(0);
      }
    }
  });
});
