import type { VaccinationRecord } from "@/types";
import type { BackendVaccination } from "./types";

export function transformVaccination(
  data: BackendVaccination
): VaccinationRecord {
  return {
    id: String(data.id ?? 0),
    petId: data.pet_id ? String(data.pet_id) : undefined,
    ownerName: data.pet?.owner?.owner_name ?? "",
    petName: data.pet?.name ?? "",
    vaccineId: String(data.vaccine_id ?? 0),
    vaccineName: data.vaccine?.name ?? "",
    doctor: data.doctor?.name ?? "",
    date: data.date ? data.date.slice(0, 10) : "",
    nextDate: data.next_date ? data.next_date.slice(0, 10) : "",
    nextScheduleType: data.next_schedule_type || undefined,
    lot1: data.lot1 || undefined,
    lot2: data.lot2 || undefined,
    lot3: data.lot3 || undefined,
    lot4: data.lot4 || undefined,
    supplemental: data.supplemental || undefined,
    remarks: data.remarks || undefined,
  };
}
