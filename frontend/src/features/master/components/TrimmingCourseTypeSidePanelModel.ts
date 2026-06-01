import type { TrimmingCourseType } from "../api/trimming-course-type";

export interface TrimmingCourseTypeFormData {
  name: string;
  isActive: boolean;
}

export function trimmingCourseTypeToFormData(
  item: TrimmingCourseType | null,
): TrimmingCourseTypeFormData {
  return {
    name: item?.name ?? "",
    isActive: item?.isActive ?? true,
  };
}
