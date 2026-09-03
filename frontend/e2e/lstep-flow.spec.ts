import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { LstepPage } from "./pages/lstep-page";

// E2E flow tests for L-step integration pages:
// /lstep/checkup-sync, /lstep/delivery-monitor, /lstep/analytics
// Covers: page load, basic navigation, interaction with filters/selectors.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe("Lステップ連携 フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/lstep/checkup-sync — 健診リマインダー抽出ページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoCheckupSync();
      await expect(lstep.checkupSyncHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(lstep.searchTargetsButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/lstep/checkup-sync — ページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoCheckupSync();
      await expect(page).toHaveURL(/\/lstep\/checkup-sync/);
      await expect(lstep.checkupTypeSelect()).toBeVisible({ timeout: 10000 });
      await expect(lstep.ltvAmountInput()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/lstep/delivery-monitor — 自動配信トリガー監視ページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoDeliveryMonitor();
      await expect(lstep.deliveryMonitorHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(lstep.refreshButton()).toBeVisible({ timeout: 10000 });
      await expect(lstep.filterFrom()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/lstep/delivery-monitor — ページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoDeliveryMonitor();
      await expect(page).toHaveURL(/\/lstep\/delivery-monitor/);
      await expect(lstep.filterTriggerType()).toBeVisible({ timeout: 10000 });
      await expect(lstep.filterStatus()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/lstep/analytics — Lステップ分析レポートページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoAnalytics();
      await expect(lstep.analyticsHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(lstep.monthlyStatsText()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/lstep/analytics — ページが表示される", async () => {
    const page = await context.newPage();
    const lstep = new LstepPage(page);
    try {
      await lstep.gotoAnalytics();
      await expect(page).toHaveURL(/\/lstep\/analytics/);
      await expect(lstep.visitRateText()).toBeVisible({ timeout: 10000 });
      await expect(lstep.csvImportHeading()).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});
