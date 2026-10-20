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
  /** EMR-457: 任意。空文字は未登録 */
  shortName: string;
  /** 0=日〜6=土。空配列は全曜日 */
  allowedWeekdays: number[];
  /** 保存する object key。空文字は画像なし */
  imageObjectKey: string;
  /** 表示専用（署名付き URL）。保存しない */
  imagePreviewUrl: string;
  /** 画像を選択／削除したか。未変更なら更新時に送らず既存値を保持する */
  imageChanged: boolean;
  /** 短縮名・曜日を編集したか。未変更なら更新時に送らず既存値を保持する */
  shortNameChanged: boolean;
  weekdaysChanged: boolean;
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
    shortName: item?.shortName ?? "",
    allowedWeekdays: item?.allowedWeekdays ?? [],
    imageObjectKey: item?.imageObjectKey ?? "",
    imagePreviewUrl: item?.imageUrl ?? "",
    imageChanged: false,
    shortNameChanged: false,
    weekdaysChanged: false,
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
