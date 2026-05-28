import { ShiftTypeOff, ShiftTypePaidLeave } from "@/types/generated/models";

import type { ShiftType } from "../types";

export function isShiftTemplateTimeHidden(shiftType: ShiftType) {
  return shiftType === ShiftTypeOff || shiftType === ShiftTypePaidLeave;
}
