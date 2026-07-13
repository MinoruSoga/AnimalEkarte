import type { BackendTrimming } from "@/types/trimming";

type TrimmingStatus = "完了" | "予約" | "進行中" | "キャンセル";

/**
 * reservation_status → トリミング表示ステータス変換
 * BE-119: appointments ベース移行により status 値が変わった
 */
const STATUS_MAP: Record<string, TrimmingStatus> = {
  confirmed: "予約",
  checked_in: "予約",
  // #233: カルテ直接新規作成で選択可能になった pending を明示（従来はフォールバックに依存）
  pending: "予約",
  in_consultation: "進行中",
  accounting: "進行中",
  completed: "完了",
  cancelled: "キャンセル",
  canceled: "キャンセル",
  no_show: "キャンセル",
};

export function transformTrimming(data: BackendTrimming) {
  // start_time から日付部分を抽出（"2025-10-10T10:00:00+09:00" → "2025-10-10"）
  const date =
    data.start_time && !String(data.start_time).startsWith("0001")
      ? String(data.start_time).split("T")[0]
      : "";

  return {
    id: String(data.id ?? 0),
    reservationTypeId: data.reservation_type_id != null ? String(data.reservation_type_id) : "",
    hasDetail: data.has_detail ?? false,
    date,
    petId: data.pet?.id != null ? String(data.pet.id) : undefined,
    ownerId: data.pet?.owner?.id != null ? String(data.pet.owner.id) : undefined,
    petNumber: data.pet?.pet_number ?? "",
    petName: data.pet?.name ?? "",
    ownerName: data.pet?.owner?.name ?? "",
    species: data.pet?.animal_species?.name ?? "",
    breed: data.pet?.breed ?? "",
    weight: data.pet?.weight != null ? String(data.pet.weight) : "",
    styleRequest: data.style_request ?? "",
    staff: data.staff?.name ?? "",
    status: (STATUS_MAP[data.status] ?? "予約") as TrimmingStatus,
    // Form fields
    staffId: data.staff_id != null ? String(data.staff_id) : "",
    courseId: data.course?.id != null ? String(data.course.id) : "",
    // 表示用コース名（一覧 API は TrimmingDetail.Course を preload するため埋まる。未設定は空文字）。
    courseName: data.course?.name ?? "",
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

export type TrimmingUI = ReturnType<typeof transformTrimming>;
