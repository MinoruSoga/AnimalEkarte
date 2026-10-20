import { describe, expect, it } from "vitest";
import type { CertificateIssueRequest, CertificatePreviewRequest } from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { certificateContextSchema, certificateEditionSchema, certificateResultSchema, emptyCertificateInputs } from "../api/certificate-schemas";
import { createMockCertificateApi } from "./certificate-mock-api";

const KEY = "11111111-1111-4111-8111-111111111111";

function normalRequest(): CertificatePreviewRequest {
  return { kind: "normal", layout: "portrait", inputs: {}, preconditions: { source_version: 7, latest_edition_id: null } };
}

async function issueNormal(api: ReturnType<typeof createMockCertificateApi>["api"], key = KEY) {
  const preview = await api.previewForIssue(101, normalRequest());
  const body: CertificateIssueRequest = {
    ...normalRequest(),
    operation_key: key,
    preview_fingerprint: preview.preview_fingerprint,
    preview_display_issue_date: preview.snapshot.display_issue_date,
  };
  return { preview, body, result: await api.issue(101, body) };
}

describe("契約モック adapter（実 API ではない）", () => {
  it("応答は契約の zod 解析を通る", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    expect(certificateContextSchema.safeParse(await api.getContext(101)).success).toBe(true);
    const { result } = await issueNormal(api);
    expect(certificateResultSchema.safeParse(result).success).toBe(true);
    const page = await api.listEditions(101, { kind: "normal" });
    expect(certificateEditionSchema.safeParse(page.data[0]).success).toBe(true);
  });

  it("rabies は eligible のときだけ。false / unknown では 422", async () => {
    for (const eligible of [false, null]) {
      const { api } = createMockCertificateApi({ rabiesEligible: eligible });
      const error = await api
        .previewForIssue(101, { kind: "rabies", layout: "rabies", inputs: {}, preconditions: { source_version: 7, latest_edition_id: null } })
        .catch((e: unknown) => e);
      expect(error).toMatchObject({ kind: "validation", httpStatus: 422 });
    }
  });

  it("normal で専用 10 欄のどれかが非 null（明示 0 を含む）なら 422", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    const error = await api
      .previewForIssue(101, { ...normalRequest(), inputs: { tag_fee: 0 } })
      .catch((e: unknown) => e);
    expect(error).toMatchObject({ kind: "validation" });
  });

  it("同 key・同内容の再送は replayed、同 key・別内容は 409 operation-key-conflict", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    const { body } = await issueNormal(api);
    const replay = await api.issue(101, body);
    expect(replay.status).toBe("replayed");
    const conflict = await api.issue(101, { ...body, layout: "landscape" }).catch((e: unknown) => e);
    expect(conflict).toMatchObject({ kind: "operation-key-conflict", httpStatus: 409 });
  });

  it("日跨ぎ(display_issue_date 不一致)・接種訂正後の古い preview は dependency-changed / version-conflict", async () => {
    const a = createMockCertificateApi({ rabiesEligible: false });
    const preview = await a.api.previewForIssue(101, normalRequest());
    a.controls.setToday("2026-10-21");
    const body = { ...normalRequest(), operation_key: KEY, preview_fingerprint: preview.preview_fingerprint, preview_display_issue_date: preview.snapshot.display_issue_date };
    await expect(a.api.issue(101, body)).rejects.toMatchObject({ kind: "dependency-changed" });

    const b = createMockCertificateApi({ rabiesEligible: false });
    const p2 = await b.api.previewForIssue(101, normalRequest());
    b.controls.bumpSourceVersion();
    await expect(
      b.api.issue(101, { ...normalRequest(), operation_key: KEY, preview_fingerprint: p2.preview_fingerprint, preview_display_issue_date: p2.snapshot.display_issue_date }),
    ).rejects.toMatchObject({ kind: "version-conflict" });
  });

  it("接種が訂正されると既発行版は旧版になり、保存プレビュー・再印刷・PDF 取得は 409", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const { result } = await issueNormal(api);
    if (result.status === "not-succeeded") throw new Error("unexpected");
    const editionId = result.result_refs.edition_id;
    const artifactId = result.result_refs.artifact_id;
    await expect(api.getArtifactContent(artifactId)).resolves.toBeInstanceOf(Blob);
    controls.bumpSourceVersion();
    expect((await api.getEdition(editionId)).validity).toBe("source_changed");
    await expect(api.previewSaved(editionId, "portrait")).rejects.toMatchObject({ kind: "dependency-changed" });
    await expect(api.getArtifactContent(artifactId)).rejects.toMatchObject({ kind: "dependency-changed" });
    await expect(
      api.reprint(editionId, { operation_key: "22222222-2222-4222-8222-222222222222", layout: "portrait", preconditions: { source_version: 8, latest_edition_id: Number(editionId) } }),
    ).rejects.toMatchObject({ kind: "dependency-changed" });
  });

  it("normal は縦横を同じ edition の別 artifact として再印刷できる（版は増えない）", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: false });
    const { result } = await issueNormal(api);
    if (result.status === "not-succeeded") throw new Error("unexpected");
    const editionId = result.result_refs.edition_id;
    const reprint = await api.reprint(editionId, {
      operation_key: "22222222-2222-4222-8222-222222222222",
      layout: "landscape",
      preconditions: { source_version: 7, latest_edition_id: Number(editionId) },
    });
    if (reprint.status === "not-succeeded") throw new Error("unexpected");
    expect(reprint.result_refs.edition_id).toBe(editionId);
    expect(reprint.result_refs.artifact_id).not.toBe(result.result_refs.artifact_id);
    expect((await api.listEditions(101, { kind: "normal" })).total).toBe(1);
  });

  it("応答だけ失われた発行は照会で成功が分かり、binding のみの失敗は not-succeeded、未 claim は 404", async () => {
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    const preview = await api.previewForIssue(101, normalRequest());
    const body: CertificateIssueRequest = { ...normalRequest(), operation_key: KEY, preview_fingerprint: preview.preview_fingerprint, preview_display_issue_date: preview.snapshot.display_issue_date };
    controls.failNext("issue", { error: new CertificateApiError("result-unavailable", { httpStatus: 503 }), commitFirst: true });
    await expect(api.issue(101, body)).rejects.toMatchObject({ kind: "result-unavailable" });
    expect((await api.lookupOperation(KEY, "issue")).status).toBe("succeeded");

    const other = "33333333-3333-4333-8333-333333333333";
    await expect(api.lookupOperation(other, "issue")).rejects.toMatchObject({ kind: "scope-not-found" });
    controls.failNext("issue", { error: new CertificateApiError("render-unavailable", { httpStatus: 503, code: "render_failed" }), claimOnly: true });
    await expect(api.issue(101, { ...body, operation_key: other })).rejects.toMatchObject({ kind: "render-unavailable" });
    expect((await api.lookupOperation(other, "issue")).status).toBe("not-succeeded");
  });

  it("初期入力は料金 4 欄とも null（会計・マスタから転記しない）", async () => {
    const { api } = createMockCertificateApi({ rabiesEligible: true });
    expect((await api.getContext(101)).initial_inputs).toEqual(emptyCertificateInputs());
  });
});
