import { describe, it, expect } from "vitest";
import { computeShowDemoAccounts } from "./show-demo-accounts";

// M-10 / SEC-CS2-F01: demo accounts UI is local Vite DEV or Vercel preview (STG).
// Production and unknown VERCEL_ENV stay hidden. The flag must not enable production.
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

  it("preview（STG）は flag 無しでも true", () => {
    expect(
      computeShowDemoAccounts({
        vercelEnv: "preview",
        isDev: false,
        showDemoAccountsFlag: undefined,
      }),
    ).toBe(true);
  });

  it("VITE_SHOW_DEMO_ACCOUNTS だけでは非 preview を有効化しない", () => {
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "true" }),
    ).toBe(false);
    expect(
      computeShowDemoAccounts({ vercelEnv: "", isDev: false, showDemoAccountsFlag: "false" }),
    ).toBe(false);
  });
});
