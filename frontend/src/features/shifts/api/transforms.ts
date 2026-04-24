import type { BackendShift } from "./types";
import type { ShiftType } from "../types";

export function transformShift(data: BackendShift) {
  return {
    id: String(data.id ?? 0),
    clinic_id: String(data.clinic_id ?? 0),
    staff_id: String(data.staff_id ?? 0),
    staff_name: data.staff?.name ?? data.staff_name ?? "",
    date: data.date ? String(data.date).split("T")[0] : "",
    shift_type: data.shift_type as ShiftType,
    start_time: data.start_time ?? "",
    end_time: data.end_time ?? "",
    notes: data.notes ?? "",
    breaks: (data.breaks ?? []).map((b) => ({
      id: String(b.id ?? 0),
      break_start: b.break_start ?? "",
      break_end: b.break_end ?? "",
    })),
    created_at: data.created_at ?? "",
    updated_at: data.updated_at ?? "",
  };
}

export type Shift = ReturnType<typeof transformShift>;
