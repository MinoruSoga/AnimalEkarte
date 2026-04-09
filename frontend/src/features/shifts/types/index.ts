/**
 * Shifts feature types
 * Backend types: {@link ShiftEntry}, {@link ShiftType as BackendShiftType} from models.ts
 */
import { C } from "@/lib/design-tokens";
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

/** 休憩時間 */
export interface ShiftBreak {
  id: string;
  break_start: string;
  break_end: string;
}

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
  breaks: ShiftBreak[];
  created_at: string;
  updated_at: string;
}

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
  note?: string;
  breaks?: ShiftBreakInput[];
}

/** @see {@link import("@/types/generated/models").ShiftEntry} */
export interface UpdateShiftInput {
  shift_type?: ShiftType;
  start_time?: string;
  end_time?: string;
  note?: string;
  breaks?: ShiftBreakInput[];
}

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  [ShiftTypeFull]: "全日",
  [ShiftTypeMorning]: "午前",
  [ShiftTypeAfternoon]: "午後",
  [ShiftTypeOff]: "休日",
  [ShiftTypePaidLeave]: "有休",
};

export const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  [ShiftTypeFull]: `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}`,
  [ShiftTypeMorning]: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  [ShiftTypeAfternoon]: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  [ShiftTypeOff]: `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`,
  [ShiftTypePaidLeave]: `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}`,
};

// シフトカレンダーで使用するスタッフの最小型（feature間importを避けるため）
export interface ShiftStaff {
  id: string;
  name: string;
}
