import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// Primary execution: scripts/run-e2e.sh (mcr.microsoft.com/playwright Docker image, connects to localhost:3003 via host.docker.internal).
// Fallback: macOS native `pnpm test:e2e` (requires playwright browsers installed on host).
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe('飼主一覧 かな検索', () => {
  test('未ログイン時は /owners にアクセスすると /login にリダイレクトされる', async ({ browser }) => {
    // Use a fresh context with no auth state.
    // domcontentloaded avoids waiting for all Vite ES-module requests; the
    // client-side redirect (/login) fires after JS loads so we poll via waitForURL.
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    await freshPage.goto('/owners', { waitUntil: 'domcontentloaded' });
    await freshPage.waitForURL(/\/login/, { timeout: 30000 });
    await freshContext.close();
  });

  test.describe('ログイン後', () => {
    let loggedInContext: BrowserContext;

    test.beforeAll(async ({ browser }) => {
      loggedInContext = await browser.newContext();
      const loginPage = await loggedInContext.newPage();
      await loginAsDemoAdmin(loginPage);
      await loginPage.close();
    });

    test.afterAll(async () => {
      await loggedInContext.close();
    });

    test('ひらがな「ぴ」で検索するとカタカナ「ピーター」が表示される', async () => {
      const page = await loggedInContext.newPage();
      try {
        await page.goto('/owners', { waitUntil: 'domcontentloaded' });
        await expect(page).toHaveURL(/\/owners/);

        // Open search bar (toggle button with aria-label="検索")
        await page.getByRole('button', { name: '検索' }).click();

        // Search input appears
        const searchInput = page.getByPlaceholder('飼主名、ペット名、飼主No、種別...');
        await expect(searchInput).toBeVisible();

        // Type hiragana ぴ — normalizeKana converts katakana ピ→ぴ, so ぴーたー matches ぴ
        await searchInput.fill('ぴ');

        // Pet name ピーター should be visible in the filtered table
        await expect(page.getByText('ピーター')).toBeVisible({ timeout: 5000 });
      } finally {
        await page.close();
      }
    });

    test('カタカナ「ピ」で検索しても「ピーター」が表示される (ひらがな・カタカナ統一検索)', async () => {
      const page = await loggedInContext.newPage();
      try {
        await page.goto('/owners', { waitUntil: 'domcontentloaded' });
        await expect(page).toHaveURL(/\/owners/);

        await page.getByRole('button', { name: '検索' }).click();
        const searchInput = page.getByPlaceholder('飼主名、ペット名、飼主No、種別...');
        await expect(searchInput).toBeVisible();

        // Type katakana ピ — normalizeKana(ピ) === normalizeKana(ぴ) so both match
        await searchInput.fill('ピ');

        await expect(page.getByText('ピーター')).toBeVisible({ timeout: 5000 });
      } finally {
        await page.close();
      }
    });
  });
});
