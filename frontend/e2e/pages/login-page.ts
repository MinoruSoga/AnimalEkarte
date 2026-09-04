import { type Locator, type Page, type Response } from "@playwright/test";
import { BasePage } from "./base-page";

/**
 * Login screen (`/login`). Encapsulates the `#login-email` / `#login-password`
 * id locators and the brand/error text that auth-flows asserted inline.
 *
 * Note: the demo-admin login *flow* used by every other spec lives in
 * `helpers/auth.ts` (safety-critical, untouched). This page object only serves
 * the auth-flows UI assertions.
 */
export class LoginPage extends BasePage {
  gotoLogin(): ReturnType<Page["goto"]> {
    return this.page.goto("/login", { waitUntil: "domcontentloaded", timeout: 60000 });
  }

  emailInput(): Locator {
    return this.page.locator("#login-email");
  }

  passwordInput(): Locator {
    return this.page.locator("#login-password");
  }

  submitButton(): Locator {
    return this.page.getByRole("button", { name: "ログイン" });
  }

  showPasswordButton(): Locator {
    return this.page.getByRole("button", { name: "パスワードを表示" });
  }

  /** Brand heading "ノア動物病院" (h1). */
  brandHeading(): Locator {
    return this.page.getByRole("heading", { name: "ノア動物病院", level: 1 });
  }

  /** Subtitle — was `locator('text=管理システムにログイン')`; getByText is the documented equivalent. */
  subtitle(): Locator {
    return this.page.getByText("管理システムにログイン");
  }

  /** Invalid-credentials error — was `locator('text=メールアドレスまたはパスワードが違います')`. */
  errorMessage(): Locator {
    return this.page.getByText("メールアドレスまたはパスワードが違います");
  }

  /** Authenticated home heading "当日の受付". */
  homeHeading(): Locator {
    return this.page.getByRole("heading", { name: "当日の受付" });
  }

  waitForLoginResponse(): Promise<Response> {
    return this.page.waitForResponse(
      (response) => response.url().includes("/v1/login") && response.request().method() === "POST",
      { timeout: 60000 },
    );
  }
}
