import { describe, it, expect } from "vitest";
import { computeShowDemoAccounts } from "./show-demo-accounts";

// M-10: 本番ビルドで VITE_SHOW_DEMO_ACCOUNTS が誤って true のままでも
// Vercel Production では必ずデモアカウントを非表示にする回帰防止テスト。
describe("computeShowDemoAccounts", () => {
  it("VERCEL_ENV=production のとき、他のフラグに関わらず常に false", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "production", isDev: true, showDemoAccountsFlag: "true" }),
    ).toBe(false);
    expect(
      computeShowDemoAccounts({ vercelEnv: "production", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
  });

  it("dev モードでは VERCEL_ENV が production 以外なら true", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: true, showDemoAccountsFlag: undefined }),
    ).toBe(true);
  });

  it("preview（STG）で VITE_SHOW_DEMO_ACCOUNTS=true のときは true", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "preview", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(true);
  });

  it("VITE_SHOW_DEMO_ACCOUNTS が未設定/false のときは false", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "preview", isDev: false, showDemoAccountsFlag: undefined }),
    ).toBe(false);
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "false" }),
    ).toBe(false);
  });
});
