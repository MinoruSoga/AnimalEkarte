/**
 * Shifts feature types
 * Backend types: {@link ShiftEntry}, {@link ShiftType as BackendShiftType} from models.ts
 */
import {
  ShiftTypeFull,
  ShiftTypeMorning,
  ShiftTypeAfternoon,
  ShiftTypeOff,
  ShiftTypePaidLeave,
} from "@/types/generated/models";

/** ShiftType — models.ts const 値と一致する union 型（型安全性のため union 維持） */
export type ShiftType =
  | typeof ShiftTypeFull
  | typeof ShiftTypeMorning
  | typeof ShiftTypeAfternoon
  | typeof ShiftTypeOff
  | typeof ShiftTypePaidLeave;

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
  [ShiftTypeFull]: "全日",
  [ShiftTypeMorning]: "午前",
  [ShiftTypeAfternoon]: "午後",
  [ShiftTypeOff]: "休日",
  [ShiftTypePaidLeave]: "有休",
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
