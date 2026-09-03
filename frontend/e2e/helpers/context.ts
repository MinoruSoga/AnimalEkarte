import { type Browser, type BrowserContext } from "@playwright/test";
import { loginAsDemoAdmin } from "./auth";

/**
 * Create a fresh browser context that is already authenticated as the demo admin.
 *
 * Extracted verbatim from the `beforeAll` block that was duplicated across every
 * flow spec:
 *
 * ```ts
 * context = await browser.newContext();
 * const loginPage = await context.newPage();
 * await loginAsDemoAdmin(loginPage);
 * await loginPage.close();
 * ```
 *
 * The context lifecycle (one shared context per file, login page opened then
 * closed) is identical — no Playwright fixture/worker-scope change.
 */
export async function createAuthedContext(browser: Browser): Promise<BrowserContext> {
  const context = await browser.newContext();
  const loginPage = await context.newPage();
  await loginAsDemoAdmin(loginPage);
  await loginPage.close();
  return context;
}
