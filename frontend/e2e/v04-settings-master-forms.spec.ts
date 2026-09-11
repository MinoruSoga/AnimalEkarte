import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { SettingsMasterPage } from "./pages/settings-master-page";

/**
 * V04 disposable master CRUD/DELETE browser coverage.
 *
 * Closes the UAT V04 UNKNOWN gap around unused chief-complaint DELETE success
 * and at least one create→update→delete cycle beyond animal-species.
 *
 * Conflict-on-in-use for chief-complaint requires an inquiry tied to a medical
 * record (PHI). That path is intentionally omitted here; payment-method system
 * DELETE denial covers a safe denied/conflict-style assertion without elevating
 * clinic 1/2 privileges or creating patient data.
 *
 * Credentials: E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD via helpers/auth.ts.
 */

function disposableName(kind: string): string {
  return `V04_${kind}_${Date.now()}`;
}

async function bestEffortDeleteRow(
  settings: SettingsMasterPage,
  path: string,
  heading: string,
  searchPlaceholder: string,
  name: string,
  confirmLabel: string | RegExp = "削除",
): Promise<void> {
  try {
    await settings.open(path);
    await expect(settings.heading(heading)).toBeVisible({ timeout: 15000 });
    await settings.searchFor(searchPlaceholder, name);
    const row = settings.rowContaining(name);
    if ((await row.count()) === 0) return;
    await settings.rowActionButton(name).click();
    await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
    await settings.deleteButton().click();
    await expect(settings.deleteDialog()).toBeVisible();
    await settings.deleteConfirmButton(confirmLabel).click();
    await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
  } catch {
    // Best-effort cleanup only — assertion failures belong to the test body.
  }
}

test.describe("V04 設定マスタ disposable CRUD/DELETE", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("動物種類: 新規作成 → 一覧 → 削除", async () => {
    test.setTimeout(90000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const speciesName = disposableName("種");
    let needsCleanup = false;

    try {
      await settings.open("/settings/animal-species");
      await expect(settings.heading("動物種類マスタ")).toBeVisible();

      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible();
      await settings.masterTitleInput().fill(speciesName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });
      needsCleanup = true;

      await settings.open("/settings/animal-species");
      await expect(settings.heading("動物種類マスタ")).toBeVisible();
      await settings.searchFor("動物種類名で検索...", speciesName);
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });

      await settings.rowActionButton(speciesName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.deleteButton().click();
      await expect(settings.deleteDialog()).toBeVisible();
      await settings.deleteConfirmButton("削除").click();
      await expect(page.getByText(speciesName)).not.toBeVisible({ timeout: 10000 });
      needsCleanup = false;
    } finally {
      if (needsCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/animal-species",
          "動物種類マスタ",
          "動物種類名で検索...",
          speciesName,
        );
      }
      await page.close();
    }
  });

  test("主訴マスタ: 未使用区分の作成 → DELETE 成功", async () => {
    test.setTimeout(90000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const complaintName = disposableName("主訴");
    let needsCleanup = false;

    try {
      await settings.open("/settings/interview/chief-complaint");
      await expect(settings.heading("主訴マスタ")).toBeVisible();

      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(complaintName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(complaintName)).toBeVisible({ timeout: 10000 });
      needsCleanup = true;

      await settings.open("/settings/interview/chief-complaint");
      await expect(settings.heading("主訴マスタ")).toBeVisible();
      await settings.searchFor("名称で検索...", complaintName);
      await expect(page.getByText(complaintName)).toBeVisible({ timeout: 10000 });

      await settings.rowActionButton(complaintName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await expect(settings.masterTitleInput()).toHaveValue(complaintName);

      const deleteResponsePromise = page.waitForResponse(
        (response) =>
          response.url().includes("/masters/chief-complaint-types/") &&
          response.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await settings.deleteButton().click();
      await expect(settings.deleteDialog()).toBeVisible();
      await settings.deleteConfirmButton("削除").click();
      const deleteResponse = await deleteResponsePromise;
      expect(deleteResponse.status(), "unused chief-complaint DELETE must succeed").toBe(204);

      await expect(page.getByText(complaintName)).not.toBeVisible({ timeout: 10000 });
      needsCleanup = false;
    } finally {
      if (needsCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/interview/chief-complaint",
          "主訴マスタ",
          "名称で検索...",
          complaintName,
        );
      }
      await page.close();
    }
  });

  test("薬剤マスタ: 作成 → 価格更新永続 → 削除", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const medicineName = disposableName("薬");
    const updatedPrice = "1234";
    let needsCleanup = false;

    try {
      await settings.open("/settings/medicine");
      await expect(settings.heading("薬剤マスタ")).toBeVisible();

      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(medicineName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      needsCleanup = true;

      await settings.open("/settings/medicine");
      await expect(settings.heading("薬剤マスタ")).toBeVisible();
      await settings.searchFor("薬品名で検索...", medicineName);
      await expect(page.getByText(medicineName)).toBeVisible({ timeout: 10000 });

      await settings.rowActionButton(medicineName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await expect(settings.masterTitleInput()).toHaveValue(medicineName);
      await settings.medicinePriceInput().fill(updatedPrice);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // C2-2 / C2-3: reload then reopen must show persisted price.
      await settings.open("/settings/medicine");
      await expect(settings.heading("薬剤マスタ")).toBeVisible();
      await settings.searchFor("薬品名で検索...", medicineName);
      await expect(page.getByText(medicineName)).toBeVisible({ timeout: 10000 });
      await settings.rowActionButton(medicineName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await expect(settings.masterTitleInput()).toHaveValue(medicineName);
      await expect(settings.medicinePriceInput()).toHaveValue(updatedPrice);

      await settings.deleteButton().click();
      await expect(settings.deleteDialog()).toBeVisible();
      await settings.deleteConfirmButton("削除する").click();
      await expect(page.getByText(medicineName)).not.toBeVisible({ timeout: 10000 });
      needsCleanup = false;
    } finally {
      if (needsCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/medicine",
          "薬剤マスタ",
          "薬品名で検索...",
          medicineName,
          "削除する",
        );
      }
      await page.close();
    }
  });

  test("支払方法: システム標準行の DELETE は拒否される", async () => {
    test.setTimeout(90000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const systemMethodName = "現金";

    try {
      await settings.open("/settings/payment-methods");
      await expect(settings.heading("支払方法マスタ")).toBeVisible();

      await settings.searchFor("支払方法名で検索...", systemMethodName);
      await expect(page.getByText(systemMethodName, { exact: true }).first()).toBeVisible({
        timeout: 10000,
      });

      await settings.rowActionButton(systemMethodName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await expect(settings.masterTitleInput()).toHaveValue(systemMethodName);

      const deleteResponsePromise = page.waitForResponse(
        (response) =>
          /\/v1\/payment-methods\/\d+/.test(response.url()) &&
          response.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await settings.deleteButton().click();
      await expect(settings.deleteDialog()).toBeVisible();
      await settings.deleteConfirmButton("削除").click();

      const deleteResponse = await deleteResponsePromise;
      expect(deleteResponse.status(), "system payment method DELETE must conflict").toBe(409);
      await expect(settings.toast()).toContainText("システム標準の支払方法は削除できません", {
        timeout: 10000,
      });

      await settings.open("/settings/payment-methods");
      await expect(settings.heading("支払方法マスタ")).toBeVisible();
      await settings.searchFor("支払方法名で検索...", systemMethodName);
      await expect(page.getByText(systemMethodName, { exact: true }).first()).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});
