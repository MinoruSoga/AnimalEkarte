import type {
  CertificateApi,
  CertificateIssueRequest,
  CertificateOperationIntent,
  CertificatePreviewRequest,
  CertificateReprintRequest,
} from "../api/certificate-api";
import { CertificateApiError } from "../api/certificate-errors";
import { certificateRecoveryQueryRequestSchema } from "../api/certificate-recovery-schemas";
import {
  emptyCertificateInputs,
  type CertificateContext,
  type CertificateEdition,
  type CertificateInputs,
  type CertificateKind,
  type CertificateLayout,
  type CertificateResult,
  type CertificateValidity,
} from "../api/certificate-schemas";
import {
  MOCK_SOURCE_VERSION,
  MOCK_TODAY,
  MOCK_VACCINATION_ID,
  buildSnapshot,
  mockHash,
  mockPdfBase64,
  mockPdfBlob,
} from "./certificate-fixtures";
import { createMockRecoveryLog, type MockRecoveryLog } from "./certificate-recovery-mock";

/**
 * EMR-487: 契約モック adapter（明示的に注入したときだけ使われる。実 API の代替ではない）。
 * - 契約（EMR-486 selected）の 9 操作の形・失敗コードを模す。BE 実装・実 renderer・実 PDF ではない。
 * - 本番コードからは import されない（index.ts から export しない）。テスト / Storybook 専用。
 */

export type MockOperationName =
  | "getContext"
  | "previewForIssue"
  | "issue"
  | "getEdition"
  | "listEditions"
  | "previewSaved"
  | "reprint"
  | "getArtifactContent"
  | "lookupOperation"
  | "queryOperationRecovery";

export interface MockInjectedFailure {
  error: CertificateApiError;
  /** true: 業務処理は確定させてから失敗を返す（結果不明＝応答だけ失われた状況の再現）。 */
  commitFirst?: boolean;
  /** true: 業務処理は行わず、binding だけ claim 済み（照会は not-succeeded を返す）。render 失敗等の再現。 */
  claimOnly?: boolean;
}

export interface MockCertificateOptions {
  /** rabies_certificate_eligible の snapshot。true / false / null(unknown) */
  rabiesEligible: boolean | null;
  canIssue?: boolean;
  sourceVersion?: number;
  /** このモックが扱う接種記録 ID。これ以外の ID への呼び出しは not-found（既定: MOCK_VACCINATION_ID）。 */
  vaccinationId?: number;
}

interface StoredOperation {
  intent: CertificateOperationIntent;
  contentSignature: string;
  result: CertificateResult;
}

interface StoredArtifact {
  editionId: number;
  layout: CertificateLayout;
}

export interface MockCertificateControls {
  /** 次の 1 回の呼び出しを失敗させる。 */
  failNext(operation: MockOperationName, failure: MockInjectedFailure): void;
  /** 接種が訂正され source_version が進んだ状況（既発行版は source_changed になる）。 */
  bumpSourceVersion(): void;
  /** 現在日付が変わった状況（preview の display_issue_date と不一致になる）。 */
  setToday(date: string): void;
  setCanIssue(value: boolean): void;
  /** 回復 discovery の claim 台帳（EMR-487 Slice A）。発見対象の状態をテストから操作する。 */
  recovery: Pick<MockRecoveryLog, "markUnavailable" | "closeWithoutEffect" | "seedUnresolved" | "get">;
  calls: { operation: MockOperationName; detail?: unknown }[];
}

export function createMockCertificateApi(options: MockCertificateOptions): {
  api: CertificateApi;
  controls: MockCertificateControls;
} {
  let sourceVersion = options.sourceVersion ?? MOCK_SOURCE_VERSION;
  let today = MOCK_TODAY;
  let canIssue = options.canIssue ?? true;
  const boundVaccinationId = options.vaccinationId ?? MOCK_VACCINATION_ID;
  let nextEditionId = 301;
  let nextArtifactId = 401;
  const editions: CertificateEdition[] = [];
  const operations = new Map<string, StoredOperation>();
  const artifacts = new Map<number, StoredArtifact>();
  const failures = new Map<MockOperationName, MockInjectedFailure>();
  // claim 済みの操作の台帳。lookup と回復 discovery の唯一の出所（二重管理しない）。
  const recoveryLog = createMockRecoveryLog();
  const calls: MockCertificateControls["calls"] = [];

  const eligibilityState = options.rabiesEligible === true ? "eligible" : options.rabiesEligible === false ? "ineligible" : "unknown";
  const allowedKinds: CertificateKind[] = options.rabiesEligible === true ? ["normal", "rabies"] : ["normal"];

  function latestOf(kind: CertificateKind): CertificateEdition | null {
    const list = editions.filter((e) => e.kind === kind);
    return list.length === 0 ? null : list[list.length - 1];
  }

  function validityOf(edition: CertificateEdition): CertificateValidity {
    if (edition.source_version !== sourceVersion) return "source_changed";
    if (latestOf(edition.kind)?.id !== edition.id) return "superseded";
    return "current";
  }

  function withValidity(edition: CertificateEdition): CertificateEdition {
    return { ...edition, validity: validityOf(edition) };
  }

  function findEdition(id: number | string): CertificateEdition {
    const found = editions.find((e) => String(e.id) === String(id));
    if (!found) throw new CertificateApiError("not-found", { httpStatus: 404, code: "not_found" });
    return found;
  }

  /** 契約どおり、別の接種記録の ID は存在しない扱い（404）にする。 */
  function assertScope(vaccinationId: number) {
    if (vaccinationId !== boundVaccinationId) {
      throw new CertificateApiError("not-found", { httpStatus: 404, code: "not_found" });
    }
  }

  function takeFailure(operation: MockOperationName): MockInjectedFailure | undefined {
    const failure = failures.get(operation);
    failures.delete(operation);
    return failure;
  }

  function track<T>(operation: MockOperationName, detail: unknown, run: () => T): T {
    calls.push({ operation, detail });
    return run();
  }

  function assertInputsValid(request: CertificatePreviewRequest): CertificateInputs {
    if (!allowedKinds.includes(request.kind)) {
      throw new CertificateApiError("validation", { httpStatus: 422, code: "validation", c06Status: "validation" });
    }
    const layoutOk =
      request.kind === "normal"
        ? request.layout === "portrait" || request.layout === "landscape"
        : request.layout === "rabies";
    if (!layoutOk) {
      throw new CertificateApiError("validation", { httpStatus: 422, code: "validation", c06Status: "validation" });
    }
    const full: CertificateInputs = { ...emptyCertificateInputs(), ...request.inputs };
    if (request.kind === "normal") {
      const dirty = Object.values(request.inputs).some((v) => v !== null && v !== undefined);
      if (dirty) {
        throw new CertificateApiError("validation", { httpStatus: 422, code: "validation", c06Status: "validation" });
      }
    }
    return full;
  }

  function assertCas(preconditions: CertificatePreviewRequest["preconditions"], kind: CertificateKind) {
    const latest = latestOf(kind);
    if (preconditions.source_version !== sourceVersion || (preconditions.latest_edition_id ?? null) !== (latest?.id ?? null)) {
      throw new CertificateApiError("version-conflict", {
        httpStatus: 409,
        code: "version-conflict",
        c06Status: "version-conflict",
      });
    }
  }

  function snapshotFor(request: CertificatePreviewRequest, inputs: CertificateInputs) {
    return buildSnapshot({
      kind: request.kind,
      layout: request.layout,
      sourceVersion,
      rabiesEligible: options.rabiesEligible,
      inputs,
      displayIssueDate: today,
      vaccinationId: boundVaccinationId,
    });
  }

  function fingerprintOf(layout: CertificateLayout, snapshot: unknown): string {
    return mockHash(JSON.stringify({ layout, snapshot }));
  }

  function contentSignature(request: CertificateIssueRequest): string {
    return JSON.stringify([
      request.kind,
      request.layout,
      { ...emptyCertificateInputs(), ...request.inputs },
      request.preview_display_issue_date,
      request.preview_fingerprint,
    ]);
  }

  function issueArtifact(editionId: number, layout: CertificateLayout): number {
    for (const [id, a] of artifacts) if (a.editionId === editionId && a.layout === layout) return id;
    const id = nextArtifactId;
    nextArtifactId += 1;
    artifacts.set(id, { editionId, layout });
    return id;
  }

  const api: CertificateApi = {
    async getContext(vaccinationId) {
      return track("getContext", vaccinationId, () => {
        const failure = takeFailure("getContext");
        if (failure) throw failure.error;
        assertScope(vaccinationId);
        const context: CertificateContext = {
          vaccination_id: vaccinationId,
          source_version: sourceVersion,
          eligibility: {
            state: eligibilityState,
            allowed_kinds: allowedKinds,
            provenance: options.rabiesEligible === null ? "legacy_unresolved" : "master_snapshot",
          },
          can_issue: canIssue,
          latest: {
            normal: latestOf("normal") ? withValidity(latestOf("normal")!) : null,
            rabies: latestOf("rabies") ? withValidity(latestOf("rabies")!) : null,
          },
          initial_inputs: emptyCertificateInputs(), // 新規料金は 4 欄とも null
        };
        return context;
      });
    },
    async previewForIssue(vaccinationId, body) {
      return track("previewForIssue", body, () => {
        const failure = takeFailure("previewForIssue");
        if (failure) throw failure.error;
        assertScope(vaccinationId);
        if (!canIssue) throw new CertificateApiError("forbidden", { httpStatus: 403, code: "not-authorized" });
        const inputs = assertInputsValid(body);
        assertCas(body.preconditions, body.kind);
        const snapshot = snapshotFor(body, inputs);
        return {
          snapshot,
          snapshot_sha256: mockHash(JSON.stringify(snapshot)),
          preview_fingerprint: fingerprintOf(body.layout, snapshot),
          source_version: sourceVersion,
          latest_edition_id: latestOf(body.kind)?.id ?? null,
          pdf_base64: mockPdfBase64(),
          media_type: "application/pdf" as const,
          template_version: "mock-template-v1",
          renderer_version: "mock-renderer-v1",
        };
      });
    },
    async issue(vaccinationId, body) {
      return track("issue", body, () => {
        const failure = takeFailure("issue");
        assertScope(vaccinationId);
        if (failure?.claimOnly) recoveryLog.claim(body.operation_key, "issue", null);
        if (failure && !failure.commitFirst) throw failure.error;
        if (!canIssue) throw new CertificateApiError("forbidden", { httpStatus: 403, code: "edit-right-lost", c06Status: "edit-right-lost" });
        const signature = contentSignature(body);
        const existing = operations.get(body.operation_key);
        if (existing) {
          if (existing.intent !== "issue" || existing.contentSignature !== signature) {
            throw new CertificateApiError("operation-key-conflict", {
              httpStatus: 409,
              code: "operation-key-conflict",
              c06Status: "operation-key-conflict",
              operationKey: body.operation_key,
            });
          }
          // 同 key 同内容: 歴史的成功 receipt を replayed で返す（現在 CAS 不一致でも未実行扱いに戻さない）。
          const replay = { ...existing.result, status: "replayed" } as CertificateResult;
          if (failure) throw failure.error;
          return replay;
        }
        const inputs = assertInputsValid(body);
        if (body.preview_display_issue_date !== today) {
          throw new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed", c06Status: "dependency-changed" });
        }
        assertCas(body.preconditions, body.kind);
        const snapshot = snapshotFor(body, inputs);
        if (fingerprintOf(body.layout, snapshot) !== body.preview_fingerprint) {
          throw new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed", c06Status: "dependency-changed" });
        }
        const previous = latestOf(body.kind);
        const id = nextEditionId;
        nextEditionId += 1;
        const edition: CertificateEdition = {
          id,
          series_id: body.kind === "normal" ? 201 : 202,
          vaccination_id: boundVaccinationId,
          kind: body.kind,
          edition: (previous?.edition ?? 0) + 1,
          source_version: sourceVersion,
          previous_edition_id: previous?.id ?? null,
          issued_at: "2026-10-20T14:30:00Z",
          issued_by: 9,
          validity: "current",
          snapshot,
          snapshot_sha256: mockHash(JSON.stringify(snapshot)),
          available_layouts: body.kind === "normal" ? ["portrait", "landscape"] : ["rabies"],
        };
        editions.push(edition);
        const artifactId = issueArtifact(id, body.layout);
        const result: CertificateResult = {
          status: "succeeded",
          operation_key: body.operation_key,
          result_refs: {
            vaccination_id: String(boundVaccinationId),
            series_id: String(edition.series_id),
            edition_id: String(id),
            artifact_id: String(artifactId),
          },
          committed_versions: { edition: String(edition.edition), source_version: String(sourceVersion) },
          recorded_at: "2026-10-20T14:30:00Z",
        };
        operations.set(body.operation_key, { intent: "issue", contentSignature: signature, result });
        recoveryLog.claim(body.operation_key, "issue", null);
        recoveryLog.succeed(body.operation_key, result);
        if (failure) throw failure.error; // commitFirst: 確定済みだが応答は失われた
        return result;
      });
    },
    async getEdition(editionId) {
      return track("getEdition", editionId, () => {
        const failure = takeFailure("getEdition");
        if (failure) throw failure.error;
        return withValidity(findEdition(editionId));
      });
    },
    async listEditions(vaccinationId, query) {
      return track("listEditions", query, () => {
        const failure = takeFailure("listEditions");
        if (failure) throw failure.error;
        assertScope(vaccinationId);
        const all = editions.filter((e) => e.kind === query.kind).map(withValidity).reverse();
        const limit = query.limit ?? 20;
        const page = query.page ?? 1;
        return { data: all.slice((page - 1) * limit, page * limit), total: all.length, page, limit };
      });
    },
    async previewSaved(editionId, layout) {
      return track("previewSaved", { editionId, layout }, () => {
        const failure = takeFailure("previewSaved");
        if (failure) throw failure.error;
        const edition = findEdition(editionId);
        if (validityOf(edition) !== "current") {
          throw new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed", c06Status: "dependency-changed" });
        }
        if (!edition.available_layouts.includes(layout)) {
          throw new CertificateApiError("validation", { httpStatus: 422, code: "validation", c06Status: "validation" });
        }
        return {
          edition_id: edition.id,
          layout,
          pdf_base64: mockPdfBase64(),
          media_type: "application/pdf" as const,
          template_version: "mock-template-v1",
          renderer_version: "mock-renderer-v1",
          content_sha256: mockHash(`${edition.id}:${layout}`),
        };
      });
    },
    async reprint(editionId, body: CertificateReprintRequest) {
      return track("reprint", { editionId, body }, () => {
        const failure = takeFailure("reprint");
        if (failure?.claimOnly) recoveryLog.claim(body.operation_key, "reprint", Number(editionId));
        if (failure && !failure.commitFirst) throw failure.error;
        const edition = findEdition(editionId);
        const signature = JSON.stringify([edition.id, body.layout]);
        const existing = operations.get(body.operation_key);
        if (existing) {
          if (existing.intent !== "reprint" || existing.contentSignature !== signature) {
            throw new CertificateApiError("operation-key-conflict", { httpStatus: 409, code: "operation-key-conflict", c06Status: "operation-key-conflict" });
          }
          if (failure) throw failure.error;
          return { ...existing.result, status: "replayed" } as CertificateResult;
        }
        if (!edition.available_layouts.includes(body.layout)) {
          throw new CertificateApiError("validation", { httpStatus: 422, code: "validation", c06Status: "validation" });
        }
        if (validityOf(edition) !== "current") {
          throw new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed", c06Status: "dependency-changed" });
        }
        if (body.preconditions.source_version !== sourceVersion || body.preconditions.latest_edition_id !== latestOf(edition.kind)?.id) {
          throw new CertificateApiError("version-conflict", { httpStatus: 409, code: "version-conflict", c06Status: "version-conflict" });
        }
        const artifactId = issueArtifact(edition.id, body.layout);
        const result: CertificateResult = {
          status: "succeeded",
          operation_key: body.operation_key,
          result_refs: {
            vaccination_id: String(edition.vaccination_id),
            series_id: String(edition.series_id),
            edition_id: String(edition.id),
            artifact_id: String(artifactId),
          },
          committed_versions: { edition: String(edition.edition), source_version: String(sourceVersion) },
          recorded_at: "2026-10-20T15:00:00Z",
        };
        operations.set(body.operation_key, { intent: "reprint", contentSignature: signature, result });
        recoveryLog.claim(body.operation_key, "reprint", edition.id);
        recoveryLog.succeed(body.operation_key, result);
        if (failure) throw failure.error;
        return result;
      });
    },
    async getArtifactContent(artifactId) {
      return track("getArtifactContent", artifactId, () => {
        const failure = takeFailure("getArtifactContent");
        if (failure) throw failure.error;
        const artifact = artifacts.get(Number(artifactId));
        if (!artifact) throw new CertificateApiError("not-found", { httpStatus: 404, code: "not_found" });
        // 旧版は有効な出力にしない（現在 current でなければ 409）。
        if (validityOf(findEdition(artifact.editionId)) !== "current") {
          throw new CertificateApiError("dependency-changed", { httpStatus: 409, code: "dependency-changed", c06Status: "dependency-changed" });
        }
        return mockPdfBlob();
      });
    },
    async queryOperationRecovery(request) {
      return track("queryOperationRecovery", request, () => {
        const failure = takeFailure("queryOperationRecovery");
        if (failure) throw failure.error;
        // 閉じた契約に反する要求は 400（reprint+null / issue+非 null を含む）。
        if (!certificateRecoveryQueryRequestSchema.safeParse(request).success) {
          throw new CertificateApiError("validation", { httpStatus: 400, code: "VALIDATION" });
        }
        // 別の接種・存在しない版は存在を開示しない 404 SCOPE_NOT_FOUND。
        const unknownEdition =
          request.intent === "reprint" && !editions.some((e) => e.id === request.target_ref.edition_id);
        if (request.target_ref.vaccination_id !== boundVaccinationId || unknownEdition) {
          throw new CertificateApiError("recovery-scope-not-found", { httpStatus: 404, code: "SCOPE_NOT_FOUND" });
        }
        return recoveryLog.query(request);
      });
    },
    async lookupOperation(operationKey, intent: CertificateOperationIntent) {
      return track("lookupOperation", { operationKey, intent }, () => {
        const failure = takeFailure("lookupOperation");
        if (failure) throw failure.error;
        // 照会の scope は (operation_key, intent)。intent が違えば別 scope なので未 claim 扱い。
        const stored = operations.get(operationKey);
        const existing = stored?.intent === intent ? stored : undefined;
        const claimed = recoveryLog.get(operationKey);
        if (!existing && claimed?.intent === intent) {
          return { status: "not-succeeded", operation_key: operationKey } as CertificateResult;
        }
        if (!existing) {
          throw new CertificateApiError("scope-not-found", { httpStatus: 404, code: "scope-not-found", c06Status: "scope-not-found" });
        }
        return existing.result;
      });
    },
  };

  const controls: MockCertificateControls = {
    failNext(operation, failure) {
      failures.set(operation, failure);
    },
    bumpSourceVersion() {
      sourceVersion += 1;
    },
    setToday(date) {
      today = date;
    },
    setCanIssue(value) {
      canIssue = value;
    },
    recovery: recoveryLog,
    calls,
  };
  return { api, controls };
}
