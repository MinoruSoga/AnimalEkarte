import type {
  CreateTrimmingCourseTypeRequest,
  UpdateTrimmingCourseTypeRequest,
} from "../api/trimming-course-type";
import type { TrimmingCourseTypeFormData } from "../lib/trimming-course-type-side-panel-model";

export function buildTrimmingCourseTypeCreateRequest(
  data: TrimmingCourseTypeFormData,
): CreateTrimmingCourseTypeRequest {
  return {
    name: data.name,
    is_active: data.isActive,
  };
}

export function buildTrimmingCourseTypeUpdateRequest(
  data: TrimmingCourseTypeFormData,
): UpdateTrimmingCourseTypeRequest {
  return buildTrimmingCourseTypeCreateRequest(data);
}
