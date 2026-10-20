/**
 * Shifts feature types
 * 契約型: {@link ShiftEntry} は backend/docs/api.yaml（PR740採用SHA）の新契約。
 * 共有生成 models.ts は ShiftType（enum値は不変）のみ利用する。
 */
// FE6-3: tygo enum_style: "union"（FE6-1/FE6-2）により生成定数が真の literal union になったため、
// 手書き union を生成型からの re-export へ移行した。drift テストは union-drift.test.ts から削除済み。
import type { ShiftType } from "@/types/generated/models";
export type { ShiftType };

/** UI-facing shift (post-transform — EMR-514 新契約) */
export type { Shift } from "../api/transforms";

/** フォームの休憩入力行（HH:mm 時刻。送信時に JST date-time へ組み立てる） */
export interface ShiftBreakInput {
  break_start: string;
  break_end: string;
}

export { SHIFT_CONTRACT_ERROR_CODES } from "../api/contract-types";
export type {
  CreateShiftRequest,
  ShiftContractError,
  ShiftContractErrorCode,
  ShiftIntervalState,
  ShiftMutationIntent,
  ShiftMutationReceipt,
  ShiftOperationResult,
  ShiftProtectedBooking,
  ShiftWorkBreak,
  ShiftWorkBreakView,
  UpdateShiftRequest,
} from "../api/contract-types";

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  full: "全日",
  morning: "午前",
  afternoon: "午後",
  off: "休日",
  paid_leave: "有休",
};

const SHIFT_TYPE_VALUES: readonly string[] = Object.keys(SHIFT_TYPE_LABELS);

/** FE-RC-037: 外部データ（FormData/API応答）を無検証で `as ShiftType` しないための型ガード */
export function isShiftType(value: unknown): value is ShiftType {
  return typeof value === "string" && SHIFT_TYPE_VALUES.includes(value);
}

// シフトカレンダーで使用するスタッフの最小型（feature間importを避けるため）
export interface ShiftStaff {
  id: string;
  name: string;
  /** occupations マスタ ID。未設定は null */
  occupationId: string | null;
  /** 職種表示名。未設定は null（UI では「未設定」） */
  occupationName: string | null;
}

// ─── シフトテンプレート型 ────────────────────────────────────────────────

/** バックエンド ShiftTemplateBreak の UI 型 */
interface ShiftTemplateBreak {
  id: string;
  shift_template_id: string;
  break_start: string;
  break_end: string;
}

/** バックエンド ShiftTemplate の UI 型 */
export interface ShiftTemplate {
  id: string;
  clinic_id: string;
  name: string;
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  notes: string;
  sort_order: number;
  is_active: boolean;
  breaks: ShiftTemplateBreak[];
  created_at: string;
  updated_at: string;
}

/** テンプレート作成入力 */
export interface CreateShiftTemplateInput {
  name: string;
  shift_type: ShiftType;
  start_time?: string;
  end_time?: string;
  notes?: string;
  sort_order?: number;
  is_active?: boolean;
  breaks?: { break_start: string; break_end: string }[];
}

/** テンプレート更新入力 */
export interface UpdateShiftTemplateInput {
  name?: string;
  shift_type?: ShiftType;
  start_time?: string | null;
  end_time?: string | null;
  notes?: string;
  sort_order?: number;
  is_active?: boolean;
  breaks?: { break_start: string; break_end: string }[];
}
