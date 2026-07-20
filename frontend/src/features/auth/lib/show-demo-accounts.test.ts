import { describe, it, expect } from "vitest";
import { computeShowDemoAccounts } from "./show-demo-accounts";

// M-10: 本番ビルドで VITE_SHOW_DEMO_ACCOUNTS が誤って true のままでも
// Vercel Production では必ずデモアカウントを非表示にする回帰防止テスト。
describe("computeShowDemoAccounts", () => {
  it("VERCEL_ENV=production かつ非 dev では flag=true でも false", () => {
    // 実 Vercel production ビルドでは import.meta.env.DEV は常に false。
    expect(
      computeShowDemoAccounts({ vercelEnv: "production", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
  });

  it("dev モードでは VERCEL_ENV に関わらず true（ローカル開発維持）", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: true, showDemoAccountsFlag: undefined }),
    ).toBe(true);
    expect(
      computeShowDemoAccounts({ vercelEnv: "production", isDev: true, showDemoAccountsFlag: "true" }),
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

  // DEC-7 / SEC-DEMO-FAILCLOSED: 非 Vercel ビルドでは __VERCEL_ENV__=""。
  // 旧式 `!== "production"` は fail-open になるため、未定義 × flag=true でも非表示にする。
  it("VERCEL_ENV 未定義（\"\"）かつ非 dev では flag=true でも false（fail-closed）", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
  });
});
