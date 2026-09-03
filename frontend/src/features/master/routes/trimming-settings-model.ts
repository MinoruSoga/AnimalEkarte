import type { CourseFormData, OptionFormData } from "../lib/trimming-side-panel-model";
import type {
  CreateTrimmingCourseRequest,
  CreateTrimmingOptionRequest,
  UpdateTrimmingCourseRequest,
  UpdateTrimmingOptionRequest,
} from "../api/trimming";

export const TRIMMING_TABS = [
  { value: "course", label: "コース" },
  { value: "option", label: "オプション" },
] as const;

export type TrimmingTabValue = (typeof TRIMMING_TABS)[number]["value"];

const TRIMMING_TAB_VALUES = new Set<string>(TRIMMING_TABS.map((tab) => tab.value));

export function toTrimmingTabValue(value: string | null): TrimmingTabValue {
  return TRIMMING_TAB_VALUES.has(value ?? "") ? (value as TrimmingTabValue) : "course";
}

function toNullableNumber(value: string): number | null {
  return value !== "" ? Number(value) : null;
}

export function buildTrimmingCourseCreateRequest(
  data: CourseFormData,
): CreateTrimmingCourseRequest {
  return {
    name: data.name,
    price: toNullableNumber(data.price),
    target_size: data.targetSize !== "" ? data.targetSize : null,
    course_type_id: data.courseTypeId !== "" ? Number(data.courseTypeId) : undefined,
    duration: toNullableNumber(data.duration),
    description: data.description || undefined,
    is_active: true,
  };
}

export function buildTrimmingCourseUpdateRequest(
  data: CourseFormData,
): UpdateTrimmingCourseRequest {
  return {
    ...buildTrimmingCourseCreateRequest(data),
    is_active: data.isActive,
  };
}

export function buildTrimmingOptionCreateRequest(
  data: OptionFormData,
): CreateTrimmingOptionRequest {
  return {
    name: data.name,
    price: toNullableNumber(data.price),
    duration: toNullableNumber(data.duration),
    is_combinable: data.combinable,
    description: data.description || undefined,
    is_active: true,
  };
}

export function buildTrimmingOptionUpdateRequest(
  data: OptionFormData,
): UpdateTrimmingOptionRequest {
  return {
    ...buildTrimmingOptionCreateRequest(data),
    is_active: data.isActive,
  };
}
