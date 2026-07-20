import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// #91: LoginForm.tsx の SHOW_DEMO はコメント上 computeShowDemoAccounts と同一ロジックを
// インライン化していると主張していたが、実装には __VERCEL_ENV__ 本番ガードが欠落していた。
// VITE_SHOW_DEMO_ACCOUNTS=true が Vercel Production にも設定された場合、デモアカウント一覧
// （実メールアドレス群）がログイン画面に露出する回帰を防止する。
describe("LoginForm SHOW_DEMO — Vercel Production 一次ガード (#91)", () => {
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
