import type { TrimmingUI } from "@/types";
import type { BackendTrimming } from "@/types/trimming";

/**
 * reservation_status → トリミング表示ステータス変換
 * BE-119: appointments ベース移行により status 値が変わった
 */
const STATUS_MAP: Record<string, TrimmingUI["status"]> = {
  confirmed: "予約",
  checked_in: "予約",
  in_consultation: "進行中",
  accounting: "進行中",
  completed: "完了",
  canceled: "キャンセル",
  no_show: "キャンセル",
};

export function transformTrimming(data: BackendTrimming): TrimmingUI {
  // start_time から日付部分を抽出（"2025-10-10T10:00:00+09:00" → "2025-10-10"）
  const date =
    data.start_time && !String(data.start_time).startsWith("0001")
      ? String(data.start_time).split("T")[0]
      : "";

  return {
    id: String(data.id ?? 0),
    date,
    petId: data.pet?.id != null ? String(data.pet.id) : undefined,
    ownerId: data.pet?.owner?.id != null ? String(data.pet.owner.id) : undefined,
    petNumber: data.pet?.pet_number ?? "",
    petName: data.pet?.name ?? "",
    ownerName: data.pet?.owner?.name ?? "",
    species: data.pet?.animal_species?.name ?? "",
    weight: data.pet?.weight != null ? String(data.pet.weight) : "",
    styleRequest: data.style_request ?? "",
    staff: data.staff?.name ?? "",
    status: STATUS_MAP[data.status] ?? "予約",
    // Form fields
    staffId: data.staff_id != null ? String(data.staff_id) : "",
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
