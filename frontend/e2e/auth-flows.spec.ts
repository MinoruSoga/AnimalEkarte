import { test, expect } from '@playwright/test';
import { LoginPage } from './pages/login-page';

// E2E tests for authentication flows: login, forgot-password, reset-password.
// These are unauthenticated-only pages; no loginAsDemoAdmin() helper used.
// Auth forms use React 19 useActionState; validation is server-side.

test.describe('認証フロー E2E', () => {
  test('/login — ログインページが表示される', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    // Main heading: "ノア動物病院"
    await expect(loginPage.brandHeading()).toBeVisible({ timeout: 15000 });
    // Subtitle
    await expect(loginPage.subtitle()).toBeVisible({ timeout: 15000 });
    // Form inputs
    await expect(loginPage.emailInput()).toBeVisible();
    await expect(loginPage.passwordInput()).toBeVisible();
    await expect(loginPage.submitButton()).toBeVisible();
  });

  test('/login — メールアドレス入力フィールドが機能する', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const emailInput = loginPage.emailInput();
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('test@example.com');
    await expect(emailInput).toHaveValue('test@example.com');
  });

  test('/login — 有効な認証情報でログインできる', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const emailInput = loginPage.emailInput();
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('admin@noavet.jp');
    await loginPage.passwordInput().fill('password');

    const loginResponsePromise = loginPage.waitForLoginResponse();
    await loginPage.submitButton().click();
    const loginResponse = await loginResponsePromise;

    expect(loginResponse.status()).toBe(200);
    // After login, should redirect to home
    await expect(loginPage.homeHeading()).toBeVisible({ timeout: 60000 });
  });

  test('/login — 無効な認証情報でエラーが表示される', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const emailInput = loginPage.emailInput();
    await emailInput.waitFor({ state: 'visible', timeout: 60000 });

    await emailInput.fill('admin@noavet.jp');
    await loginPage.passwordInput().fill('wrongpassword');

    const loginResponsePromise = loginPage.waitForLoginResponse();
    await loginPage.submitButton().click();
    const loginResponse = await loginResponsePromise;

    // Invalid credentials return 401
    expect(loginResponse.status()).toBe(401);
    // Should remain on login page
    await expect(page).toHaveURL(/\/login/);
    // Error message should be visible
    await expect(loginPage.errorMessage()).toBeVisible({ timeout: 15000 });
  });

  test('/login — パスワード表示切り替えボタンが機能する', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const passwordInput = loginPage.passwordInput();
    await passwordInput.waitFor({ state: 'visible', timeout: 60000 });

    // Initially password type="password"
    await expect(passwordInput).toHaveAttribute('type', 'password');

    // Click eye button to show
    await loginPage.showPasswordButton().click();

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
