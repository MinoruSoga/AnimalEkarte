import type { CreateTrimmingRequest, UpdateTrimmingRequest, TrimmingFormData } from "@/types/trimming";

const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

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

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

export function formatJSTDate(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}`;
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
    .sort((a, b) => a.sort_order - b.sort_order)[0]?.id;
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
