import type { VaccinationRecord } from "@/types";
import type { BackendVaccination } from "./types";

export function transformVaccination(
  data: BackendVaccination
): VaccinationRecord {
  return {
    id: String(data.id ?? 0),
    ownerName: "",
    petName: data.pet?.name ?? "",
    vaccineName: data.vaccine?.name ?? "",
    date: data.date ?? "",
    nextDate: data.next_date ?? "",
  };
}
