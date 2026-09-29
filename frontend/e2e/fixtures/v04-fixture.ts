import { expect, type Browser, type BrowserContext } from "@playwright/test";

/**
 * V04 view-only account fixture contract.
 *
 * `E2E_CLINICAL_FIXTURE` is the one-line JSON emitted by
 * `cmd/clinical-e2e-fixture setup`. The nested `v04` object carries the
 * view-only account email and the V04-prefixed master rows. It deliberately
 * has no `clinicId` key — run-e2e.sh extracts `clinicId` with a greedy sed
 * over the whole line, so a duplicate key would corrupt the shell parse.
 */
export interface V04Fixture {
  clinicId: number;
  viewOnlyEmail: string;
  viewOnlyResources: string[];
  cageName: string;
  labDeviceName: string;
}

const V04_ROW_PREFIX = "V04-e2e-";
const VIEW_ONLY_EMAIL_SUFFIX = "@example.test";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function readName(value: unknown): string | null {
  return typeof value === "string" && value.trim() !== "" ? value : null;
}

function readClinicId(value: unknown): number | null {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0) {
    return null;
  }
  return value;
}

function readResources(value: unknown): string[] | null {
  if (!Array.isArray(value) || value.length === 0) {
    return null;
  }
  const resources: string[] = [];
  for (const item of value) {
    if (typeof item !== "string" || item.trim() === "") {
      return null;
    }
    resources.push(item);
  }
  return resources;
}

export function parseV04Fixture(raw: string | undefined): V04Fixture {
  if (raw === undefined || raw.trim() === "") {
    throw new Error("clinical e2e fixture is missing");
  }
  let decoded: unknown;
  try {
    decoded = JSON.parse(raw) as unknown;
  } catch {
    throw new Error("clinical e2e fixture is not JSON");
  }
  if (!isRecord(decoded)) {
    throw new Error("clinical e2e fixture is not an object");
  }
  const clinicId = readClinicId(decoded.clinicId);
  if (clinicId === null) {
    throw new Error("clinical e2e fixture clinicId is invalid");
  }
  if (clinicId === 1 || clinicId === 2) {
    throw new Error("clinical e2e clinicId is reserved");
  }
  const v04 = decoded.v04;
  if (!isRecord(v04)) {
    throw new Error("clinical e2e fixture v04 block is missing");
  }
  const viewOnlyEmail = readName(v04.viewOnlyEmail);
  if (viewOnlyEmail === null || !viewOnlyEmail.endsWith(VIEW_ONLY_EMAIL_SUFFIX)) {
    throw new Error("v04 viewOnlyEmail is not a synthetic example.test address");
  }
  const viewOnlyResources = readResources(v04.viewOnlyResources);
  if (viewOnlyResources === null) {
    throw new Error("v04 viewOnlyResources must be a non-empty string array");
  }
  const cageName = readName(v04.cageName);
  if (cageName === null || !cageName.startsWith(V04_ROW_PREFIX)) {
    throw new Error("v04 cageName is missing or lacks the V04-e2e- prefix");
  }
  const labDeviceName = readName(v04.labDeviceName);
  if (labDeviceName === null || !labDeviceName.startsWith(V04_ROW_PREFIX)) {
    throw new Error("v04 labDeviceName is missing or lacks the V04-e2e- prefix");
  }
  return { clinicId, viewOnlyEmail, viewOnlyResources, cageName, labDeviceName };
}

/** null when the fixture env is absent (non-clinical suites); throws when set but invalid. */
export function readV04FixtureFromEnv(): V04Fixture | null {
  const raw = process.env.E2E_CLINICAL_FIXTURE;
  if (raw === undefined || raw.trim() === "") {
    return null;
  }
  return parseV04Fixture(raw);
}

/**
 * Fresh browser context logged in as the V04 view-only account.
 *
 * Deliberately does NOT reuse the shared admin storage state from
 * helpers/auth.ts — the view-only session must be its own cookie jar. Also
 * does NOT wait for 当日の受付: the view-only account has no reception
 * permission, so the post-login landing page is not the reception board.
 */
export async function createViewOnlyContext(
  browser: Browser,
  fixture: V04Fixture,
): Promise<BrowserContext> {
  const password = process.env.E2E_LOGIN_PASSWORD ?? "";
  if (password === "") {
    throw new Error("E2E_LOGIN_PASSWORD must be set for the V04 view-only account");
  }

  const context = await browser.newContext();
  const loginPage = await context.newPage();
  try {
    await loginPage.goto("/login", { waitUntil: "domcontentloaded" });

    // The login form renders after initialAuthPromise resolves; fresh contexts
    // also download all Vite modules through the Docker bridge (~35-50 s worst).
    const emailInput = loginPage.locator("#login-email");
    await emailInput.waitFor({ state: "visible", timeout: 60000 });
    await emailInput.fill(fixture.viewOnlyEmail);
    await loginPage.locator("#login-password").fill(password);

    const loginResponsePromise = loginPage.waitForResponse(
      (response) => response.url().includes("/v1/login") && response.request().method() === "POST",
      { timeout: 60000 },
    );
    await loginPage.getByRole("button", { name: "ログイン" }).click();
    const loginResponse = await loginResponsePromise;
    expect(loginResponse.status()).toBe(200);
  } finally {
    await loginPage.close();
  }
  return context;
}
