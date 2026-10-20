import type {
  CertificateContext,
  CertificateEdition,
  CertificateEditionPage,
  CertificateInputs,
  CertificateKind,
  CertificateLayout,
  CertificatePreviewResponse,
  CertificateResult,
  CertificateSavedPreview,
} from "./certificate-schemas";

/**
 * EMR-487: 証明書 API の注入可能な境界。
 * 実 API 実装（certificate-http-client）とモック実装（../mock）は別ファイル・別オブジェクトで、
 * 呼び出し側が明示的に選ぶ。バックエンド失敗時のモックへの自動フォールバックは存在しない。
 */

/** 現在照合用 CAS。Content とは分離され、再送 identity に含めない。 */
export interface CertificatePreconditions {
  source_version: number;
  latest_edition_id: number | null;
}

export interface CertificatePreviewRequest {
  kind: CertificateKind;
  layout: CertificateLayout;
  /** normal は `{}`（10 欄すべて null/省略のみ許可）。rabies は 10 欄を明示 null 含め送る。 */
  inputs: Partial<CertificateInputs>;
  preconditions: CertificatePreconditions;
}

export interface CertificateIssueRequest extends CertificatePreviewRequest {
  operation_key: string;
  preview_fingerprint: string;
  preview_display_issue_date: string;
}

export interface CertificateReprintRequest {
  operation_key: string;
  layout: CertificateLayout;
  preconditions: { source_version: number; latest_edition_id: number };
}

export type CertificateOperationIntent = "issue" | "reprint";

export interface CertificateApi {
  /** 1. GET certificate-context */
  getContext(vaccinationId: number, signal?: AbortSignal): Promise<CertificateContext>;
  /** 2. POST certificate-preview（発行前・純計算。receipt/版/監査を作らない） */
  previewForIssue(
    vaccinationId: number,
    body: CertificatePreviewRequest,
    signal?: AbortSignal,
  ): Promise<CertificatePreviewResponse>;
  /** 3. POST certificate-editions（発行。C06 ResultResponse のみ返る） */
  issue(vaccinationId: number, body: CertificateIssueRequest): Promise<CertificateResult>;
  /** 4. GET certificate-editions/{edition_id} */
  getEdition(editionId: string, signal?: AbortSignal): Promise<CertificateEdition>;
  /** 5. GET certificate-editions?kind=&page=&limit= */
  listEditions(
    vaccinationId: number,
    query: { kind: CertificateKind; page?: number; limit?: number },
    signal?: AbortSignal,
  ): Promise<CertificateEditionPage>;
  /** 6. POST certificate-editions/{id}/preview（保存版・現在有効版のみ） */
  previewSaved(
    editionId: string,
    layout: CertificateLayout,
    signal?: AbortSignal,
  ): Promise<CertificateSavedPreview>;
  /** 7. POST certificate-editions/{id}/reprints */
  reprint(editionId: string, body: CertificateReprintRequest): Promise<CertificateResult>;
  /** 8. GET artifacts/{artifact_id}/content（印刷に使う唯一の PDF） */
  getArtifactContent(artifactId: string, signal?: AbortSignal): Promise<Blob>;
  /** 9. GET certificate-operations/{operation_key}?intent= */
  lookupOperation(
    operationKey: string,
    intent: CertificateOperationIntent,
    signal?: AbortSignal,
  ): Promise<CertificateResult>;
}
