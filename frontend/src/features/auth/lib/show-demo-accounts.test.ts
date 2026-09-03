import { describe, it, expect } from "vitest";
import { computeShowDemoAccounts } from "./show-demo-accounts";

// M-10 / SEC-CS2-F01: demo accounts UI is local Vite DEV only.
// Vercel preview / production / flag must never expose privileged demo accounts.
describe("computeShowDemoAccounts", () => {
  it("VERCEL_ENV=production かつ非 dev では flag=true でも false", () => {
    // 実 Vercel production ビルドでは import.meta.env.DEV は常に false。
    expect(
      computeShowDemoAccounts({
        vercelEnv: "production",
        isDev: false,
        showDemoAccountsFlag: "true",
      }),
    ).toBe(false);
  });

  it("dev モードでは VERCEL_ENV に関わらず true（ローカル開発維持）", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: true, showDemoAccountsFlag: undefined }),
    ).toBe(true);
    expect(
      computeShowDemoAccounts({
        vercelEnv: "production",
        isDev: true,
        showDemoAccountsFlag: "true",
      }),
    ).toBe(true);
    expect(
      computeShowDemoAccounts({ vercelEnv: "preview", isDev: true, showDemoAccountsFlag: "true" }),
    ).toBe(true);
  });

  // SEC-CS2-F01: STG/preview ではデモアカウント UI を出さない（特権 demo 資格情報の露出防止）。
  it("preview（STG）で VITE_SHOW_DEMO_ACCOUNTS=true でも false（local DEV only）", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "preview", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
  });

  it("VITE_SHOW_DEMO_ACCOUNTS が未設定/false のときは false", () => {
    expect(
      computeShowDemoAccounts({
        vercelEnv: "preview",
        isDev: false,
        showDemoAccountsFlag: undefined,
      }),
    ).toBe(false);
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "false" }),
    ).toBe(false);
  });

  // DEC-7 / SEC-DEMO-FAILCLOSED: 非 Vercel ビルドでは __VERCEL_ENV__=""。
  it('VERCEL_ENV 未定義（""）かつ非 dev では flag=true でも false（fail-closed）', () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
  });
});
