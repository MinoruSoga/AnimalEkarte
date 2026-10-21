import { describe, expect, it } from "vitest";
import { buildSnapshot, mockHash, mockPdfBase64 } from "../mock/certificate-fixtures";
import { boundSucceededRefs, previewMatchesRequest, resultMatchesOperation } from "./certificate-binding";
import {
  emptyCertificateInputs,
  type CertificatePreviewResponse,
  type CertificateSucceededResult,
} from "./certificate-schemas";

const KEY = "11111111-1111-4111-8111-111111111111";
const OTHER = "22222222-2222-4222-8222-222222222222";

const preview: CertificatePreviewResponse = {
  snapshot: buildSnapshot({ kind: "normal", layout: "portrait", sourceVersion: 7, rabiesEligible: false, inputs: emptyCertificateInputs() }),
  snapshot_sha256: mockHash("a"),
  preview_fingerprint: mockHash("b"),
  source_version: 7,
  latest_edition_id: null,
  pdf_base64: mockPdfBase64(),
  media_type: "application/pdf",
  template_version: "t",
  renderer_version: "r",
};
const expected = { vaccinationId: 101, kind: "normal" as const, sourceVersion: 7, latestEditionId: null };

describe("発行前 preview の束縛", () => {
  it("接種・様式・source_version・最新 edition が要求と一致するときだけ true", () => {
    expect(previewMatchesRequest(preview, expected)).toBe(true);
    expect(previewMatchesRequest(preview, { ...expected, vaccinationId: 102 })).toBe(false);
    expect(previewMatchesRequest(preview, { ...expected, kind: "rabies" })).toBe(false);
    expect(previewMatchesRequest(preview, { ...expected, sourceVersion: 8 })).toBe(false);
    expect(previewMatchesRequest(preview, { ...expected, latestEditionId: 5 })).toBe(false);
    expect(previewMatchesRequest({ ...preview, source_version: 8 }, { ...expected, sourceVersion: 8 })).toBe(false);
  });
});

const succeeded: CertificateSucceededResult = {
  status: "succeeded",
  operation_key: KEY,
  result_refs: { vaccination_id: "101", series_id: "201", edition_id: "301", artifact_id: "401" },
  committed_versions: { edition: "1", source_version: "7" },
  recorded_at: "2026-10-20T14:30:00Z",
};

describe("C06 結果の束縛", () => {
  it("operation_key は成功・成功なしの両方で照合する", () => {
    expect(resultMatchesOperation(succeeded, KEY)).toBe(true);
    expect(resultMatchesOperation({ status: "not-succeeded", operation_key: OTHER }, KEY)).toBe(false);
  });

  it("key・接種・（指定時は）edition が一致する成功だけ refs を返す", () => {
    expect(boundSucceededRefs(succeeded, { operationKey: KEY, vaccinationId: 101 })).toEqual({ editionId: "301", artifactId: "401" });
    expect(boundSucceededRefs(succeeded, { operationKey: OTHER, vaccinationId: 101 })).toBeNull();
    expect(boundSucceededRefs(succeeded, { operationKey: KEY, vaccinationId: 102 })).toBeNull();
    expect(boundSucceededRefs(succeeded, { operationKey: KEY, vaccinationId: 101, editionId: "301" })).not.toBeNull();
    expect(boundSucceededRefs(succeeded, { operationKey: KEY, vaccinationId: 101, editionId: "302" })).toBeNull();
    const noVaccination = { ...succeeded, result_refs: { edition_id: "301", artifact_id: "401" } };
    expect(boundSucceededRefs(noVaccination, { operationKey: KEY, vaccinationId: 101 })).toBeNull();
  });
});
