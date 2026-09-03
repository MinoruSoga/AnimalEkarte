import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { ShiftsPage } from "./pages/shifts-page";

// E2E flow tests for shifts (/shifts) calendar page.
// Covers: page load, calendar navigation, basic interaction.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe("シフト管理 フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/shifts — シフト管理カレンダーが表示される", async () => {
    const page = await context.newPage();
    const shifts = new ShiftsPage(page);
    try {
      await shifts.gotoCalendar();
      await expect(shifts.calendarHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(shifts.prevMonthButton()).toBeVisible({ timeout: 10000 });
      await expect(shifts.nextMonthButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/shifts — カレンダーナビゲーションが存在する", async () => {
    const page = await context.newPage();
    const shifts = new ShiftsPage(page);
    try {
      await shifts.gotoCalendar();
      await expect(shifts.calendarHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(shifts.monthLabel()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/shifts — スタッフセレクタが表示される", async () => {
    const page = await context.newPage();
    const shifts = new ShiftsPage(page);
    try {
      await shifts.gotoCalendar();
      await expect(shifts.calendarHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(shifts.firstCombobox()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/shifts — カレンダーを前月に移動できる", async () => {
    const page = await context.newPage();
    const shifts = new ShiftsPage(page);
    try {
      await shifts.gotoCalendar();
      await expect(shifts.calendarHeading()).toBeVisible({
        timeout: 15000,
      });
      const monthDisplay = shifts.monthLabel();
      const currentText = await monthDisplay.textContent();

      await shifts.prevMonthButton().click();

      await expect(monthDisplay).not.toHaveText(currentText ?? "", { timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("/shifts — スタッフフィルタが機能する", async () => {
    const page = await context.newPage();
    const shifts = new ShiftsPage(page);
    try {
      await shifts.gotoCalendar();
      await expect(shifts.calendarHeading()).toBeVisible({
        timeout: 15000,
      });

      await shifts.firstCombobox().click();
      await expect(shifts.firstOption()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
