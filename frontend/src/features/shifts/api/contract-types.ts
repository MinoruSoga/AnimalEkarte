/**
 * EMR-514: 勤務日時CRUD 固定API契約の feature ローカル型。
 *
 * 正本: `backend/docs/api.yaml`（PR740 採用SHA 30538608da0850be2d6a41e88ecec0ca86a7a328）。
 * 共有生成 `@/types/generated/models.ts` は旧契約のままのため編集せず、
 * 新契約（ShiftEntry / ShiftWorkBreak / ShiftWorkBreakView / CreateShiftRequest /
 * UpdateShiftRequest / ShiftMutationReceipt / ShiftOperationResult / ShiftContractError）
 * をここに宣言する。契約schemaとの対応は shifts/api 契約対応テストで確認する。
 *
 * wire上の注意（api.yaml description より）:
 * - id / staff_id / clinic_id は先頭ゼロなし正十進「文字列」。JS Numberへ変換しない
 * - version は aggregate 版（休憩変更を含む）。readOnly
 * - date は勤務開始側の JST 日付（YYYY-MM-DD）
 * - starts_at/ends_at は明示 date-time（JSTオフセット付き）。NULL/未指定を同一視しない
 */
import type { ShiftType } from "../types";

/** 勤務区間/休憩の解決状態。unknown は旧日時が未解決（NULL を休日・勤務可と解釈しない） */
export type ShiftIntervalState = "known" | "unknown";

/** 保存要求用の休憩（明示日時）。starts_at < ends_at、勤務区間内 */
export interface ShiftWorkBreak {
  starts_at: string;
  ends_at: string;
}

/** 応答専用の休憩。unknown でも判明している片側/両側の日時は保持する */
export type ShiftWorkBreakView =
  | {
      interval_state: "known";
      starts_at: string;
      ends_at: string;
    }
  | {
      interval_state: "unknown";
      starts_at: string | null;
      ends_at: string | null;
    };

/** GET /shifts・GET /shifts/{id} の応答DTO（選択医院の勤務） */
export interface ShiftEntry {
  id: string;
  clinic_id: string;
  staff_id: string;
  staff_name: string;
  date: string;
  shift_type: ShiftType;
  starts_at: string | null;
  ends_at: string | null;
  /** 勤務aggregate版。休憩変更を含めて進む（readOnly） */
  version: number;
  interval_state: ShiftIntervalState;
  notes: string;
  breaks: ShiftWorkBreakView[];
  created_at: string;
  updated_at: string;
}

/** POST /shifts。全入力を明示する（NULL/未指定を同一視しない）。off/paid_leave は両日時 NULL・breaks [] */
export interface CreateShiftRequest {
  operation_key: string;
  staff_id: string;
  date: string;
  shift_type: ShiftType;
  starts_at: string | null;
  ends_at: string | null;
  breaks: ShiftWorkBreak[];
  notes: string;
}

/** PATCH /shifts/{id}。未指定は維持。starts_at/ends_at の明示NULLのみ許可（終日非勤務化）。breaks は配列全置換 */
export interface UpdateShiftRequest {
  operation_key: string;
  expected_version: number;
  shift_type?: ShiftType;
  starts_at?: string | null;
  ends_at?: string | null;
  breaks?: ShiftWorkBreak[];
  notes?: string;
}

export type ShiftMutationIntent = "create" | "update" | "delete";

/** 保存と同じ transaction の確定結果。delete は shift=NULL・resulting_version=NULL */
export interface ShiftMutationReceipt {
  operation_key: string;
  intent: ShiftMutationIntent;
  shift_id: string;
  resulting_version: number | null;
  shift: ShiftEntry | null;
}

/** 結果照会。unknown は未実行・失敗・不存在の証明ではない（新key再送禁止） */
export type ShiftOperationResult =
  | { operation_key: string; state: "succeeded"; receipt: ShiftMutationReceipt }
  | { operation_key: string; state: "unknown"; receipt: null };

/** protected_bookings 409 に同梱される認可済み対象予約 */
export interface ShiftProtectedBooking {
  appointment_id: string;
  starts_at: string;
  ends_at: string;
  patient_label: string;
  reservation_type_label: string;
}

/** 409 競合/未解決。code で処理し error 文字列を分岐に使わない */
export const SHIFT_CONTRACT_ERROR_CODES = [
  "version_conflict",
  "operation_key_reused",
  "shift_exists",
  "cross_clinic_overlap",
  "protected_bookings",
  "source_unresolved",
] as const;

export type ShiftContractErrorCode = (typeof SHIFT_CONTRACT_ERROR_CODES)[number];

export interface ShiftContractError {
  error: string;
  code: ShiftContractErrorCode;
  /** code=cross_clinic_overlap で返る相手医院名と重複時間（他院勤務全体は開示されない） */
  conflicting_clinic_name?: string;
  overlap_starts_at?: string;
  overlap_ends_at?: string;
  /** code=protected_bookings で返る認可済み対象予約のみ */
  protected_bookings?: ShiftProtectedBooking[];
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

/** 応答payloadが契約エラーshapeかを検証する（無検査の as を避けるための型ガード） */
export function isShiftContractErrorPayload(value: unknown): value is ShiftContractError {
  if (!isRecord(value)) return false;
  if (typeof value.error !== "string" || typeof value.code !== "string") {
    return false;
  }
  return (SHIFT_CONTRACT_ERROR_CODES as readonly string[]).includes(value.code);
}

/** axios等の例外から契約エラーpayloadを取り出す。契約外の形は null */
export function extractShiftContractError(err: unknown): ShiftContractError | null {
  if (!isRecord(err)) return null;
  const response = isRecord(err.response) ? err.response : null;
  const data = response !== null ? response.data : err.data;
  return isShiftContractErrorPayload(data) ? data : null;
}
