import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
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
  it('__VERCEL_ENV__="" かつ非 DEV では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false', async () => {
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

  it("VITE_DEMO_LOGIN_PASSWORD は import.meta.env.DEV の真分岐内でのみ読む", () => {
    const src = readFileSync(
      join(dirname(fileURLToPath(import.meta.url)), "LoginForm.tsx"),
      "utf8",
    );
    const fnStart = src.indexOf("function readDemoLoginPassword");
    expect(fnStart).toBeGreaterThanOrEqual(0);
    const brace = src.indexOf("{", fnStart);
    let depth = 0;
    let fnEnd = -1;
    for (let i = brace; i < src.length; i += 1) {
      if (src[i] === "{") depth += 1;
      if (src[i] === "}") {
        depth -= 1;
        if (depth === 0) {
          fnEnd = i;
          break;
        }
      }
    }
    expect(fnEnd).toBeGreaterThan(fnStart);
    const fn = src.slice(fnStart, fnEnd + 1);
    expect(fn).toMatch(/if\s*\(\s*import\.meta\.env\.DEV\s*\)/);
    const envRefs = [...src.matchAll(/import\.meta\.env\.VITE_DEMO_LOGIN_PASSWORD/g)];
    expect(envRefs.length).toBeGreaterThan(0);
    for (const match of envRefs) {
      expect(match.index).toBeGreaterThan(fnStart);
      expect(match.index).toBeLessThan(fnEnd);
    }
  });
});
