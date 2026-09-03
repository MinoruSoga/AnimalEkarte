import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";
import { formatJSTDate } from "@/lib/jst-date";

export { formatJSTDate };

export const DEFAULT_TRIMMING_FORM_DATA: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

export function parseTrimmingAppointmentId(appointmentId: unknown, searchValue: string | null): {
  appointmentIdFromState: number;
  existingAppointmentId: string;
} {
  const appointmentIdFromState = typeof appointmentId === "string"
    ? Number(appointmentId)
    : typeof appointmentId === "number"
      ? appointmentId
      : Number(searchValue ?? NaN);
  const existingAppointmentId = Number.isFinite(appointmentIdFromState)
    ? String(appointmentIdFromState)
    : "";
  return { appointmentIdFromState, existingAppointmentId };
}

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

const JST_OFFSET_MS = 9 * 60 * 60 * 1000;
/** トリミング施術の想定所要時間（record_shortcut の終了時刻算出に使用） */
const DEFAULT_TRIMMING_DURATION_MS = 90 * 60 * 1000;

/** record_shortcut の既定時刻。固定 10:00 だと uk_appointment_staff_time で同スタッフ同日が 409 になる (BUG-010)。 */
export function defaultRecordShortcutTimes(
  date: string,
  now = new Date(),
): { start: string; end: string } {
  const jst = new Date(now.getTime() + JST_OFFSET_MS);
  const pad = (value: number, width = 2) => String(value).padStart(width, "0");
  const time = `${pad(jst.getUTCHours())}:${pad(jst.getUTCMinutes())}:${pad(jst.getUTCSeconds())}.${pad(jst.getUTCMilliseconds(), 3)}`;
  const start = `${date}T${time}+09:00`;
  const endAt = new Date(Date.parse(start) + DEFAULT_TRIMMING_DURATION_MS);
  const endJst = new Date(endAt.getTime() + JST_OFFSET_MS);
  const end = `${endJst.getUTCFullYear()}-${pad(endJst.getUTCMonth() + 1)}-${pad(endJst.getUTCDate())}T${pad(endJst.getUTCHours())}:${pad(endJst.getUTCMinutes())}:${pad(endJst.getUTCSeconds())}.${pad(endJst.getUTCMilliseconds(), 3)}+09:00`;
  return { start, end };
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
