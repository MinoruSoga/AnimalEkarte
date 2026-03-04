import type { TrimmingRecord } from "@/types";
import type { BackendTrimming } from "./types";

export function transformTrimming(data: BackendTrimming): TrimmingRecord {
  return {
    id: data.id,
    date: data.appointment_date,
    petNumber: data.pet?.pet_number ?? "",
    petName: data.pet?.name ?? "",
    ownerName: data.owner?.name ?? "",
    species: data.pet?.species ?? "",
    weight: data.pet?.weight != null ? String(data.pet.weight) : "",
    styleRequest: data.style_request ?? "",
    staff: data.staff_id ?? "",
    status: data.status,
  };
}
