import type { Shift } from "../types";
import type { BackendShift } from "./types";

export function transformShift(data: BackendShift): Shift {
  return {
    id: String(data.id ?? 0),
    clinic_id: String(data.clinic_id ?? 0),
    staff_id: String(data.staff_id ?? 0),
    staff_name: data.staff?.name ?? data.staff_name ?? "",
    date: data.date ? String(data.date).split("T")[0] : "",
    shift_type: data.shift_type,
    start_time: data.start_time ?? "",
    end_time: data.end_time ?? "",
    note: data.note ?? "",
    created_at: data.created_at ?? "",
    updated_at: data.updated_at ?? "",
  };
}
