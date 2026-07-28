import {
  transformExamResult as transformSharedExamResult,
} from "@/lib/transforms/examination";
import type { BackendExamResult } from "./types";

export function transformExamResult(item: BackendExamResult) {
  return {
    ...transformSharedExamResult(item),
    isAssessed: item.is_assessed,
  };
}

export type ExamResult = ReturnType<typeof transformExamResult>;

export {
  transformExamination,
  type ExaminationRecord,
} from "@/lib/transforms/examination";
