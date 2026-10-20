/**
 * EMR-487: 証明書 API 失敗の分類。機械判定は HTTP status と body の code / status のみ
 * （本文 `error` はローカライズ可のため判定に使わない）。
 */
export type CertificateErrorKind =
  | "unauthenticated" // 401
  | "forbidden" // 403 not-authorized / edit-right-lost（保存元対象不明も 403）
  | "not-found" // 404 not_found
  | "scope-not-found" // 404 lookup で binding 不在 = 未 claim
  | "validation" // 400 / 422
  | "version-conflict" // 409
  | "dependency-changed" // 409
  | "operation-key-conflict" // 409
  | "operation-in-progress" // 409
  | "result-unavailable" // 503 sticky unknown
  | "render-unavailable" // 503 render_failed / artifact_unavailable
  | "network" // 応答なし・タイムアウト
  | "server" // 500 ほか
  | "malformed"; // 応答形式が契約外

export class CertificateApiError extends Error {
  readonly kind: CertificateErrorKind;
  readonly httpStatus: number | undefined;
  readonly code: string | undefined;
  readonly c06Status: string | undefined;
  readonly operationKey: string | undefined;

  constructor(
    kind: CertificateErrorKind,
    init: {
      message?: string;
      httpStatus?: number;
      code?: string;
      c06Status?: string;
      operationKey?: string;
    } = {},
  ) {
    super(init.message ?? kind);
    this.name = "CertificateApiError";
    this.kind = kind;
    this.httpStatus = init.httpStatus;
    this.code = init.code;
    this.c06Status = init.c06Status;
    this.operationKey = init.operationKey;
  }
}

export interface ErrorBody {
  error?: string;
  code?: string;
  status?: string;
  operation_key?: string;
}

/** HTTP status と flat error body から分類する。未知の組合せは fail-closed で server。 */
export function classifyCertificateError(
  httpStatus: number | undefined,
  body: ErrorBody | null,
): CertificateApiError {
  const code = body?.code;
  const c06Status = body?.status;
  const init = {
    message: body?.error,
    httpStatus,
    code,
    c06Status,
    operationKey: body?.operation_key,
  };
  const key = c06Status ?? code;
  if (httpStatus === undefined) return new CertificateApiError("network", init);
  if (httpStatus === 401) return new CertificateApiError("unauthenticated", init);
  if (httpStatus === 403) return new CertificateApiError("forbidden", init);
  if (httpStatus === 400 || httpStatus === 422) return new CertificateApiError("validation", init);
  if (httpStatus === 404) {
    return new CertificateApiError(key === "scope-not-found" ? "scope-not-found" : "not-found", init);
  }
  if (httpStatus === 409) {
    if (key === "version-conflict") return new CertificateApiError("version-conflict", init);
    if (key === "dependency-changed") return new CertificateApiError("dependency-changed", init);
    if (key === "operation-key-conflict") {
      return new CertificateApiError("operation-key-conflict", init);
    }
    if (key === "operation-in-progress") {
      return new CertificateApiError("operation-in-progress", init);
    }
    return new CertificateApiError("server", init);
  }
  if (httpStatus === 503) {
    if (code === "render_failed" || code === "artifact_unavailable") {
      return new CertificateApiError("render-unavailable", init);
    }
    return new CertificateApiError("result-unavailable", init);
  }
  return new CertificateApiError("server", init);
}
