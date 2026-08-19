import { expect, test, type BrowserContext, type Page } from "@playwright/test";

import {
  createSyntheticMeResponse,
  SYNTHETIC_CLINICAL_SCENARIOS,
  type SyntheticClinicalKey,
} from "./fixtures/ui-design-clinical";
import { loginAsDemoAdmin } from "./helpers/auth";
import {
  installSyntheticApiInterceptor,
  isBusinessNonGet,
  type SyntheticApiInterceptor,
  type SyntheticEndpoint,
} from "./helpers/synthetic-api";
import {
  assertSyntheticRenderedState,
  inspectPage,
  type SyntheticRenderedAssertion,
} from "./helpers/ui-design-audit";

const VIEWPORTS = [1440, 1200, 800, 500] as const;
const AUDIT_HEIGHT = 900;
const AUDIT_ORIGIN = new URL(
  process.env.PLAYWRIGHT_TEST_BASE_URL ?? "http://localhost:3003",
).origin;

type FixtureKey =
  | "owner"
  | "medicalRecord"
  | "examination"
  | "pet";

interface RouteTemplate {
  label: string;
  path: string;
  fixture?: FixtureKey;
  suffix?: string;
  expectedText?: string;
  renderedPath?: string;
  allowSearch?: boolean;
  useEditableEstimate?: boolean;
  syntheticClinical?: SyntheticClinicalKey;
}

const PUBLIC_ROUTES: RouteTemplate[] = [
  { label: "ログイン", path: "/login" },
  { label: "パスワードを忘れた方", path: "/forgot-password" },
  { label: "パスワード再設定", path: "/reset-password" },
];

const PROTECTED_ROUTES: RouteTemplate[] = [
  { label: "飼主カルテレポート", path: "/owners/:id/report", fixture: "owner", suffix: "/report" },
  { label: "会計一覧", path: "/accounting" },
  { label: "会計ペット選択", path: "/accounting/select-pet" },
  { label: "会計登録", path: "/accounting/new", fixture: "pet" },
  { label: "会計詳細", path: "/accounting/1024128" },
  { label: "レジ締め", path: "/accounting/close" },
  { label: "レジ締め履歴", path: "/accounting/close/history" },
  { label: "月次集計レポート", path: "/accounting/reports" },
  { label: "在庫一覧", path: "/inventory" },
  { label: "在庫登録", path: "/inventory/new" },
  { label: "在庫編集", path: "/inventory/7", expectedText: "在庫切れ" },
  { label: "見積一覧", path: "/estimates" },
  { label: "見積作成", path: "/estimates/new" },
  { label: "見積詳細", path: "/estimates/1", expectedText: "期限切れ" },
  { label: "見積編集", path: "/estimates/1/edit", useEditableEstimate: true },
  { label: "シフトカレンダー", path: "/shifts" },
  { label: "カルテ一覧", path: "/medical-records" },
  { label: "カルテペット選択", path: "/medical-records/select-pet" },
  {
    label: "カルテ作成",
    path: "/medical-records/new",
    syntheticClinical: "medicalRecordCreate",
  },
  { label: "カルテ編集", path: "/medical-records/:id", fixture: "medicalRecord" },
  { label: "入院・ホテル一覧", path: "/hospitalization" },
  { label: "入院・ホテル ペット選択", path: "/hospitalization/select-pet" },
  { label: "入院・ホテル登録", path: "/hospitalization/new", fixture: "pet" },
  {
    label: "入院・ホテル詳細",
    path: "/hospitalization/:id",
    syntheticClinical: "hospitalizationDetail",
  },
  {
    label: "入院・ホテル編集",
    path: "/hospitalization/:id/edit",
    syntheticClinical: "hospitalizationEdit",
  },
  { label: "トリミング一覧", path: "/trimming" },
  { label: "トリミング ペット選択", path: "/trimming/select-pet" },
  { label: "トリミング登録", path: "/trimming/new", fixture: "pet" },
  {
    label: "トリミング編集",
    path: "/trimming/:id",
    syntheticClinical: "trimmingEdit",
  },
  { label: "検査一覧", path: "/examinations" },
  { label: "検査ペット選択", path: "/examinations/select-pet" },
  { label: "検査編集", path: "/examinations/:id", fixture: "examination" },
  { label: "検査受信", path: "/lab-device" },
  { label: "ワクチン一覧", path: "/vaccinations" },
  { label: "ワクチン ペット選択", path: "/vaccinations/select-pet" },
  { label: "ワクチン登録", path: "/vaccinations/new", fixture: "pet" },
  {
    label: "ワクチン編集",
    path: "/vaccinations/:id",
    syntheticClinical: "vaccinationEdit",
  },
  { label: "定期健診一覧", path: "/checkups" },
  { label: "定期健診ペット選択", path: "/checkups/select-pet" },
  { label: "定期健診登録", path: "/checkups/new", fixture: "pet" },
  { label: "受付", path: "/" },
  { label: "飼主一覧", path: "/owners" },
  { label: "飼主登録", path: "/owners/new" },
  { label: "飼主編集", path: "/owners/:id", fixture: "owner" },
  { label: "集計ダッシュボード", path: "/aggregation", renderedPath: "/aggregation?tab=revenue" },
  { label: "予約管理", path: "/reservations" },
  { label: "Lステップ健診連携", path: "/lstep/checkup-sync" },
  { label: "Lステップ配信モニター", path: "/lstep/delivery-monitor" },
  { label: "Lステップ分析", path: "/lstep/analytics" },
  { label: "LINE予約設定 index", path: "/line-reservation" },
  { label: "LINE予約設定", path: "/line-reservation/settings" },
  { label: "LINE予約ページエディタ", path: "/line-reservation/page-editor" },
  { label: "LINE予約枠設定", path: "/line-reservation/slots", allowSearch: true },
  { label: "医院マスタ設定", path: "/settings/clinic" },
  {
    label: "マニュアルトップ",
    path: "/manual",
    renderedPath: "/manual/screens/00-overview",
  },
  { label: "マニュアル記事", path: "/manual/screens/00-overview" },
  { label: "設定トップ", path: "/settings" },
  { label: "職員マスタ", path: "/settings/staff" },
  { label: "診療項目マスタ", path: "/settings/treatment-items" },
  { label: "診断名マスタ", path: "/settings/diagnosis" },
  { label: "動物種マスタ", path: "/settings/animal-species" },
  { label: "トリミングマスタ", path: "/settings/trimming" },
  { label: "トリミングコース種別", path: "/settings/trimming-course-type" },
  { label: "薬剤マスタ", path: "/settings/medicine" },
  { label: "予約種別マスタ", path: "/settings/reservation-type" },
  { label: "入院・ホテルマスタ", path: "/settings/hospitalization" },
  { label: "ケージマスタ", path: "/settings/cage" },
  { label: "物販品マスタ", path: "/settings/merchandise-items" },
  { label: "保険マスタ", path: "/settings/insurance" },
  { label: "職種マスタ", path: "/settings/occupations" },
  { label: "権限グループマスタ", path: "/settings/permission-groups" },
  { label: "問診テンプレート", path: "/settings/inquiry-templates" },
  { label: "主訴マスタ", path: "/settings/interview/chief-complaint" },
  { label: "問診テンプレート（interview）", path: "/settings/interview/templates" },
  { label: "シフトテンプレート", path: "/settings/shift-templates" },
  { label: "締め時間設定", path: "/settings/closing-time" },
  { label: "支払方法マスタ", path: "/settings/payment-methods" },
  { label: "割引キャンペーン", path: "/settings/campaigns" },
  { label: "検査機器マスタ", path: "/settings/lab-device-item-masters" },
  { label: "Lステップ連携設定", path: "/settings/integrations/lstep" },
  { label: "Lステップタグ管理", path: "/settings/lstep/tags" },
  { label: "同一飼主・ペット連携", path: "/identity-links" },
];

type FixturePaths = Partial<Record<FixtureKey, string>>;
type AuthenticatedStorageState = Awaited<ReturnType<BrowserContext["storageState"]>>;

function resolveRoute(template: RouteTemplate, fixtures: FixturePaths): string | undefined {
  if (!template.fixture) return template.path;
  const fixturePath = fixtures[template.fixture];
  if (!fixturePath) return undefined;
  if (template.fixture === "pet") {
    const separator = template.path.includes("?") ? "&" : "?";
    return `${template.path}${separator}petId=${encodeURIComponent(fixturePath)}`;
  }
  return `${fixturePath}${template.suffix ?? ""}`;
}

async function firstDetailPath(page: Page, listPath: string, pattern: RegExp): Promise<string> {
  await page.goto(listPath, { waitUntil: "domcontentloaded" });
  await page.locator("h1").first().waitFor({ state: "visible", timeout: 30000 });
  const hrefs = await page.locator("a[href]").evaluateAll((links) =>
    links.map((link) => link.getAttribute("href")).filter((href): href is string => href !== null),
  );
  const detailPath = hrefs.find((href) => pattern.test(href));
  if (!detailPath) throw new Error(`${listPath}: no production detail link matched ${pattern}`);
  return detailPath;
}

function readPetId(value: unknown): string | undefined {
  if (typeof value !== "object" || value === null) return undefined;
  const record = value as Record<string, unknown>;
  const direct = record.pet_id ?? record.petId;
  if (typeof direct === "string" || typeof direct === "number") return String(direct);
  const pet = record.pet;
  if (typeof pet === "object" && pet !== null) {
    const id = (pet as Record<string, unknown>).id;
    if (typeof id === "string" || typeof id === "number") return String(id);
  }
  return readPetId(record.data);
}

function createEditableEstimateResponse(raw: unknown): unknown {
  if (typeof raw !== "object" || raw === null) {
    throw new Error("/v1/estimates/1 did not return an object for the edit-route audit");
  }
  const record = raw as Record<string, unknown>;
  if ("status" in record) return { ...record, status: "draft" };
  if ("data" in record) return { ...record, data: createEditableEstimateResponse(record.data) };
  throw new Error("/v1/estimates/1 response did not contain an estimate status");
}

async function discoverFixtures(page: Page): Promise<FixturePaths> {
  const owner = await firstDetailPath(page, "/owners", /^\/owners\/[^/]+$/);
  const medicalRecord = await firstDetailPath(
    page,
    "/medical-records",
    /^\/medical-records\/[^/]+$/,
  );
  const examination = await firstDetailPath(
    page,
    "/examinations",
    /^\/examinations\/[^/]+$/,
  );

  const response = await page.request.get(`/api/v1${medicalRecord}`);
  if (!response.ok()) {
    throw new Error(`${medicalRecord}: fixture detail request failed with HTTP ${response.status()}`);
  }
  const pet = readPetId(await response.json());
  if (!pet) throw new Error(`${medicalRecord}: fixture response did not contain a pet id`);

  return { owner, medicalRecord, examination, pet };
}

interface AuditRouteOptions {
  allowLogin?: boolean;
  allowSearch?: boolean;
  expectedText?: string;
  expectedRenderedPath?: string;
  disabledSelectors?: readonly string[];
  forbiddenButtons?: readonly string[];
  forbiddenGetPaths?: readonly string[];
  getResponses?: Readonly<Record<string, unknown>>;
  meResponse?: unknown;
  syntheticEndpoints?: readonly SyntheticEndpoint[];
  expectedLocalBusinessNonGet?: readonly string[];
  currentClinicId?: string;
  fixedTime?: string;
  renderedAssertions?: readonly SyntheticRenderedAssertion[];
}

interface AuditSignals {
  consoleErrors: readonly string[];
  failedResponses: readonly string[];
  failedRequests: readonly string[];
  pageErrors: readonly string[];
}

function assertRequestSafety(
  path: string,
  context: string,
  interceptor: SyntheticApiInterceptor,
  options: AuditRouteOptions,
  signals: AuditSignals,
): void {
  expect(signals.consoleErrors, `${path} ${context}: console errors`).toEqual([]);
  expect(signals.failedResponses, `${path} ${context}: failed responses`).toEqual([]);
  expect(signals.failedRequests, `${path} ${context}: failed requests`).toEqual([]);
  expect(signals.pageErrors, `${path} ${context}: page errors`).toEqual([]);
  expect(interceptor.ledger.validationFailures, `${path} ${context}: fixture validation`).toEqual([]);
  expect(interceptor.ledger.blocked, `${path} ${context}: unexpected requests`).toEqual([]);
  expect(
    interceptor.ledger.continuedToBackend.filter(isBusinessNonGet),
    `${path} ${context}: continued business non-GET`,
  ).toEqual([]);
  expect(
    interceptor.ledger.locallyFulfilled.filter(isBusinessNonGet),
    `${path} ${context}: locally fulfilled business non-GET`,
  ).toEqual(options.expectedLocalBusinessNonGet ?? []);
  if (options.syntheticEndpoints) {
    expect(
      interceptor.ledger.continuedToBackend,
      `${path} ${context}: fully synthetic route must not reach the backend`,
    ).toEqual([]);
  }
  expect(
    interceptor.ledger.attempted.length,
    `${path} ${context}: every request has one terminal classification`,
  ).toBe(
    interceptor.ledger.locallyFulfilled.length +
      interceptor.ledger.blocked.length +
      interceptor.ledger.continuedToBackend.length,
  );
  for (const forbiddenPath of options.forbiddenGetPaths ?? []) {
    expect(
      interceptor.ledger.attempted.some((request) => request.startsWith(`GET:${forbiddenPath}`)),
      `${path} ${context}: forbidden GET ${forbiddenPath}`,
    ).toBe(false);
  }
}

async function auditRoute(
  page: Page,
  path: string,
  options: AuditRouteOptions = {},
): Promise<void> {
  let consoleErrors: readonly string[] = [];
  let failedResponses: readonly string[] = [];
  let failedRequests: readonly string[] = [];
  let pageErrors: readonly string[] = [];
  let navigating = false;

  page.on("console", (message) => {
    if (message.type() === "error") consoleErrors = [...consoleErrors, "console:error"];
  });
  page.on("response", (response) => {
    if (response.status() >= 400) {
      failedResponses = [
        ...failedResponses,
        `${response.request().method()}:${new URL(response.url()).pathname}:${response.status()}`,
      ];
    }
  });
  page.on("requestfailed", (request) => {
    const failureText = request.failure()?.errorText ?? "unknown";
    const isExpectedNavigationAbort =
      navigating &&
      request.isNavigationRequest() &&
      request.method() === "GET" &&
      failureText === "net::ERR_ABORTED";
    if (isExpectedNavigationAbort) return;
    failedRequests = [
      ...failedRequests,
      `${request.method()}:${new URL(request.url()).pathname}:${failureText}`,
    ];
  });
  page.on("pageerror", (error) => {
    pageErrors = [...pageErrors, error.name];
  });
  const endpoints: SyntheticEndpoint[] = [
    ...Object.entries(options.getResponses ?? {}).map(([pathname, response]) => ({
      method: "GET" as const,
      pathname,
      response,
    })),
    ...(options.meResponse === undefined
      ? []
      : [{ method: "GET" as const, pathname: "/api/v1/me", response: options.meResponse }]),
    ...(options.syntheticEndpoints ?? []),
  ];
  const interceptor = await installSyntheticApiInterceptor(page, endpoints, {
    expectedOrigin: options.syntheticEndpoints ? AUDIT_ORIGIN : undefined,
    expectedClinicId: options.syntheticEndpoints ? options.currentClinicId : undefined,
  });

  try {
    if (options.currentClinicId) {
      await page.addInitScript(
        ({ key, value }) => window.localStorage.setItem(key, value),
        { key: "auth_current_clinic:v1", value: options.currentClinicId },
      );
    }
    if (options.fixedTime) await page.clock.setFixedTime(new Date(options.fixedTime));
    for (const width of VIEWPORTS) {
      consoleErrors = [];
      failedResponses = [];
      failedRequests = [];
      pageErrors = [];
      interceptor.reset();
      await page.setViewportSize({ width, height: AUDIT_HEIGHT });
      navigating = true;
      try {
        await page.goto(path, { waitUntil: "domcontentloaded", timeout: 60000 });
      } finally {
        navigating = false;
      }
      if (!options.allowLogin) {
        await expect(page).not.toHaveURL(/\/login(?:\?|$)/, { timeout: 30000 });
      }
      const expectedRenderedPath = options.expectedRenderedPath ?? path;
      await expect(page).toHaveURL(
        (url) =>
          (options.allowSearch ? url.pathname : `${url.pathname}${url.search}`) ===
          expectedRenderedPath,
        { timeout: 30000 },
      );
      await expect(page.locator("h1").first(), `${path} @ ${width}px: h1`).toBeVisible({ timeout: 30000 });
      const currentUrl = new URL(page.url());
      expect(
        options.allowSearch ? currentUrl.pathname : `${currentUrl.pathname}${currentUrl.search}`,
        `${path} @ ${width}px: route did not render the requested page`,
      ).toBe(expectedRenderedPath);
      if (options.expectedText) {
        await expect(
          page.getByText(options.expectedText, { exact: false }).filter({ visible: true }).first(),
          `${path} @ ${width}px: required state text`,
        ).toBeVisible({ timeout: 30000 });
      }
      if (options.renderedAssertions) {
        await assertSyntheticRenderedState(
          page,
          options.renderedAssertions,
          `${path} @ ${width}px`,
        );
      }
      for (const buttonName of options.forbiddenButtons ?? []) {
        await expect(
          page.getByRole("button", { name: buttonName, exact: true }),
          `${path} @ ${width}px: forbidden button ${buttonName}`,
        ).toHaveCount(0);
      }
      for (const selector of options.disabledSelectors ?? []) {
        await expect(
          page.locator(selector).first(),
          `${path} @ ${width}px: read-only control ${selector}`,
        ).toHaveAttribute("disabled", "");
      }
      await page.waitForTimeout(350);
      await page.waitForLoadState("networkidle", { timeout: 15_000 });

      if (options.syntheticEndpoints) {
        const toastTypes = await page.locator("[data-sonner-toast]").evaluateAll((toasts) =>
          toasts.map((toast) => toast.getAttribute("data-type") ?? "unknown"),
        );
        expect(toastTypes, `${path} @ ${width}px: unexpected toast types`).toEqual([]);
      }

      const result = await inspectPage(page);
      expect(result.overflow, `${path} @ ${width}px: document overflow`).toBe(0);
      expect(result.unnamed, `${path} @ ${width}px: accessible names`).toEqual([]);
      expect(result.undersized, `${path} @ ${width}px: 44px targets`).toEqual([]);
      assertRequestSafety(path, `@ ${width}px`, interceptor, options, {
        consoleErrors,
        failedResponses,
        failedRequests,
        pageErrors,
      });
    }
    await page.close();
    assertRequestSafety(path, "during close", interceptor, options, {
      consoleErrors,
      failedResponses,
      failedRequests,
      pageErrors,
    });
  } finally {
    if (!page.isClosed()) await page.close();
    await interceptor.dispose();
  }
}

test("route matrix inventory is exactly 86 product pages", () => {
  expect(PUBLIC_ROUTES).toHaveLength(3);
  expect(PROTECTED_ROUTES).toHaveLength(82);
  expect([...PUBLIC_ROUTES, ...PROTECTED_ROUTES]).toHaveLength(86);
  expect(PROTECTED_ROUTES.filter((route) => route.syntheticClinical)).toHaveLength(5);
});

test.describe("public auth routes: read-only 4 viewport audit", () => {
  for (const route of PUBLIC_ROUTES) {
    test(route.label, async ({ page }) => {
      await auditRoute(page, route.path, { allowLogin: route.path === "/login" });
    });
  }
});

test.describe("protected routes: read-only 4 viewport audit", () => {
  let fixtures: FixturePaths;
  let editableEstimateResponse: unknown;
  let authenticatedStorageState: AuthenticatedStorageState | undefined;

  test.beforeAll(async ({ browser }) => {
    const setupContext = await browser.newContext();
    try {
      const setupPage = await setupContext.newPage();
      await loginAsDemoAdmin(setupPage);
      const estimateResponse = await setupPage.request.get("/api/v1/estimates/1");
      expect(estimateResponse.ok(), "edit-route audit requires an existing estimate shape").toBe(true);
      editableEstimateResponse = createEditableEstimateResponse(await estimateResponse.json());
      fixtures = await discoverFixtures(setupPage);
      authenticatedStorageState = await setupContext.storageState();
      await setupPage.close();
    } finally {
      await setupContext.close();
    }
  });

  for (const route of PROTECTED_ROUTES) {
    test(route.label, async ({ browser }) => {
      test.setTimeout(180000);
      const syntheticScenario = route.syntheticClinical
        ? SYNTHETIC_CLINICAL_SCENARIOS[route.syntheticClinical]
        : undefined;
      const path = syntheticScenario?.entryPath ?? resolveRoute(route, fixtures);
      expect(path, `${route.path}: production fixture unavailable`).toBeDefined();
      if (!authenticatedStorageState) throw new Error("protected audit auth state was not initialized");
      const routeContext = await browser.newContext({ storageState: authenticatedStorageState });
      const page = await routeContext.newPage();
      try {
        await auditRoute(page, path ?? route.path, {
          expectedText: route.expectedText,
          expectedRenderedPath: syntheticScenario?.renderedPath ?? route.renderedPath,
          allowSearch: route.allowSearch,
          getResponses: route.useEditableEstimate
            ? { "/api/v1/estimates/1": editableEstimateResponse }
            : undefined,
          syntheticEndpoints: syntheticScenario?.endpoints,
          expectedLocalBusinessNonGet: syntheticScenario?.expectedLocalBusinessNonGet,
          currentClinicId: syntheticScenario?.currentClinicId,
          fixedTime: syntheticScenario?.fixedTime,
          renderedAssertions: syntheticScenario?.renderedAssertions,
        });
      } finally {
        if (!page.isClosed()) await page.close();
        await routeContext.close();
      }
    });
  }
});

test.describe("known RBAC regressions: synthetic read-only /me", () => {
  let editableEstimateResponse: unknown;
  let authenticatedStorageState: AuthenticatedStorageState | undefined;

  test.beforeAll(async ({ browser }) => {
    const setupContext = await browser.newContext();
    try {
      const setupPage = await setupContext.newPage();
      await loginAsDemoAdmin(setupPage);
      const estimateResponse = await setupPage.request.get("/api/v1/estimates/1");
      expect(estimateResponse.ok(), "RBAC audit requires an existing estimate shape").toBe(true);
      editableEstimateResponse = createEditableEstimateResponse(await estimateResponse.json());
      authenticatedStorageState = await setupContext.storageState();
      await setupPage.close();
    } finally {
      await setupContext.close();
    }
  });

  test("会計詳細はレジ締め閲覧権限がない場合にfail-closedとなる", async ({ browser }) => {
    test.setTimeout(180000);
    if (!authenticatedStorageState) throw new Error("RBAC audit auth state was not initialized");
    const routeContext = await browser.newContext({ storageState: authenticatedStorageState });
    const page = await routeContext.newPage();
    try {
      await auditRoute(page, "/accounting/1024128", {
        expectedText: "レジ締め状態を確認する権限がないため変更できません",
        forbiddenGetPaths: ["/api/v1/cash-register/closes"],
        meResponse: createSyntheticMeResponse({
          accounting: { view: true, create: true, edit: true, delete: true },
        }),
      });
    } finally {
      if (!page.isClosed()) await page.close();
      await routeContext.close();
    }
  });

  test("在庫編集はedit権限がない場合に閲覧専用となる", async ({ browser }) => {
    test.setTimeout(180000);
    if (!authenticatedStorageState) throw new Error("RBAC audit auth state was not initialized");
    const routeContext = await browser.newContext({ storageState: authenticatedStorageState });
    const page = await routeContext.newPage();
    try {
      await auditRoute(page, "/inventory/7", {
        expectedText: "閲覧専用 — 編集権限がないため変更できません",
        meResponse: createSyntheticMeResponse({
          inventory: { view: true, create: false, edit: false, delete: false },
        }),
      });
    } finally {
      if (!page.isClosed()) await page.close();
      await routeContext.close();
    }
  });

  test("見積詳細はedit/delete権限がない場合に変更操作を表示しない", async ({ browser }) => {
    test.setTimeout(180000);
    if (!authenticatedStorageState) throw new Error("RBAC audit auth state was not initialized");
    const routeContext = await browser.newContext({ storageState: authenticatedStorageState });
    const page = await routeContext.newPage();
    try {
      await auditRoute(page, "/estimates/1", {
        expectedText: "期限切れ",
        forbiddenButtons: ["編集", "削除"],
        getResponses: { "/api/v1/estimates/1": editableEstimateResponse },
        meResponse: createSyntheticMeResponse({
          estimates: { view: true, create: false, edit: false, delete: false },
        }),
      });
    } finally {
      if (!page.isClosed()) await page.close();
      await routeContext.close();
    }
  });
});

test.describe("clinical P10: view-only /me cannot mutate clinical records", () => {
  let authenticatedStorageState: AuthenticatedStorageState | undefined;

  test.beforeAll(async ({ browser }) => {
    const setupContext = await browser.newContext();
    try {
      const setupPage = await setupContext.newPage();
      await loginAsDemoAdmin(setupPage);
      authenticatedStorageState = await setupContext.storageState();
      await setupPage.close();
    } finally {
      await setupContext.close();
    }
  });

  const cases = [
    {
      label: "カルテ作成",
      scenarioKey: "medicalRecordCreate",
      resource: "medical-records",
      forbiddenButtons: ["保存"],
      accessDenied: true,
    },
    {
      label: "入院詳細",
      scenarioKey: "hospitalizationDetail",
      resource: "hospitalization",
      forbiddenButtons: ["退院処理", "入院情報の編集"],
    },
    {
      label: "入院編集",
      scenarioKey: "hospitalizationEdit",
      resource: "hospitalization",
      forbiddenButtons: ["削除", "更新"],
      accessDenied: true,
    },
    {
      label: "トリミング編集",
      scenarioKey: "trimmingEdit",
      resource: "trimming",
      forbiddenButtons: ["削除", "保存"],
      disabledSelectors: ["fieldset"],
    },
    {
      label: "ワクチン編集",
      scenarioKey: "vaccinationEdit",
      resource: "vaccinations",
      forbiddenButtons: ["削除", "保存"],
      disabledSelectors: ["fieldset"],
    },
  ] as const;

  for (const readOnlyCase of cases) {
    test(readOnlyCase.label, async ({ browser }) => {
      test.setTimeout(180000);
      if (!authenticatedStorageState) throw new Error("clinical RBAC auth state was not initialized");
      const scenario = SYNTHETIC_CLINICAL_SCENARIOS[readOnlyCase.scenarioKey];
      const viewOnlyMe = createSyntheticMeResponse({
        [readOnlyCase.resource]: { view: true, create: false, edit: false, delete: false },
      });
      const endpoints = scenario.endpoints.map((endpoint) =>
        endpoint.method === "GET" && endpoint.pathname === "/api/v1/me"
          ? { ...endpoint, response: viewOnlyMe }
          : endpoint,
      );
      const renderedAssertions: readonly SyntheticRenderedAssertion[] =
        "accessDenied" in readOnlyCase
          ? [{ kind: "visibleText", value: "アクセス権限がありません" }]
          : scenario.renderedAssertions.filter(
              (assertion) =>
                assertion.kind !== "selectedOption" &&
                !(assertion.kind === "visibleText" && assertion.value === "退院処理"),
            );
      const routeContext = await browser.newContext({ storageState: authenticatedStorageState });
      const page = await routeContext.newPage();
      try {
        await auditRoute(page, scenario.entryPath, {
          expectedRenderedPath: "accessDenied" in readOnlyCase
            ? scenario.entryPath
            : scenario.renderedPath,
          syntheticEndpoints: endpoints,
          expectedLocalBusinessNonGet: [],
          currentClinicId: scenario.currentClinicId,
          fixedTime: scenario.fixedTime,
          renderedAssertions,
          forbiddenButtons: readOnlyCase.forbiddenButtons,
          disabledSelectors: "disabledSelectors" in readOnlyCase
            ? readOnlyCase.disabledSelectors
            : undefined,
        });
      } finally {
        if (!page.isClosed()) await page.close();
        await routeContext.close();
      }
    });
  }
});
