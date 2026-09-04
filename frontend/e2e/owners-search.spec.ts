import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { DEMO_PETER_PET } from "./helpers/demo-seed";
import { OwnersPage } from "./pages/owners-page";

// Primary execution: scripts/run-e2e.sh (mcr.microsoft.com/playwright Docker image, connects to localhost:3003 via host.docker.internal).
// Fallback: macOS native `pnpm test:e2e` (requires playwright browsers installed on host).
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe("飼主一覧 かな検索", () => {
  test("未ログイン時は /owners にアクセスすると /login にリダイレクトされる", async ({
    browser,
  }) => {
    // Use a fresh context with no auth state.
    // domcontentloaded avoids waiting for all Vite ES-module requests; the
    // Browser redirect (/login) fires after JS loads, so poll via waitForURL.
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    await new OwnersPage(freshPage).gotoList();
    await freshPage.waitForURL(/\/login/, { timeout: 30000 });
    await freshContext.close();
  });

  test.describe("ログイン後", () => {
    let loggedInContext: BrowserContext;

    test.beforeAll(async ({ browser }) => {
      loggedInContext = await createAuthedContext(browser);
    });

    test.afterAll(async () => {
      await loggedInContext.close();
    });

    test("ひらがな「ぴーたー」で検索するとカタカナ「ピーター」が表示される", async () => {
      const page = await loggedInContext.newPage();
      const owners = new OwnersPage(page);
      try {
        await owners.gotoList();
        await expect(page).toHaveURL(/\/owners/);

        // Open search bar (toggle button with aria-label="検索")
        await page.getByRole("button", { name: "検索" }).click();

        // Search input appears
        const searchInput = owners.searchInput();
        await expect(searchInput).toBeVisible();

        // Full hiragana form — single-char「ぴ」matches 500+ pets and may omit ピーター on page 1.
        // Owners list loader uses GET /v1/pets with NormalizeKana symmetry.
        await searchInput.fill(DEMO_PETER_PET.hiraganaSearch);

        await expect(owners.peterText()).toBeVisible({ timeout: 15000 });
      } finally {
        await page.close();
      }
    });

    test("カタカナ「ピーター」で検索しても「ピーター」が表示される (ひらがな・カタカナ統一検索)", async () => {
      const page = await loggedInContext.newPage();
      const owners = new OwnersPage(page);
      try {
        await owners.gotoList();
        await expect(page).toHaveURL(/\/owners/);

        await page.getByRole("button", { name: "検索" }).click();
        const searchInput = owners.searchInput();
        await expect(searchInput).toBeVisible();

        await searchInput.fill(DEMO_PETER_PET.katakanaSearch);

        await expect(owners.peterText()).toBeVisible({ timeout: 15000 });
      } finally {
        await page.close();
      }
    });
  });
});
