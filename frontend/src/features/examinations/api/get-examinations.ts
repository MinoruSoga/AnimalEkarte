// R-F2-S8: shared hook (@/hooks/use-examinations) へ昇格。
// medical-records / examinations feature 内部の両方がこのバレル経由で参照する
// （ExaminationImportDialog.tsx は現状 @/hooks/use-examinations を直接 import 済みで
// このコメントの前提は一部陳腐化しているが、feature 内部からの参照は本バレル経由に統一する）。
export {
  useGetExaminations,
  useGetExaminationsPage,
  type ExaminationFilters,
} from "@/hooks/use-examinations";
