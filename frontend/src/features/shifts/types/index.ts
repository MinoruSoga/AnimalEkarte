/**
 * Shifts feature types
 * Backend types: {@link ShiftEntry}, {@link ShiftType as BackendShiftType} from models.ts
 */
/**
 * 手書き literal union（FE4-3）。
 * tygo 生成定数は `export const X: ShiftType = "full"` 形式（ShiftType = string）のため
 * `typeof 生成定数` は string に退化する（旧 JSDoc の「型安全性のため union 維持」という記述は
 * 実態に反していた）。生成側は編集禁止のため、手書き literal union +
 * ランタイム値集合の drift テスト（./union-drift.test.ts）で生成定数の値集合との一致を機械固定する。
 */
export type ShiftType = "full" | "morning" | "afternoon" | "off" | "paid_leave";

/** UI-facing shift (string IDs — post-transform) */
export type { Shift } from "../api/transforms";

/** @see {@link import("@/types/generated/models").ShiftEntry} */
export interface ShiftBreakInput {
  break_start: string;
  break_end: string;
}

export interface CreateShiftInput {
  staff_id: string;
  date: string; // YYYY-MM-DD
  shift_type: ShiftType;
  start_time?: string;
  end_time?: string;
  notes?: string;
  breaks?: ShiftBreakInput[];
}

/** @see {@link import("@/types/generated/models").ShiftEntry} */
export interface UpdateShiftInput {
  shift_type?: ShiftType;
  start_time?: string;
  end_time?: string;
  notes?: string;
  breaks?: ShiftBreakInput[];
}

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  full: "全日",
  morning: "午前",
  afternoon: "午後",
  off: "休日",
  paid_leave: "有休",
};


// シフトカレンダーで使用するスタッフの最小型（feature間importを避けるため）
export interface ShiftStaff {
  id: string;
  name: string;
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
