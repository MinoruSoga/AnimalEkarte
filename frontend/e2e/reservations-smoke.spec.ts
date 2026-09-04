import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { loginAsDemoAdmin } from "./helpers/auth";

// The filterCalendarAppointments function (cancelled → hidden, no_show → visible) is
// unit-tested in:
//   frontend/src/features/reservations/routes/__tests__/reservation-management.filter.test.ts
//
// E2E seed is time-dependent (calendar shows the current week), so seeding appointments
// for a stable date is impractical. This file covers:
//   1. auth guard (unauthenticated → /login)
//   2. page smoke (calendar navigation renders)
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe("予約管理 smoke E2E", () => {
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

  test("未ログイン時は /reservations にアクセスすると /login にリダイレクトされる", async ({
    browser,
  }) => {
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    await freshPage.goto("/reservations", { waitUntil: "domcontentloaded" });
    await freshPage.waitForURL(/\/login/, { timeout: 30000 });
    await freshContext.close();
  });

  test("予約管理 (/reservations) が表示される — カレンダーナビゲーション確認", async () => {
    const page = await context.newPage();
    try {
      await page.goto("/reservations", { waitUntil: "domcontentloaded" });
      await expect(page).toHaveURL(/\/reservations/, { timeout: 10000 });
      // ReservationManagementCalendar always renders a "今日" button in the header
      await expect(page.getByRole("button", { name: "今日" })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
