import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/testing/mocks/node";
import { CertificateApiError } from "./certificate-errors";
import { createHttpCertificateApi } from "./certificate-http-client";

/**
 * 実 API クライアント（契約モックではない）の形を MSW で確認する。
 * MSW はこのテスト内の通信応答を差し替えるだけで、クライアントにモックへの分岐は無い。
 * 契約は selected / runtime 未実装のため、実 BE との結合は未検証。
 */
const BASE = "/api/v1";
const KEY = "11111111-1111-4111-8111-111111111111";

describe("実 API クライアント", () => {
  const api = createHttpCertificateApi();

  it("context を /v1/vaccinations/{id}/certificate-context から読み、契約外の応答は malformed で拒否する", async () => {
    server.use(http.get(`${BASE}/vaccinations/101/certificate-context`, () => HttpResponse.json({ nope: true })));
    await expect(api.getContext(101)).rejects.toMatchObject({ kind: "malformed" });
  });

  it("発行 POST は /certificate-editions へ operation_key 付きで送り、C06 結果をそのまま返す", async () => {
    let received: unknown;
    server.use(
      http.post(`${BASE}/vaccinations/101/certificate-editions`, async ({ request }) => {
        received = await request.json();
        return HttpResponse.json(
          {
            status: "succeeded",
            operation_key: KEY,
            result_refs: { vaccination_id: "101", series_id: "201", edition_id: "301", artifact_id: "401" },
            committed_versions: { edition: "1", source_version: "7" },
            recorded_at: "2026-10-20T14:30:00Z",
          },
          { status: 201 },
        );
      }),
    );
    const body = {
      operation_key: KEY,
      kind: "normal" as const,
      layout: "portrait" as const,
      inputs: {},
      preview_fingerprint: "a".repeat(64),
      preview_display_issue_date: "2026-10-20",
      preconditions: { source_version: 7, latest_edition_id: null },
    };
    const result = await api.issue(101, body);
    expect(received).toEqual(body);
    expect(result.status).toBe("succeeded");
  });

  it("flat error を code / status で分類する（409 dependency-changed）", async () => {
    server.use(
      http.post(`${BASE}/vaccinations/101/certificate-editions`, () =>
        HttpResponse.json(
          { error: "changed", code: "dependency-changed", status: "dependency-changed", operation_key: KEY },
          { status: 409 },
        ),
      ),
    );
    const error = await api
      .issue(101, {
        operation_key: KEY,
        kind: "normal",
        layout: "portrait",
        inputs: {},
        preview_fingerprint: "a".repeat(64),
        preview_display_issue_date: "2026-10-20",
        preconditions: { source_version: 7, latest_edition_id: null },
      })
      .catch((e: unknown) => e);
    expect(error).toBeInstanceOf(CertificateApiError);
    expect(error).toMatchObject({ kind: "dependency-changed", httpStatus: 409 });
  });

  it("lookup は intent を query に付け、404 scope-not-found を未 claim として分類する", async () => {
    let intent: string | null = null;
    server.use(
      http.get(`${BASE}/vaccination-certificate-operations/${KEY}`, ({ request }) => {
        intent = new URL(request.url).searchParams.get("intent");
        return HttpResponse.json({ error: "x", code: "scope-not-found", status: "scope-not-found" }, { status: 404 });
      }),
    );
    await expect(api.lookupOperation(KEY, "issue")).rejects.toMatchObject({ kind: "scope-not-found" });
    expect(intent).toBe("issue");
  });

  it("履歴は kind / page / limit を query に付ける", async () => {
    let search = "";
    server.use(
      http.get(`${BASE}/vaccinations/101/certificate-editions`, ({ request }) => {
        search = new URL(request.url).search;
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
      }),
    );
    await api.listEditions(101, { kind: "normal" });
    expect(search).toContain("kind=normal");
    expect(search).toContain("page=1");
    expect(search).toContain("limit=20");
  });

  it("PDF 本体は /content から Blob で取得する", async () => {
    server.use(
      http.get(`${BASE}/vaccination-certificate-artifacts/401/content`, () =>
        new HttpResponse("%PDF-1.1 mock", { headers: { "Content-Type": "application/pdf" } }),
      ),
    );
    const blob = await api.getArtifactContent("401");
    expect(blob.size).toBeGreaterThan(0);
  });

  it("200 でも PDF ではない応答（HTML・JSON・magic 不一致）は malformed で拒否する", async () => {
    const url = `${BASE}/vaccination-certificate-artifacts/401/content`;
    server.use(http.get(url, () => new HttpResponse("<html>login</html>", { headers: { "Content-Type": "text/html" } })));
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });

    server.use(http.get(url, () => HttpResponse.json({ ok: true })));
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });

    // Content-Type だけ PDF を名乗っても、本体が PDF でなければ拒否する。
    server.use(http.get(url, () => new HttpResponse("<html>login</html>", { headers: { "Content-Type": "application/pdf" } })));
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });
  });

  it("本体が PDF なら Content-Type の誤ラベルでは拒否せず、application/pdf に正規化して返す", async () => {
    server.use(
      http.get(`${BASE}/vaccination-certificate-artifacts/401/content`, () =>
        new HttpResponse("%PDF-1.1 mock", { headers: { "Content-Type": "application/octet-stream" } }),
      ),
    );
    const blob = await api.getArtifactContent("401");
    expect(blob.type).toBe("application/pdf");
    expect(blob.size).toBeGreaterThan(0);
  });

  it("PDF 取得の失敗 body（Blob で届く JSON）も code で分類する", async () => {
    server.use(
      http.get(`${BASE}/vaccination-certificate-artifacts/401/content`, () =>
        HttpResponse.json({ error: "x", code: "artifact_unavailable" }, { status: 503 }),
      ),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "render-unavailable" });
  });
});
