// R-F2-S8: shared hook (@/hooks/use-update-examination) へ昇格。
// medical-records から examinations feature への直接 import を避けるための re-export。
// 注意: このモジュールの UpdateExaminationRequest は shared hook 内の narrow 定義。
// feature 内部の正本は ./types.ts の同名 interface（BE 契約は同一）。
export {
  useUpdateExamination,
} from "@/hooks/use-update-examination";
