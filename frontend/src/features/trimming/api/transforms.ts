import type { TrimmingRecord } from "@/types";
import type { BackendTrimming } from "./types";

export function transformTrimming(data: BackendTrimming): TrimmingRecord {
  const statusMap: Record<string, TrimmingRecord["status"]> = {
    completed: "完了",
    reserved: "予約",
    in_progress: "進行中",
  };

  return {
    id: String(data.id ?? 0),
    date: data.date ? String(data.date).split("T")[0] : "",
    petId: data.pet?.id != null ? String(data.pet.id) : undefined,
    ownerId: data.pet?.owner?.id != null ? String(data.pet.owner.id) : undefined,
    petNumber: data.pet?.pet_number ?? "",
    petName: data.pet?.name ?? "",
    ownerName: data.pet?.owner?.owner_name ?? "",
    species: data.pet?.animal_species?.name ?? "",
    weight: data.pet?.weight != null ? String(data.pet.weight) : "",
    styleRequest: data.style_request ?? "",
    staff: data.staff?.name ?? "",
    status: statusMap[data.status] ?? "予約",
  };
}
