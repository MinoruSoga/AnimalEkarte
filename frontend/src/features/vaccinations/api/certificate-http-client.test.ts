import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vitest";
import { server } from "@/testing/mocks/node";
import { emptyCertificateInputs } from "./certificate-schemas";
import { buildSnapshot, mockHash, mockPdfBase64 } from "../mock/certificate-fixtures";
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
    server.use(
      http.get(`${BASE}/vaccinations/101/certificate-context`, () =>
        HttpResponse.json({ nope: true }),
      ),
    );
    await expect(api.getContext(101)).rejects.toMatchObject({ kind: "malformed" });
  });

  describe("edition の版の整合（EMR-487）", () => {
    const sourceVersion = 7;
    const edition = (overrides: { source_version?: number } = {}) => ({
      id: 301,
      series_id: 201,
      vaccination_id: 101,
      kind: "normal",
      edition: 1,
      source_version: sourceVersion,
      previous_edition_id: null,
      issued_at: "2026-10-20T14:30:00Z",
      issued_by: 9,
      validity: "current",
      snapshot: buildSnapshot({ kind: "normal", layout: "portrait", sourceVersion, rabiesEligible: false, inputs: emptyCertificateInputs() }),
      snapshot_sha256: mockHash("a"),
      available_layouts: ["portrait"],
      ...overrides,
    });
    const context = (latestNormal: unknown) => ({
      eligibility: { state: "ineligible", allowed_kinds: ["normal"], provenance: "master_snapshot" },
      can_issue: true,
      latest: { normal: latestNormal, rabies: null },
      initial_inputs: emptyCertificateInputs(),
    });

    it("getEdition は外側の版と snapshot の版が食い違う応答を malformed で拒否し、一致する応答は返す", async () => {
      const url = `${BASE}/vaccination-certificate-editions/301`;
      server.use(http.get(url, () => HttpResponse.json(edition())));
      await expect(api.getEdition("301")).resolves.toMatchObject({ id: 301, source_version: 7 });

      server.use(http.get(url, () => HttpResponse.json(edition({ source_version: 8 }))));
      await expect(api.getEdition("301")).rejects.toMatchObject({ kind: "malformed" });
    });

    it("getContext は latest の edition の版が食い違う応答を malformed で拒否し、null と一致する edition は受け入れる", async () => {
      const url = `${BASE}/vaccinations/101/certificate-context`;
      server.use(http.get(url, () => HttpResponse.json(context(null))));
      await expect(api.getContext(101)).resolves.toMatchObject({ latest: { normal: null, rabies: null } });

      server.use(http.get(url, () => HttpResponse.json(context(edition()))));
      await expect(api.getContext(101)).resolves.toMatchObject({ latest: { normal: { id: 301 } } });

      server.use(http.get(url, () => HttpResponse.json(context(edition({ source_version: 8 })))));
      await expect(api.getContext(101)).rejects.toMatchObject({ kind: "malformed" });
    });

    it("現在の接種版と違う過去の発行版の edition は、edition 自身が整合していれば正常として返す", async () => {
      // 発行時の版は 7。この接種の現在の版が進んでいても、edition は過去の発行版として正常。
      server.use(http.get(`${BASE}/vaccination-certificate-editions/301`, () => HttpResponse.json(edition())));
      await expect(api.getEdition("301")).resolves.toMatchObject({ source_version: 7 });
    });
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
            result_refs: {
              vaccination_id: "101",
              series_id: "201",
              edition_id: "301",
              artifact_id: "401",
            },
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
          {
            error: "changed",
            code: "dependency-changed",
            status: "dependency-changed",
            operation_key: KEY,
          },
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
        return HttpResponse.json(
          { error: "x", code: "scope-not-found", status: "scope-not-found" },
          { status: 404 },
        );
      }),
    );
    await expect(api.lookupOperation(KEY, "issue")).rejects.toMatchObject({
      kind: "scope-not-found",
    });
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

  it("PDF 本体は /content から Accept: application/pdf の Blob で取得する", async () => {
    let accept: string | null = null;
    server.use(
      http.get(`${BASE}/vaccination-certificate-artifacts/401/content`, ({ request }) => {
        accept = request.headers.get("Accept");
        return new HttpResponse("%PDF-1.1 mock", {
          headers: { "Content-Type": "application/pdf" },
        });
      }),
    );
    const blob = await api.getArtifactContent("401");
    expect(blob.size).toBeGreaterThan(0);
    expect(blob.type).toBe("application/pdf");
    expect(accept).toBe("application/pdf");
  });

  it("200 でも PDF ではない応答（HTML・JSON・magic 不一致）は malformed で拒否する", async () => {
    const url = `${BASE}/vaccination-certificate-artifacts/401/content`;
    server.use(
      http.get(
        url,
        () => new HttpResponse("<html>login</html>", { headers: { "Content-Type": "text/html" } }),
      ),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });

    server.use(http.get(url, () => HttpResponse.json({ ok: true })));
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });

    // Content-Type だけ PDF を名乗っても、本体が PDF でなければ拒否する。
    server.use(
      http.get(
        url,
        () =>
          new HttpResponse("<html>login</html>", {
            headers: { "Content-Type": "application/pdf" },
          }),
      ),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });
  });

  it("200 でも Content-Type が PDF でなければ、本体が %PDF- でも契約不正として拒否する（EMR-525）", async () => {
    const url = `${BASE}/vaccination-certificate-artifacts/401/content`;
    server.use(
      http.get(
        url,
        () =>
          new HttpResponse("%PDF-1.1 mock", {
            headers: { "Content-Type": "application/octet-stream" },
          }),
      ),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });

    server.use(
      http.get(url, () => new HttpResponse("%PDF-1.1 mock", { headers: { "Content-Type": "" } })),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({ kind: "malformed" });
  });

  it("Content-Type の parameter と大小文字は PDF 判定に影響しない", async () => {
    server.use(
      http.get(
        `${BASE}/vaccination-certificate-artifacts/401/content`,
        () =>
          new HttpResponse("%PDF-1.1 mock", {
            headers: { "Content-Type": "Application/PDF; charset=binary" },
          }),
      ),
    );
    const blob = await api.getArtifactContent("401");
    expect(blob.size).toBeGreaterThan(0);
  });

  it("PDF 取得の失敗 body（Blob で届く JSON）も code で分類する", async () => {
    server.use(
      http.get(`${BASE}/vaccination-certificate-artifacts/401/content`, () =>
        HttpResponse.json({ error: "x", code: "artifact_unavailable" }, { status: 503 }),
      ),
    );
    await expect(api.getArtifactContent("401")).rejects.toMatchObject({
      kind: "render-unavailable",
    });
  });

  describe("応答の対象束縛（形が正しいだけの別対象を載せない）", () => {
    const editionFor = (over: { vaccination_id?: number; kind?: "normal" | "rabies"; snapshotVaccinationId?: number }) => {
      const kind = over.kind ?? "normal";
      const snapshot = buildSnapshot({ kind, layout: kind === "normal" ? "portrait" : "rabies", sourceVersion: 7, rabiesEligible: kind === "rabies", inputs: emptyCertificateInputs() });
      return {
        id: 301,
        series_id: 201,
        vaccination_id: over.vaccination_id ?? 101,
        kind,
        edition: 1,
        source_version: 7,
        previous_edition_id: null,
        issued_at: "2026-10-20T14:30:00Z",
        issued_by: 9,
        validity: "current",
        snapshot: { ...snapshot, source: { ...snapshot.source, vaccination_id: over.snapshotVaccinationId ?? 101 } },
        snapshot_sha256: mockHash("x"),
        available_layouts: ["portrait", "landscape"],
      };
    };
    const useHistory = (item: ReturnType<typeof editionFor>) =>
      server.use(
        http.get(`${BASE}/vaccinations/101/certificate-editions`, () =>
          HttpResponse.json({ data: [item], total: 1, page: 1, limit: 20 }),
        ),
      );

    it("履歴: 要求した接種・様式の edition だけを返す", async () => {
      useHistory(editionFor({}));
      const page = await api.listEditions(101, { kind: "normal" });
      expect(page.data).toHaveLength(1);
    });

    it("履歴: 別接種・別様式・snapshot が別接種の edition を含む応答は malformed で拒否する", async () => {
      useHistory(editionFor({ vaccination_id: 999 }));
      await expect(api.listEditions(101, { kind: "normal" })).rejects.toMatchObject({ kind: "malformed" });
      useHistory(editionFor({ kind: "rabies" }));
      await expect(api.listEditions(101, { kind: "normal" })).rejects.toMatchObject({ kind: "malformed" });
      useHistory(editionFor({ snapshotVaccinationId: 999 }));
      await expect(api.listEditions(101, { kind: "normal" })).rejects.toMatchObject({ kind: "malformed" });
    });

    const savedBody = (over: { edition_id?: number; layout?: string }) => ({
      edition_id: over.edition_id ?? 301,
      layout: over.layout ?? "landscape",
      pdf_base64: mockPdfBase64(),
      media_type: "application/pdf",
      template_version: "t",
      renderer_version: "r",
      content_sha256: mockHash("z"),
    });

    it("保存版 preview: 要求した版・向きの応答だけを返し、別の版・向きは malformed で拒否する", async () => {
      const url = `${BASE}/vaccination-certificate-editions/301/preview`;
      server.use(http.post(url, () => HttpResponse.json(savedBody({}))));
      await expect(api.previewSaved("301", "landscape")).resolves.toMatchObject({ edition_id: 301 });

      server.use(http.post(url, () => HttpResponse.json(savedBody({ edition_id: 302 }))));
      await expect(api.previewSaved("301", "landscape")).rejects.toMatchObject({ kind: "malformed" });
      server.use(http.post(url, () => HttpResponse.json(savedBody({ layout: "portrait" }))));
      await expect(api.previewSaved("301", "landscape")).rejects.toMatchObject({ kind: "malformed" });
    });
  });
});
