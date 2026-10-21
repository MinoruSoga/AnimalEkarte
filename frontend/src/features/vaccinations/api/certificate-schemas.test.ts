import { describe, expect, it } from "vitest";
import { classifyCertificateError } from "./certificate-errors";
import {
  certificateContextSchema,
  certificateEditionSchema,
  certificatePreviewResponseSchema,
  certificateResultSchema,
  certificateSavedPreviewSchema,
  isPdfBase64,
  readSucceededRefs,
  type CertificateSucceededResult,
} from "./certificate-schemas";
import { buildSnapshot, mockHash, mockPdfBase64, sampleRabiesInputs } from "../mock/certificate-fixtures";

const edition = {
  id: 301,
  series_id: 202,
  vaccination_id: 101,
  kind: "rabies",
  edition: 1,
  source_version: 7,
  previous_edition_id: null,
  issued_at: "2026-10-20T14:30:00Z",
  issued_by: 9,
  validity: "current",
  snapshot: buildSnapshot({ kind: "rabies", layout: "rabies", sourceVersion: 7, rabiesEligible: true, inputs: sampleRabiesInputs() }),
  snapshot_sha256: mockHash("x"),
  available_layouts: ["rabies"],
};

describe("証明書 wire の解析", () => {
  it("契約の合成 snapshot（全 null キー・4 LOT）を受理する", () => {
    expect(certificateEditionSchema.safeParse(edition).success).toBe(true);
  });

  it("int64 が JS で安全に表せない ID は丸めずに拒否する", () => {
    expect(certificateEditionSchema.safeParse({ ...edition, id: Number.MAX_SAFE_INTEGER + 2 }).success).toBe(false);
  });

  it("snapshot の必須キー欠落を補完せず拒否する", () => {
    const withoutOwner = { ...edition.snapshot, owner: undefined };
    expect(certificateEditionSchema.safeParse({ ...edition, snapshot: withoutOwner }).success).toBe(false);
  });

  it("context.initial_inputs は料金 4 欄 null を受理し、省略欄を null に正規化する", () => {
    const parsed = certificateContextSchema.parse({
      vaccination_id: 101,
      source_version: 7,
      eligibility: { state: "unknown", allowed_kinds: ["normal"], provenance: "legacy_unresolved" },
      can_issue: true,
      latest: { normal: null, rabies: null },
      initial_inputs: { registration_fee: null, injection_fee: null, tag_fee: null, total: null },
    });
    expect(parsed.initial_inputs.certificate_number).toBeNull();
    expect(parsed.initial_inputs.total).toBeNull();
  });

  it("成功 result の参照は string のまま取り出し、欠落は null", () => {
    const ok = certificateResultSchema.parse({
      status: "succeeded",
      operation_key: "11111111-1111-4111-8111-111111111111",
      result_refs: { vaccination_id: "101", series_id: "201", edition_id: "301", artifact_id: "401" },
      committed_versions: { edition: "1", source_version: "7" },
      recorded_at: "2026-10-20T14:30:00Z",
    }) as CertificateSucceededResult;
    expect(readSucceededRefs(ok)).toEqual({ editionId: "301", artifactId: "401" });
    expect(readSucceededRefs({ ...ok, result_refs: { edition_id: "301" } })).toBeNull();
  });

  it("not-succeeded は status と operation_key のみ", () => {
    expect(
      certificateResultSchema.safeParse({ status: "not-succeeded", operation_key: "11111111-1111-4111-8111-111111111111" }).success,
    ).toBe(true);
  });
});

describe("証明書 API 失敗の分類", () => {
  it("契約表の HTTP / code の組合せを機械判定する", () => {
    const k = (status: number | undefined, body: object | null = null) =>
      classifyCertificateError(status, body as never).kind;
    expect(k(401)).toBe("unauthenticated");
    expect(k(403, { code: "edit-right-lost", status: "edit-right-lost" })).toBe("forbidden");
    expect(k(404, { code: "scope-not-found", status: "scope-not-found" })).toBe("scope-not-found");
    expect(k(404, { code: "not_found" })).toBe("not-found");
    expect(k(422, { code: "validation", status: "validation" })).toBe("validation");
    expect(k(409, { code: "version-conflict", status: "version-conflict" })).toBe("version-conflict");
    expect(k(409, { code: "dependency-changed", status: "dependency-changed" })).toBe("dependency-changed");
    expect(k(409, { code: "operation-key-conflict", status: "operation-key-conflict" })).toBe("operation-key-conflict");
    expect(k(409, { code: "operation-in-progress", status: "operation-in-progress" })).toBe("operation-in-progress");
    expect(k(503, { code: "result-unavailable", status: "result-unavailable" })).toBe("result-unavailable");
    expect(k(503, { code: "render_failed" })).toBe("render-unavailable");
    expect(k(503, { code: "artifact_unavailable" })).toBe("render-unavailable");
    expect(k(500)).toBe("server");
    expect(k(undefined)).toBe("network");
  });

  it("未知の 409 は成功とも前提変更とも推定せず server 扱い", () => {
    expect(classifyCertificateError(409, { code: "something-else" }).kind).toBe("server");
  });
});

describe("preview PDF payload の検証", () => {
  it("base64 として正しく先頭が %PDF- のものだけ受理する", () => {
    expect(isPdfBase64(mockPdfBase64())).toBe(true);
    expect(isPdfBase64("")).toBe(false);
    expect(isPdfBase64("not base64!!")).toBe(false);
    expect(isPdfBase64(btoa("<html>login</html>"))).toBe(false);
    expect(isPdfBase64(`${mockPdfBase64()}=`)).toBe(false); // 長さが 4 の倍数でない
  });

  it("確認できない PDF を含む preview 応答は両 schema で拒否する", () => {
    const snapshot = buildSnapshot({ kind: "normal", layout: "portrait", sourceVersion: 7, rabiesEligible: false });
    const base = {
      snapshot,
      snapshot_sha256: mockHash("x"),
      preview_fingerprint: mockHash("y"),
      source_version: 7,
      latest_edition_id: null,
      media_type: "application/pdf",
      template_version: "t",
      renderer_version: "r",
    };
    expect(certificatePreviewResponseSchema.safeParse({ ...base, pdf_base64: mockPdfBase64() }).success).toBe(true);
    expect(certificatePreviewResponseSchema.safeParse({ ...base, pdf_base64: "AAAA" }).success).toBe(false);
    expect(certificatePreviewResponseSchema.safeParse({ ...base, pdf_base64: "%%%%%%%%" }).success).toBe(false);

    const saved = { edition_id: 301, layout: "portrait", media_type: "application/pdf", template_version: "t", renderer_version: "r", content_sha256: mockHash("z") };
    expect(certificateSavedPreviewSchema.safeParse({ ...saved, pdf_base64: mockPdfBase64() }).success).toBe(true);
    expect(certificateSavedPreviewSchema.safeParse({ ...saved, pdf_base64: btoa("<html>login</html>") }).success).toBe(false);
  });
});
