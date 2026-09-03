import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { AccountingPage } from "./pages/accounting-page";

// E2E flow tests for accounting (/accounting) navigation.
// Smoke coverage (tab visibility, kana search) lives in accounting-smoke.spec.ts.
// This file adds detail navigation: list link → AccountingDetail page.
// Row body click does not navigate (only DataTableRowLink on the date cell).
// Iris has no billing rows in 003_demo — use DEMO_ACCOUNTING_KANA_PET instead.
//
// Design: fresh page per test within shared context.

test.describe("会計フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/accounting — 会計一覧の詳細リンクで会計精算画面に遷移する", async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      // AccountingList mounts after API resolves (may take 30 s with Docker lag)
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await expect(accounting.firstDetailLink()).toBeVisible({ timeout: 15000 });

      await accounting.firstDetailLink().click();
      // AccountingDetail renders PageLayout with title="会計精算" only after API resolves
      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test("/accounting — かな検索したペット行の詳細リンクで会計精算画面に遷移する", async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });

      const searchToggle = accounting.searchToggle();
      await expect(searchToggle).toBeVisible({ timeout: 15000 });
      await searchToggle.click();
      const searchInput = accounting.searchInput();
      await expect(searchInput).toBeVisible({ timeout: 5000 });
      // page-scoped client filter — DEMO_ACCOUNTING_KANA_PET is on page 1
      await searchInput.fill("さき");

      await expect(accounting.kanaPetCell()).toBeVisible({ timeout: 10000 });
      await accounting.kanaPetDetailLink().click();

      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);
    } finally {
      await page.close();
    }
  });

  test("/accounting/reports — 月次集計レポートにセレクタが表示される", async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoReports({ waitUntil: "domcontentloaded" });
      await expect(accounting.reportsHeading()).toBeVisible({
        timeout: 15000,
      });
      // Reports page always renders year/month selectors
      await expect(accounting.firstCombobox()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/accounting/:id — 会計精算フォームの確定ボタンが表示される", async () => {
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await accounting.gotoList();
      await expect(accounting.listTab()).toBeVisible({ timeout: 30000 });
      await expect(accounting.firstDetailLink()).toBeVisible({ timeout: 15000 });

      await accounting.firstDetailLink().click();
      await expect(accounting.detailHeading()).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/accounting\/\d+/);

      // pending seed rows show 「会計を確定する」(may be disabled until payment complete)
      await expect(accounting.confirmButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
