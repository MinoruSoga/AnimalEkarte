import { ShiftTypeOff, ShiftTypePaidLeave } from "@/types/generated/models";

import type { ShiftType } from "../types";

/**
 * EMR-241: カテゴリ名による時刻フィールドの非表示化は廃止。
 * この判定は「時刻入力が必須か」（backend sharedkernel.RequiresTimeSlot と同一）のみに使い、
 * フィールドの表示・非表示や値の送信可否の分岐には使わない。
 */
export function requiresShiftTimes(shiftType: ShiftType): boolean {
  return shiftType !== ShiftTypeOff && shiftType !== ShiftTypePaidLeave;
}
