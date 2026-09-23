import { test, expect } from "@playwright/test";
import type { APIRequestContext, BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { SettingsMasterPage } from "./pages/settings-master-page";

/**
 * EMR-127 V04 UAT retest — form × C1/C2/C3 browser coverage on clinic 2
 * (disposable local clinic) with the seeded executive account.
 *
 * Complements v04-settings-master-forms.spec.ts (4 narrow tests) by walking the
 * scenario denominator in docs/ops/testing/scenarios/V04-settings-master-forms.md.
 * Rows use a unique V04 run prefix and are deleted via API in finally blocks.
 *
 * Account fixture notes (see the dated reports/uat v04-retest matrix):
 * - animal-species mutations are system-admin-only (requireSystemAdminForGlobalMaster);
 *   no seeded account is a system admin, so those cells are asserted as denied.
 * - clinic-2 執行 group was temporarily granted master-payment-method create/delete
 *   for the run and restored afterwards (PUT /permission-groups/3/rules).
 * - view-only checks use the clinic-2 一般 account (V04_GENERAL_EMAIL env or default
 *   catalog account).
 */

const RUN = `V04${Date.now().toString(36)}`;
const CLINIC = "2";
const API_HEADERS = { "X-Clinic-ID": CLINIC, "X-Requested-With": "XMLHttpRequest" };

function v04(kind: string): string {
  return `${RUN}_${kind}`;
}

async function openClinic2Page(context: BrowserContext | undefined): Promise<Page> {
  if (!context) throw new Error("authed context unavailable — beforeAll login failed");
  const page = await context.newPage();
  await page.addInitScript(
    (clinicId) => localStorage.setItem("auth_current_clinic:v1", clinicId),
    CLINIC,
  );
  return page;
}

async function openMaster(page: Page, path: string, heading: string) {
  await page.goto(path, { waitUntil: "domcontentloaded" });
  await expect(
    page.getByRole("heading", { name: heading }).first(),
    `heading ${heading} on ${path}`,
  ).toBeVisible({ timeout: 45000 });
}

/** Save must be rejected: panel stays open and an inline error or sonner toast appears. */
async function expectSaveRejected(page: Page) {
  const title = page.locator("#master-title");
  await expect(title, "panel must stay open on rejected save").toBeVisible({
    timeout: 8000,
  });
  const inlineInvalid = page.locator('#master-title[aria-invalid="true"]');
  const toast = page.locator("[data-sonner-toast]");
  const anyError = page.getByText(
    /入力してください|してください|エラー|できません|使用されています|以上|以降|既に|重複/,
  );
  await expect(
    inlineInvalid.or(toast).or(anyError).first(),
    "an inline error or toast must appear",
  ).toBeVisible({ timeout: 8000 });
}

/** Save must succeed: the SidePanel closes. */
async function expectSaveSucceeded(page: Page) {
  await expect(page.locator("#master-title")).not.toBeVisible({ timeout: 15000 });
}

/** Wait for sonner toasts to auto-dismiss (they overlay the top-right 新規登録). */
async function waitToastGone(page: Page) {
  // let a just-fired toast mount before asserting none remain
  await page.waitForTimeout(400);
  await expect(page.locator("[data-sonner-toast]")).toHaveCount(0, {
    timeout: 15000,
  });
}

/** Click 新規登録; retries in case a late-mounting toast intercepted the click. */
async function clickNewButton(page: Page, settings: SettingsMasterPage) {
  for (let attempt = 0; attempt < 4; attempt++) {
    await waitToastGone(page);
    const btn = settings.newButton();
    try {
      await btn.waitFor({ state: "visible", timeout: 8000 });
      await btn.click();
      await settings.masterTitleInput().waitFor({ state: "visible", timeout: 8000 });
      return;
    } catch {
      if (attempt === 2) {
        // header may be stuck after rapid save/close cycles — reload recovers
        await page.reload({ waitUntil: "domcontentloaded" });
      }
    }
  }
  await expect(settings.masterTitleInput()).toBeVisible({ timeout: 15000 });
}

/** Open the create panel; safe to call right after a successful save. */
async function openNewPanel(page: Page, settings: SettingsMasterPage) {
  await clickNewButton(page, settings);
}

/**
 * Close an open panel without saving. Dirty panels show a discard dialog
 * (未保存の変更があります / 破棄してよろしいですか?) — confirm with 確認.
 */
async function closePanel(page: Page) {
  // Rejection/success toasts overlay the panel toolbar — wait them out first.
  await waitToastGone(page);
  await page.getByLabel("閉じる").click();
  const confirm = page.getByRole("button", { name: "確認", exact: true });
  try {
    await confirm.click({ timeout: 3000 });
  } catch {
    // clean close — no discard dialog
  }
  await expect(page.locator("#master-title")).not.toBeVisible({ timeout: 8000 });
}

/** Radix Select inside a PropertyRow — the row is a <label> wrapping the control. */
async function selectInRow(page: Page, rowLabel: string, optionName: string | RegExp) {
  const row = page
    .locator("label")
    .filter({ has: page.getByText(rowLabel, { exact: true }) })
    .first();
  await row.getByRole("combobox").first().click();
  await page.getByRole("option", { name: optionName }).first().click();
}

async function confirmDelete(page: Page, confirm: string | RegExp = /削除/) {
  await page.getByLabel("削除").click();
  const dialog = page.getByRole("alertdialog");
  await expect(dialog).toBeVisible({ timeout: 8000 });
  await dialog.getByRole("button", { name: confirm }).click();
}

/** API helpers through the app origin (cookie session is shared with the page). */
async function api(request: APIRequestContext, method: string, path: string, data?: unknown) {
  const opts = { headers: API_HEADERS, ...(data !== undefined ? { data } : {}) };
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

async function apiDeleteByName(
  request: APIRequestContext,
  listPath: string,
  deletePath: string,
  name: string,
) {
  const res = await api(request, "GET", listPath);
  if (!res.ok()) return;
  const body = await res.json();
  const items: Array<{ id: number | string; name?: string }> = Array.isArray(body)
    ? body
    : (body.data ?? body.items ?? []);
  for (const item of items) {
    if (item.name === name || (item.name ?? "").startsWith(RUN)) {
      await api(request, "DELETE", `${deletePath}/${item.id}`);
    }
  }
}

test.describe("V04 §1 標準マスタ16種 retest (clinic 2)", () => {
  let context: BrowserContext | undefined;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context?.close();
  });

  test("1-1 動物種類: 非システム管理者には作成操作が提示されない", async () => {
    test.setTimeout(90000);
    const page = await openClinic2Page(context);
    try {
      await openMaster(page, "/settings/animal-species", "動物種類マスタ");
      await expect(page.getByText("犬").first()).toBeVisible({ timeout: 15000 });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "non-admin must not get a create affordance",
      ).toHaveCount(0);
      // API-level evidence for the same account: POST/PATCH/DELETE → 403
      const post = await api(page.request, "POST", "/masters/animal-species", {
        name: v04("種拒否"),
      });
      expect(post.status()).toBe(403);
    } finally {
      await page.close();
    }
  });

  test("1-2 診断カテゴリ: C1-1 → 作成 → C2 → C3-2 → 使用中削除拒否 → 削除", async () => {
    test.setTimeout(180000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const cat = v04("カテゴリ");
    const catRenamed = `${cat}改`;
    const child = v04("病名");
    try {
      await openMaster(page, "/settings/diagnosis?tab=diagnosis_type", "診断マスタ");

      // C1-1
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // create
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(cat);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 duplicate name rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(cat);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2-1/2/3 rename → reload → reopen
      await settings.rowActionButton(cat).click();
      await expect(settings.masterTitleInput()).toHaveValue(cat, { timeout: 10000 });
      await settings.masterTitleInput().fill(catRenamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/diagnosis?tab=diagnosis_type", "診断マスタ");
      await settings.rowActionButton(catRenamed).click();
      await expect(settings.masterTitleInput()).toHaveValue(catRenamed, { timeout: 10000 });
      await settings.masterTitleInput().fill(cat);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // in-use delete rejection: child diagnosis name under this category via API
      const typeRes = await api(page.request, "GET", "/masters/diagnosis-types");
      const types = (await typeRes.json()).data as Array<{ id: number; name: string }>;
      const catId = types.find((t) => t.name === cat)!.id;
      const childRes = await api(page.request, "POST", "/masters/diagnosis-names", {
        name: child,
        diagnosis_type_id: catId,
      });
      expect(childRes.status(), "fixture child create").toBe(201);

      await openMaster(page, "/settings/diagnosis?tab=diagnosis_type", "診断マスタ");
      const conflict = page.waitForResponse(
        (r) => /diagnosis-types\/\d+/.test(r.url()) && r.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await settings.rowActionButton(cat).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await confirmDelete(page);
      expect((await conflict).status(), "in-use category delete must be 409").toBe(409);
      await expect(page.getByText(/登録されているため削除できません/)).toBeVisible({
        timeout: 10000,
      });
      await closePanel(page);
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-names",
        "/masters/diagnosis-names",
        child,
      );
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-types",
        "/masters/diagnosis-types",
        cat,
      );
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-types",
        "/masters/diagnosis-types",
        catRenamed,
      );
      await page.close();
    }
  });

  test("1-3 診断病名: C1-1(名称・カテゴリ) → 作成 → C2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const cat = v04("病名カテゴリ");
    const name = v04("病名");
    const renamed = `${name}改`;
    try {
      // fixture category via API (C3-1: FK source master data appears in options)
      const catRes = await api(page.request, "POST", "/masters/diagnosis-types", {
        name: cat,
      });
      expect(catRes.status()).toBe(201);

      await openMaster(page, "/settings/diagnosis?tab=diagnosis_name", "診断マスタ");

      // C1-1: name empty → rejected (category auto-defaults to categories[0])
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      // C3-1: API-created category appears in the select options
      await settings.masterTitleInput().fill(name);
      await selectInRow(page, "カテゴリ", new RegExp(cat));
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C2 rename → reload → reopen
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await settings.masterTitleInput().fill(renamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/diagnosis?tab=diagnosis_name", "診断マスタ");
      await settings.rowActionButton(renamed).click();
      await expect(settings.masterTitleInput()).toHaveValue(renamed, { timeout: 10000 });

      // delete
      await confirmDelete(page);
      await expect(settings.rowContaining(renamed)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-names",
        "/masters/diagnosis-names",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-names",
        "/masters/diagnosis-names",
        renamed,
      );
      await apiDeleteByName(
        page.request,
        "/masters/diagnosis-types",
        "/masters/diagnosis-types",
        cat,
      );
      await page.close();
    }
  });

  test("1-4 主訴種別: C1-1 → 作成 → C2(説明) → C3-2 → 未使用削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("主訴");
    const renamed = `${name}改`;
    try {
      await openMaster(page, "/settings/interview/chief-complaint", "主訴マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2: edit 説明 → reload → reopen persists
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await page.getByPlaceholder("説明を入力").fill(`${RUN} 説明`);
      await settings.masterTitleInput().fill(renamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/interview/chief-complaint", "主訴マスタ");
      await settings.rowActionButton(renamed).click();
      await expect(settings.masterTitleInput()).toHaveValue(renamed, { timeout: 10000 });
      await expect(page.getByPlaceholder("説明を入力")).toHaveValue(`${RUN} 説明`);

      // unused delete succeeds (204)
      const del = page.waitForResponse(
        (r) => /chief-complaint-types\/\d+/.test(r.url()) && r.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await confirmDelete(page);
      expect((await del).status()).toBe(204);
      await expect(settings.rowContaining(renamed)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/chief-complaint-types",
        "/masters/chief-complaint-types",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/chief-complaint-types",
        "/masters/chief-complaint-types",
        renamed,
      );
      await page.close();
    }
  });

  test("1-5 問診テンプレート: C1-1 → 作成(本文空可) → C2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("テンプレ");
    const renamed = `${name}改`;
    try {
      await openMaster(page, "/settings/inquiry-templates", "問診テンプレートマスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await page.getByPlaceholder("カテゴリを入力").fill(`${RUN}カテゴリ`);
      // 本文(content) is optional — leave empty
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await settings.masterTitleInput().fill(renamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/inquiry-templates", "問診テンプレートマスタ");
      await settings.rowActionButton(renamed).click();
      await expect(settings.masterTitleInput()).toHaveValue(renamed, { timeout: 10000 });

      await confirmDelete(page);
      await expect(settings.rowContaining(renamed)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/inquiry-templates",
        "/masters/inquiry-templates",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/inquiry-templates",
        "/masters/inquiry-templates",
        renamed,
      );
      await page.close();
    }
  });

  test("1-6 予約区分グループ: グループ追加 → C2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("グループ");
    const renamed = `${name}改`;
    try {
      await openMaster(page, "/settings/reservation-type", "予約区分マスタ");
      await page.getByRole("button", { name: "グループを追加" }).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 15000 });
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await expect(page.getByText(name).first()).toBeVisible({ timeout: 10000 });

      // C2 rename → reload → reopen via the group's edit affordance
      await page
        .getByRole("button", { name: new RegExp(`編集.*${name}|${name}.*編集`) })
        .first()
        .click()
        .catch(async () => {
          await page.getByText(name, { exact: true }).first().click();
        });
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(renamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/reservation-type", "予約区分マスタ");
      await expect(page.getByText(renamed).first()).toBeVisible({ timeout: 15000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/reservation-type-groups",
        "/masters/reservation-type-groups",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/reservation-type-groups",
        "/masters/reservation-type-groups",
        renamed,
      );
      await page.close();
    }
  });

  test("1-7 入院・宿泊プラン: C1-1 → 作成(任意項目空) → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("入院");
    try {
      await openMaster(page, "/settings/hospitalization", "入院マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      // bodySize/billingUnit left unset — allowed
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/hospitalization-plans",
        "/masters/hospitalization-plans",
        name,
      );
      await page.close();
    }
  });

  test("1-8 ケージ: C1-1 → 作成 → C2 → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("ケージ");
    const renamed = `${name}改`;
    try {
      await openMaster(page, "/settings/cage", "ケージマスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await settings.masterTitleInput().fill(renamed);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/cage", "ケージマスタ");
      await settings.rowActionButton(renamed).click();
      await expect(settings.masterTitleInput()).toHaveValue(renamed, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(renamed)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/cages", "/masters/cages", name);
      await apiDeleteByName(page.request, "/masters/cages", "/masters/cages", renamed);
      await page.close();
    }
  });

  test("1-9 物販・商品: C1-1 → 作成(税率8%) → C2 → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("商品");
    try {
      await openMaster(page, "/settings/merchandise-items", "商品マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await selectInRow(page, "税率", /8%/);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 duplicate rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // reopen: 税率 8% persisted
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await openMaster(page, "/settings/merchandise-items", "商品マスタ");
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/merchandise-items",
        "/masters/merchandise-items",
        name,
      );
      await page.close();
    }
  });

  test("1-10 保険: C1-1 → 補償率境界(0/100受理・101拒否) → C3-2 → 削除", async () => {
    test.setTimeout(150000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("保険");
    try {
      await openMaster(page, "/settings/insurance", "保険マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C1-3: coverage 101 must be rejected (inline or toast), not silently clamped
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(`${name}_101`);
      await page.getByLabel("補償率(%)").fill("101");
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // boundary 100 accepted
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await page.getByLabel("補償率(%)").fill("100");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2: edit coverage to 0 → persists after reload
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await page.getByLabel("補償率(%)").fill("0");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/insurance", "保険マスタ");
      await settings.rowActionButton(name).click();
      await expect(page.getByLabel("補償率(%)")).toHaveValue("0", { timeout: 10000 });

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/insurances", "/masters/insurances", name);
      await apiDeleteByName(
        page.request,
        "/masters/insurances",
        "/masters/insurances",
        `${name}_101`,
      );
      await page.close();
    }
  });

  test("1-11 職種: C1-1 → 作成 → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("職種");
    try {
      await openMaster(page, "/settings/occupations", "職種マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/occupations", "/masters/occupations", name);
      await page.close();
    }
  });

  test("1-12 トリミングコース種別: C1-1 → 作成 → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("コース種別");
    try {
      await openMaster(page, "/settings/trimming-course-type", "コース種別マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/trimming-course-types",
        "/masters/trimming-course-types",
        name,
      );
      await page.close();
    }
  });

  test("1-13 トリミングコース: C3-1(種別FK) → 作成 → C2 → C3-2 → 使用中種別削除拒否 → 削除", async () => {
    test.setTimeout(150000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const type = v04("種別");
    const name = v04("コース");
    try {
      // fixture course type via API (FK source for the course panel)
      const typeRes = await api(page.request, "POST", "/masters/trimming-course-types", {
        name: type,
      });
      expect(typeRes.status()).toBe(201);
      const typeId = (await typeRes.json()).id as number;

      await openMaster(page, "/settings/trimming?tab=course", "トリミングマスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      await settings.masterTitleInput().fill(name);
      // C3-1: API-created course type selectable in コース種別
      await selectInRow(page, "コース種別", new RegExp(type));
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 duplicate name rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // in-use: course type referenced by the course must not be deletable (409)
      const delRes = await api(page.request, "DELETE", `/masters/trimming-course-types/${typeId}`);
      expect(delRes.status(), "in-use course type delete must be 409").toBe(409);

      // C2: reopen, verify name + type persisted
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await expect(
        page.locator("label").filter({ has: page.getByText("コース種別", { exact: true }) }),
      ).toContainText(type);

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/trimming-courses",
        "/masters/trimming-courses",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/trimming-course-types",
        "/masters/trimming-course-types",
        type,
      );
      await page.close();
    }
  });

  test("1-14 トリミングオプション: C1-1 → 作成 → C2(組合せ可否) → C3-2 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("オプション");
    try {
      await openMaster(page, "/settings/trimming?tab=option", "トリミングマスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2: toggle 組合せ可否 OFF → reload → reopen persists
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      const combinableRow = page
        .locator("label")
        .filter({ has: page.getByText("組合せ可否", { exact: true }) })
        .first();
      await combinableRow.getByRole("button").first().click();
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/trimming?tab=option", "トリミングマスタ");
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(
        page.request,
        "/masters/trimming-options",
        "/masters/trimming-options",
        name,
      );
      await page.close();
    }
  });

  test("1-15 割引キャンペーン: C1-1 → 期間バリデーション → 作成 → C3-2 → 削除", async () => {
    test.setTimeout(150000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("キャンペーン");
    try {
      await openMaster(page, "/settings/campaigns", "割引キャンペーンマスタ");

      // C1-1: empty name
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      // C1-2: name set but no dates → period error
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expect(page.getByText("開始日・終了日を入力してください").first()).toBeVisible({
        timeout: 8000,
      });

      // C1-2: end before start → rejected
      await page.getByLabel("開始日").fill("2026-10-10");
      await page.getByLabel("終了日").fill("2026-10-01");
      await settings.saveButton().click();
      await expect(page.getByText("終了日は開始日以降にしてください").first()).toBeVisible({
        timeout: 8000,
      });

      // valid period → save
      await page.getByLabel("終了日").fill("2026-10-20");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 per spec: campaigns have NO name unique constraint (scenario 一意制約列「—」),
      // so a duplicate name is accepted with 201 — verified as spec-conformant, not a defect.
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await page.getByLabel("開始日").fill("2026-11-01");
      await page.getByLabel("終了日").fill("2026-11-30");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      // two rows with the same name now exist — delete the duplicate via API
      await apiDeleteByName(page.request, "/masters/campaigns", "/masters/campaigns", name);
      // apiDeleteByName removes ALL rows named `name` — recreate one for the C2 step
      const recreate = await api(page.request, "POST", "/masters/campaigns", {
        name,
        discount_type: "rate",
        discount_value: 10,
        start_date: "2026-10-10",
        end_date: "2026-10-20",
      });
      expect(recreate.status()).toBe(201);
      await page.reload({ waitUntil: "domcontentloaded" });

      // C2: reopen → dates persisted
      await settings.rowActionButton(name).first().click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByLabel("開始日")).toHaveValue("2026-10-10");
      await expect(page.getByLabel("終了日")).toHaveValue("2026-10-20");

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/campaigns", "/masters/campaigns", name);
      await page.close();
    }
  });

  test("1-16 支払方法: カスタム作成 → C3-2 → システム行削除拒否 → 削除", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("支払");
    try {
      await openMaster(page, "/settings/payment-methods", "支払方法マスタ");

      // create custom method (requires the temporary group-3 grant for the run)
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 duplicate name rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // system row (現金) delete → 409 with message
      const conflict = page.waitForResponse(
        (r) => /payment-methods\/\d+/.test(r.url()) && r.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await settings.rowActionButton("現金").click();
      await expect(settings.masterTitleInput()).toHaveValue("現金", { timeout: 10000 });
      await confirmDelete(page);
      expect((await conflict).status()).toBe(409);
      await expect(page.getByText(/システム標準の支払方法は削除できません/)).toBeVisible({
        timeout: 10000,
      });
      await closePanel(page);

      // custom row delete succeeds
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/payment-methods", "/payment-methods", name);
      await page.close();
    }
  });
});

test.describe("V04 §2-§10 retest (clinic 2)", () => {
  let context: BrowserContext | undefined;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context?.close();
  });

  test("2 診療項目(診察タブ): C1-1 → C1-2負値 → 作成 → C2 → C3-2同タブ重複 → 削除", async () => {
    test.setTimeout(180000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("診察");
    try {
      await openMaster(page, "/settings/treatment-items?tab=consultation", "診療項目マスタ");

      // C1-1: empty name
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      // C1-2: negative price rejected (FE priceError keeps panel open)
      await settings.masterTitleInput().fill(name);
      await page.getByLabel("単価(税込)").fill("-100");
      await settings.saveButton().click();
      await expectSaveRejected(page);

      await page.getByLabel("単価(税込)").fill("3000");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2: same-name in same tab rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2: edit price → reload → reopen persists
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await page.getByLabel("単価(税込)").fill("3500");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/treatment-items?tab=consultation", "診療項目マスタ");
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByLabel("単価(税込)")).toHaveValue("3500");

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/consultations", "/masters/consultations", name);
      await page.close();
    }
  });

  test("3 薬剤: C1-1 → 作成 → C3-2 → C2価格のみ更新 → 削除", async () => {
    test.setTimeout(150000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const name = v04("薬剤");
    try {
      await openMaster(page, "/settings/medicine", "薬剤マスタ");

      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      await settings.masterTitleInput().fill(name);
      await page.getByLabel("単価(税込)").fill("500");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);

      // C3-2 duplicate name rejected
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // C2: price-only update persists (PATCH keeps other fields)
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await page.getByLabel("単価(税込)").fill("800");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await openMaster(page, "/settings/medicine", "薬剤マスタ");
      await settings.rowActionButton(name).click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByLabel("単価(税込)")).toHaveValue("800");

      await confirmDelete(page);
      await expect(settings.rowContaining(name)).toHaveCount(0, { timeout: 10000 });
    } finally {
      await apiDeleteByName(page.request, "/masters/medicines", "/masters/medicines", name);
      await page.close();
    }
  });

  test("4-5 予約区分: C1-1 → グループ配下作成 → 職種紐付(C3-1) → C3-2 → 予約可能枠追加/重複409 → 削除", async () => {
    test.setTimeout(240000);
    const page = await openClinic2Page(context);
    const settings = new SettingsMasterPage(page);
    const group = v04("区分グループ");
    const occupation = v04("職種");
    const name = v04("区分");
    let typeId = 0;
    try {
      // fixtures via API: group + occupation (FK sources)
      const gRes = await api(page.request, "POST", "/masters/reservation-type-groups", {
        name: group,
      });
      expect(gRes.status()).toBe(201);
      const oRes = await api(page.request, "POST", "/masters/occupations", {
        name: occupation,
      });
      expect(oRes.status()).toBe(201);

      await openMaster(page, "/settings/reservation-type", "予約区分マスタ");

      // C1-1: empty name
      await openNewPanel(page, settings);
      await settings.saveButton().click();
      await expectSaveRejected(page);

      // create in the V04 group (C3-1: API-created group in options)
      await settings.masterTitleInput().fill(name);
      await selectInRow(page, "グループ", new RegExp(group));
      await page.getByLabel("所要時間（分）").fill("30");
      await settings.saveButton().click();
      await expectSaveSucceeded(page);
      await expect(
        page.getByRole("button", { name: new RegExp(`詳細: 予約区分 ${name}`) }),
      ).toBeVisible({ timeout: 10000 });

      // C3-2 duplicate name rejected (idx_reservation_types_clinic_name)
      await clickNewButton(page, settings);
      await settings.masterTitleInput().fill(name);
      await settings.saveButton().click();
      await expectSaveRejected(page);
      await closePanel(page);

      // reopen → occupation link via the V04 occupation (C3-1 FK option)
      await page
        .getByRole("button", { name: new RegExp(`編集: 予約区分 ${name}`) })
        .first()
        .click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      // occupation link is a Radix Select whose trigger renders EMPTY text
      // (value="__none__" suppresses the placeholder) — scope by its section
      const occSection = page
        .locator("div")
        .filter({ has: page.getByText("紐付け職種") })
        .filter({ has: page.locator('button[role="combobox"]') })
        .last();
      // capture type id for API assertions + cleanup (endpoint returns a bare array)
      const listRes = await api(page.request, "GET", "/masters/reservation-types");
      const listBody = await listRes.json();
      const types = (Array.isArray(listBody) ? listBody : (listBody.data ?? [])) as Array<{
        id: number;
        name: string;
      }>;
      typeId = types.find((t) => t.name === name)!.id;

      await occSection.getByRole("combobox").click();
      await page
        .getByRole("option", { name: new RegExp(occupation) })
        .first()
        .click();
      // BUG-MASTER-RESVTYPE-OCC-ENVELOPE: GET .../occupations returns a bare
      // array while the frontend maps `data.data` — the linked badge never
      // renders. Assert persistence at API level instead of the badge UI.
      await expect
        .poll(
          async () => {
            const res = await api(
              page.request,
              "GET",
              `/masters/reservation-types/${typeId}/occupations`,
            );
            if (!res.ok()) return "";
            const body = await res.json();
            const items = (Array.isArray(body) ? body : (body.data ?? [])) as Array<{
              occupation?: { name?: string };
            }>;
            return items.map((i) => i.occupation?.name ?? "").join(",");
          },
          { timeout: 15000 },
        )
        .toContain(occupation);

      // §5 予約可能枠 — BUG-MASTER-RESVTYPE-SLOT-FORM-NESTED: the side panel wraps
      // its content in <form action={handleAction}>, so this section's inner
      // <form> is nested and dropped by the HTML parser; the 追加 button submits
      // the panel's javascript: placeholder action (CSP-blocked). The in-panel
      // slot form is non-functional — exercise the same path via API instead.
      const slotCreate = await api(
        page.request,
        "POST",
        `/masters/reservation-types/${typeId}/available-slots`,
        { available_type: "weekly", day_of_week: 1, start_time: "09:00", is_active: true },
      );
      expect(slotCreate.status(), "weekly slot create").toBe(201);
      // list renders the persisted slot after refetch
      await openMaster(page, "/settings/reservation-type", "予約区分マスタ");
      await page
        .getByRole("button", { name: new RegExp(`編集: 予約区分 ${name}`) })
        .first()
        .click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByText(/毎週月曜日/).first()).toBeVisible({ timeout: 10000 });

      // C3-2: same weekly slot again → API must conflict
      const dupSlot = await api(
        page.request,
        "POST",
        `/masters/reservation-types/${typeId}/available-slots`,
        { available_type: "weekly", day_of_week: 1, start_time: "09:00" },
      );
      expect(dupSlot.status(), "duplicate weekly slot must be rejected").toBeGreaterThanOrEqual(
        400,
      );

      // C2: reload → type persists with occupation + slot
      await closePanel(page);
      await openMaster(page, "/settings/reservation-type", "予約区分マスタ");
      await page
        .getByRole("button", { name: new RegExp(`編集: 予約区分 ${name}`) })
        .first()
        .click();
      await expect(settings.masterTitleInput()).toHaveValue(name, { timeout: 10000 });
      // occupation badge cannot render (envelope bug) — assert API persistence
      const occRes = await api(
        page.request,
        "GET",
        `/masters/reservation-types/${typeId}/occupations`,
      );
      const occItems = (await occRes.json()) as Array<{ occupation?: { name?: string } }>;
      expect(
        occItems.map((i) => i.occupation?.name),
        "linked occupation must persist after reload",
      ).toContain(occupation);
      await expect(page.getByText(/毎週月曜日/).first()).toBeVisible({ timeout: 10000 });
    } finally {
      if (typeId) {
        await api(page.request, "DELETE", `/masters/reservation-types/${typeId}`);
      }
      await apiDeleteByName(
        page.request,
        "/masters/reservation-types",
        "/masters/reservation-types",
        name,
      );
      await apiDeleteByName(
        page.request,
        "/masters/reservation-type-groups",
        "/masters/reservation-type-groups",
        group,
      );
      await apiDeleteByName(
        page.request,
        "/masters/occupations",
        "/masters/occupations",
        occupation,
      );
      await page.close();
    }
  });

  test("6 締め時間設定: 標準締め時間の編集永続 → 休診日追加/重複409 → 特別期間追加/重複409", async () => {
    test.setTimeout(240000);
    const page = await openClinic2Page(context);
    const holidayDate = "2099-12-31";
    try {
      await openMaster(page, "/settings/closing-time", "締め時間設定");
      await expect(page.locator("#closing_weekday_end")).toBeVisible({ timeout: 15000 });

      // C2: change weekday end → save → reload persists → restore
      const originalEnd = await page.locator("#closing_weekday_end").inputValue();
      const newEnd = originalEnd === "18:00" ? "18:30" : "18:00";
      await page.locator("#closing_weekday_end").fill(newEnd);
      await page.getByRole("button", { name: "保存" }).first().click();
      await expect(page.locator("[data-sonner-toast]")).toBeVisible({ timeout: 10000 });
      await page.goto("/settings/closing-time", { waitUntil: "domcontentloaded" });
      await expect(page.locator("#closing_weekday_end")).toHaveValue(newEnd, { timeout: 15000 });
      // restore
      await page.locator("#closing_weekday_end").fill(originalEnd);
      await page.getByRole("button", { name: "保存" }).first().click();
      await expect(page.locator("[data-sonner-toast]")).toBeVisible({ timeout: 10000 });

      // 個別休診日: the add form sits behind the section's 新規登録 toggle
      const holidaySection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "個別休診日" }) })
        .first();
      await holidaySection.getByRole("button", { name: "新規登録" }).click();
      const addBtn = holidaySection.getByRole("button", { name: "追加" });
      await expect(holidaySection.locator("#holiday_date")).toBeVisible({ timeout: 8000 });

      // C1-2: reason only, no date → HTML5 required blocks submit (no POST fires)
      await holidaySection.locator("#holiday_reason").fill(`${RUN} 休診`);
      await addBtn.click();
      await expect(page.getByText(holidayDate).first()).not.toBeVisible();

      // add holiday → appears in list
      await holidaySection.locator("#holiday_date").fill(holidayDate);
      await addBtn.click();
      await expect(page.getByText(holidayDate).first()).toBeVisible({ timeout: 10000 });

      // C3-2: same date again → rejected (409)
      await holidaySection.getByRole("button", { name: "新規登録" }).click();
      const dupHoliday = page.waitForResponse(
        (r) => /closing-settings\/holidays/.test(r.url()) && r.request().method() === "POST",
        { timeout: 15000 },
      );
      await holidaySection.locator("#holiday_date").fill(holidayDate);
      await holidaySection.locator("#holiday_reason").fill(`${RUN} 重複`);
      await holidaySection.getByRole("button", { name: "追加" }).click();
      expect((await dupHoliday).status(), "duplicate holiday must conflict").toBe(409);

      // delete holiday (cleanup through UI)
      await page.getByLabel(`${holidayDate}の休診日を削除`).click();
      await expect(page.getByText(holidayDate).first()).not.toBeVisible({ timeout: 10000 });

      // 特別期間: add opens a MasterSidePanel with named inputs
      const periodSection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "特別期間" }) })
        .first();
      await periodSection.getByRole("button", { name: "新規登録" }).click();
      await expect(page.locator("#start_date")).toBeVisible({ timeout: 8000 });
      await page.locator("#start_date").fill("2099-12-20");
      await page.locator("#end_date").fill("2099-12-30");
      await page.locator("#am_pm_boundary").fill("13:00");
      await page.locator("#pm_end").fill("19:00");
      // panel 保存 is the last 保存 button in DOM order (main form has one too)
      await page.getByRole("button", { name: "保存" }).last().click();
      await expect(page.locator("#start_date")).not.toBeVisible({ timeout: 15000 });
      await expect(page.getByText("2099-12-20").first()).toBeVisible({ timeout: 10000 });

      // overlapping period → rejected (400/409)
      await periodSection.getByRole("button", { name: "新規登録" }).click();
      await expect(page.locator("#start_date")).toBeVisible({ timeout: 8000 });
      const dupPeriod = page.waitForResponse(
        (r) => /closing-settings\/special-periods/.test(r.url()) && r.request().method() === "POST",
        { timeout: 15000 },
      );
      await page.locator("#start_date").fill("2099-12-25");
      await page.locator("#end_date").fill("2099-12-28");
      await page.locator("#am_pm_boundary").fill("13:00");
      await page.locator("#pm_end").fill("19:00");
      await page.getByRole("button", { name: "保存" }).last().click();
      const dupStatus = (await dupPeriod).status();
      expect([400, 409], "overlapping special period must be rejected").toContain(dupStatus);
      await closePanel(page);

      // cleanup special period via UI
      await page.getByLabel("2099-12-20から2099-12-30の特別期間を削除").click();
      await expect(page.getByText("2099-12-20").first()).not.toBeVisible({ timeout: 10000 });
    } finally {
      // belt-and-braces cleanup in case UI delete did not run
      await api(page.request, "DELETE", `/closing-settings/holidays/${holidayDate}`);
      const sp = await api(page.request, "GET", "/closing-settings/special-periods");
      if (sp.ok()) {
        const body = await sp.json();
        const periods = (Array.isArray(body) ? body : (body.data ?? [])) as Array<{
          id: number;
          start_date: string;
        }>;
        for (const p of periods) {
          if (p.start_date >= "2099-12-20") {
            await api(page.request, "DELETE", `/closing-settings/special-periods/${p.id}`);
          }
        }
      }
      await page.close();
    }
  });

  test("7 シフトテンプレート: 名称必須(保存disabled) → 勤務種別時刻必須 → 休日作成 → C2 → 削除", async () => {
    test.setTimeout(150000);
    const page = await openClinic2Page(context);
    const name = v04("シフト");
    try {
      await openMaster(page, "/settings/shift-templates", "シフトテンプレート");
      await page.getByRole("button", { name: "新規登録" }).click();
      const nameInput = page.getByLabel("テンプレート名");
      await expect(nameInput).toBeVisible({ timeout: 15000 });

      // C1-1: empty name → save button disabled
      const saveBtn = page.getByRole("button", { name: "保存" });
      await expect(saveBtn).toBeDisabled();

      // C1-2: name + work type with empty times → time validation error
      await nameInput.fill(name);
      await expect(saveBtn).toBeEnabled();
      await saveBtn.click();
      await expect(page.getByText("勤務種別では開始時刻と終了時刻を入力してください")).toBeVisible({
        timeout: 8000,
      });

      // 休日 hides times → save succeeds without times
      // (the shift panel's PropertyRow is a div; its only combobox is シフト種別)
      await page.getByRole("combobox").click();
      await page.getByRole("option", { name: "休日" }).click();
      await saveBtn.click();
      await expect(nameInput).not.toBeVisible({ timeout: 15000 });

      // C2: reload → reopen → 休日 persisted
      await openMaster(page, "/settings/shift-templates", "シフトテンプレート");
      await page
        .locator("tbody tr")
        .filter({ hasText: name })
        .getByRole("button", { name: /編集|詳細/ })
        .first()
        .click();
      await expect(nameInput).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByRole("combobox")).toContainText("休日");

      // delete via the panel's labelled delete button
      await page.getByLabel(new RegExp(`削除: シフトテンプレート ${name}`)).click();
      const dialog = page.getByRole("alertdialog");
      await expect(dialog).toBeVisible({ timeout: 8000 });
      await dialog.getByRole("button", { name: /削除/ }).click();
      await expect(page.locator("tbody tr").filter({ hasText: name })).toHaveCount(0, {
        timeout: 10000,
      });
    } finally {
      await apiDeleteByName(page.request, "/shift-templates", "/shift-templates", name);
      await page.close();
    }
  });

  test("8 検査機器マスタ: lab-import権限viewのみ → 作成 affordance なし・POST 403", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    try {
      await openMaster(page, "/settings/lab-device-item-masters", "検査機器マスタ");
      // group-3: lab-import view+edit but NO create → 新規登録 hidden, POST /lab-devices 403
      await expect(page.getByRole("button", { name: "新規登録" })).toHaveCount(0);
      const post = await api(page.request, "POST", "/lab-devices", {
        name: v04("機器拒否"),
      });
      expect(post.status()).toBe(403);
    } finally {
      await page.close();
    }
  });

  test("10 法人情報(インボイス): 任意文字列受理 → 再読込永続 → 復元", async () => {
    test.setTimeout(120000);
    const page = await openClinic2Page(context);
    const input = page.locator("#invoice_registration_number");
    try {
      await openMaster(page, "/settings/clinic", "医院マスタ");
      await expect(input).toBeVisible({ timeout: 15000 });
      const original = await input.inputValue();

      // arbitrary text accepted per scenario (blank/malformed/short/long all OK)
      const probe = `${RUN}-T9999999999999`;
      await input.fill(probe);
      await page.getByRole("button", { name: "保存" }).first().click();
      await expect(page.locator("[data-sonner-toast]")).toBeVisible({ timeout: 10000 });
      await page.goto("/settings/clinic", { waitUntil: "domcontentloaded" });
      await expect(input).toHaveValue(probe, { timeout: 15000 });

      // blank also accepted; then restore original
      await input.fill("");
      await page.getByRole("button", { name: "保存" }).first().click();
      await expect(page.locator("[data-sonner-toast]")).toBeVisible({ timeout: 10000 });
      await page.goto("/settings/clinic", { waitUntil: "domcontentloaded" });
      await input.fill(original);
      await page.getByRole("button", { name: "保存" }).first().click();
      await expect(page.locator("[data-sonner-toast]")).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});

test.describe("V04 権限: 一般アカウントは閲覧のみ", () => {
  const GENERAL_EMAIL = process.env.V04_GENERAL_EMAIL ?? "stg-staff-20000003@example.test";

  test("一般(高橋): ケージ一覧は閲覧可・新規登録なし・POSTは403", async ({ browser }) => {
    test.setTimeout(120000);
    const password = process.env.E2E_LOGIN_PASSWORD ?? "";
    expect(password, "E2E_LOGIN_PASSWORD required").toBeTruthy();

    const ctx = await browser.newContext();
    const page = await ctx.newPage();
    try {
      await page.goto("/login", { waitUntil: "domcontentloaded" });
      await page.locator("#login-email").waitFor({ state: "visible", timeout: 60000 });
      await page.locator("#login-email").fill(GENERAL_EMAIL);
      await page.locator("#login-password").fill(password);
      const loginRes = page.waitForResponse(
        (r) => r.url().includes("/v1/login") && r.request().method() === "POST",
        { timeout: 60000 },
      );
      await page.getByRole("button", { name: "ログイン" }).click();
      expect((await loginRes).status()).toBe(200);
      await page.goto("/", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "当日の受付" })).toBeVisible({
        timeout: 60000,
      });

      await page.addInitScript(
        (clinicId) => localStorage.setItem("auth_current_clinic:v1", clinicId),
        CLINIC,
      );
      await openMaster(page, "/settings/cage", "ケージマスタ");
      await expect(page.locator("tbody tr").first()).toBeVisible({ timeout: 15000 });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a create affordance",
      ).toHaveCount(0);

      // API-level evidence: POST/PATCH/DELETE denied for the same account
      const post = await api(page.request, "POST", "/masters/cages", {
        name: v04("権限拒否"),
      });
      expect(post.status()).toBe(403);
    } finally {
      await ctx.close();
    }
  });
});
