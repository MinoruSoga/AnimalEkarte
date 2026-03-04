import type { VaccinationRecord } from "@/types";
import type { BackendVaccination } from "./types";

export function transformVaccination(data: BackendVaccination): VaccinationRecord {
  return {
    id: data.id,
    ownerName: data.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    vaccineName: data.vaccine_name,
    date: data.vaccination_date,
    nextDate: data.next_date || "",
  };
}
