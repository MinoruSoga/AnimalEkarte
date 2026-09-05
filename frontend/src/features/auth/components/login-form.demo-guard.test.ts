import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// #91 / DEC-7 / SEC-CS2-F01: SHOW_DEMO is local Vite DEV or Vercel preview (STG).
// Production and unknown VERCEL_ENV stay fail-closed. VITE_SHOW_DEMO_ACCOUNTS
// must not enable production (.env.production currently sets the flag true).
describe("LoginForm SHOW_DEMO — DEV or Vercel preview (#91 / SEC-CS2-F01)", () => {
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
    vi.stubEnv("VITE_VERCEL_ENV", "production");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });

  it("__VERCEL_ENV__=preview かつ非 DEV では SHOW_DEMO=true（STG）", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "preview");
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "false");
    vi.stubEnv("VITE_VERCEL_ENV", "preview");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(true);
  });

  // DEC-7: 非 Vercel ビルド（__VERCEL_ENV__=""）は fail-closed。DEV=false の production ビルド相当。
  it('__VERCEL_ENV__="" かつ非 DEV では VITE_SHOW_DEMO_ACCOUNTS=true でも SHOW_DEMO=false', async () => {
    vi.stubGlobal("__VERCEL_ENV__", "");
    vi.stubEnv("DEV", false);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "true");
    vi.stubEnv("VITE_VERCEL_ENV", "");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(false);
  });

  it("DEV=true では SHOW_DEMO=true（ローカル開発）", async () => {
    vi.stubGlobal("__VERCEL_ENV__", "");
    vi.stubEnv("DEV", true);
    vi.stubEnv("VITE_SHOW_DEMO_ACCOUNTS", "false");
    vi.stubEnv("VITE_VERCEL_ENV", "");

    const mod = await import("./LoginForm");

    expect(mod.SHOW_DEMO).toBe(true);
  });

  it("vite.config は preview/production の API を Cloudflare ホストに define する", () => {
    const frontendRoot = join(dirname(fileURLToPath(import.meta.url)), "../../../..");
    const src = readFileSync(join(frontendRoot, "vite.config.ts"), "utf8");
    expect(src).toContain('define["import.meta.env.VITE_API_URL"]');
    expect(src).toContain("https://api.stg.noah-karte.com/api");
    expect(src).toContain("https://api.noah-karte.com/api");
    expect(src).not.toContain("elb.amazonaws.com");

    const envProduction = readFileSync(join(frontendRoot, ".env.production"), "utf8");
    expect(envProduction).toContain("https://api.stg.noah-karte.com/api");
    expect(envProduction).not.toContain("elb.amazonaws.com");
  });

  it("デモパスワードは SHOW_DEMO が真のときだけ返す", () => {
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
    expect(fn).toMatch(/if\s*\(\s*!SHOW_DEMO\s*\)/);
    expect(fn).toContain('"password"');
    expect(src).not.toContain("VITE_DEMO_LOGIN_PASSWORD");
  });
});
