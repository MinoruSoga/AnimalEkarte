import type { ShiftEntry, ShiftIntervalState, ShiftWorkBreakView } from "./contract-types";
import type { ShiftType } from "../types";

/**
 * UI-facing shift（EMR-514 新契約の応答をそのままUIへ渡す形）。
 * 契約上 id/staff_id/clinic_id は既に正十進文字列のため Number 変換を行わない。
 * unknown の休憩（interval_state=unknown）は配列から黙って除外しない。
 */
export function transformShift(data: ShiftEntry) {
  return {
    id: data.id,
    clinic_id: data.clinic_id,
    staff_id: data.staff_id,
    staff_name: data.staff_name ?? "",
    date: data.date,
    shift_type: data.shift_type,
    starts_at: data.starts_at,
    ends_at: data.ends_at,
    version: data.version,
    interval_state: data.interval_state,
    notes: data.notes,
    breaks: data.breaks,
    created_at: data.created_at,
    updated_at: data.updated_at,
  };
}

export type Shift = ReturnType<typeof transformShift>;
export type { ShiftIntervalState, ShiftWorkBreakView, ShiftType };
