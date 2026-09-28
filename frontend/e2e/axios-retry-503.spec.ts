import { test, expect, type Page, type Route } from "@playwright/test";

// EMR-203 / PERF-E5 synthetic field check: in a real browser, the app's shared
// axios instance (frontend/src/lib/axios.ts, loaded from the running Vite dev
// server) must NOT retry a GET that receives 503, while 502 stays retryable
// (MAX_RETRIES = 2, backoff 1s + 2s).
//
// Every /api/ request is answered by page.route, so nothing reaches the backend
// or the shared DB. The only app-issued GET before login is the startup
// session restore (GET /v1/me), which opts out of all retries regardless of
// status, so it cannot discriminate the 503 rule; the check therefore issues a
// regular GET through the app's own axios module on the public /login page.

const AXIOS_MODULE_PATH = "/src/lib/axios.ts";
const TARGET_PATH = "/v1/masters/animal-species";
const TARGET_URL_PATTERN = `**/api${TARGET_PATH}`;
const API_URL_PATTERN = "**/api/**";
const MAX_RETRIES = 2;
const RETRY_WINDOW_MS = 3000;

interface GetOutcome {
  outcome: "resolved" | "rejected";
  status: number | null;
  elapsedMs: number;
}

async function stubApi(page: Page, targetStatus: number): Promise<{ hits: () => number }> {
  let targetHits = 0;
  // Registered first so the more specific target route (registered later) wins.
  await page.route(API_URL_PATTERN, (route: Route) =>
    route.fulfill({
      status: 401,
      contentType: "application/json",
      json: { error: "unauthorized" },
    }),
  );
  await page.route(TARGET_URL_PATTERN, (route: Route) => {
    if (route.request().method() !== "GET") return route.fallback();
    targetHits += 1;
    return route.fulfill({
      status: targetStatus,
      contentType: "application/json",
      json: { error: targetStatus === 503 ? "service_unavailable" : "bad_gateway" },
    });
  });
  return { hits: () => targetHits };
}

async function getViaAppAxios(page: Page): Promise<GetOutcome> {
  return page.evaluate(
    async ({ modulePath, path }) => {
      const mod: unknown = await import(/* @vite-ignore */ modulePath);
      const client =
        typeof mod === "object" && mod !== null && "axios" in mod
          ? (mod as { axios: { get: (url: string) => Promise<{ status: number }> } }).axios
          : null;
      if (client === null) throw new Error(`axios export not found in ${modulePath}`);
      const started = performance.now();
      try {
        const response = await client.get(path);
        return {
          outcome: "resolved",
          status: response.status,
          elapsedMs: performance.now() - started,
        };
      } catch (error: unknown) {
        const status =
          typeof error === "object" &&
          error !== null &&
          "response" in error &&
          typeof (error as { response?: { status?: unknown } }).response?.status === "number"
            ? (error as { response: { status: number } }).response.status
            : null;
        return { outcome: "rejected", status, elapsedMs: performance.now() - started };
      }
    },
    { modulePath: AXIOS_MODULE_PATH, path: TARGET_PATH },
  ) as Promise<GetOutcome>;
}

test.describe("EMR-203 axios GET retry — synthetic browser check", () => {
  test("GET receiving 503 is sent once and not retried", async ({ page }, testInfo) => {
    const stub = await stubApi(page, 503);
    await page.goto("/login");

    const result = await getViaAppAxios(page);
    // Wait past the whole retry window to prove no late retry is scheduled.
    await page.waitForTimeout(RETRY_WINDOW_MS + 1000);
    const hits = stub.hits();

    await testInfo.attach("emr-203-503", {
      body: JSON.stringify({ ...result, requestCount: hits }),
      contentType: "application/json",
    });
    console.log(`[EMR-203] 503 case: ${JSON.stringify({ ...result, requestCount: hits })}`);

    expect(result.outcome).toBe("rejected");
    expect(result.status).toBe(503);
    expect(result.elapsedMs).toBeLessThan(1000);
    expect(hits).toBe(1);
  });

  test("GET receiving 502 is retried MAX_RETRIES times (control)", async ({ page }, testInfo) => {
    const stub = await stubApi(page, 502);
    await page.goto("/login");

    const result = await getViaAppAxios(page);
    const hits = stub.hits();

    await testInfo.attach("emr-203-502", {
      body: JSON.stringify({ ...result, requestCount: hits }),
      contentType: "application/json",
    });
    console.log(`[EMR-203] 502 case: ${JSON.stringify({ ...result, requestCount: hits })}`);

    expect(result.outcome).toBe("rejected");
    expect(result.status).toBe(502);
    expect(result.elapsedMs).toBeGreaterThanOrEqual(RETRY_WINDOW_MS - 100);
    expect(hits).toBe(1 + MAX_RETRIES);
  });
});
