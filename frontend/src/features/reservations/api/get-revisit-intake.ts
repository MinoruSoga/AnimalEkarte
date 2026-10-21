import { AxiosHeaders, isAxiosError } from "axios";
import { useInfiniteQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

/**
 * EMR-475: GET /v1/reservations/{appointmentId}/revisit-intake（staff 読取専用）。
 * 契約正本: backend/docs/api.yaml `RevisitIntakeStaffReservationRead`（EMR-471 selected v3）。
 * `types/generated` に未生成のため feature 内で最小の wire 型を持つ。
 * ID / revision / generation は BIGINT 文字列のまま扱い、Number へ変換しない。
 */
export type RevisitIntakeEnergyAppetiteCode = string | null;

export interface RevisitIntakeFullAnswers {
  visit_purpose: string;
  onset: string | null;
  energy: RevisitIntakeEnergyAppetiteCode;
  appetite: RevisitIntakeEnergyAppetiteCode;
  current_symptoms: string | null;
  improvement: string | null;
  examination_preference: string | null;
  main_concern: string | null;
}

export interface RevisitIntakeAnswer {
  answer_id: string;
  revision: string;
  form_definition_version: number;
  received_at: string;
  provenance: "owner_submission";
  answers: RevisitIntakeFullAnswers;
}

export interface RevisitIntakeTargetRef {
  target_id: string | null;
  appointment_id: string;
  pet_id: string;
  target_generation: string;
}

export interface RevisitIntakeCurrentTarget {
  ref: RevisitIntakeTargetRef;
  head_revision: string | null;
  answer_state: "unanswered" | "answered";
  latest_answer: RevisitIntakeAnswer | null;
  content_available: boolean;
}

export interface RevisitIntakeHistoryEntry {
  target: RevisitIntakeTargetRef;
  answer: RevisitIntakeAnswer;
  target_relation: "current" | "previous";
}

export interface RevisitIntakeStaffReservationRead {
  availability: "available" | "target_unresolved";
  form_definition_version: number;
  current_target: RevisitIntakeCurrentTarget | null;
  history: RevisitIntakeHistoryEntry[];
  next_cursor: string | null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isStr(value: unknown): value is string {
  return typeof value === "string";
}

function isNullableStr(value: unknown): value is string | null {
  return value === null || typeof value === "string";
}

function isNullableText(value: unknown): boolean {
  return isNullableStr(value);
}

function isTargetRef(value: unknown, allowNullTargetId: boolean): value is RevisitIntakeTargetRef {
  return (
    isRecord(value) &&
    (allowNullTargetId ? isNullableStr(value.target_id) : isStr(value.target_id)) &&
    isStr(value.appointment_id) &&
    isStr(value.pet_id) &&
    isStr(value.target_generation)
  );
}

function isAnswer(value: unknown): value is RevisitIntakeAnswer {
  if (!isRecord(value) || !isRecord(value.answers)) return false;
  const a = value.answers;
  return (
    isStr(value.answer_id) &&
    isStr(value.revision) &&
    typeof value.form_definition_version === "number" &&
    isStr(value.received_at) &&
    value.provenance === "owner_submission" &&
    isStr(a.visit_purpose) &&
    isNullableText(a.onset) &&
    isNullableStr(a.energy) &&
    isNullableStr(a.appetite) &&
    isNullableText(a.current_symptoms) &&
    isNullableStr(a.improvement) &&
    isNullableStr(a.examination_preference) &&
    isNullableText(a.main_concern)
  );
}

function isCurrentTarget(value: unknown): value is RevisitIntakeCurrentTarget {
  if (!isRecord(value) || !isTargetRef(value.ref, true)) return false;
  if (typeof value.content_available !== "boolean" || !isNullableStr(value.head_revision)) {
    return false;
  }
  const latest = value.latest_answer;
  if (latest !== null && !isAnswer(latest)) return false;
  if (value.answer_state === "unanswered") {
    return value.head_revision === null && latest === null && value.content_available === false;
  }
  if (value.answer_state !== "answered" || value.head_revision === null) return false;
  // 可視 head（本文あり）か不可視 head（本文なし）のどちらかのみ。
  return value.content_available ? latest !== null : latest === null;
}

function isHistoryEntry(value: unknown): value is RevisitIntakeHistoryEntry {
  return (
    isRecord(value) &&
    isTargetRef(value.target, false) &&
    isAnswer(value.answer) &&
    (value.target_relation === "current" || value.target_relation === "previous")
  );
}

/**
 * target_relation=current の履歴は current_target と同一対象（予約・ペット・世代・target）でなければならない。
 * 別ペット／別世代が current と偽装された応答を「訂正履歴」として出さない（fail-closed）。
 */
function isCurrentRelationConsistent(
  entry: RevisitIntakeHistoryEntry,
  currentTarget: RevisitIntakeCurrentTarget | null,
): boolean {
  if (entry.target_relation !== "current") return true;
  if (currentTarget === null) return false;
  const expected = currentTarget.ref;
  return (
    entry.target.appointment_id === expected.appointment_id &&
    entry.target.pet_id === expected.pet_id &&
    entry.target.target_generation === expected.target_generation &&
    (expected.target_id === null || entry.target.target_id === expected.target_id)
  );
}

function isStaffRead(
  value: unknown,
  expectedAppointmentId: string,
): value is RevisitIntakeStaffReservationRead {
  if (!isRecord(value)) return false;
  if (!(value.current_target === null || isCurrentTarget(value.current_target))) return false;
  const currentTarget = value.current_target as RevisitIntakeCurrentTarget | null;
  // 要求した予約以外の原文は、形が正しくても表示しない（中間層の取り違え対策・fail-closed）。
  if (currentTarget !== null && currentTarget.ref.appointment_id !== expectedAppointmentId) {
    return false;
  }
  return (
    (value.availability === "available" || value.availability === "target_unresolved") &&
    Array.isArray(value.history) &&
    value.history.every(isHistoryEntry) &&
    (value.history as RevisitIntakeHistoryEntry[]).every(
      (h) =>
        h.target.appointment_id === expectedAppointmentId &&
        isCurrentRelationConsistent(h, currentTarget),
    ) &&
    // 空文字 cursor は「次ページあり」と誤解され 1 ページ目を重複取得するため契約外として拒否する。
    (value.next_cursor === null ||
      (typeof value.next_cursor === "string" && value.next_cursor !== ""))
  );
}

export interface GetRevisitIntakeParams {
  appointmentId: string;
  clinicId?: string;
  /** 履歴ページサイズ（BE 既定 50 / 最大 100）。省略時は BE 既定。 */
  limit?: number;
}

export async function getRevisitIntake(
  { appointmentId, clinicId, limit }: GetRevisitIntakeParams,
  cursor: string | null,
  signal?: AbortSignal,
): Promise<RevisitIntakeStaffReservationRead> {
  const params: Record<string, string | number> = {};
  if (cursor) params.cursor = cursor;
  if (limit !== undefined) params.limit = limit;
  const { data } = await axios.get<unknown>(
    `/v1/reservations/${encodeURIComponent(appointmentId)}/revisit-intake`,
    {
      params,
      signal,
      headers: clinicId ? { "X-Clinic-ID": clinicId } : undefined,
    },
  );
  // 契約外の応答は原文として扱わない（fail-closed）。
  if (!isStaffRead(data, appointmentId)) throw new Error("再診問診の応答形式が不正です");
  return data;
}

/**
 * query key は `lib/query-keys.ts` に未登録のため feature 内で定義する（共通化は要求事項として報告）。
 * 予約 ID と clinicId を必ず含め、別予約・別医院の原文を共有キャッシュへ載せない。
 */
export function revisitIntakeQueryKey(appointmentId: string, clinicId?: string) {
  return ["reservations", "revisit-intake", appointmentId, clinicId ?? ""] as const;
}

/** 失敗の HTTP status（axios 以外・通信断は undefined）。 */
export function getRevisitIntakeErrorStatus(error: unknown): number | undefined {
  return isAxiosError(error) ? error.response?.status : undefined;
}

/** 自動再試行で待つ Retry-After の上限。超える場合は自動再試行せず、手動の再試行に委ねる。 */
const MAX_AUTO_RETRY_AFTER_MS = 5_000;

/**
 * 429/503 の Retry-After（秒または HTTP-date）をミリ秒へ。無い・解釈不能は undefined。
 * 契約（api.yaml 429/503）は Retry-After があれば尊重することを求める。
 */
export function getRetryAfterMs(error: unknown, now: number = Date.now()): number | undefined {
  if (!isAxiosError(error)) return undefined;
  const status = error.response?.status;
  if (status !== 429 && status !== 503) return undefined;
  // ヘッダ名の大文字小文字に依存しない（AxiosHeaders は元の表記で保持する）。
  const raw: unknown = error.response?.headers
    ? AxiosHeaders.from(error.response.headers as AxiosHeaders).get("retry-after")
    : undefined;
  if (typeof raw !== "string" || raw.trim() === "") return undefined;
  const value = raw.trim();
  if (/^\d+$/.test(value)) return Number(value) * 1000;
  const at = Date.parse(value);
  return Number.isNaN(at) ? undefined : Math.max(0, at - now);
}

/**
 * 原文は認証 scope の機微情報（Cache-Control: no-store）。
 * staleTime 0 / gcTime 0 で、unmount（モーダル閉鎖・予約切替）時にキャッシュを破棄し再利用しない。
 */
export function useGetRevisitIntake(params: GetRevisitIntakeParams & { enabled?: boolean }) {
  const { appointmentId, clinicId, limit, enabled = true } = params;
  return useInfiniteQuery({
    queryKey: revisitIntakeQueryKey(appointmentId, clinicId),
    queryFn: ({ pageParam, signal }) =>
      getRevisitIntake({ appointmentId, clinicId, limit }, pageParam, signal),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    enabled: enabled && appointmentId !== "",
    staleTime: 0,
    gcTime: 0,
    refetchOnMount: "always",
    // 403/404 は再試行しても結果が変わらず、権限の存在確認連打にもなるため即確定。
    retry: (failureCount, error) => {
      const status = getRevisitIntakeErrorStatus(error);
      if (status === 400 || status === 401 || status === 403 || status === 404) return false;
      if (failureCount >= 1) return false;
      // Retry-After が長い場合は早すぎる自動再試行で過負荷・レート制限を延ばさない。
      const retryAfter = getRetryAfterMs(error);
      return retryAfter === undefined || retryAfter <= MAX_AUTO_RETRY_AFTER_MS;
    },
    retryDelay: (failureCount, error) =>
      getRetryAfterMs(error) ?? Math.min(1000 * 2 ** failureCount, 30_000),
    // セクション内で表示するためグローバル toast は抑止する。
    meta: { silentError: true },
  });
}
