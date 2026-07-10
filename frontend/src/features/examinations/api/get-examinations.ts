// R-F2-S8: shared hook (@/hooks/use-examinations) へ昇格。
// medical-records から examinations feature への直接 import を避けるための re-export。
export {
  useGetExaminations,
  type ExaminationFilters,
} from "@/hooks/use-examinations";
