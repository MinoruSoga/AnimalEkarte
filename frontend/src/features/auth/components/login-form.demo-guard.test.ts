import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// #91 / DEC-7: SHOW_DEMO は computeShowDemoAccounts と同ロジックのリテラル式。
// 旧実装は `__VERCEL_ENV__ !== "production" && (DEV || flag)` で production ガードは存在した。
// 欠陥は未定義 ""（非 Vercel ビルド）時に `!== "production"` が true になり fail-open すること。
// 本スイートは production 非表示・preview 表示・未定義時 fail-closed を固定する。
describe("LoginForm SHOW_DEMO — fail-closed ゲート (#91 / DEC-7)", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("__VERCEL_ENV__=production では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "production");
    // vitest 実行時は import.meta.env.DEV が true のため、本番ビルド相当にする
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });

  it("__VERCEL_ENV__=preview では VITE_SHOW_DEMO_ACCOUNTS=true のとき SHOW_DEMO=true", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "preview");
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(true);
  });

  // DEC-7: 非 Vercel ビルド（__VERCEL_ENV__=""）は fail-closed。DEV=false の production ビルド相当。
  it("__VERCEL_ENV__=\"\" かつ非 DEV では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "");
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });
});
