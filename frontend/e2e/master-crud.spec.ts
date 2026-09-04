import os from "node:os";
import { test, expect } from "@playwright/test";
import type { BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { TreatmentItemsPage } from "./pages/treatment-items-page";

// Note: arm64/aarch64 runtime (e.g., Docker on Apple Silicon) will skip these tests
// because Playwright Chromium does not run reliably in that environment.
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

const isArm64Runtime = process.platform === "linux" && os.arch() === "arm64";

async function gotoTreatmentItems(treatment: TreatmentItemsPage, page: Page, tab: string) {
  await treatment.gotoTab(tab);
  await expect(page).toHaveURL(treatment.tabUrlPattern(tab));
  // Lazy chunk for settings may take 15-20 s through Docker — rely on expect.timeout (15 s).
  await expect(treatment.planMasterHeading()).toBeVisible();
}

test.describe("Master CRUD E2E Tests", () => {
  let context: BrowserContext;

  test.skip(isArm64Runtime, "linux arm64");

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("A: Chief Complaint Navigation", async () => {
    const page = await context.newPage();
    const treatment = new TreatmentItemsPage(page);
    try {
      // Navigate to settings page
      await treatment.gotoSettings();

      // Check if Chief Complaint card is visible
      const chiefComplaintCard = treatment.chiefComplaintCard();
      await expect(chiefComplaintCard).toBeVisible();

      // Click and navigate to chief complaint page
      await chiefComplaintCard.click();
      await page.waitForURL("**/chief-complaint");
      await expect(page).toHaveURL(/\/settings\/interview\/chief-complaint/);

      // Verify page content loaded
      await expect(treatment.body()).toContainText("主訴カテゴリ");
    } finally {
      await page.close();
    }
  });

  test("B: Treatment Plan - Procedure Tab and Parent Selector", async () => {
    const page = await context.newPage();
    const treatment = new TreatmentItemsPage(page);
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(treatment, page, "procedure");

      // Click on 処置 (Procedure) tab
      const procedureTab = treatment.tab("処置");
      await expect(procedureTab).toBeVisible();
      await expect(procedureTab).toHaveAttribute("data-state", "active");

      // Look for parent category selector in the form/panel
      // The selector should be visible when new item creation panel opens
      await treatment.newButton().click();
      await expect(treatment.parentCategoryLabel()).toBeVisible();

      // Parent selector should be visible or selectable
      await expect(treatment.parentCategoryCombobox()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("C: Treatment Plan - Root with Children Cannot Change Parent", async () => {
    const page = await context.newPage();
    const treatment = new TreatmentItemsPage(page);
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(treatment, page, "procedure");

      // Click on 処置 (Procedure) tab
      const procedureTab = treatment.tab("処置");
      await expect(procedureTab).toHaveAttribute("data-state", "active");

      // Seed data has "注射" as a root category with children.
      const rootItem = treatment.injectionRootItem();
      await expect(rootItem).toBeVisible();
      await rootItem.click();

      // Check for the "cannot change parent" message
      await expect(treatment.cannotChangeParentText()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("D: All 5 Tabs Exist and Show Parent Selector", async () => {
    const page = await context.newPage();
    const treatment = new TreatmentItemsPage(page);
    try {
      // Navigate to treatment plan page
      await gotoTreatmentItems(treatment, page, "consultation");

      const tabs = ["診察", "検査", "予防接種", "処置", "定期健診"];

      for (const tabName of tabs) {
        const tab = treatment.tab(tabName);
        await expect(tab, `Tab ${tabName} should be visible`).toBeVisible();

        await tab.click();
        await expect(tab).toHaveAttribute("data-state", "active");

        // For each tab, the parent selector should be accessible in creation panel
        await treatment.newButton().click();
        await expect(treatment.parentCategoryLabel()).toBeVisible();
        await expect(treatment.parentCategoryCombobox()).toBeVisible();
        await treatment.closeButton().click();
      }
    } finally {
      await page.close();
    }
  });
});
