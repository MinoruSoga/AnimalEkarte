import { expect, test, type BrowserContext, type Page } from "@playwright/test";

import { loginAsDemoAdmin } from "./helpers/auth";

const VIEWPORTS = [1440, 1200, 800, 500] as const;
const AUDIT_HEIGHT = 900;

type FixtureKey =
  | "owner"
  | "medicalRecord"
  | "hospitalization"
  | "trimming"
  | "examination"
  | "vaccination"
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
  blocker?: "fixture-unavailable" | "mount-non-get";
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
    fixture: "pet",
    blocker: "mount-non-get",
  },
  { label: "カルテ編集", path: "/medical-records/:id", fixture: "medicalRecord" },
  { label: "入院・ホテル一覧", path: "/hospitalization" },
  { label: "入院・ホテル ペット選択", path: "/hospitalization/select-pet" },
  { label: "入院・ホテル登録", path: "/hospitalization/new", fixture: "pet" },
  {
    label: "入院・ホテル詳細",
    path: "/hospitalization/:id",
    fixture: "hospitalization",
    blocker: "fixture-unavailable",
  },
  {
    label: "入院・ホテル編集",
    path: "/hospitalization/:id/edit",
    fixture: "hospitalization",
    suffix: "/edit",
    blocker: "fixture-unavailable",
  },
  { label: "トリミング一覧", path: "/trimming" },
  { label: "トリミング ペット選択", path: "/trimming/select-pet" },
  { label: "トリミング登録", path: "/trimming/new", fixture: "pet" },
  {
    label: "トリミング編集",
    path: "/trimming/:id",
    fixture: "trimming",
    blocker: "fixture-unavailable",
  },
  { label: "検査一覧", path: "/examinations" },
  { label: "検査ペット選択", path: "/examinations/select-pet" },
  { label: "検査登録", path: "/examinations/new", fixture: "pet" },
  { label: "検査編集", path: "/examinations/:id", fixture: "examination" },
  { label: "ワクチン一覧", path: "/vaccinations" },
  { label: "ワクチン ペット選択", path: "/vaccinations/select-pet" },
  { label: "ワクチン登録", path: "/vaccinations/new", fixture: "pet" },
  {
    label: "ワクチン編集",
    path: "/vaccinations/:id",
    fixture: "vaccination",
    blocker: "fixture-unavailable",
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
  { label: "Lステップ連携設定", path: "/settings/integrations/lstep" },
  { label: "Lステップタグ管理", path: "/settings/lstep/tags" },
];

type FixturePaths = Partial<Record<FixtureKey, string>>;

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

async function firstDetailPath(page: Page, listPath: string, pattern: RegExp): Promise<string | undefined> {
  await page.goto(listPath, { waitUntil: "domcontentloaded" });
  await page.locator("h1").first().waitFor({ state: "visible", timeout: 30000 });
  const hrefs = await page.locator("a[href]").evaluateAll((links) =>
    links.map((link) => link.getAttribute("href")).filter((href): href is string => href !== null),
  );
  return hrefs.find((href) => pattern.test(href));
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

function readFirstId(value: unknown): string | undefined {
  if (Array.isArray(value)) {
    for (const item of value) {
      const id = readFirstId(item);
      if (id) return id;
    }
    return undefined;
  }
  if (typeof value !== "object" || value === null) return undefined;
  const record = value as Record<string, unknown>;
  if (typeof record.id === "string" || typeof record.id === "number") {
    return String(record.id);
  }
  return readFirstId(record.data);
}

async function firstApiDetailPath(
  page: Page,
  endpoint: string,
  routePrefix: string,
): Promise<string | undefined> {
  const response = await page.request.get(endpoint);
  if (!response.ok()) return undefined;
  const id = readFirstId(await response.json());
  return id ? `${routePrefix}/${encodeURIComponent(id)}` : undefined;
}

interface SyntheticResourcePermission {
  view: boolean;
  create: boolean;
  edit: boolean;
  delete: boolean;
}

function createSyntheticMeResponse(
  raw: unknown,
  permissions: Readonly<Record<string, SyntheticResourcePermission>>,
): unknown {
  if (typeof raw !== "object" || raw === null) {
    throw new Error("/v1/me did not return an object for the RBAC audit");
  }

  return {
    ...raw,
    id: "ui-audit-read-only",
    email: "ui-audit-read-only@example.invalid",
    display_name: "UI監査 閲覧専用",
    is_system_admin: false,
    permissions,
  };
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
  const hospitalizationFromPage = await firstDetailPath(
    page,
    "/hospitalization",
    /^\/hospitalization\/[^/]+$/,
  );
  const hospitalization = hospitalizationFromPage ?? await firstApiDetailPath(
    page,
    "/api/v1/hospitalizations",
    "/hospitalization",
  );
  const trimmingFromPage = await firstDetailPath(page, "/trimming", /^\/trimming\/[^/]+$/);
  const trimming = trimmingFromPage ?? await firstApiDetailPath(
    page,
    "/api/v1/trimmings?page=1&limit=20",
    "/trimming",
  );
  const examination = await firstDetailPath(
    page,
    "/examinations",
    /^\/examinations\/[^/]+$/,
  );
  const vaccinationFromPage = await firstDetailPath(
    page,
    "/vaccinations",
    /^\/vaccinations\/[^/]+$/,
  );
  const vaccination = vaccinationFromPage ?? await firstApiDetailPath(
    page,
    "/api/v1/vaccinations",
    "/vaccinations",
  );

  let pet: string | undefined;
  if (medicalRecord) {
    const response = await page.request.get(`/api/v1${medicalRecord}`);
    if (response.ok()) pet = readPetId(await response.json());
  }

  return { owner, medicalRecord, hospitalization, trimming, examination, vaccination, pet };
}

interface AuditResult {
  unnamed: string[];
  undersized: string[];
  overflow: number;
}

async function inspectPage(page: Page): Promise<AuditResult> {
  return page.evaluate(() => {
    const unnamed: string[] = [];
    const undersized: string[] = [];
    const elements = document.querySelectorAll<HTMLElement>(
      "button, input:not([type='hidden']), select, textarea, [role='button'], [role='tab'], [role='switch']",
    );

    for (const element of elements) {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      if (
        element.getAttribute("aria-hidden") === "true" ||
        style.display === "none" ||
        style.visibility === "hidden" ||
        rect.width === 0 ||
        rect.height === 0
      ) {
        continue;
      }

      const labelledBy = element.getAttribute("aria-labelledby");
      const labelledText = labelledBy
        ?.split(/\s+/)
        .map((id) => document.getElementById(id)?.textContent ?? "")
        .join(" ");
      const explicitLabels = element.id
        ? [...document.querySelectorAll<HTMLLabelElement>("label[for]")]
            .filter((label) => label.htmlFor === element.id)
            .map((label) => label.textContent ?? "")
            .join(" ")
        : "";
      const wrappingLabelElement = element.closest("label");
      const wrappingLabel = wrappingLabelElement?.textContent ?? "";
      const name = [
        element.getAttribute("aria-label"),
        labelledText,
        explicitLabels,
        wrappingLabel,
        element.textContent,
        element.getAttribute("title"),
        element.getAttribute("placeholder"),
      ].find((candidate) => candidate?.trim());
      const descriptor = [
        element.tagName.toLowerCase(),
        element.id ? `#${element.id}` : "",
        `[role=${element.getAttribute("role") ?? "native"}]`,
        element.getAttribute("name") ? `[name=${element.getAttribute("name")}]` : "",
        element.getAttribute("type") ? `[type=${element.getAttribute("type")}]` : "",
        element.dataset.testid ? `[testid=${element.dataset.testid}]` : "",
        element.classList.length > 0 ? `[class=${[...element.classList].slice(0, 6).join(".")}]` : "",
      ].join("");
      if (!name) unnamed.push(descriptor);

      let targetRect = rect;
      const role = element.getAttribute("role");
      const usesAssociatedLabelTarget =
        role === "checkbox" ||
        role === "radio" ||
        role === "switch" ||
        (element instanceof HTMLInputElement &&
          (element.type === "checkbox" || element.type === "radio" || element.type === "file"));
      if (usesAssociatedLabelTarget) {
        const candidates = [
          rect,
          ...[...document.querySelectorAll<HTMLLabelElement>("label[for]")]
            .filter((label) => element.id !== "" && label.htmlFor === element.id)
            .map((label) => label.getBoundingClientRect()),
          ...(wrappingLabelElement ? [wrappingLabelElement.getBoundingClientRect()] : []),
        ];
        targetRect = candidates.reduce((largest, candidate) =>
          Math.min(candidate.width, candidate.height) > Math.min(largest.width, largest.height)
            ? candidate
            : largest,
        );
      }
      if (targetRect.width < 44 || targetRect.height < 44) {
        undersized.push(`${descriptor}:${Math.round(targetRect.width)}x${Math.round(targetRect.height)}`);
      }
    }

    return {
      unnamed: [...new Set(unnamed)],
      undersized: [...new Set(undersized)],
      overflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
    };
  });
}

async function auditRoute(
  page: Page,
  path: string,
  options: {
    allowLogin?: boolean;
    allowSearch?: boolean;
    expectedText?: string;
    expectedRenderedPath?: string;
    forbiddenButtons?: readonly string[];
    forbiddenGetPaths?: readonly string[];
    getResponses?: Readonly<Record<string, unknown>>;
    meResponse?: unknown;
  } = {},
): Promise<void> {
  const consoleErrors: string[] = [];
  const failedResponses: string[] = [];
  const nonGetAttempts: string[] = [];
  const getRequests: string[] = [];

  page.on("console", (message) => {
    if (message.type() === "error") consoleErrors.push("console:error");
  });
  page.on("response", (response) => {
    if (response.status() >= 400) {
      failedResponses.push(`${response.request().method()}:${new URL(response.url()).pathname}:${response.status()}`);
    }
  });
  await page.route("**/api/**", async (route) => {
    const method = route.request().method();
    if (["GET", "HEAD", "OPTIONS"].includes(method)) {
      const pathname = new URL(route.request().url()).pathname;
      if (method === "GET") getRequests.push(pathname);
      const getResponse = options.getResponses?.[pathname];
      if (method === "GET" && getResponse !== undefined) {
        await route.fulfill({ json: getResponse });
        return;
      }
      if (method === "GET" && pathname.endsWith("/v1/me") && options.meResponse !== undefined) {
        await route.fulfill({ json: options.meResponse });
        return;
      }
      await route.continue();
      return;
    }
    nonGetAttempts.push(`${method}:${new URL(route.request().url()).pathname}`);
    await route.abort("blockedbyclient");
  });

  for (const width of VIEWPORTS) {
    consoleErrors.length = 0;
    failedResponses.length = 0;
    nonGetAttempts.length = 0;
    getRequests.length = 0;
    await page.setViewportSize({ width, height: AUDIT_HEIGHT });
    await page.goto(path, { waitUntil: "domcontentloaded", timeout: 60000 });
    if (!options.allowLogin) {
      await expect(page).not.toHaveURL(/\/login(?:\?|$)/, { timeout: 30000 });
    }
    await expect(page.locator("h1").first(), `${path} @ ${width}px: h1`).toBeVisible({ timeout: 30000 });
    const currentUrl = new URL(page.url());
    expect(
      options.allowSearch ? currentUrl.pathname : `${currentUrl.pathname}${currentUrl.search}`,
      `${path} @ ${width}px: route did not render the requested page`,
    ).toBe(options.expectedRenderedPath ?? path);
    if (options.expectedText) {
      await expect(
        page.getByText(options.expectedText, { exact: false }).first(),
        `${path} @ ${width}px: required state text`,
      ).toBeVisible({ timeout: 30000 });
    }
    for (const buttonName of options.forbiddenButtons ?? []) {
      await expect(
        page.getByRole("button", { name: buttonName, exact: true }),
        `${path} @ ${width}px: forbidden button ${buttonName}`,
      ).toHaveCount(0);
    }
    await page.waitForTimeout(350);

    const result = await inspectPage(page);
    expect(result.overflow, `${path} @ ${width}px: document overflow`).toBeLessThanOrEqual(1);
    expect(result.unnamed, `${path} @ ${width}px: accessible names`).toEqual([]);
    expect(result.undersized, `${path} @ ${width}px: 44px targets`).toEqual([]);
    expect(consoleErrors, `${path} @ ${width}px: console errors`).toEqual([]);
    expect(failedResponses, `${path} @ ${width}px: failed responses`).toEqual([]);
    expect(nonGetAttempts, `${path} @ ${width}px: business non-GET attempts`).toEqual([]);
    for (const forbiddenPath of options.forbiddenGetPaths ?? []) {
      expect(getRequests, `${path} @ ${width}px: forbidden GET ${forbiddenPath}`).not.toContain(
        forbiddenPath,
      );
    }
  }
}

test("route matrix inventory is exactly 83 product pages", () => {
  expect(PUBLIC_ROUTES).toHaveLength(3);
  expect(PROTECTED_ROUTES).toHaveLength(80);
  expect([...PUBLIC_ROUTES, ...PROTECTED_ROUTES]).toHaveLength(83);
});

test.describe("public auth routes: read-only 4 viewport audit", () => {
  for (const route of PUBLIC_ROUTES) {
    test(route.label, async ({ page }) => {
      await auditRoute(page, route.path, { allowLogin: route.path === "/login" });
    });
  }
});

test.describe("protected routes: read-only 4 viewport audit", () => {
  let context: BrowserContext | undefined;
  let fixtures: FixturePaths;
  let editableEstimateResponse: unknown;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const setupPage = await context.newPage();
    await loginAsDemoAdmin(setupPage);
    const estimateResponse = await setupPage.request.get("/api/v1/estimates/1");
    expect(estimateResponse.ok(), "edit-route audit requires an existing estimate shape").toBe(true);
    editableEstimateResponse = createEditableEstimateResponse(await estimateResponse.json());
    fixtures = await discoverFixtures(setupPage);
    await setupPage.close();
  });

  test.afterAll(async () => {
    await context?.close();
  });

  for (const route of PROTECTED_ROUTES.filter((candidate) => candidate.blocker === undefined)) {
    test(route.label, async () => {
      test.setTimeout(180000);
      const path = resolveRoute(route, fixtures);
      expect(path, `${route.path}: production fixture unavailable`).toBeDefined();
      if (!context) throw new Error("protected audit context was not initialized");
      const page = await context.newPage();
      try {
        await auditRoute(page, path ?? route.path, {
          expectedText: route.expectedText,
          expectedRenderedPath: route.renderedPath,
          allowSearch: route.allowSearch,
          getResponses: route.useEditableEstimate
            ? { "/api/v1/estimates/1": editableEstimateResponse }
            : undefined,
        });
      } finally {
        await page.close();
      }
    });
  }
});

test.describe("blocked routes: deterministic safety evidence", () => {
  let context: BrowserContext | undefined;
  let fixtures: FixturePaths;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const setupPage = await context.newPage();
    await loginAsDemoAdmin(setupPage);
    fixtures = await discoverFixtures(setupPage);
    await setupPage.close();
  });

  test.afterAll(async () => {
    await context?.close();
  });

  for (const route of PROTECTED_ROUTES.filter(
    (candidate) => candidate.blocker === "fixture-unavailable",
  )) {
    test(`${route.label}: current list and API contain no safe fixture`, () => {
      expect(resolveRoute(route, fixtures), `${route.path}: fixture unexpectedly became available`).toBeUndefined();
    });
  }

  test("カルテ作成: mount時business POSTを遮断しDB変更0を証明する", async () => {
    test.setTimeout(180000);
    if (!context) throw new Error("blocked-route audit context was not initialized");
    const route = PROTECTED_ROUTES.find((candidate) => candidate.blocker === "mount-non-get");
    if (!route) throw new Error("mount-non-get route was not registered");
    const path = resolveRoute(route, fixtures);
    expect(path, `${route.path}: pet fixture unavailable`).toBeDefined();

    const nonGetAttempts: string[] = [];
    const page = await context.newPage();
    await page.route("**/api/**", async (requestRoute) => {
      const method = requestRoute.request().method();
      if (["GET", "HEAD", "OPTIONS"].includes(method)) {
        await requestRoute.continue();
        return;
      }
      nonGetAttempts.push(`${method}:${new URL(requestRoute.request().url()).pathname}`);
      await requestRoute.abort("blockedbyclient");
    });

    try {
      await page.goto(path ?? route.path, { waitUntil: "domcontentloaded", timeout: 60000 });
      await expect(page.locator("h1").first()).toBeVisible({ timeout: 30000 });
      await expect.poll(() => nonGetAttempts, { timeout: 30000 }).toContain(
        "POST:/api/v1/reservations",
      );
      expect(nonGetAttempts).toEqual(["POST:/api/v1/reservations"]);
    } finally {
      await page.close();
    }
  });
});

test.describe("known RBAC regressions: synthetic read-only /me", () => {
  let context: BrowserContext | undefined;
  let meResponse: unknown;
  let editableEstimateResponse: unknown;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const setupPage = await context.newPage();
    await loginAsDemoAdmin(setupPage);
    const response = await setupPage.request.get("/api/v1/me");
    expect(response.ok(), "RBAC audit requires the authenticated /v1/me shape").toBe(true);
    meResponse = await response.json();
    const estimateResponse = await setupPage.request.get("/api/v1/estimates/1");
    expect(estimateResponse.ok(), "RBAC audit requires an existing estimate shape").toBe(true);
    editableEstimateResponse = createEditableEstimateResponse(await estimateResponse.json());
    await setupPage.close();
  });

  test.afterAll(async () => {
    await context?.close();
  });

  test("会計詳細はレジ締め閲覧権限がない場合にfail-closedとなる", async () => {
    test.setTimeout(180000);
    if (!context) throw new Error("RBAC audit context was not initialized");
    const page = await context.newPage();
    try {
      await auditRoute(page, "/accounting/1024128", {
        expectedText: "レジ締め状態を確認する権限がないため変更できません",
        forbiddenGetPaths: ["/api/v1/cash-register/closes"],
        meResponse: createSyntheticMeResponse(meResponse, {
          accounting: { view: true, create: true, edit: true, delete: true },
        }),
      });
    } finally {
      await page.close();
    }
  });

  test("在庫編集はedit権限がない場合に閲覧専用となる", async () => {
    test.setTimeout(180000);
    if (!context) throw new Error("RBAC audit context was not initialized");
    const page = await context.newPage();
    try {
      await auditRoute(page, "/inventory/7", {
        expectedText: "閲覧専用 — 編集権限がないため変更できません",
        meResponse: createSyntheticMeResponse(meResponse, {
          inventory: { view: true, create: false, edit: false, delete: false },
        }),
      });
    } finally {
      await page.close();
    }
  });

  test("見積詳細はedit/delete権限がない場合に変更操作を表示しない", async () => {
    test.setTimeout(180000);
    if (!context) throw new Error("RBAC audit context was not initialized");
    const page = await context.newPage();
    try {
      await auditRoute(page, "/estimates/1", {
        expectedText: "期限切れ",
        forbiddenButtons: ["編集", "削除"],
        getResponses: { "/api/v1/estimates/1": editableEstimateResponse },
        meResponse: createSyntheticMeResponse(meResponse, {
          estimates: { view: true, create: false, edit: false, delete: false },
        }),
      });
    } finally {
      await page.close();
    }
  });
});
