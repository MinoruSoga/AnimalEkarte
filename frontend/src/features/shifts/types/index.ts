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
export interface Shift {
  id: string;
  clinic_id: string;
  staff_id: string;
  staff_name?: string;
  date: string; // YYYY-MM-DD
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  note: string;
  created_at: string;
  updated_at: string;
}

/** @see {@link import("@/types/generated/models").ShiftEntry} */
export interface CreateShiftInput {
  staff_id: string;
  date: string; // YYYY-MM-DD
  shift_type: ShiftType;
  start_time?: string;
  end_time?: string;
  note?: string;
}

/** @see {@link import("@/types/generated/models").ShiftEntry} */
export interface UpdateShiftInput {
  shift_type?: ShiftType;
  start_time?: string;
  end_time?: string;
  note?: string;
}

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  [ShiftTypeFull]: "全日",
  [ShiftTypeMorning]: "午前",
  [ShiftTypeAfternoon]: "午後",
  [ShiftTypeOff]: "休日",
  [ShiftTypePaidLeave]: "有休",
};

export const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  [ShiftTypeFull]: "bg-blue-100 text-blue-800 border-blue-200",
  [ShiftTypeMorning]: "bg-green-100 text-green-800 border-green-200",
  [ShiftTypeAfternoon]: "bg-teal-100 text-teal-800 border-teal-200",
  [ShiftTypeOff]: "bg-gray-100 text-gray-600 border-gray-200",
  [ShiftTypePaidLeave]: "bg-purple-100 text-purple-800 border-purple-200",
};

// シフトカレンダーで使用するスタッフの最小型（feature間importを避けるため）
export interface ShiftStaff {
  id: string;
  name: string;
}
