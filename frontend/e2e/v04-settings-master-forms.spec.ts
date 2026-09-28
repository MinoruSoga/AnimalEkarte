import { test, expect } from "@playwright/test";
import type { APIRequestContext, BrowserContext, Page } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import {
  createViewOnlyContext,
  readV04FixtureFromEnv,
  type V04Fixture,
} from "./fixtures/v04-fixture";
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

  test("予約区分: 予約可能枠を追加 → 再読込で永続（EMR-208 回帰）", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const typeName = disposableName("区分枠");
    let needsCleanup = false;

    // セクション見出し <p> の 2 つ上の div がセクション本体（追加フォーム+一覧を含む）。
    const slotsSection = () => page.getByText("予約可能枠", { exact: true }).locator("xpath=../..");

    try {
      await settings.open("/settings/reservation-type");
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });

      // 子セクションは既存の leaf 区分を開き直した時だけ描画されるため、
      // 先に区分を作成して保存 → 一覧からパネルを開き直す。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(typeName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(typeName)).toBeVisible({ timeout: 10000 });
      needsCleanup = true;

      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });

      // EMR-208 回帰点: 「追加」がネスト form ではなく mutation を直接呼び、
      // POST /available-slots が発行されること（既定値: 毎週月曜日 09:45）。
      const createSlotResponsePromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/reservation-types\/\d+\/available-slots/.test(response.url()) &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await slotsSection().getByRole("button", { name: "追加", exact: true }).click();
      const createSlotResponse = await createSlotResponsePromise;
      expect(createSlotResponse.status(), "available-slot POST must be issued and succeed").toBe(
        201,
      );
      await expect(slotsSection().getByText("毎週月曜日")).toBeVisible({ timeout: 10000 });

      // C2-2/C2-3: ブラウザ再読込 → パネル再オープンでスロットが永続している。
      await page.reload({ waitUntil: "domcontentloaded" });
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });
      const persistedRow = slotsSection().locator("div", { hasText: "毎週月曜日" }).last();
      await expect(persistedRow).toBeVisible({ timeout: 10000 });
      await expect(persistedRow).toContainText("09:45");

      // 後始末: スロット・予約不可時間行にも aria-label="削除" が居るためツールバー側を .first() で取る。
      await settings.deleteButton().first().click();
      await expect(settings.deleteDialog()).toBeVisible();
      await settings.deleteConfirmButton("削除").click();
      await expect(settings.rowContaining(typeName)).toHaveCount(0, { timeout: 10000 });
      needsCleanup = false;
    } finally {
      if (needsCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/reservation-type",
          "予約区分マスタ",
          "予約区分名で検索...",
          typeName,
        );
      }
      await page.close();
    }
  });

  test("予約区分: 職種を紐付け → パネルにバッジ表示（EMR-209 回帰）", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const typeName = disposableName("区分職種");
    const occupationName = disposableName("職種");
    let needsTypeCleanup = false;
    let needsOccupationCleanup = false;

    // <p>紐付け職種…</p> の 2 つ上の div がセクション本体（バッジ群+追加 select を含む）。
    const occupationsSection = () =>
      page.getByText("紐付け職種", { exact: false }).locator("xpath=../..");

    try {
      // 紐付け対象の職種を disposable clinic に作成（fixture は持たないため UI で作成）。
      await settings.open("/settings/occupations");
      await expect(settings.heading("職種マスタ")).toBeVisible({ timeout: 15000 });
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(occupationName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      needsOccupationCleanup = true;

      await settings.open("/settings/reservation-type");
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(typeName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      needsTypeCleanup = true;

      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });

      const linkResponsePromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/reservation-types\/\d+\/occupations/.test(response.url()) &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await occupationsSection().getByRole("combobox").click();
      await page.getByRole("option", { name: occupationName, exact: true }).click();
      const linkResponse = await linkResponsePromise;
      expect(linkResponse.status(), "occupation link POST must succeed").toBe(201);

      // EMR-209 回帰点: 紐付け一覧 GET の {data} エンベロープがパースされ、
      // 紐付け職種セクションに職種名のバッジが描画される。
      await expect(occupationsSection().getByText(occupationName)).toBeVisible({
        timeout: 10000,
      });
      await expect(
        occupationsSection().getByLabel(`${occupationName} の紐付けを解除`),
      ).toBeVisible();

      // 再読込 → パネル再オープンでもバッジが残る（GET パース経路の永続確認）。
      await page.reload({ waitUntil: "domcontentloaded" });
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });
      await expect(occupationsSection().getByText(occupationName)).toBeVisible({
        timeout: 10000,
      });
    } finally {
      if (needsTypeCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/reservation-type",
          "予約区分マスタ",
          "予約区分名で検索...",
          typeName,
        );
      }
      if (needsOccupationCleanup) {
        await bestEffortDeleteRow(
          settings,
          "/settings/occupations",
          "職種マスタ",
          "職種名で検索...",
          occupationName,
        );
      }
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

/** `id` を単一行レスポンスまたは `{data: {...}}` エンベロープから取り出す。 */
function readRowId(body: unknown): number | null {
  const row = isRecord(body) && isRecord(body.data) ? body.data : body;
  if (!isRecord(row)) return null;
  const id = row.id;
  if (typeof id === "number" && Number.isSafeInteger(id)) return id;
  if (typeof id === "string" && /^\d+$/.test(id)) return Number(id);
  return null;
}

/**
 * view-only 横展開の対象画面。一覧行は admin account で disposable clinic に
 * シードする（`seedBody` 指定時）か、既存行（グローバル seed・trigger 作成の
 * システム標準行）を `rowName` / 先頭行で解決する。
 */
interface ViewOnlyMasterScreen {
  /** fixture の viewOnlyResources に含まれるべき resource 名 */
  resource: string;
  /** テスト名用の画面ラベル */
  screen: string;
  path: string;
  heading: string;
  /** 行ボタンの `詳細: <entityLabel> <name> (ID <id>)` に使う名詞 */
  entityLabel: string;
  /** 代表 API パス（GET=一覧、POST/PATCH/DELETE=403 確認） */
  apiPath: string;
  /** 指定時は admin で `{name, ...seedBody}` を POST して行を用意する */
  seedBody?: Record<string, unknown>;
  /** 指定時はシードせず GET 一覧から同名の行を使う */
  rowName?: string;
}

const VIEW_ONLY_MASTER_SCREENS: ViewOnlyMasterScreen[] = [
  {
    // グローバル seed の行を先頭から使う（clinic_id を持たない共有マスタのためシードしない）。
    resource: "master-animal-species",
    screen: "動物種類",
    path: "/settings/animal-species",
    heading: "動物種類マスタ",
    entityLabel: "動物種類",
    apiPath: "/masters/animal-species",
  },
  {
    resource: "master-medical",
    screen: "診療項目",
    path: "/settings/treatment-items",
    heading: "診療項目マスタ",
    entityLabel: "治療プラン",
    apiPath: "/masters/consultations",
    seedBody: { is_active: true },
  },
  {
    resource: "master-reservation-type",
    screen: "予約区分",
    path: "/settings/reservation-type",
    heading: "予約区分マスタ",
    entityLabel: "予約区分",
    apiPath: "/masters/reservation-types",
    seedBody: { is_active: true },
  },
  {
    resource: "master-hospitalization",
    screen: "入院・宿泊",
    path: "/settings/hospitalization",
    heading: "入院マスタ",
    entityLabel: "入院プラン",
    apiPath: "/masters/hospitalization-plans",
    seedBody: { is_active: true },
  },
  {
    resource: "master-trimming",
    screen: "トリミング",
    path: "/settings/trimming?tab=course",
    heading: "トリミングマスタ",
    entityLabel: "トリミングコース",
    apiPath: "/masters/trimming-courses",
    seedBody: { is_active: true },
  },
  {
    resource: "master-insurance",
    screen: "保険",
    path: "/settings/insurance",
    heading: "保険マスタ",
    entityLabel: "保険",
    apiPath: "/masters/insurances",
    seedBody: { coverage_rate: 50, is_active: true },
  },
  {
    resource: "master-merchandise",
    screen: "物販・商品",
    path: "/settings/merchandise-items",
    heading: "商品マスタ",
    entityLabel: "品目",
    apiPath: "/masters/merchandise-items",
    seedBody: { category: "goods", unit_price: 1000, tax_type: "excluded", is_active: true },
  },
  {
    // clinic 作成トリガーのシステム標準行「現金」を使う。
    resource: "master-payment-method",
    screen: "支払方法",
    path: "/settings/payment-methods",
    heading: "支払方法マスタ",
    entityLabel: "支払方法",
    apiPath: "/payment-methods",
    rowName: "現金",
  },
];

/** disposable clinic admin session で行をシードし、採番された id を返す。 */
async function seedMasterRow(
  request: APIRequestContext,
  clinicId: number,
  path: string,
  body: Record<string, unknown>,
): Promise<number> {
  const response = await v04Api(request, clinicId, "POST", path, body);
  const id = await response
    .json()
    .then((parsed) => readRowId(parsed))
    .catch(() => null);
  expect(id, `admin seed POST ${path} must succeed (status=${response.status()})`).not.toBeNull();
  return id as number;
}

/** シード行の後始末。失敗は握りつぶす（best-effort cleanup）。 */
async function deleteSeededRow(
  request: APIRequestContext,
  clinicId: number,
  path: string,
): Promise<void> {
  try {
    await v04Api(request, clinicId, "DELETE", path);
  } catch {
    // Best-effort cleanup only.
  }
}

test.describe("V04 権限: view のみ account（disposable clinic fixture）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext | undefined;
  // シード・後始末用の admin session（view-only account では行を作れない）。
  let adminContext: BrowserContext | undefined;

  test.beforeAll(async ({ browser }) => {
    if (v04 === null) return;
    context = await createViewOnlyContext(browser, v04);
    adminContext = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context?.close();
    await adminContext?.close();
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
      const del = await v04Api(page.request, v04.clinicId, "DELETE", `/masters/cages/${cageId}`);
      expect(del.status(), "DELETE /masters/cages/:id must be denied").toBe(403);
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
      // lab-devices has no DELETE route (GET/POST/PATCH/PUT configuration only), so it is not probed.
    } finally {
      await page.close();
    }
  });

  for (const master of VIEW_ONLY_MASTER_SCREENS) {
    test(`${master.screen}: 一覧と詳細は閲覧可・作成/保存/削除の導線なし・API write は 403`, async () => {
      test.setTimeout(120000);
      if (v04 === null) throw new Error("v04 fixture unavailable");
      if (adminContext === undefined) throw new Error("admin context unavailable");
      expect(v04.viewOnlyResources, `fixture must grant ${master.resource} view`).toContain(
        master.resource,
      );
      const page = await openViewOnlyPage(context, v04);
      let seededId: number | null = null;
      try {
        // 一覧の読取 + 対象行の解決（seedBody 指定時は admin で行を用意する）。
        const list = await v04Api(page.request, v04.clinicId, "GET", master.apiPath);
        expect(list.status(), `GET ${master.apiPath} must be allowed for view`).toBe(200);
        let rowName: string;
        let rowId: number;
        if (master.seedBody !== undefined) {
          rowName = `V04-e2e-${master.entityLabel}-${Date.now()}`;
          rowId = await seedMasterRow(adminContext.request, v04.clinicId, master.apiPath, {
            name: rowName,
            ...master.seedBody,
          });
          seededId = rowId;
        } else {
          const rows = listRows(await list.json());
          const found =
            master.rowName !== undefined ? rows.find((r) => r.name === master.rowName) : rows[0];
          const foundId = found === undefined ? null : readRowId(found);
          expect(foundId, `${master.apiPath} must list a readable row`).not.toBeNull();
          rowId = foundId as number;
          rowName = String(found?.name);
        }

        await page.goto(master.path, { waitUntil: "domcontentloaded" });
        await expect(page.getByRole("heading", { name: master.heading }).first()).toBeVisible({
          timeout: 45000,
        });

        const rowButton = page.getByRole("button", {
          name: `詳細: ${master.entityLabel} ${rowName} (ID ${rowId})`,
        });
        await expect(rowButton, "対象行が一覧に表示されること").toBeVisible({ timeout: 30000 });
        await expect(
          page.getByRole("button", { name: "新規登録" }),
          "view-only account must not get a create affordance",
        ).toHaveCount(0);

        await rowButton.click();
        const title = page.locator("#master-title");
        await expect(title).toBeVisible({ timeout: 15000 });
        await expect(title).toHaveValue(rowName);
        await expect(
          page.getByRole("button", { name: "保存" }),
          "read-only panel must not render a save button",
        ).toHaveCount(0);
        await expect(
          page.getByLabel("削除"),
          "read-only panel must not render a delete button",
        ).toHaveCount(0);

        const deniedName = `V04-e2e-denied-${Date.now()}`;
        const post = await v04Api(page.request, v04.clinicId, "POST", master.apiPath, {
          name: deniedName,
          ...(master.seedBody ?? {}),
        });
        expect(post.status(), `POST ${master.apiPath} must be denied`).toBe(403);
        const patch = await v04Api(
          page.request,
          v04.clinicId,
          "PATCH",
          `${master.apiPath}/${rowId}`,
          { name: deniedName },
        );
        expect(patch.status(), `PATCH ${master.apiPath}/:id must be denied`).toBe(403);
        const del = await v04Api(
          page.request,
          v04.clinicId,
          "DELETE",
          `${master.apiPath}/${rowId}`,
        );
        expect(del.status(), `DELETE ${master.apiPath}/:id must be denied`).toBe(403);
      } finally {
        if (seededId !== null && adminContext !== undefined) {
          await deleteSeededRow(
            adminContext.request,
            v04.clinicId,
            `${master.apiPath}/${seededId}`,
          );
        }
        await page.close();
      }
    });
  }

  test("締め時間設定: 閲覧可・入力は disabled・保存/新規登録の導線なし・API write は 403", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    expect(v04.viewOnlyResources).toContain("closing-settings");
    const page = await openViewOnlyPage(context, v04);
    try {
      await page.goto("/settings/closing-time", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "締め時間設定" }).first()).toBeVisible({
        timeout: 45000,
      });

      // 標準締め時間: fieldset が disabled 化され、保存 submit は描画されない。
      await expect(page.locator("#closing_weekday_end")).toBeDisabled();
      await expect(
        page.getByRole("button", { name: "保存" }),
        "closing-time must not render a save button for view-only",
      ).toHaveCount(0);

      // 個別休診日 / 特別期間: 作成権限がないため「新規登録」ボタン自体を描画しない
      // （他マスタと同じ canCreate ゲート。canEditRef のトーストガードは二重防御として残る）。
      const holidaySection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "個別休診日" }) });
      const specialPeriodSection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "特別期間" }) });
      await expect(
        holidaySection.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a holiday create affordance",
      ).toHaveCount(0);
      await expect(
        specialPeriodSection.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a special-period create affordance",
      ).toHaveCount(0);

      const settings = await v04Api(page.request, v04.clinicId, "GET", "/closing-settings");
      expect(settings.status(), "GET /closing-settings must be allowed for view").toBe(200);
      const patch = await v04Api(page.request, v04.clinicId, "PATCH", "/closing-settings", {
        am_pm_boundary: "13:00",
        weekday_end: "19:00",
        sunday_end: "18:00",
      });
      expect(patch.status(), "PATCH /closing-settings must be denied").toBe(403);
      const postPeriod = await v04Api(
        page.request,
        v04.clinicId,
        "POST",
        "/closing-settings/special-periods",
        {
          start_date: "2099-12-20",
          end_date: "2099-12-21",
          am_pm_boundary: "13:00",
          pm_end: "19:00",
        },
      );
      expect(postPeriod.status(), "POST /closing-settings/special-periods must be denied").toBe(
        403,
      );
      const postHoliday = await v04Api(
        page.request,
        v04.clinicId,
        "POST",
        "/closing-settings/holidays",
        { date: "2099-12-31", reason: "V04-e2e-view-only" },
      );
      expect(postHoliday.status(), "POST /closing-settings/holidays must be denied").toBe(403);
      const delPeriod = await v04Api(
        page.request,
        v04.clinicId,
        "DELETE",
        "/closing-settings/special-periods/0",
      );
      expect(
        delPeriod.status(),
        "DELETE /closing-settings/special-periods/:id must be denied",
      ).toBe(403);
    } finally {
      await page.close();
    }
  });

  test("シフトテンプレート: 一覧と詳細は閲覧可・作成/保存/削除の導線なし・API write は 403", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    if (adminContext === undefined) throw new Error("admin context unavailable");
    expect(v04.viewOnlyResources).toContain("shifts");
    const templateName = `V04-e2e-シフト-${Date.now()}`;
    const templateId = await seedMasterRow(adminContext.request, v04.clinicId, "/shift-templates", {
      name: templateName,
      shift_type: "full",
      start_time: "09:00",
      end_time: "18:00",
    });
    const page = await openViewOnlyPage(context, v04);
    try {
      await page.goto("/settings/shift-templates", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "シフトテンプレートマスタ" })).toBeVisible({
        timeout: 45000,
      });

      const rowButton = page.getByRole("button", {
        name: `詳細: シフトテンプレート ${templateName} (ID ${templateId})`,
      });
      await expect(rowButton, "seeded template row must be listed").toBeVisible({
        timeout: 30000,
      });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a create affordance",
      ).toHaveCount(0);

      await rowButton.click();
      const nameInput = page.getByLabel("テンプレート名");
      await expect(nameInput).toBeVisible({ timeout: 15000 });
      await expect(nameInput).toHaveValue(templateName);
      await expect(nameInput).toHaveAttribute("readonly", "");
      await expect(
        page.getByRole("button", { name: "保存" }),
        "read-only panel must not render a save button",
      ).toHaveCount(0);
      await expect(
        page.getByLabel(`削除: シフトテンプレート ${templateName} (ID ${templateId})`),
        "read-only panel must not render a delete button",
      ).toHaveCount(0);
      // The read-only panel renders both the header icon (aria-label) and a footer text button.
      await expect(page.getByRole("button", { name: "閉じる", exact: true }).first()).toBeVisible();

      const list = await v04Api(page.request, v04.clinicId, "GET", "/shift-templates");
      expect(list.status(), "GET /shift-templates must be allowed for view").toBe(200);
      const post = await v04Api(page.request, v04.clinicId, "POST", "/shift-templates", {
        name: `V04-e2e-denied-${Date.now()}`,
        shift_type: "full",
        start_time: "09:00",
        end_time: "18:00",
      });
      expect(post.status(), "POST /shift-templates must be denied").toBe(403);
      const patch = await v04Api(
        page.request,
        v04.clinicId,
        "PATCH",
        `/shift-templates/${templateId}`,
        { name: "V04-e2e-denied" },
      );
      expect(patch.status(), "PATCH /shift-templates/:id must be denied").toBe(403);
      const del = await v04Api(
        page.request,
        v04.clinicId,
        "DELETE",
        `/shift-templates/${templateId}`,
      );
      expect(del.status(), "DELETE /shift-templates/:id must be denied").toBe(403);
    } finally {
      await deleteSeededRow(adminContext.request, v04.clinicId, `/shift-templates/${templateId}`);
      await page.close();
    }
  });

  test("医院マスタ: 一覧と詳細は閲覧可・作成/保存/削除の導線なし・API write は 403", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    expect(v04.viewOnlyResources).toContain("hospital-settings");
    const page = await openViewOnlyPage(context, v04);
    try {
      const list = await v04Api(page.request, v04.clinicId, "GET", "/clinics");
      expect(list.status(), "GET /clinics must be allowed for view").toBe(200);
      const rows = listRows(await list.json());
      const selfRow = rows.find((row) => readRowId(row) === v04.clinicId);
      expect(selfRow, "GET /clinics must list the fixture clinic").not.toBeUndefined();
      const clinicName = String(selfRow?.name);

      await page.goto("/settings/clinic", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "医院マスタ" })).toBeVisible({
        timeout: 45000,
      });

      // 法人インボイス欄も canEdit で disabled 化される。
      await expect(page.locator("#invoice_registration_number")).toBeDisabled();

      const rowButton = page.getByRole("button", {
        name: `詳細: 医院 ${clinicName} (ID ${v04.clinicId})`,
      });
      await expect(rowButton, "fixture clinic row must be listed").toBeVisible({
        timeout: 30000,
      });
      await expect(
        page.getByRole("button", { name: "新規登録" }),
        "view-only account must not get a create affordance",
      ).toHaveCount(0);

      await rowButton.click();
      const nameInput = page.getByLabel("無題");
      await expect(nameInput).toBeVisible({ timeout: 15000 });
      await expect(nameInput).toHaveValue(clinicName);
      await expect(nameInput).toBeDisabled();
      await expect(
        page.getByRole("button", { name: "保存" }),
        "read-only panel must not render a save button",
      ).toHaveCount(0);

      const post = await v04Api(page.request, v04.clinicId, "POST", "/clinics", {
        name: `V04-e2e-denied-${Date.now()}`,
      });
      expect(post.status(), "POST /clinics must be denied").toBe(403);
      const patch = await v04Api(page.request, v04.clinicId, "PATCH", `/clinics/${v04.clinicId}`, {
        name: "V04-e2e-denied",
      });
      expect(patch.status(), "PATCH /clinics/:id must be denied").toBe(403);
    } finally {
      await page.close();
    }
  });
});

// ─────────────────────────────────────────────────
// admin account: §5 予約可能枠 #3〜#6（EMR-127d）
// ─────────────────────────────────────────────────

/** 予約区分と配下の予約可能枠を API で後始末する（子→親の順。失敗は握りつぶす）。 */
async function deleteReservationTypeWithSlots(
  request: APIRequestContext,
  clinicId: number,
  typeId: number,
): Promise<void> {
  try {
    const slots = await v04Api(
      request,
      clinicId,
      "GET",
      `/masters/reservation-types/${typeId}/available-slots`,
    );
    if (slots.ok()) {
      for (const row of listRows(await slots.json())) {
        const slotId = readRowId(row);
        if (slotId !== null) {
          await deleteSeededRow(
            request,
            clinicId,
            `/masters/reservation-types/${typeId}/available-slots/${slotId}`,
          );
        }
      }
    }
    await deleteSeededRow(request, clinicId, `/masters/reservation-types/${typeId}`);
  } catch {
    // Best-effort cleanup only.
  }
}

test.describe("V04 設定マスタ §5 予約可能枠（admin）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("予約可能枠: 特定日スロット追加・永続・パネル再オープン・削除 (#3・#4・#6)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const typeName = disposableName("区分特定日");
    const specificDate = "2099-01-15";
    let typeId: number | null = null;

    // セクション見出し <p> の 2 つ上の div がセクション本体（追加フォーム+一覧を含む）。
    const slotsSection = () => page.getByText("予約可能枠", { exact: true }).locator("xpath=../..");

    try {
      // leaf 区分を作成して保存 → パネルを開き直す（子セクションは既存行のみ描画）。
      await settings.open("/settings/reservation-type");
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(typeName);
      await settings.saveButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(typeName)).toBeVisible({ timeout: 10000 });

      const listResponse = await v04Api(
        page.request,
        v04.clinicId,
        "GET",
        "/masters/reservation-types",
      );
      typeId = findRowId(listRows(await listResponse.json()), typeName);
      expect(typeId, "created reservation type must be listed").not.toBeNull();

      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });

      // 毎週枠を 1 件追加（既定値: 毎週月曜日 09:45）。
      const weeklyPostPromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/reservation-types\/\d+\/available-slots/.test(response.url()) &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await slotsSection().getByRole("button", { name: "追加", exact: true }).click();
      expect((await weeklyPostPromise).status(), "weekly slot POST must succeed").toBe(201);
      await expect(slotsSection().getByText("毎週月曜日")).toBeVisible({ timeout: 10000 });

      // #3: モードを「特定日」に切替 → 日付を入れて追加 → 永続。
      // パネル内の他セクション（予約不可時間など）にも aria-label="特定日" の
      // date input があるため、予約可能枠セクションにスコープして一意にする。
      await slotsSection().getByRole("combobox").first().click();
      await page.getByRole("option", { name: "特定日", exact: true }).click();
      await slotsSection().getByLabel("特定日").fill(specificDate);
      const specificPostPromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/reservation-types\/\d+\/available-slots/.test(response.url()) &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await slotsSection().getByRole("button", { name: "追加", exact: true }).click();
      expect((await specificPostPromise).status(), "specific-date slot POST must succeed").toBe(
        201,
      );
      await expect(slotsSection().getByText(specificDate)).toBeVisible({ timeout: 10000 });

      // #6: パネルを閉じて開き直す → 毎週・特定日の保存済み枠が再表示される。
      await page.getByLabel("閉じる").first().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });
      await expect(slotsSection().getByText("毎週月曜日")).toBeVisible({ timeout: 10000 });
      await expect(slotsSection().getByText(specificDate)).toBeVisible();

      // ブラウザ再読込でも永続。
      await page.reload({ waitUntil: "domcontentloaded" });
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });
      await expect(slotsSection().getByText("毎週月曜日")).toBeVisible({ timeout: 10000 });
      await expect(slotsSection().getByText(specificDate)).toBeVisible();

      // #4: 特定日スロットを削除 → 再読込でも消えている。
      const slotRow = page.getByText(specificDate).locator("xpath=..");
      const deleteSlotPromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/reservation-types\/\d+\/available-slots\/\d+/.test(response.url()) &&
          response.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await slotRow.getByRole("button", { name: "削除" }).click();
      expect((await deleteSlotPromise).status(), "slot DELETE must succeed").toBe(204);
      await expect(slotsSection().getByText(specificDate)).toHaveCount(0, { timeout: 10000 });

      await page.reload({ waitUntil: "domcontentloaded" });
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });
      await settings.rowActionButton(typeName).click();
      await expect(settings.masterTitleInput()).toHaveValue(typeName, { timeout: 10000 });
      await expect(slotsSection().getByText(specificDate)).toHaveCount(0, { timeout: 10000 });
      await expect(slotsSection().getByText("毎週月曜日")).toBeVisible();
    } finally {
      if (typeId !== null) {
        await deleteReservationTypeWithSlots(page.request, v04.clinicId, typeId);
      }
      await page.close();
    }
  });

  test("予約可能枠: 区分セレクタはリーフ区分のみ選択可 (#5)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const parentName = disposableName("区分親");
    const leafName = disposableName("区分子");
    let parentId: number | null = null;
    let leafId: number | null = null;

    try {
      // 親区分 + 配下の leaf 区分を API で用意（ツリー描画の対象）。
      parentId = await seedMasterRow(page.request, v04.clinicId, "/masters/reservation-types", {
        name: parentName,
        is_active: true,
      });
      leafId = await seedMasterRow(page.request, v04.clinicId, "/masters/reservation-types", {
        name: leafName,
        is_active: true,
        parent_id: parentId,
      });

      await page.goto("/line-reservation/slots", { waitUntil: "domcontentloaded" });
      await expect(page.getByRole("heading", { name: "LINE予約枠" }).first()).toBeVisible({
        timeout: 45000,
      });

      // 親区分は展開トグル（"{name} グループ"）だけで、選択肢としては出ない。
      // 初期の aria-expanded は「展開済み」になっているため固定値は前提にせず、
      // クリックで値がトグルし、かつ親は選択されない（typeId が付かない）ことを見る。
      const parentButton = page.getByRole("button", { name: `${parentName} グループ` });
      await expect(parentButton).toBeVisible({ timeout: 30000 });
      const expandedBefore = await parentButton.getAttribute("aria-expanded");
      await parentButton.click();
      await expect(parentButton).toHaveAttribute(
        "aria-expanded",
        expandedBefore === "true" ? "false" : "true",
      );
      expect(
        new URL(page.url()).searchParams.get("typeId"),
        "parent group click must not select the parent type",
      ).not.toBe(String(parentId));

      // leaf を出すため、トグル後が折りたたみならもう一度クリックして展開する。
      if ((await parentButton.getAttribute("aria-expanded")) !== "true") {
        await parentButton.click();
        await expect(parentButton).toHaveAttribute("aria-expanded", "true");
      }

      // 展開したグループ内の leaf を選択すると typeId が指し、カレンダーが描画される。
      const leafButton = page.getByRole("button", { name: leafName, exact: true });
      await expect(leafButton).toBeVisible({ timeout: 10000 });
      await leafButton.click();
      await expect(page).toHaveURL(new RegExp(`typeId=${leafId}`), { timeout: 10000 });
      await expect(page.getByText(`${parentName} / ${leafName}`, { exact: true })).toBeVisible();
    } finally {
      if (leafId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/reservation-types/${leafId}`);
      }
      if (parentId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/reservation-types/${parentId}`);
      }
      await page.close();
    }
  });
});

// ─────────────────────────────────────────────────
// admin account: §6 締め時間 #2/#3/#4/#8（EMR-127d）
// ─────────────────────────────────────────────────

test.describe("V04 設定マスタ §6 締め時間（admin）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("標準締め時間: 空欄は送信されず・境界逆転は PATCH 400 + トースト (#2・#3)", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);

    try {
      await settings.open("/settings/closing-time");
      await expect(settings.heading("締め時間設定")).toBeVisible({ timeout: 15000 });

      const standardSection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "標準締め時間" }) });

      // #2: 必須欄を空にして保存 → ネイティブ required で PATCH は発行されない。
      await page.locator("#closing_am_pm_boundary").fill("");
      const blockedPatchPromise = page.waitForRequest(
        (request) => request.url().includes("/closing-settings") && request.method() === "PATCH",
        { timeout: 5000 },
      );
      await standardSection.getByRole("button", { name: "保存" }).click();
      await expect(
        blockedPatchPromise,
        "required empty field must not issue PATCH /closing-settings",
      ).rejects.toThrow();
      await expect(page.locator("#closing_am_pm_boundary")).toHaveJSProperty(
        "validity.valueMissing",
        true,
      );

      // #3: 境界逆転（区切り 14:00・平日終了 12:00）→ PATCH 400 + トースト。
      await page.locator("#closing_am_pm_boundary").fill("14:00");
      await page.locator("#closing_weekday_end").fill("12:00");
      const patchResponsePromise = page.waitForResponse(
        (response) =>
          response.url().includes("/closing-settings") && response.request().method() === "PATCH",
        { timeout: 15000 },
      );
      await standardSection.getByRole("button", { name: "保存" }).click();
      expect(
        (await patchResponsePromise).status(),
        "reversed boundary PATCH must be rejected",
      ).toBe(400);
      await expect(settings.toast()).toContainText("境界時刻", { timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test("個別休診日: 日付空では追加されない (#4)", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);

    try {
      await settings.open("/settings/closing-time");
      await expect(settings.heading("締め時間設定")).toBeVisible({ timeout: 15000 });

      const holidaySection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "個別休診日" }) });
      await holidaySection.getByRole("button", { name: "新規登録" }).click();
      await expect(page.locator("#holiday_date")).toBeVisible({ timeout: 10000 });

      const blockedPostPromise = page.waitForRequest(
        (request) =>
          request.url().includes("/closing-settings/holidays") && request.method() === "POST",
        { timeout: 5000 },
      );
      await holidaySection.getByRole("button", { name: "追加" }).click();
      await expect(
        blockedPostPromise,
        "empty date must not issue POST /closing-settings/holidays",
      ).rejects.toThrow();
      await expect(page.locator("#holiday_date")).toHaveJSProperty("validity.valueMissing", true);
    } finally {
      await page.close();
    }
  });

  test("特別期間: 開始日>終了日・境界>=終了時刻は POST 400 + トースト (#8)", async () => {
    test.setTimeout(120000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);

    try {
      await settings.open("/settings/closing-time");
      await expect(settings.heading("締め時間設定")).toBeVisible({ timeout: 15000 });

      const periodSection = page
        .locator("section")
        .filter({ has: page.getByRole("heading", { name: "特別期間" }) });
      await periodSection.getByRole("button", { name: "新規登録" }).click();
      await expect(page.locator("#start_date")).toBeVisible({ timeout: 10000 });
      // MasterSidePanel の <form action noValidate> — 必須空欄でも submit する。
      const periodForm = page.locator("form").filter({ has: page.locator("#start_date") });

      // 開始日 > 終了日 → POST 400「開始日は終了日以前に設定してください」。
      await page.locator("#start_date").fill("2099-12-20");
      await page.locator("#end_date").fill("2099-12-19");
      await page.locator("#am_pm_boundary").fill("13:00");
      await page.locator("#pm_end").fill("19:00");
      const postStartGtEndPromise = page.waitForResponse(
        (response) =>
          response.url().includes("/closing-settings/special-periods") &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await periodForm.getByRole("button", { name: "保存" }).click();
      expect((await postStartGtEndPromise).status(), "start>end POST must be rejected").toBe(400);
      await expect(settings.toast().filter({ hasText: "開始日は終了日以前" })).toBeVisible({
        timeout: 10000,
      });

      // 区切り時刻 >= 終了時刻 → POST 400「…は境界時刻…より後に設定してください」。
      // 直前のトーストが残っているため hasText で絞る（strict mode 回避）。
      // エラー時もフォームは閉じないため、そのまま値を入れ直して再送する。
      await page.locator("#start_date").fill("2099-12-10");
      await page.locator("#end_date").fill("2099-12-19");
      await page.locator("#am_pm_boundary").fill("19:00");
      await page.locator("#pm_end").fill("13:00");
      const postBoundaryPromise = page.waitForResponse(
        (response) =>
          response.url().includes("/closing-settings/special-periods") &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await periodForm.getByRole("button", { name: "保存" }).click();
      expect((await postBoundaryPromise).status(), "boundary>=end POST must be rejected").toBe(400);
      await expect(settings.toast().filter({ hasText: "境界時刻" })).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});

// ─────────────────────────────────────────────────
// admin account: §7 シフトパターン #4/#5/#6（EMR-127d）
// ─────────────────────────────────────────────────

test.describe("V04 設定マスタ §7 シフトテンプレート（admin）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("シフトテンプレート: 休憩2件・勤務時間外休憩も保存され再オープンで保持 (#4・#5)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const name = disposableName("シフト");
    let templateId: number | null = null;

    try {
      await settings.open("/settings/shift-templates");
      await expect(settings.heading("シフトテンプレートマスタ")).toBeVisible({ timeout: 15000 });

      await settings.newButton().click();
      await expect(page.getByLabel("テンプレート名")).toBeVisible({ timeout: 10000 });
      await page.getByLabel("テンプレート名").fill(name);
      await page.getByLabel("開始時刻", { exact: true }).fill("09:00");
      await page.getByLabel("終了時刻", { exact: true }).fill("18:00");

      // #4: 休憩を 2 件追加。2 件目は勤務時間外（19:00–20:00）— #5 により受理される契約。
      const breakAddButton = page.getByRole("button", { name: "追加", exact: true });
      await breakAddButton.click();
      await expect(page.getByLabel("休憩1 開始時刻")).toBeVisible();
      await breakAddButton.click();
      await expect(page.getByLabel("休憩2 開始時刻")).toBeVisible();
      await page.getByLabel("休憩1 開始時刻").fill("12:00");
      await page.getByLabel("休憩1 終了時刻").fill("13:00");
      await page.getByLabel("休憩2 開始時刻").fill("19:00");
      await page.getByLabel("休憩2 終了時刻").fill("20:00");

      const postPromise = page.waitForResponse(
        (response) =>
          response.url().includes("/v1/shift-templates") && response.request().method() === "POST",
        { timeout: 15000 },
      );
      await page.getByRole("button", { name: "保存", exact: true }).click();
      const postResponse = await postPromise;
      expect(postResponse.status(), "shift template POST must succeed").toBe(201);

      // #5: レスポンスの breaks に勤務時間外の 19:00–20:00 が保存されている。
      const createdBody: unknown = await postResponse.json();
      const createdRow =
        isRecord(createdBody) && isRecord(createdBody.data) ? createdBody.data : createdBody;
      templateId = readRowId(createdRow);
      const createdBreaks =
        isRecord(createdRow) && Array.isArray(createdRow.breaks) ? createdRow.breaks : [];
      expect(createdBreaks.length, "two breaks must be persisted").toBe(2);
      expect(
        createdBreaks.some(
          (b) =>
            isRecord(b) &&
            String(b.break_start).startsWith("19:00") &&
            String(b.break_end).startsWith("20:00"),
        ),
        "out-of-hours break 19:00–20:00 must be persisted",
      ).toBe(true);

      await expect(settings.toast()).toContainText("テンプレートを作成しました", {
        timeout: 10000,
      });
      await expect(page.getByLabel("テンプレート名")).not.toBeVisible({ timeout: 10000 });
      await expect(settings.rowContaining(name)).toBeVisible({ timeout: 10000 });

      // #4: パネル再オープン → 休憩 2 件が初期表示で保持される。
      await settings.rowActionButton(name).click();
      await expect(page.getByLabel("テンプレート名")).toHaveValue(name, { timeout: 10000 });
      await expect(page.getByLabel("休憩1 開始時刻")).toHaveValue(/^12:00/, { timeout: 10000 });
      await expect(page.getByLabel("休憩1 終了時刻")).toHaveValue(/^13:00/);
      await expect(page.getByLabel("休憩2 開始時刻")).toHaveValue(/^19:00/);
      await expect(page.getByLabel("休憩2 終了時刻")).toHaveValue(/^20:00/);
    } finally {
      if (templateId === null) {
        const listResponse = await v04Api(
          page.request,
          v04.clinicId,
          "GET",
          "/shift-templates",
        ).catch(() => null);
        if (listResponse !== null && listResponse.ok()) {
          templateId = findRowId(listRows(await listResponse.json()), name);
        }
      }
      if (templateId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/shift-templates/${templateId}`);
      }
      await page.close();
    }
  });

  test("シフトテンプレート: 同名登録は 409 + トーストで拒否 (#6)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const name = disposableName("シフト同名");
    let templateId: number | null = null;

    try {
      // uk_shift_templates_clinic_name の対象行を API でシード。
      templateId = await seedMasterRow(page.request, v04.clinicId, "/shift-templates", {
        name,
        shift_type: "full",
        start_time: "09:00",
        end_time: "18:00",
        is_active: true,
        breaks: [],
      });

      await settings.open("/settings/shift-templates");
      await expect(settings.heading("シフトテンプレートマスタ")).toBeVisible({ timeout: 15000 });
      await expect(settings.rowContaining(name)).toBeVisible({ timeout: 10000 });

      // 同名で新規作成 → POST 409 + トースト。パネルは開いたまま、行は増えない。
      await settings.newButton().click();
      await expect(page.getByLabel("テンプレート名")).toBeVisible({ timeout: 10000 });
      await page.getByLabel("テンプレート名").fill(name);
      await page.getByLabel("開始時刻", { exact: true }).fill("10:00");
      await page.getByLabel("終了時刻", { exact: true }).fill("19:00");

      const dupPostPromise = page.waitForResponse(
        (response) =>
          response.url().includes("/v1/shift-templates") && response.request().method() === "POST",
        { timeout: 15000 },
      );
      await page.getByRole("button", { name: "保存", exact: true }).click();
      expect((await dupPostPromise).status(), "duplicate name POST must be rejected").toBe(409);
      await expect(settings.toast()).toContainText("既に使用されています", { timeout: 10000 });
      await expect(page.getByLabel("テンプレート名")).toBeVisible();
      await expect(
        settings.rowContaining(name),
        "duplicate-name create must not add a second row",
      ).toHaveCount(1);
    } finally {
      if (templateId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/shift-templates/${templateId}`);
      }
      await page.close();
    }
  });
});

// ─────────────────────────────────────────────────
// admin account: §2 診療項目 5 タブ（EMR-127d）
// ─────────────────────────────────────────────────

test.describe("V04 設定マスタ §2 診療項目（admin）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("診療項目: 4タブを UI で作成・別タブへの同名登録は受理 (#4)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const sharedName = disposableName("同名項目");
    const created: { path: string; id: number }[] = [];

    /** 指定タブへ遷移して名称+価格だけで新規作成し、採番 id を返す。 */
    async function createViaUi(
      tabLabel: string,
      tabValue: string,
      apiPath: string,
      name: string,
    ): Promise<number> {
      await settings.tab(tabLabel).click();
      // 既定タブ（consultation）は URL に tab= が付かないため aria-selected で確認する。
      if (tabValue === "consultation") {
        await expect(settings.tab(tabLabel)).toHaveAttribute("aria-selected", "true");
      } else {
        await expect(page).toHaveURL(new RegExp(`tab=${tabValue}`), { timeout: 10000 });
      }
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(name);
      await settings.medicinePriceInput().fill("1500");
      const postPromise = page.waitForResponse(
        (response) =>
          response.url().includes(`/v1${apiPath}`) &&
          response.request().method() === "POST" &&
          !response.url().includes("/reorder"),
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      const postResponse = await postPromise;
      expect(postResponse.status(), `${tabLabel} create POST must succeed`).toBe(201);
      const id = readRowId(await postResponse.json());
      expect(id, `${tabLabel} create response must carry an id`).not.toBeNull();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(
        settings.rowContaining(name),
        `${tabLabel} row must appear in the list`,
      ).toBeVisible({ timeout: 10000 });
      return id as number;
    }

    try {
      await settings.open("/settings/treatment-items");
      await expect(settings.heading("診療項目マスタ")).toBeVisible({ timeout: 15000 });

      // 診察タブで同名行を作成（別タブ同名の比較元。診察の UI 作成は R 2 済みだが
      // ここではクロスタブ受理の前提行として使う）。
      const consultationId = await createViaUi(
        "診察",
        "consultation",
        "/masters/consultations",
        sharedName,
      );
      created.push({ path: "/masters/consultations", id: consultationId });

      // #4: 別タブ（検査）に同名 → 受理される（タブごとに別テーブルで一意）。
      const examinationId = await createViaUi(
        "検査",
        "examination",
        "/masters/examination-types",
        sharedName,
      );
      created.push({ path: "/masters/examination-types", id: examinationId });

      // 残り 3 タブの UI 作成（各タブ固有 API へ POST 201）。
      for (const { label, value, apiPath, kind } of [
        { label: "処置", value: "procedure", apiPath: "/masters/procedures", kind: "処置" },
        { label: "予防接種", value: "vaccine", apiPath: "/masters/vaccines", kind: "予防接種" },
        { label: "定期健診", value: "checkup", apiPath: "/masters/checkup-types", kind: "健診" },
      ]) {
        const id = await createViaUi(label, value, apiPath, disposableName(kind));
        created.push({ path: apiPath, id });
      }
    } finally {
      for (const { path, id } of created) {
        await deleteSeededRow(page.request, v04.clinicId, `${path}/${id}`);
      }
      await page.close();
    }
  });

  test("診療項目: 処置タブの親子階層 — 親セレクタ実データ・ツリー表示・親変更禁止 (#5・#6)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const parentName = disposableName("処置親");
    const childName = disposableName("処置子");
    let parentId: number | null = null;
    let childId: number | null = null;

    const postProcedure = () =>
      page.waitForResponse(
        (response) =>
          response.url().includes("/v1/masters/procedures") &&
          response.request().method() === "POST" &&
          !response.url().includes("/reorder"),
        { timeout: 15000 },
      );
    const parentCategoryRow = () =>
      page.getByText("親カテゴリ", { exact: true }).locator("xpath=..");

    try {
      await settings.open("/settings/treatment-items?tab=procedure");
      await expect(settings.heading("診療項目マスタ")).toBeVisible({ timeout: 15000 });
      await expect(settings.tab("処置")).toHaveAttribute("data-state", "active");

      // 親項目を UI で作成。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(parentName);
      let postPromise = postProcedure();
      await settings.saveButton().click();
      const parentResponse = await postPromise;
      expect(parentResponse.status(), "parent create must succeed").toBe(201);
      parentId = readRowId(await parentResponse.json());
      expect(parentId).not.toBeNull();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // #5: 子項目を作成し、親カテゴリセレクタ（同タブ実データ由来）で親を指定。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(childName);
      await parentCategoryRow().getByRole("combobox").click();
      await page.getByRole("option", { name: parentName }).click();
      postPromise = postProcedure();
      await settings.saveButton().click();
      const childResponse = await postPromise;
      expect(childResponse.status(), "child create must succeed").toBe(201);
      const childBody = await childResponse.json();
      childId = readRowId(childBody);
      expect(childId).not.toBeNull();
      const childRow = isRecord(childBody) && isRecord(childBody.data) ? childBody.data : childBody;
      expect(
        Number(isRecord(childRow) ? childRow.parent_id : null),
        "created child must carry the parent_id",
      ).toBe(parentId);
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // #5: 一覧で親行を展開すると子項目がツリー表示される。
      await page.getByRole("button", { name: new RegExp(`${parentName}.*子項目を展開`) }).click();
      await expect(settings.rowContaining(childName)).toBeVisible({ timeout: 10000 });

      // #6: 子を持つ親を編集 → 親カテゴリは変更不可表示でセレクタなし。
      await settings.rowActionButton(parentName).click();
      await expect(settings.masterTitleInput()).toHaveValue(parentName, { timeout: 10000 });
      await expect(parentCategoryRow().getByText("子項目があるため変更できません")).toBeVisible();
      await expect(parentCategoryRow().getByRole("combobox")).toHaveCount(0);
      await settings.cancelButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
    } finally {
      if (childId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/procedures/${childId}`);
      }
      if (parentId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/procedures/${parentId}`);
      }
      await page.close();
    }
  });
});

// ─────────────────────────────────────────────────
// admin account: §1 標準マスタ C1/C2 系（EMR-127d）
// ─────────────────────────────────────────────────

test.describe("V04 設定マスタ §1 標準マスタ（admin）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("動物種類: 空名は送信されず・グローバル一意の同名は 409 + トースト (C1-1・C3-2)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const speciesName = disposableName("種");
    let speciesId: number | null = null;

    try {
      await settings.open("/settings/animal-species");
      await expect(settings.heading("動物種類マスタ")).toBeVisible({ timeout: 15000 });

      // C1-1: 名称を空のまま保存 → FE バリデーションで POST は発行されない。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      const blockedPostPromise = page.waitForRequest(
        (request) =>
          request.url().includes("/masters/animal-species") && request.method() === "POST",
        { timeout: 5000 },
      );
      await settings.saveButton().click();
      await expect(
        blockedPostPromise,
        "empty name must not issue POST /masters/animal-species",
      ).rejects.toThrow();
      await expect(page.getByText("名称を入力してください")).toBeVisible();

      // C3-2: グローバル一意（clinic_id なし・WHERE is_active=true）— 既存名で作成は 409。
      speciesId = await seedMasterRow(page.request, v04.clinicId, "/masters/animal-species", {
        name: speciesName,
        is_active: true,
      });
      await settings.masterTitleInput().fill(speciesName);
      const conflictPromise = page.waitForResponse(
        (response) =>
          response.url().includes("/masters/animal-species") &&
          response.request().method() === "POST",
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      expect(
        (await conflictPromise).status(),
        "duplicate animal species name must be rejected with 409",
      ).toBe(409);
      await expect(settings.toast().filter({ hasText: "は既に使用されています" })).toBeVisible({
        timeout: 10000,
      });
    } finally {
      if (speciesId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/animal-species/${speciesId}`);
      }
      await page.close();
    }
  });

  test("予約区分グループ: 空名は送信されず・インラインエラー (C1-1)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);

    try {
      await settings.open("/settings/reservation-type");
      await expect(settings.heading("予約区分マスタ")).toBeVisible({ timeout: 15000 });

      await page.getByRole("button", { name: "グループを追加" }).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      const blockedPostPromise = page.waitForRequest(
        (request) =>
          request.url().includes("/masters/reservation-type-groups") && request.method() === "POST",
        { timeout: 5000 },
      );
      await settings.saveButton().click();
      await expect(
        blockedPostPromise,
        "empty group name must not issue POST /masters/reservation-type-groups",
      ).rejects.toThrow();
      await expect(page.getByText("名称を入力してください")).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("保険: 補償率 -1 は送信されず・編集保存が永続する (C1-3・C2)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const insuranceName = disposableName("保険");
    let insuranceId: number | null = null;

    try {
      insuranceId = await seedMasterRow(page.request, v04.clinicId, "/masters/insurances", {
        name: insuranceName,
        coverage_rate: 50,
        is_active: true,
      });

      await settings.open("/settings/insurance");
      await expect(settings.heading("保険マスタ")).toBeVisible({ timeout: 15000 });

      // C2: 補償率を変更して保存 → PATCH 200 → 再オープンで保持。
      await settings.rowActionButton(insuranceName).click();
      await expect(settings.masterTitleInput()).toHaveValue(insuranceName, { timeout: 10000 });
      await page.getByLabel("補償率(%)").fill("60");
      const patchPromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/insurances\/\d+/.test(response.url()) &&
          response.request().method() === "PATCH",
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      expect((await patchPromise).status(), "insurance update PATCH must succeed").toBe(200);
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      await settings.rowActionButton(insuranceName).click();
      await expect(page.getByLabel("補償率(%)")).toHaveValue("60", { timeout: 10000 });
      await settings.cancelButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // C1-3: 補償率 -1 → FE バリデーション（BE と同一境界 0〜100）で POST されず
      // インラインエラーが出る（HTML min による無音ブロックではない）。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(disposableName("保険"));
      await page.getByLabel("補償率(%)").fill("-1");
      const blockedPostPromise = page.waitForRequest(
        (request) => request.url().includes("/masters/insurances") && request.method() === "POST",
        { timeout: 5000 },
      );
      await settings.saveButton().click();
      await expect(
        blockedPostPromise,
        "coverage_rate -1 must not issue POST /masters/insurances",
      ).rejects.toThrow();
      await expect(page.getByText("補償率は0〜100の範囲で入力してください")).toBeVisible();
      await expect(page.getByLabel("補償率(%)")).toHaveAttribute("aria-invalid", "true");
    } finally {
      if (insuranceId !== null) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/insurances/${insuranceId}`);
      }
      await page.close();
    }
  });

  test("物販・商品: 無効化した行と同名を再登録できる (is_active 部分一意)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const itemName = disposableName("品目");
    const createdIds: number[] = [];

    const postMerchandise = () =>
      page.waitForResponse(
        (response) =>
          response.url().includes("/v1/masters/merchandise-items") &&
          response.request().method() === "POST" &&
          !response.url().includes("/reorder"),
        { timeout: 15000 },
      );

    try {
      await settings.open("/settings/merchandise-items");
      await expect(settings.heading("商品マスタ")).toBeVisible({ timeout: 15000 });

      // UI で作成（名称のみ — カテゴリ/価格/税率は既定値）。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(itemName);
      let postPromise = postMerchandise();
      await settings.saveButton().click();
      const createResponse = await postPromise;
      expect(createResponse.status(), "merchandise create must succeed").toBe(201);
      const firstId = readRowId(await createResponse.json());
      expect(firstId).not.toBeNull();
      createdIds.push(firstId as number);
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // 無効化 → (clinic_id,name) WHERE is_active=true の一意制約から外れる。
      await settings.rowActionButton(itemName).click();
      await expect(settings.masterTitleInput()).toHaveValue(itemName, { timeout: 10000 });
      await page.getByLabel("ステータスを切り替え").click();
      const patchPromise = page.waitForResponse(
        (response) =>
          /\/v1\/masters\/merchandise-items\/\d+/.test(response.url()) &&
          response.request().method() === "PATCH",
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      expect((await patchPromise).status(), "deactivate PATCH must succeed").toBe(200);
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // 同名を再登録 → 受理される（部分一意のため無効行とは衝突しない）。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      await settings.masterTitleInput().fill(itemName);
      postPromise = postMerchandise();
      await settings.saveButton().click();
      const reRegisterResponse = await postPromise;
      expect(
        reRegisterResponse.status(),
        "re-registering a disabled row's name must be accepted",
      ).toBe(201);
      const secondId = readRowId(await reRegisterResponse.json());
      expect(secondId).not.toBeNull();
      createdIds.push(secondId as number);
      await expect(settings.rowContaining(itemName).first()).toBeVisible({ timeout: 10000 });
    } finally {
      for (const id of createdIds) {
        await deleteSeededRow(page.request, v04.clinicId, `/masters/merchandise-items/${id}`);
      }
      await page.close();
    }
  });

  test("支払方法: 空名は送信されず・標準行の名称変更は可・無効化は 409 (C1-1・C2・特記)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const originalName = "現金";
    const renamedName = disposableName("支払");
    let methodId: number | null = null;

    try {
      await settings.open("/settings/payment-methods");
      await expect(settings.heading("支払方法マスタ")).toBeVisible({ timeout: 15000 });

      // C1-1: 名称を空のまま保存 → FE バリデーションで POST は発行されない。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      const blockedPostPromise = page.waitForRequest(
        (request) => request.url().includes("/payment-methods") && request.method() === "POST",
        { timeout: 5000 },
      );
      await settings.saveButton().click();
      await expect(
        blockedPostPromise,
        "empty name must not issue POST /payment-methods",
      ).rejects.toThrow();
      await expect(page.getByText("名称を入力してください")).toBeVisible();
      await settings.cancelButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });

      // C2: システム標準行「現金」の名称変更は許可される（system_key 保持行・名称は可）。
      await settings.rowActionButton(originalName).click();
      await expect(settings.masterTitleInput()).toHaveValue(originalName, { timeout: 10000 });
      await settings.masterTitleInput().fill(renamedName);
      const renamePromise = page.waitForResponse(
        (response) =>
          /\/v1\/payment-methods\/\d+/.test(response.url()) &&
          response.request().method() === "PATCH",
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      const renameResponse = await renamePromise;
      expect(renameResponse.status(), "renaming a system row must be allowed").toBe(200);
      methodId = readRowId(await renameResponse.json());
      expect(methodId).not.toBeNull();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(settings.rowContaining(renamedName)).toBeVisible({ timeout: 10000 });

      // 特記: システム標準行の無効化は拒否される（BE Conflict）。
      await settings.rowActionButton(renamedName).click();
      await expect(settings.masterTitleInput()).toHaveValue(renamedName, { timeout: 10000 });
      await page.getByLabel("ステータスを切り替え").click();
      const deactivatePromise = page.waitForResponse(
        (response) =>
          /\/v1\/payment-methods\/\d+/.test(response.url()) &&
          response.request().method() === "PATCH",
        { timeout: 15000 },
      );
      await settings.saveButton().click();
      expect(
        (await deactivatePromise).status(),
        "deactivating a system row must be rejected with 409",
      ).toBe(409);
      await expect(settings.toast().filter({ hasText: "無効化できません" })).toBeVisible({
        timeout: 10000,
      });
    } finally {
      if (methodId !== null) {
        try {
          await v04Api(page.request, v04.clinicId, "PATCH", `/payment-methods/${methodId}`, {
            name: originalName,
          });
        } catch {
          // Best-effort restore only.
        }
      }
      await page.close();
    }
  });

  test("ケージ: enum 不正値は POST 400 で拒否される (BE oneof)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");

    // 種別/サイズは UI では select のため不正値を送れない → API で直接検証する。
    const name = disposableName("ケージ");
    for (const body of [
      { name, cage_type: "dragon", cage_size: "small" },
      { name, cage_type: "dog", cage_size: "huge" },
    ]) {
      const response = await v04Api(context.request, v04.clinicId, "POST", "/masters/cages", body);
      expect(response.status(), `POST /masters/cages ${JSON.stringify(body)} must be 400`).toBe(
        400,
      );
    }
  });
});

// ─────────────────────────────────────────────────
// view-only account: §1 追加 6 画面（EMR-127f）
// ─────────────────────────────────────────────────

/**
 * EMR-127f で追加する view-only 画面。`ViewOnlyMasterScreen` と同じ観点に、
 * 行ラベルのフィールド名が `name` ではないケース（問診テンプレートの `title`）と
 * FK が要るケース（診断病名の `diagnosis_type_id`）を載せる拡張。
 */
interface ViewOnlyMasterScreenExtra extends ViewOnlyMasterScreen {
  /** 行ラベルとして使うフィールド名（既定 `"name"`） */
  seedNameField?: string;
  /** 代表行のシードを独自に行う（FK 行が要る場合）。戻り値は削除対象パス列 */
  seedRow?: (
    request: APIRequestContext,
    clinicId: number,
    rowName: string,
  ) => Promise<{ rowId: number; cleanupPaths: string[] }>;
}

const VIEW_ONLY_MASTER_SCREENS_EXTRA: ViewOnlyMasterScreenExtra[] = [
  {
    resource: "master-medical",
    screen: "診断カテゴリ",
    path: "/settings/diagnosis?tab=diagnosis_type",
    heading: "診断マスタ",
    entityLabel: "診断カテゴリ",
    apiPath: "/masters/diagnosis-types",
    seedBody: { is_active: true },
  },
  {
    resource: "master-medical",
    screen: "診断病名",
    path: "/settings/diagnosis?tab=diagnosis_name",
    heading: "診断マスタ",
    entityLabel: "診断病名",
    apiPath: "/masters/diagnosis-names",
    // 病名は diagnosis_type_id が必須 — 先にカテゴリをシードし、病名→カテゴリの順で削除する。
    seedRow: async (request, clinicId, rowName) => {
      const typeId = await seedMasterRow(request, clinicId, "/masters/diagnosis-types", {
        name: `${rowName}-category`,
        is_active: true,
      });
      const nameId = await seedMasterRow(request, clinicId, "/masters/diagnosis-names", {
        name: rowName,
        diagnosis_type_id: typeId,
        is_active: true,
      });
      return {
        rowId: nameId,
        cleanupPaths: [`/masters/diagnosis-names/${nameId}`, `/masters/diagnosis-types/${typeId}`],
      };
    },
  },
  {
    resource: "master-medical",
    screen: "主訴種別",
    path: "/settings/interview/chief-complaint",
    heading: "主訴マスタ",
    entityLabel: "主訴",
    apiPath: "/masters/chief-complaint-types",
    seedBody: { is_active: true },
  },
  {
    resource: "master-medical",
    screen: "問診・定型文テンプレート",
    path: "/settings/inquiry-templates",
    heading: "問診テンプレートマスタ",
    entityLabel: "問診テンプレート",
    apiPath: "/masters/inquiry-templates",
    // 一覧行のラベルは name ではなく title。
    seedNameField: "title",
    seedBody: { category: "chief_complaint", is_active: true },
  },
  {
    resource: "master-trimming",
    screen: "トリミングオプション",
    path: "/settings/trimming?tab=option",
    heading: "トリミングマスタ",
    entityLabel: "トリミングオプション",
    apiPath: "/masters/trimming-options",
    seedBody: { is_active: true },
  },
  {
    resource: "master-trimming",
    screen: "トリミングコース種別",
    path: "/settings/trimming-course-type",
    heading: "コース種別マスタ",
    entityLabel: "コース種別",
    apiPath: "/masters/trimming-course-types",
    seedBody: { is_active: true },
  },
];

test.describe("V04 権限: view のみ account 追加画面（EMR-127f）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext | undefined;
  // シード・後始末用の admin session（view-only account では行を作れない）。
  let adminContext: BrowserContext | undefined;

  test.beforeAll(async ({ browser }) => {
    if (v04 === null) return;
    context = await createViewOnlyContext(browser, v04);
    adminContext = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context?.close();
    await adminContext?.close();
  });

  for (const master of VIEW_ONLY_MASTER_SCREENS_EXTRA) {
    test(`${master.screen}: 一覧と詳細は閲覧可・作成/保存/削除の導線なし・API write は 403`, async () => {
      test.setTimeout(120000);
      if (v04 === null) throw new Error("v04 fixture unavailable");
      if (adminContext === undefined) throw new Error("admin context unavailable");
      expect(v04.viewOnlyResources, `fixture must grant ${master.resource} view`).toContain(
        master.resource,
      );
      const page = await openViewOnlyPage(context, v04);
      let cleanupPaths: string[] = [];
      try {
        // 一覧の読取 + 代表行を admin でシード（FK が要る画面は seedRow に委譲）。
        const list = await v04Api(page.request, v04.clinicId, "GET", master.apiPath);
        expect(list.status(), `GET ${master.apiPath} must be allowed for view`).toBe(200);

        const rowName = `V04-e2e-${master.entityLabel}-${Date.now()}`;
        let rowId: number;
        if (master.seedRow !== undefined) {
          const seeded = await master.seedRow(adminContext.request, v04.clinicId, rowName);
          rowId = seeded.rowId;
          cleanupPaths = seeded.cleanupPaths;
        } else {
          rowId = await seedMasterRow(adminContext.request, v04.clinicId, master.apiPath, {
            [master.seedNameField ?? "name"]: rowName,
            ...(master.seedBody ?? {}),
          });
          cleanupPaths = [`${master.apiPath}/${rowId}`];
        }

        await page.goto(master.path, { waitUntil: "domcontentloaded" });
        await expect(page.getByRole("heading", { name: master.heading }).first()).toBeVisible({
          timeout: 45000,
        });

        const rowButton = page.getByRole("button", {
          name: `詳細: ${master.entityLabel} ${rowName} (ID ${rowId})`,
        });
        await expect(rowButton, "対象行が一覧に表示されること").toBeVisible({ timeout: 30000 });
        await expect(
          page.getByRole("button", { name: "新規登録" }),
          "view-only account must not get a create affordance",
        ).toHaveCount(0);

        await rowButton.click();
        const title = page.locator("#master-title");
        await expect(title).toBeVisible({ timeout: 15000 });
        await expect(title).toHaveValue(rowName);
        await expect(
          page.getByRole("button", { name: "保存" }),
          "read-only panel must not render a save button",
        ).toHaveCount(0);
        await expect(
          page.getByLabel("削除"),
          "read-only panel must not render a delete button",
        ).toHaveCount(0);

        const deniedName = `V04-e2e-denied-${Date.now()}`;
        const post = await v04Api(page.request, v04.clinicId, "POST", master.apiPath, {
          [master.seedNameField ?? "name"]: deniedName,
          ...(master.seedBody ?? {}),
        });
        expect(post.status(), `POST ${master.apiPath} must be denied`).toBe(403);
        const patch = await v04Api(
          page.request,
          v04.clinicId,
          "PATCH",
          `${master.apiPath}/${rowId}`,
          { [master.seedNameField ?? "name"]: deniedName },
        );
        expect(patch.status(), `PATCH ${master.apiPath}/:id must be denied`).toBe(403);
        const del = await v04Api(
          page.request,
          v04.clinicId,
          "DELETE",
          `${master.apiPath}/${rowId}`,
        );
        expect(del.status(), `DELETE ${master.apiPath}/:id must be denied`).toBe(403);
      } finally {
        if (adminContext !== undefined) {
          for (const path of cleanupPaths) {
            await deleteSeededRow(adminContext.request, v04.clinicId, path);
          }
        }
        await page.close();
      }
    });
  }
});

// ─────────────────────────────────────────────────
// admin account: §1 C2 追加（職種・トリミングコース種別）（EMR-127f）
// ─────────────────────────────────────────────────

/** C2 更新永続の対象画面。seed→名称変更→再読込→再オープンの流れは共通。 */
interface C2MasterScreen {
  screen: string;
  path: string;
  heading: string;
  apiPath: string;
}

const C2_MASTER_SCREENS: C2MasterScreen[] = [
  {
    screen: "職種",
    path: "/settings/occupations",
    heading: "職種マスタ",
    apiPath: "/masters/occupations",
  },
  {
    screen: "トリミングコース種別",
    path: "/settings/trimming-course-type",
    heading: "コース種別マスタ",
    apiPath: "/masters/trimming-course-types",
  },
];

test.describe("V04 設定マスタ §1 C2 更新永続（admin）（EMR-127f）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  for (const master of C2_MASTER_SCREENS) {
    test(`${master.screen}: 名称変更が保存・再読込・再オープンで永続する (C2)`, async () => {
      test.setTimeout(120000);
      if (v04 === null) throw new Error("v04 fixture unavailable");
      const page = await context.newPage();
      const settings = new SettingsMasterPage(page);
      const originalName = disposableName(master.screen);
      const renamedName = disposableName(`${master.screen}改`);
      let rowId: number | null = null;

      try {
        rowId = await seedMasterRow(page.request, v04.clinicId, master.apiPath, {
          name: originalName,
          is_active: true,
        });

        await settings.open(master.path);
        await expect(settings.heading(master.heading)).toBeVisible({ timeout: 15000 });

        // C2-1: 名称を変更して保存 → PATCH 200。
        await settings.rowActionButton(originalName).click();
        await expect(settings.masterTitleInput()).toHaveValue(originalName, { timeout: 10000 });
        await settings.masterTitleInput().fill(renamedName);
        const patchPromise = page.waitForResponse(
          (response) =>
            new RegExp(`/v1${master.apiPath}/\\d+`).test(response.url()) &&
            response.request().method() === "PATCH",
          { timeout: 15000 },
        );
        await settings.saveButton().click();
        expect((await patchPromise).status(), `${master.screen} rename PATCH must succeed`).toBe(
          200,
        );
        await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
        await expect(settings.rowContaining(renamedName)).toBeVisible({ timeout: 10000 });

        // C2-2/C2-3: ブラウザ再読込 → 行を開き直すと変更後の値が初期表示される。
        await page.reload({ waitUntil: "domcontentloaded" });
        await expect(settings.heading(master.heading)).toBeVisible({ timeout: 15000 });
        await settings.rowActionButton(renamedName).click();
        await expect(settings.masterTitleInput()).toHaveValue(renamedName, { timeout: 10000 });
        await settings.cancelButton().click();
        await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      } finally {
        if (rowId !== null) {
          await deleteSeededRow(page.request, v04.clinicId, `${master.apiPath}/${rowId}`);
        }
        await page.close();
      }
    });
  }
});

// ─────────────────────────────────────────────────
// admin account: §1 C1 追加（トリミングコース）（EMR-127f）
// ─────────────────────────────────────────────────

test.describe("V04 設定マスタ §1 C1 必須バリデーション（admin）（EMR-127f）", () => {
  const v04 = readV04FixtureFromEnv();
  test.skip(v04 === null, "E2E_CLINICAL_FIXTURE 未設定（suite=v04 以外）");

  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("トリミングコース: 空名は送信されず・インラインエラー (C1-1)", async () => {
    test.setTimeout(120000);
    if (v04 === null) throw new Error("v04 fixture unavailable");
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);

    try {
      await settings.open("/settings/trimming?tab=course");
      await expect(settings.heading("トリミングマスタ")).toBeVisible({ timeout: 15000 });

      // C1-1: 名称を空のまま保存 → FE バリデーションで POST は発行されない。
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });
      const blockedPostPromise = page.waitForRequest(
        (request) =>
          request.url().includes("/masters/trimming-courses") && request.method() === "POST",
        { timeout: 5000 },
      );
      await settings.saveButton().click();
      await expect(
        blockedPostPromise,
        "empty name must not issue POST /masters/trimming-courses",
      ).rejects.toThrow();
      await expect(page.getByText("名称を入力してください")).toBeVisible();
      await settings.cancelButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
