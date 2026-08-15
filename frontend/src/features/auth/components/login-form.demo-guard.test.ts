import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// #91 / DEC-7 / SEC-CS2-F01: SHOW_DEMO is local Vite DEV only.
// Vercel preview + VITE_SHOW_DEMO_ACCOUNTS no longer enables the demo UI.
describe("LoginForm SHOW_DEMO — local DEV only (#91 / SEC-CS2-F01)", () => {
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

  // SEC-CS2-F01: preview でも demo UI は出さない（local DEV only）。
  it("__VERCEL_ENV__=preview では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "preview");
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });

  // DEC-7: 非 Vercel ビルド（__VERCEL_ENV__=""）は fail-closed。DEV=false の production ビルド相当。
  it("__VERCEL_ENV__=\"\" かつ非 DEV では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "");
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });

  it("DEV=true では SHOW_DEMO=true（ローカル開発）", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "");
    vi.stubEnv("DEV", true);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "false");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(true);
  });
});
