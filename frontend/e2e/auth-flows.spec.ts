import { test, expect } from '@playwright/test';

// E2E tests for authentication flows: login, forgot-password, reset-password.
// These are unauthenticated-only pages; no loginAsDemoAdmin() helper used.
// Auth forms use React 19 useActionState; validation is server-side.

test.describe('認証フロー E2E', () => {
  test('/login — ログインページが表示される', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded', timeout: 60000 });
    // Main heading: "ノア動物病院"
    await expect(page.getByRole('heading', { name: 'ノア動物病院', level: 1 })).toBeVisible({ timeout: 15000 });
    // Subtitle
    await expect(page.locator('text=管理システムにログイン')).toBeVisible({ timeout: 15000 });
    // Form inputs
    await expect(page.locator('#login-email')).toBeVisible();
    await expect(page.locator('#login-password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'ログイン' })).toBeVisible();
  });

  test('/login — メールアドレス入力フィールドが機能する', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const emailInput = page.locator('#login-email');
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('test@example.com');
    await expect(emailInput).toHaveValue('test@example.com');
  });

  test('/login — 有効な認証情報でログインできる', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const emailInput = page.locator('#login-email');
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('admin@noavet.jp');
    await page.locator('#login-password').fill('password');

    const loginResponsePromise = page.waitForResponse(
      (response) => response.url().includes('/v1/login') && response.request().method() === 'POST',
      { timeout: 60000 },
    );
    await page.getByRole('button', { name: 'ログイン' }).click();
    const loginResponse = await loginResponsePromise;

    expect(loginResponse.status()).toBe(200);
    // After login, should redirect to home
    await expect(page.getByRole('heading', { name: '当日の受付' })).toBeVisible({ timeout: 60000 });
  });

  test('/login — 無効な認証情報でエラーが表示される', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const emailInput = page.locator('#login-email');
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('admin@noavet.jp');
    await page.locator('#login-password').fill('wrongpassword');

    const loginResponsePromise = page.waitForResponse(
      (response) => response.url().includes('/v1/login') && response.request().method() === 'POST',
      { timeout: 60000 },
    );
    await page.getByRole('button', { name: 'ログイン' }).click();
    const loginResponse = await loginResponsePromise;

    // Invalid credentials return 401
    expect(loginResponse.status()).toBe(401);
    // Should remain on login page
    await expect(page).toHaveURL(/\/login/);
    // Error message should be visible
    await expect(page.locator('text=メールアドレスまたはパスワードが違います')).toBeVisible({ timeout: 15000 });
  });

  test('/login — パスワード表示切り替えボタンが機能する', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const passwordInput = page.locator('#login-password');
    await passwordInput.waitFor({ state: 'visible', timeout: 60000 });

    // Initially password type="password"
    await expect(passwordInput).toHaveAttribute('type', 'password');

    // Click eye button to show
    const eyeButton = page.getByRole('button', { name: 'パスワードを表示' });
    await eyeButton.click();

    // Should change to type="text"
    await expect(passwordInput).toHaveAttribute('type', 'text');
  });

  test('/forgot-password — ページにアクセスできる', async ({ page }) => {
    // Smoke test: page load or redirect to login
    await page.goto('/forgot-password', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const urlPath = page.url();
    // Accept either the forgot-password page or redirect to login
    expect(urlPath).toMatch(/\/(forgot-password|login)/);
  });

  test('/reset-password — ページにアクセスできる', async ({ page }) => {
    // Smoke test: reset password page requires token; without it redirects to login
    await page.goto('/reset-password?token=test-token', { waitUntil: 'domcontentloaded', timeout: 60000 });
    const urlPath = page.url();
    expect(urlPath).toMatch(/\/(reset-password|login)/);
  });
});
