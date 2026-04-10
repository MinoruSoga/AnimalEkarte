import type { TrimmingUI } from "@/types";
import type { BackendTrimming } from "@/types/trimming";

export function transformTrimming(data: BackendTrimming): TrimmingUI {
  const statusMap: Record<string, TrimmingUI["status"]> = {
    completed: "完了",
    reserved: "予約",
    in_progress: "進行中",
  };

  return {
    id: String(data.id ?? 0),
    date: data.date && !String(data.date).startsWith("0001") ? String(data.date).split("T")[0] : "",
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
    // Form fields
    staffId: data.staff?.id != null ? String(data.staff.id) : "",
    courseId: data.course?.id != null ? String(data.course.id) : "",
    optionIds: data.options?.map((o) => String(o.id)) ?? [],
    bw: data.bw != null ? String(data.bw) : "",
    bwUnit: (data.bw_unit as "Kg" | "g") || "Kg",
    bt: data.bt != null ? String(data.bt) : "",
    usedShampoo: data.used_shampoo ?? "",
    usedRibbon: data.used_ribbon ?? "",
    remarks: data.remarks ?? "",
    styleImage: data.style_image || undefined,
    completedImage: data.completed_image || undefined,
  };
}
