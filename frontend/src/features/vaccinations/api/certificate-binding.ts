import type {
  CertificateEditionPage,
  CertificateKind,
  CertificatePreviewResponse,
  CertificateResult,
  CertificateSavedPreview,
  CertificateSucceededResult,
} from "./certificate-schemas";
import { readSucceededRefs, type CertificateLayout } from "./certificate-schemas";

/**
 * EMR-487: 応答が「要求した対象」のものかを確かめる純関数。
 * zod は形しか保証しない。形が正しいだけの別対象（別接種・別版・別操作）の応答を、
 * 画面に載せたり印刷対象にしたりしないための束縛検査。不一致は呼び出し側が契約違反として扱う。
 */

/** 発行前 preview の応答が、要求した接種・様式・CAS 前提に対するものか。 */
export function previewMatchesRequest(
  preview: CertificatePreviewResponse,
  expected: {
    vaccinationId: number;
    kind: CertificateKind;
    sourceVersion: number;
    latestEditionId: number | null;
  },
): boolean {
  const { snapshot } = preview;
  return (
    snapshot.source.vaccination_id === expected.vaccinationId &&
    snapshot.kind === expected.kind &&
    snapshot.source.version === preview.source_version &&
    preview.source_version === expected.sourceVersion &&
    preview.latest_edition_id === expected.latestEditionId
  );
}

/** 履歴 1 ページの全 edition が、要求した接種・様式のものか。 */
export function editionPageMatches(
  page: CertificateEditionPage,
  expected: { vaccinationId: number; kind: CertificateKind },
): boolean {
  return page.data.every(
    (edition) =>
      edition.vaccination_id === expected.vaccinationId &&
      edition.kind === expected.kind &&
      edition.snapshot.kind === expected.kind &&
      edition.snapshot.source.vaccination_id === expected.vaccinationId,
  );
}

/** 保存版 preview の応答が、要求した版・向きのものか。 */
export function savedPreviewMatches(
  preview: CertificateSavedPreview,
  expected: { editionId: string; layout: CertificateLayout },
): boolean {
  return String(preview.edition_id) === expected.editionId && preview.layout === expected.layout;
}

/** C06 結果が、待っている操作（operation_key）の結果か。成功・成功なしの両方に使う。 */
export function resultMatchesOperation(result: CertificateResult, operationKey: string): boolean {
  return result.operation_key === operationKey;
}

/**
 * 成功結果から edition / artifact の ID を取り出す。操作の key と、結果が指す接種（・版）が
 * 期待と一致するときだけ返す。欠落・不一致は null（契約違反。印刷対象にしない）。
 */
export function boundSucceededRefs(
  result: CertificateSucceededResult,
  expected: { operationKey: string; vaccinationId: number; editionId?: string },
): { editionId: string; artifactId: string } | null {
  if (!resultMatchesOperation(result, expected.operationKey)) return null;
  if (result.result_refs.vaccination_id !== String(expected.vaccinationId)) return null;
  const refs = readSucceededRefs(result);
  if (refs === null) return null;
  if (expected.editionId !== undefined && refs.editionId !== expected.editionId) return null;
  return refs;
}
