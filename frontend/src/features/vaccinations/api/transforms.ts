import { formatJSTDate } from "@/lib/jst-date";
import type { BackendVaccination } from "./types";

/** SD-19: 絶対時刻の UTC 切り出しではなく JST 壁日付 YYYY-MM-DD。 */
function toJSTDateOnly(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
  return formatJSTDate(d);
}

export function transformVaccination(data: BackendVaccination) {
  return {
    id: String(data.id ?? 0),
    petId: data.pet_id ? String(data.pet_id) : undefined,
    ownerName: data.pet?.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    vaccineId: String(data.vaccine_id ?? 0),
    vaccineName: data.vaccine?.name ?? "",
    doctor: data.doctor?.name ?? "",
    date: toJSTDateOnly(data.date),
    nextDate: toJSTDateOnly(data.next_date),
    nextScheduleType: data.next_schedule_type || undefined,
    lot1: data.lot1 || undefined,
    lot2: data.lot2 || undefined,
    lot3: data.lot3 || undefined,
    lot4: data.lot4 || undefined,
    supplemental: data.supplemental || undefined,
    remarks: data.remarks || undefined,
  };
}

export type VaccinationRecord = ReturnType<typeof transformVaccination>;
