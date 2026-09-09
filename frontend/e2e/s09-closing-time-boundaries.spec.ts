import {
  test,
  expect,
  type APIRequestContext,
  type BrowserContext,
  type Page,
} from "@playwright/test";
import { LoginPage } from "./pages/login-page";
import { AccountingPage } from "./pages/accounting-page";

/**
 * S09 #2–#6: closing attribution preview against disposable synthetic fixture.
 * Setup/teardown: POST/DELETE /api/v1/uat/synthetic-closings (backend helper).
 * Password: UAT_SYNTHETIC_CLOSING_PASSWORD only — never logged or committed.
 */

interface SyntheticClosingFixture {
  clinicId: number;
  loginEmail: string;
  billingIds: number[];
  completedAt: string[];
  cleanupToken: string;
  targetDate: string;
}

function pad2(value: number): string {
  return String(value).padStart(2, "0");
}

/** Most recent Monday on/before today (JST) — fixture rejects Sat/Sun. */
function fixtureTargetDateJST(): string {
  const jst = new Date(Date.now() + 9 * 60 * 60 * 1000);
  const day = jst.getUTCDay(); // 0=Sun .. 6=Sat
  const back = day === 0 ? 6 : day - 1;
  const monday = new Date(
    Date.UTC(jst.getUTCFullYear(), jst.getUTCMonth(), jst.getUTCDate() - back),
  );
  return `${monday.getUTCFullYear()}-${pad2(monday.getUTCMonth() + 1)}-${pad2(monday.getUTCDate())}`;
}

function nextCalendarDay(isoDate: string): string {
  const [year, month, day] = isoDate.split("-").map(Number);
  const next = new Date(Date.UTC(year, month - 1, day + 1));
  return `${next.getUTCFullYear()}-${pad2(next.getUTCMonth() + 1)}-${pad2(next.getUTCDate())}`;
}

/**
 * UAT synthetic-closing HTTP origin.
 *
 * Default: same origin as PLAYWRIGHT_TEST_BASE_URL (frontend). Vite proxies `/api`
 * to `backend:8080`, so the backend Request.Host is allowlisted (`backend`) even
 * when Playwright Docker reaches the app via `host.docker.internal:3003`.
 * Direct `:8080` via host.docker.internal fails AllowUATSyntheticClosingHTTPHost.
 * Override with UAT_SYNTHETIC_CLOSING_API_BASE only when you intentionally bypass the proxy.
 */
function syntheticClosingApiBase(): string {
  const explicit = process.env.UAT_SYNTHETIC_CLOSING_API_BASE?.trim();
  if (explicit) return explicit.replace(/\/$/, "");
  const playwrightBase = process.env.PLAYWRIGHT_TEST_BASE_URL ?? "http://localhost:3003";
  try {
    return new URL(playwrightBase).origin;
  } catch {
    return "http://localhost:3003";
  }
}

function requireSyntheticClosingPassword(): string {
  const password = process.env.UAT_SYNTHETIC_CLOSING_PASSWORD ?? "";
  if (!password) {
    throw new Error(
      "UAT_SYNTHETIC_CLOSING_PASSWORD must be set for S09 fixture e2e (no in-repo fallback; value is never logged)",
    );
  }
  return password;
}

function parseFixture(body: unknown, targetDate: string): SyntheticClosingFixture {
  if (typeof body !== "object" || body === null) {
    throw new Error("synthetic-closings response is not an object");
  }
  const record = body as Record<string, unknown>;
  const clinicId = record.clinicId;
  const loginEmail = record.loginEmail;
  const billingIds = record.billingIds;
  const completedAt = record.completedAt;
  const cleanupToken = record.cleanupToken;
  if (
    typeof clinicId !== "number" ||
    !Number.isSafeInteger(clinicId) ||
    clinicId <= 0 ||
    typeof loginEmail !== "string" ||
    loginEmail.trim() === "" ||
    !Array.isArray(billingIds) ||
    billingIds.length !== 5 ||
    !Array.isArray(completedAt) ||
    completedAt.length !== 5 ||
    typeof cleanupToken !== "string" ||
    cleanupToken === ""
  ) {
    throw new Error("synthetic-closings response is incomplete");
  }
  if (clinicId === 1 || clinicId === 2) {
    throw new Error("synthetic-closings returned reserved clinic id");
  }
  return {
    clinicId,
    loginEmail: loginEmail.trim(),
    billingIds: billingIds as number[],
    completedAt: completedAt as string[],
    cleanupToken,
    targetDate,
  };
}

async function createSyntheticClosingFixture(
  request: APIRequestContext,
  targetDate: string,
): Promise<SyntheticClosingFixture> {
  const response = await request.post(
    `${syntheticClosingApiBase()}/api/v1/uat/synthetic-closings`,
    {
      data: { targetDate },
      failOnStatusCode: false,
    },
  );
  if (response.status() === 404) {
    throw new Error(
      "POST /api/v1/uat/synthetic-closings returned 404 (stack APP_ENV/host allowlist, or backend down)",
    );
  }
  if (response.status() !== 201) {
    throw new Error(`POST /api/v1/uat/synthetic-closings failed with status ${response.status()}`);
  }
  return parseFixture(await response.json(), targetDate);
}

async function deleteSyntheticClosingFixture(
  request: APIRequestContext,
  fixture: SyntheticClosingFixture,
): Promise<void> {
  const response = await request.delete(
    `${syntheticClosingApiBase()}/api/v1/uat/synthetic-closings/${fixture.clinicId}`,
    {
      headers: { "X-UAT-Cleanup-Token": fixture.cleanupToken },
      failOnStatusCode: false,
    },
  );
  if (response.status() !== 204 && response.status() !== 404) {
    throw new Error(
      `DELETE /api/v1/uat/synthetic-closings/${fixture.clinicId} failed with status ${response.status()}`,
    );
  }
}

/** Fixture staff login — do not reuse demo-admin storage state. */
async function loginAsFixtureStaff(page: Page, email: string, password: string): Promise<void> {
  const loginPage = new LoginPage(page);
  await loginPage.gotoLogin();
  const emailInput = loginPage.emailInput();
  await emailInput.waitFor({ state: "visible", timeout: 60000 });
  await emailInput.fill(email);
  await loginPage.passwordInput().fill(password);

  const loginResponsePromise = loginPage.waitForLoginResponse();
  await loginPage.submitButton().click();
  const loginResponse = await loginResponsePromise;
  expect(loginResponse.status()).toBe(200);
  await expect(loginPage.homeHeading()).toBeVisible({ timeout: 60000 });
}

async function openPreview(
  accounting: AccountingPage,
  page: Page,
  date: string,
  period: "午前" | "午後" | "緊急",
): Promise<void> {
  await accounting.gotoClose();
  await expect(accounting.targetDateInput()).toBeVisible({ timeout: 30000 });
  await accounting.targetDateInput().fill(date);
  await accounting.periodButton(period).click();

  const previewResponse = page.waitForResponse(
    (response) =>
      response.url().includes("/v1/cash-register/preview") &&
      response.request().method() === "GET" &&
      response.ok(),
    { timeout: 60000 },
  );
  await accounting.previewButton().click();
  await previewResponse;
  await expect(accounting.billingDetailsHeading()).toBeVisible({ timeout: 30000 });
}

test.describe("S09 closing time boundaries (#2–#6 attribution preview)", () => {
  let context: BrowserContext | undefined;
  let fixture: SyntheticClosingFixture | null = null;
  let apiRequest: APIRequestContext | undefined;

  test.beforeAll(async ({ browser, playwright }) => {
    const password = requireSyntheticClosingPassword();
    apiRequest = await playwright.request.newContext();
    const targetDate = fixtureTargetDateJST();

    try {
      fixture = await createSyntheticClosingFixture(apiRequest, targetDate);
      context = await browser.newContext();
      const loginPage = await context.newPage();
      try {
        await loginAsFixtureStaff(loginPage, fixture.loginEmail, password);
      } finally {
        await loginPage.close();
      }
    } catch (error) {
      if (fixture && apiRequest) {
        await deleteSyntheticClosingFixture(apiRequest, fixture).catch(() => undefined);
        fixture = null;
      }
      if (apiRequest) await apiRequest.dispose().catch(() => undefined);
      apiRequest = undefined;
      throw error;
    }
  });

  test.afterAll(async () => {
    try {
      if (fixture && apiRequest) {
        await deleteSyntheticClosingFixture(apiRequest, fixture);
        fixture = null;
      }
    } finally {
      if (context) await context.close();
      if (apiRequest) await apiRequest.dispose();
    }
  });

  test("#2 AM preview includes 10:00 only (excludes PM/EMG times)", async () => {
    if (!fixture || !context) throw new Error("synthetic closing fixture missing");
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await openPreview(accounting, page, fixture.targetDate, "午前");
      await expect(accounting.billingDetailTimeCell("10:00")).toHaveCount(1);
      await expect(accounting.billingDetailTimeCell("13:30")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("14:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("20:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("02:00")).toHaveCount(0);
      await expect(accounting.billingDetailsHeading()).toContainText("1件");
      // Scenario #2: payment-method theoretical sales visible (fixture uses cash).
      await expect(page.getByText("現金", { exact: true }).first()).toBeVisible();
      await expect(page.getByText("理論現金").first()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("#3 PM preview includes 14:00; excludes 10:00, 20:00, and overnight 02:00", async () => {
    if (!fixture || !context) throw new Error("synthetic closing fixture missing");
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await openPreview(accounting, page, fixture.targetDate, "午後");
      await expect(accounting.billingDetailTimeCell("14:00")).toHaveCount(1);
      await expect(accounting.billingDetailTimeCell("13:30")).toHaveCount(1);
      await expect(accounting.billingDetailTimeCell("10:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("20:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("02:00")).toHaveCount(0);
    } finally {
      await page.close();
    }
  });

  test("#4 EMG preview includes 20:00 and next-day 02:00 on same target day", async () => {
    if (!fixture || !context) throw new Error("synthetic closing fixture missing");
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await openPreview(accounting, page, fixture.targetDate, "緊急");
      await expect(accounting.billingDetailTimeCell("20:00")).toHaveCount(1);
      await expect(accounting.billingDetailTimeCell("02:00")).toHaveCount(1);
      await expect(accounting.billingDetailTimeCell("10:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("14:00")).toHaveCount(0);
    } finally {
      await page.close();
    }
  });

  test("#5 13:30:00 is PM (half-open boundary), not AM", async () => {
    if (!fixture || !context) throw new Error("synthetic closing fixture missing");
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      await openPreview(accounting, page, fixture.targetDate, "午前");
      await expect(accounting.billingDetailTimeCell("13:30")).toHaveCount(0);

      await openPreview(accounting, page, fixture.targetDate, "午後");
      await expect(accounting.billingDetailTimeCell("13:30")).toHaveCount(1);
    } finally {
      await page.close();
    }
  });

  test("#6 next-day EMG preview does not include overnight 02:00", async () => {
    if (!fixture || !context) throw new Error("synthetic closing fixture missing");
    const page = await context.newPage();
    const accounting = new AccountingPage(page);
    try {
      const nextDay = nextCalendarDay(fixture.targetDate);
      await openPreview(accounting, page, nextDay, "緊急");
      await expect(accounting.billingDetailTimeCell("02:00")).toHaveCount(0);
      await expect(accounting.billingDetailTimeCell("20:00")).toHaveCount(0);
    } finally {
      await page.close();
    }
  });
});
