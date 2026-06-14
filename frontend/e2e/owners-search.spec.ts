import { test, expect } from '@playwright/test';

// Primary execution: scripts/run-e2e.sh (mcr.microsoft.com/playwright Docker image, connects to localhost:3003 via host.docker.internal).
// Fallback: macOS native `pnpm test:e2e` (requires playwright browsers installed on host).

// Demo credentials (seed: 003_seed_demo.sql)
const DEMO_EMAIL = 'admin@noavet.jp';
const DEMO_PASSWORD = 'password';

test.describe('飼主一覧 かな検索', () => {
  test('未ログイン時は /owners にアクセスすると /login にリダイレクトされる', async ({ page, browser }) => {
    // Use a fresh context with no auth state
    const freshContext = await browser.newContext();
    const freshPage = await freshContext.newPage();
    await freshPage.goto('/owners');
    await expect(freshPage).toHaveURL(/\/login/);
    await freshContext.close();
  });

  test.describe('ログイン後', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login');
      await page.locator('#login-email').fill(DEMO_EMAIL);
      await page.locator('#login-password').fill(DEMO_PASSWORD);
      await page.getByRole('button', { name: 'ログイン' }).click();
      // Successful login navigates away from /login
      await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
    });

    test('ひらがな「ぴ」で検索するとカタカナ「ピーター」が表示される', async ({ page }) => {
      await page.goto('/owners');
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
    });

    test('カタカナ「ピ」で検索しても「ピーター」が表示される (ひらがな・カタカナ統一検索)', async ({ page }) => {
      await page.goto('/owners');
      await expect(page).toHaveURL(/\/owners/);

      await page.getByRole('button', { name: '検索' }).click();
      const searchInput = page.getByPlaceholder('飼主名、ペット名、飼主No、種別...');
      await expect(searchInput).toBeVisible();

      // Type katakana ピ — normalizeKana(ピ) === normalizeKana(ぴ) so both match
      await searchInput.fill('ピ');

      await expect(page.getByText('ピーター')).toBeVisible({ timeout: 5000 });
    });
  });
});
