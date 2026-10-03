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

  // EMR-255: API 呼出は frontend Worker の service binding 経由の same-origin /api が
  // 正本。vite.config が VITE_API_URL を注入したり .env.production が絶対 URL を
  // 持つと、PROD ビルドが STG API を叩く事故に戻るため禁止する。
  it("vite.config は VITE_API_URL を注入せず APP_ENV で環境を分岐する（EMR-255）", () => {
    const frontendRoot = join(dirname(fileURLToPath(import.meta.url)), "../../../..");
    const src = readFileSync(join(frontendRoot, "vite.config.ts"), "utf8");
    expect(src).not.toContain('define["import.meta.env.VITE_API_URL"]');
    expect(src).toContain("process.env.APP_ENV");
    // APP_ENV=stg は demo 表示契約の "preview" へ写像する（fail-closed 維持）。
    expect(src).toContain('"preview"');
    expect(src).not.toContain("elb.amazonaws.com");

    const envProduction = readFileSync(join(frontendRoot, ".env.production"), "utf8");
    expect(envProduction).not.toMatch(/^\s*VITE_API_URL\s*=/m);
    expect(envProduction).not.toContain("elb.amazonaws.com");
  });

  // EMR-265: STG(preview)のデモ共通パスワードは vite.config.ts の define が
  // GitHub secret STG_DEMO_PASSWORD を appEnv==="stg" のときだけ焼き込む。
  // production/未知環境では "" が焼き込まれる(fail-closed)。
  it("デモパスワード自動入力: DEV は公開定数、preview は VITE_DEMO_LOGIN_PASSWORD（STG はシークレット配布）", () => {
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
    // ローカル開発は公開定数、STG(preview)はビルド時注入値を返すことを pin する。
    expect(fn).toMatch(
      /import\.meta\.env\.DEV\s*\?\s*"password"\s*:\s*import\.meta\.env\.VITE_DEMO_LOGIN_PASSWORD/,
    );
    // 公開定数 "password" を STG ビルドに焼く経路がないこと（SEEDLOGIN_DEMO_PASSWORD
    // が正本で repo-public 値は STG では認証されない）。
    expect(fn).not.toMatch(/import\.meta\.env\.DEV\s*\|\|/);
  });

  it('vite.config は VITE_DEMO_LOGIN_PASSWORD を stg のみ注入し production では "" を焼き込む（EMR-265）', () => {
    const frontendRoot = join(dirname(fileURLToPath(import.meta.url)), "../../../..");
    const src = readFileSync(join(frontendRoot, "vite.config.ts"), "utf8");
    // define 経由で appEnv==="stg" のときだけ process.env.STG_DEMO_PASSWORD を焼く。
    expect(src).toContain('"import.meta.env.VITE_DEMO_LOGIN_PASSWORD"');
    expect(src).toMatch(
      /appEnv\s*===\s*"stg"\s*\?\s*\(process\.env\.STG_DEMO_PASSWORD\s*\?\?\s*""\)\s*:\s*""/,
    );
  });
});
