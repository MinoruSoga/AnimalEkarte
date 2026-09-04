import {
  resolveTrimmingActiveFlag,
  type TargetSize,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";

export const TARGET_SIZE_EMPTY_VALUE = "__none__";
export const COURSE_TYPE_EMPTY_VALUE = "__none__";

export interface CourseFormData {
  name: string;
  price: string;
  targetSize: TargetSize | "";
  courseTypeId: string;
  duration: string;
  description: string;
  isActive: boolean;
}

export interface OptionFormData {
  name: string;
  price: string;
  duration: string;
  combinable: boolean;
  description: string;
  isActive: boolean;
}

export function trimmingCourseToFormData(item: TrimmingCourse | null): CourseFormData {
  return {
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    targetSize: item?.targetSize ?? "",
    courseTypeId: item?.courseTypeId ?? "",
    duration: item?.duration != null ? String(item.duration) : "",
    description: item?.description ?? "",
    isActive: item ? resolveTrimmingActiveFlag(item) : true,
  };
}

export function trimmingOptionToFormData(item: TrimmingOption | null): OptionFormData {
  return {
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    duration: item?.duration != null ? String(item.duration) : "",
    combinable: item?.combinable ?? true,
    description: item?.description ?? "",
    isActive: item ? resolveTrimmingActiveFlag(item) : true,
  };
}
