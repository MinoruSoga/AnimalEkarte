import { transformExamResult as transformSharedExamResult } from "@/lib/transforms/examination";
import type { BackendExamResult } from "../types";

export function transformExamResult(item: BackendExamResult) {
  return {
    ...transformSharedExamResult(item),
    isAssessed: item.is_assessed,
    // 保存済み定性基準の snapshot（マスタ再解決しない）。命名は master qualitativeMin に揃える。
    qualitativeMin: item.qualitative_min ?? undefined,
    qualitativeMax: item.qualitative_max ?? undefined,
  };
}

export type ExamResult = ReturnType<typeof transformExamResult>;

export { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
