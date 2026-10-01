// frontend/worker/index.ts の経路規則を binding mock で検証する。
// vercel.json 由来の契約: filesystem 優先 -> /line-reserve/* fallback -> /* fallback、
// /api/* は backend Worker へ素通し。
// NOTE: 環境 pragma (vitest-environment) を node にしない — setup.ts が jsdom の
// document に触れるため全テストが失敗する。jsdom 環境でも Request/Response/URL は利用可。

import { describe, expect, it, vi } from "vitest";

import worker, { type FrontendWorkerEnv } from "./index";

function makeEnv(options: { assetStatus?: number; apiResponse?: Response }): {
  env: FrontendWorkerEnv;
  apiFetch: ReturnType<typeof vi.fn>;
} {
  const apiFetch = vi.fn(async () => options.apiResponse ?? new Response("api", { status: 200 }));
  const assetsFetch = async (request: Request) => {
    const { pathname } = new URL(request.url);
    // Workers assets は index.html への fetch をディレクトリ URL へ 307 するため、
    // SPA fallback はディレクトリパスを要求する(worker/index.ts 参照)。
    const exists =
      pathname === "/" ||
      pathname === "/line-reserve/" ||
      pathname === "/liff/" ||
      pathname === "/assets/app.js";
    return new Response(`asset:${pathname}`, {
      status: exists ? 200 : (options.assetStatus ?? 404),
    });
  };
  return {
    env: { ASSETS: { fetch: assetsFetch }, API: { fetch: apiFetch } },
    apiFetch,
  };
}

describe("frontend worker routing", () => {
  it("serves existing static assets directly", async () => {
    const { env } = makeEnv({});
    const res = await worker.fetch(new Request("https://stg.example.test/assets/app.js"), env);
    expect(res.status).toBe(200);
    expect(await res.text()).toBe("asset:/assets/app.js");
  });

  it("falls back unknown paths to /index.html", async () => {
    const { env } = makeEnv({});
    const res = await worker.fetch(new Request("https://stg.example.test/owners"), env);
    expect(res.status).toBe(200);
    expect(await res.text()).toBe("asset:/");
  });

  it("falls back /line-reserve/{clinicId}/ to its own entrypoint", async () => {
    const { env } = makeEnv({});
    const res = await worker.fetch(
      new Request("https://stg.example.test/line-reserve/clinic-1/"),
      env,
    );
    expect(await res.text()).toBe("asset:/line-reserve/");
  });

  it("falls back /liff/* to /liff/index.html", async () => {
    const { env } = makeEnv({});
    const res = await worker.fetch(
      new Request("https://stg.example.test/liff/clinic-1/health"),
      env,
    );
    expect(await res.text()).toBe("asset:/liff/");
  });

  it("proxies /api/* to the API service binding without touching ASSETS", async () => {
    const { env, apiFetch } = makeEnv({
      apiResponse: new Response("{}", { status: 201 }),
    });
    const res = await worker.fetch(
      new Request("https://stg.example.test/api/liff/c1/settings", {
        method: "POST",
        body: "{}",
      }),
      env,
    );
    expect(res.status).toBe(201);
    expect(apiFetch).toHaveBeenCalledOnce();
    expect(new URL(apiFetch.mock.calls[0][0].url).pathname).toBe("/api/liff/c1/settings");
  });

  it("does not SPA-fallback non-GET requests", async () => {
    const { env } = makeEnv({});
    const res = await worker.fetch(
      new Request("https://stg.example.test/nope", { method: "POST" }),
      env,
    );
    expect(res.status).toBe(404);
  });
});
