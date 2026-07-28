import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";
import { formatJSTDate } from "@/lib/jst-date";

export { formatJSTDate };

interface TrimmingReservationType {
  id: number;
  category: string;
  is_internal: boolean;
  sort_order: number;
}

export interface TrimmingReservationTypeGroup {
  types: TrimmingReservationType[];
}

function optionalNumber(value: string): number | undefined {
  if (value === "") return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function optionalDateTime(value: string): string | undefined {
  return value === "" ? undefined : value;
}

export function normalizeVisitDate(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  return /^\d{4}-\d{2}-\d{2}$/.test(value) ? value : undefined;
}

export function findDefaultTrimmingReservationTypeId(
  groups: readonly TrimmingReservationTypeGroup[] | undefined,
): number | undefined {
  return groups
    ?.flatMap((group) => group.types)
    .filter((type) => type.category === "trimming" && !type.is_internal)
    .toSorted((a, b) => a.sort_order - b.sort_order)[0]?.id;
}

interface FilterableMasterItem {
  id: string;
  name: string;
  status?: string;
}

/**
 * #228: コース/オプション選択肢を active のみに絞り込む。ただし selectedIds に含まれる
 * 無効アイテムは名前に「（無効）」を付与して維持する（編集中カルテのデータを消さないため）。
 */
export function filterActiveOrSelectedMasterItems<T extends FilterableMasterItem>(
  items: T[],
  selectedIds: string[],
): T[] {
  const active = items.filter((item) => item.status === "active");
  const activeIds = new Set(active.map((item) => item.id));
  const itemsById = new Map(items.map((item) => [item.id, item]));
  const uniqueSelectedIds = [...new Set(selectedIds)];
  const retainedInactive = uniqueSelectedIds
    .filter((id) => id !== "" && !activeIds.has(id))
    .map((id) => itemsById.get(id))
    .filter((item): item is T => item != null)
    .map((item) => ({ ...item, name: `${item.name}（無効）` }));
  return [...active, ...retainedInactive];
}

export function buildUpdateTrimmingRequest(formData: TrimmingFormData): UpdateTrimmingRequest {
  return {
    start_time: optionalDateTime(formData.startTime),
    end_time: optionalDateTime(formData.endTime),
    staff_id: optionalNumber(formData.staffId),
    course_id: optionalNumber(formData.courseId),
    style_request: formData.styleRequest,
    bw: optionalNumber(formData.bw),
    bw_unit: formData.bwUnit,
    bt: optionalNumber(formData.bt),
    used_shampoo: formData.usedShampoo,
    used_ribbon: formData.usedRibbon,
    remarks: formData.remarks,
    option_ids: (formData.optionIds ?? []).map(Number),
  };
}

export function buildCreateTrimmingRequest(
  formData: TrimmingFormData,
  petID: number,
  reservationTypeID: number,
  startTime: string | undefined,
  endTime: string | undefined,
  appointmentID?: number,
): CreateTrimmingRequest {
  return {
    appointment_id: appointmentID,
    reservation_type_id: reservationTypeID,
    start_time: startTime,
    end_time: endTime,
    pet_id: petID,
    staff_id: optionalNumber(formData.staffId),
    course_id: optionalNumber(formData.courseId),
    style_request: formData.styleRequest,
    bw: optionalNumber(formData.bw),
    bw_unit: formData.bwUnit,
    bt: optionalNumber(formData.bt),
    used_shampoo: formData.usedShampoo,
    used_ribbon: formData.usedRibbon,
    remarks: formData.remarks,
    option_ids: (formData.optionIds ?? []).map(Number),
  };
}
