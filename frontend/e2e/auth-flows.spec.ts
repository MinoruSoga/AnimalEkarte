import { test, expect } from "@playwright/test";
import { LoginPage } from "./pages/login-page";

// E2E tests for authentication flows: login, forgot-password, reset-password.
// These are unauthenticated-only pages; no loginAsDemoAdmin() helper used.
// Auth forms use React 19 useActionState; validation is server-side.
// SEC-CS2-F01: valid-login credentials come from E2E_LOGIN_* env only (no in-repo secrets).

function requireE2ELoginCredentials(): { email: string; password: string } {
  const email = process.env.E2E_LOGIN_EMAIL?.trim() ?? "";
  const password = process.env.E2E_LOGIN_PASSWORD ?? "";
  if (!email || !password) {
    throw new Error(
      "E2E_LOGIN_EMAIL and E2E_LOGIN_PASSWORD must be set (no in-repo credential fallback; see frontend/e2e/README.md)",
    );
  }
  return { email, password };
}

test.describe("認証フロー E2E", () => {
  test("/login — ログインページが表示される", async ({ page }) => {
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

  test("/login — メールアドレス入力フィールドが機能する", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const emailInput = loginPage.emailInput();
    await emailInput.waitFor({ state: "visible", timeout: 60000 });

    await emailInput.fill("test@example.com");
    await expect(emailInput).toHaveValue("test@example.com");
  });

  test("/login — 有効な認証情報でログインできる", async ({ page }) => {
    const { email, password } = requireE2ELoginCredentials();
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
    // After login, should redirect to home
    await expect(loginPage.homeHeading()).toBeVisible({ timeout: 60000 });
  });

  test("/login — 無効な認証情報でエラーが表示される", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const emailInput = loginPage.emailInput();
    await emailInput.waitFor({ state: "visible", timeout: 60000 });

    // Intentionally non-secret invalid pair — must not use repository-known passwords.
    await emailInput.fill("invalid-login@example.invalid");
    await loginPage.passwordInput().fill("not-a-valid-password");

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

  test("/login — パスワード表示切り替えボタンが機能する", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();
    const passwordInput = loginPage.passwordInput();
    await passwordInput.waitFor({ state: "visible", timeout: 60000 });

    // Initially password type="password"
    await expect(passwordInput).toHaveAttribute("type", "password");

    // Click eye button to show
    await loginPage.showPasswordButton().click();

    // Should change to type="text"
    await expect(passwordInput).toHaveAttribute("type", "text");
  });

  test("/login — パスワード再設定リンクから公開画面へ到達できる", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.gotoLogin();

    await page.getByRole("link", { name: "パスワードをお忘れですか？" }).click();

    await expect(page).toHaveURL(/\/forgot-password$/);
    await expect(page.getByRole("heading", { name: "パスワードのリセット" })).toBeVisible();
  });

  test("/forgot-password — ページにアクセスできる", async ({ page }) => {
    await page.goto("/forgot-password/", { waitUntil: "domcontentloaded", timeout: 60000 });

    await expect(page).toHaveURL(/\/forgot-password\/?$/);
    await expect(page.getByRole("heading", { name: "パスワードのリセット" })).toBeVisible();
  });

  test("/reset-password — ページにアクセスできる", async ({ page }) => {
    await page.goto("/reset-password/#token=test-token", {
      waitUntil: "domcontentloaded",
      timeout: 60000,
    });

    await expect(page).toHaveURL(/\/reset-password$/);
    await expect(page.getByRole("heading", { name: "新しいパスワードの設定" })).toBeVisible();
  });

  test("/reset-password — token欠落は画面固有エラーのまま留まる", async ({ page }) => {
    await page.goto("/reset-password/", { waitUntil: "domcontentloaded", timeout: 60000 });

    await expect(page).toHaveURL(/\/reset-password\/$/);
    await expect(page.getByRole("heading", { name: "無効なリンクです" })).toBeVisible();
  });

  test("/reset-password — 不正tokenのAPIエラーでもloginへ遷移しない", async ({ page }) => {
    await page.route("**/v1/auth/reset-password", async (route) => {
      await route.fulfill({
        status: 400,
        contentType: "application/json",
        body: JSON.stringify({ message: "invalid or expired token" }),
      });
    });
    await page.goto("/reset-password?token=invalid-token", {
      waitUntil: "domcontentloaded",
      timeout: 60000,
    });

    await page.getByLabel("新しいパスワード").fill("password123");
    await page.getByLabel("パスワード（確認）").fill("password123");
    await page.getByRole("button", { name: "パスワードを設定する" }).click();

    await expect(page).toHaveURL(/\/reset-password$/);
    await expect(
      page.getByText(
        "パスワードのリセットに失敗しました。リンクの有効期限が切れている可能性があります。",
      ),
    ).toBeVisible();
  });
});
