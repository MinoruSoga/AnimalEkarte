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

function sortedWeekdays(days: number[]): number[] {
  return [...new Set(days)].sort((a, b) => a - b);
}

export function buildTrimmingCourseCreateRequest(
  data: CourseFormData,
): CreateTrimmingCourseRequest {
  const shortName = data.shortName.trim();
  const weekdays = sortedWeekdays(data.allowedWeekdays);
  return {
    name: data.name,
    price: toNullableNumber(data.price),
    target_size: data.targetSize !== "" ? data.targetSize : null,
    course_type_id: data.courseTypeId !== "" ? Number(data.courseTypeId) : undefined,
    duration: toNullableNumber(data.duration),
    description: data.description || undefined,
    is_active: true,
    short_name: shortName !== "" ? shortName : undefined,
    allowed_weekdays: weekdays.length > 0 ? weekdays : undefined,
    image_object_key: data.imageObjectKey !== "" ? data.imageObjectKey : undefined,
  };
}

/**
 * 更新は編集した属性だけ送る（省略 = 既存値保持）。編集した短縮名・曜日は
 * "" / [] で未登録・全曜日へ戻す。画像も変更時だけ送る。
 */
export function buildTrimmingCourseUpdateRequest(
  data: CourseFormData,
): UpdateTrimmingCourseRequest {
  const {
    image_object_key: _createImage,
    short_name: _createShortName,
    allowed_weekdays: _createWeekdays,
    ...base
  } = buildTrimmingCourseCreateRequest(data);
  return {
    ...base,
    // EMR-448: 備考を空へ変更できるよう、空文字も省略せず送る（BE は "" を空で保存）。
    description: data.description,
    is_active: data.isActive,
    ...(data.shortNameChanged ? { short_name: data.shortName.trim() } : {}),
    ...(data.weekdaysChanged ? { allowed_weekdays: sortedWeekdays(data.allowedWeekdays) } : {}),
    ...(data.imageChanged ? { image_object_key: data.imageObjectKey } : {}),
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
    // EMR-448: 備考を空へ変更できるよう、空文字も省略せず送る（BE は "" を空で保存）。
    description: data.description,
    is_active: data.isActive,
  };
}
