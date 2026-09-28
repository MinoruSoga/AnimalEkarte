import { test, expect } from "@playwright/test";
import type { APIRequestContext, BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import {
  createViewOnlyContext,
  readV04FixtureFromEnv,
  type V04Fixture,
} from "./helpers/v04-fixture";
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
 * A second describe ("V04 権限: view のみ account") uses the disposable
 * clinical fixture's view-only account (E2E_CLINICAL_FIXTURE → v04.*) to prove
 * settings/master lists are readable while every create/update/delete is denied.
 * It only runs under the clinical/V04 suite where the fixture env is present.
 *
 * Credentials: admin suite uses E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD via
 * helpers/auth.ts; the view-only suite logs in with v04.viewOnlyEmail and
 * E2E_LOGIN_PASSWORD in its own browser context (no shared storage state).
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

// ─────────────────────────────────────────────────
// view-only account: list readable, writes denied
// ─────────────────────────────────────────────────

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function listRows(body: unknown): Record<string, unknown>[] {
  const rows = Array.isArray(body)
    ? body
    : isRecord(body)
      ? Array.isArray(body.data)
        ? body.data
        : Array.isArray(body.items)
          ? body.items
          : []
      : [];
  return rows.filter(isRecord);
}

function findRowId(rows: Record<string, unknown>[], name: string): number | null {
  for (const row of rows) {
    if (row.name !== name) continue;
    const id = row.id;
    if (typeof id === "number" && Number.isSafeInteger(id)) return id;
    if (typeof id === "string" && /^\d+$/.test(id)) return Number(id);
  }
  return null;
}

/** API calls through the app origin; the page shares the view-only session cookies. */
async function v04Api(
  request: APIRequestContext,
  clinicId: number,
  method: string,
  path: string,
  data?: unknown,
) {
  const opts = {
    headers: { "X-Clinic-ID": String(clinicId), "X-Requested-With": "XMLHttpRequest" },
    ...(data !== undefined ? { data } : {}),
  };
  const url = `/api/v1${path}`;
  switch (method) {
    case "GET":
      return request.get(url, opts);
    case "POST":
      return request.post(url, opts);
    case "PATCH":
      return request.patch(url, opts);
    case "DELETE":
      return request.delete(url, opts);
    default:
      throw new Error(method);
  }
}

async function openViewOnlyPage(
  context: BrowserContext | undefined,
  fixture: V04Fixture,
): Promise<Page> {
  if (context === undefined) {
    throw new Error("view-only context unavailable — beforeAll login failed");
  }
  const page = await context.newPage();
  // Same mechanism as the retest spec: pin the clinic before app scripts run.
  await page.addInitScript(
    (clinicId) => localStorage.setItem("auth_current_clinic:v1", clinicId),
    String(fixture.clinicId),
  );
  return page;
}

test.describe("V04 権限: view のみ account（disposable clinic fixture）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext | undefined;

  test.beforeAll(async ({ browser }) => {
    if (v04 === null) return;
    context = await createViewOnlyContext(browser, v04);
  });

  test.afterAll(async () => {
    await context?.close();
  });

  test("ケージマスタ: 一覧と詳細は閲覧可・登録/保存/削除の導線なし・API write は 403", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await openViewOnlyPage(context, v04);
    try {
      await page.goto("/settings/cage", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "ケージマスタ" }).first()).toBeVisible({
        timeout: 45000,
      });

      const rowButton = page.getByRole("button", { name: `詳細: ケージ ${v04.cageName}` });
      await expect(rowButton, "fixture cage row must be listed").toBeVisible({ timeout: 30000 });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a create affordance",
      ).toHaveCount(0);

      await rowButton.click();
      const title = page.locator("#master-title");
      await expect(title).toBeVisible({ timeout: 15000 });
      await expect(title).toHaveValue(v04.cageName);
      await expect(
        page.getByRole("button", { name: "保存" }),
        "read-only panel must not render a save button",
      ).toHaveCount(0);
      await expect(
        page.getByLabel("削除"),
        "read-only panel must not render a delete button",
      ).toHaveCount(0);

      const list = await v04Api(page.request, v04.clinicId, "GET", "/masters/cages");
      expect(list.status(), "GET /masters/cages must be allowed for view").toBe(200);
      const cageId = findRowId(listRows(await list.json()), v04.cageName);
      expect(cageId, "fixture cage must appear in GET /masters/cages").not.toBeNull();

      const deniedName = `V04-e2e-denied-${Date.now()}`;
      const post = await v04Api(page.request, v04.clinicId, "POST", "/masters/cages", {
        name: deniedName,
        cage_type: "general",
        cage_size: "medium",
      });
      expect(post.status(), "POST /masters/cages must be denied").toBe(403);
      const patch = await v04Api(page.request, v04.clinicId, "PATCH", `/masters/cages/${cageId}`, {
        name: `${v04.cageName}-denied`,
      });
      expect(patch.status(), "PATCH /masters/cages/:id must be denied").toBe(403);
    } finally {
      await page.close();
    }
  });

  test("検査機器マスタ: 一覧と詳細は閲覧可・登録/既定項目/保存の導線なし・API write は 403", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await openViewOnlyPage(context, v04);
    try {
      await page.goto("/settings/lab-device-item-masters", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "検査機器マスタ" }).first()).toBeVisible({
        timeout: 45000,
      });

      const rowButton = page.getByRole("button", {
        name: `詳細: 検査機器 ${v04.labDeviceName}`,
      });
      await expect(rowButton, "fixture lab device row must be listed").toBeVisible({
        timeout: 30000,
      });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a create affordance",
      ).toHaveCount(0);
      await expect(
        page.getByRole("button", { name: "既定項目を用意" }),
        "view-only account must not get the ensure-defaults affordance",
      ).toHaveCount(0);

      await rowButton.click();
      const title = page.locator("#master-title");
      await expect(title).toBeVisible({ timeout: 15000 });
      await expect(title).toHaveValue(v04.labDeviceName);
      await expect(
        page.getByRole("button", { name: "保存" }),
        "read-only panel must not render a save button",
      ).toHaveCount(0);
      // readOnly disables the panel's editable controls (SidePeekTitleInput
      // itself stays focusable but there is no save path to persist it).
      await expect(
        page.getByRole("combobox", { name: "検査" }),
        "read-only panel must disable the exam-type select",
      ).toBeDisabled();

      const list = await v04Api(page.request, v04.clinicId, "GET", "/lab-devices");
      expect(list.status(), "GET /lab-devices must be allowed for view").toBe(200);
      const deviceId = findRowId(listRows(await list.json()), v04.labDeviceName);
      expect(deviceId, "fixture lab device must appear in GET /lab-devices").not.toBeNull();

      const post = await v04Api(page.request, v04.clinicId, "POST", "/lab-devices", {
        name: v04.labDeviceName,
        source_type: "fuji_au10v",
      });
      expect(post.status(), "POST /lab-devices must be denied").toBe(403);
      const patch = await v04Api(page.request, v04.clinicId, "PATCH", `/lab-devices/${deviceId}`, {
        name: `V04-e2e-denied-${Date.now()}`,
        is_active: true,
        sort_order: 0,
      });
      expect(patch.status(), "PATCH /lab-devices/:id must be denied").toBe(403);
    } finally {
      await page.close();
    }
  });
});
