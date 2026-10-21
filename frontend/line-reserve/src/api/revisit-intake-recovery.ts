// EMR-474 operation-recovery Slice A: 未確定操作の発見 query の通信 adapter。
//
// 【未接続・runtime 未実装】対象 route は BE 未実装（api.yaml x-runtime-status: not-implemented）。
// 利用開始は LIFF_REVISIT_INTAKE_ACTIVE=false の背後に置かれ、mock へのフォールバックもしない。
// 契約: docs/work/emr/emr-474-487-operation-recovery-contract-20261021.md §4。
//
// - 純 read の POST。idToken は既存 LIFF と同じ Authorization: Bearer。clinic は URL path の trusted 値。
// - response は no-store 必須。closed schema を満たさないものは契約違反として fail-closed に投げる。
// - 失敗は全て RevisitIntakeRecoveryQueryError（kind で分類）。どの kind も「安全に次操作できる」根拠にならない。
import axios, {
  AxiosHeaders,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { API_BASE_URL } from "../lib/liff-config";
import { sanitizeNullBytes } from "@/shared-liff/sanitize";
import { devError } from "@/shared-liff/dev-log";
import {
  isPageConsistentWithScope,
  isPageWithinTarget,
  recoveryErrorBodySchema,
  recoveryPageSchema,
  recoveryQueryRequestSchema,
  type RevisitIntakeRecoveryPage,
  type RevisitIntakeRecoveryQueryRequest,
} from "./revisit-intake-recovery-schemas";

/** この adapter が叩く route の runtime 状態。BE 実装・配信確認までは "not-implemented"。 */
export const REVISIT_INTAKE_RECOVERY_RUNTIME_STATUS = "not-implemented" as const;

export type RevisitIntakeRecoveryErrorKind =
  | "validation" // 400 VALIDATION（client 側の request 不備。再試行しても同じ）
  | "unauthenticated" // 401
  | "not_authorized" // 403 NOT_AUTHORIZED
  | "scope_not_found" // 404 SCOPE_NOT_FOUND（存在非開示）
  | "recovery_conflict" // 409 RECOVERY_*
  | "result_unavailable" // 503 RESULT_UNAVAILABLE
  | "unclassified" // 上記以外の HTTP 応答（status と code の組が契約外を含む）
  | "contract_violation" // 200 だが closed schema・scope 整合・paging 規則に反する
  | "network" // 応答なし
  | "aborted";

export class RevisitIntakeRecoveryQueryError extends Error {
  readonly kind: RevisitIntakeRecoveryErrorKind;
  constructor(kind: RevisitIntakeRecoveryErrorKind, message: string) {
    super(message);
    this.name = "RevisitIntakeRecoveryQueryError";
    this.kind = kind;
  }
}

/**
 * 1 request の上限。無いと止まった接続で発見が永久に pending になり、保留 UI も再確認も出ない。
 * 超過は応答なしの axios error（ECONNABORTED）となり "network" に分類される。
 */
const RECOVERY_REQUEST_TIMEOUT_MS = 20_000;

const httpClient = axios.create({ baseURL: API_BASE_URL });

// liff-api.ts と同じ NULL バイトサニタイズ（本 adapter は liff-api.ts を編集しないため独自 instance に付ける）。
httpClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (config.method?.toLowerCase() === "post" && config.data) {
    config.data = sanitizeNullBytes(config.data);
  }
  return config;
});

function recoveryPath(clinicId: string): string {
  return `/api/liff/${encodeURIComponent(clinicId)}/revisit-intake/operation-recovery/query`;
}

/** HTTP status と body の code の組が契約どおりのときだけ分類する。それ以外は unclassified。 */
export function classifyRecoveryHttpError(status: number, body: unknown): RevisitIntakeRecoveryErrorKind {
  if (status === 401) return "unauthenticated"; // 認証 middleware の既存応答（body は検査しない）
  const parsed = recoveryErrorBodySchema.safeParse(body);
  if (!parsed.success) return "unclassified";
  const code = parsed.data.code;
  if (status === 400 && code === "VALIDATION") return "validation";
  if (status === 403 && code === "NOT_AUTHORIZED") return "not_authorized";
  if (status === 404 && code === "SCOPE_NOT_FOUND") return "scope_not_found";
  if (status === 409 && /^RECOVERY_[A-Z_]+$/.test(code)) return "recovery_conflict";
  if (status === 503 && code === "RESULT_UNAVAILABLE") return "result_unavailable";
  return "unclassified";
}

function hasNoStore(headers: unknown): boolean {
  // AxiosHeaders.get は大文字小文字を区別しない。
  const value = AxiosHeaders.from(headers as ConstructorParameters<typeof AxiosHeaders>[0]).get(
    "cache-control",
  );
  return typeof value === "string" && /(^|[\s,])no-store($|[\s,;])/i.test(value);
}

interface QueryRevisitIntakeRecoveryArgs {
  clinicId: string;
  idToken: string;
  request: RevisitIntakeRecoveryQueryRequest;
  signal?: AbortSignal;
  /** テスト用。既定は RECOVERY_REQUEST_TIMEOUT_MS。 */
  timeoutMs?: number;
}

/**
 * 1 page 分の発見 query。items の順序・cursor の連鎖・件数上限の検査は
 * lib/revisit-intake-recovery-discovery.ts が行う（page 単体では判定できないため）。
 */
export async function queryRevisitIntakeOperationRecoveryPage(
  args: QueryRevisitIntakeRecoveryArgs,
): Promise<RevisitIntakeRecoveryPage> {
  const request = recoveryQueryRequestSchema.safeParse(args.request);
  if (!request.success) {
    // 自分が組み立てた request の不備。送らずに止める。
    throw new RevisitIntakeRecoveryQueryError("validation", "recovery query request is invalid");
  }
  let res: AxiosResponse<unknown>;
  try {
    res = await httpClient.post<unknown>(recoveryPath(args.clinicId), request.data, {
      headers: { Authorization: `Bearer ${args.idToken}`, "Content-Type": "application/json" },
      signal: args.signal,
      timeout: args.timeoutMs ?? RECOVERY_REQUEST_TIMEOUT_MS,
    });
  } catch (err: unknown) {
    if (axios.isCancel(err)) throw new RevisitIntakeRecoveryQueryError("aborted", "aborted");
    if (axios.isAxiosError(err) && err.response) {
      throw new RevisitIntakeRecoveryQueryError(
        classifyRecoveryHttpError(err.response.status, err.response.data),
        `recovery query failed (${err.response.status})`,
      );
    }
    throw new RevisitIntakeRecoveryQueryError("network", "recovery query got no response");
  }
  if (res.status !== 200 || !hasNoStore(res.headers)) {
    throw new RevisitIntakeRecoveryQueryError("contract_violation", "unexpected status or cache policy");
  }
  const page = recoveryPageSchema.safeParse(res.data);
  const req = args.request;
  if (
    !page.success ||
    !isPageConsistentWithScope(req.scope, page.data) ||
    (req.scope === "target" && !isPageWithinTarget(req.target_ref, page.data))
  ) {
    devError("[revisit-intake-recovery] invalid response shape:", page.success ? "scope" : page.error);
    throw new RevisitIntakeRecoveryQueryError("contract_violation", "recovery response is invalid");
  }
  return page.data;
}
