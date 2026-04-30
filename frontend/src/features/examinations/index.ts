export { ExaminationsList } from "./routes/ExaminationsList";
export { ExaminationForm } from "./routes/ExaminationForm";
export { ExaminationPetSelection } from "./routes/ExaminationPetSelection";
export { useGetExaminations } from "./api/get-examinations";
export { useUpdateExamination } from "./api/update-examination";
export { useGetExaminationItems } from "./api/get-examination-items";
export { useUpdateExaminationItems } from "./api/update-examination-items";
export { ExamItemsTable } from "./components/ExamItemsTable";
export type { ExamItemRow } from "./components/ExamItemsTable";
// Shared transform / types — single source of truth for examination view models.
// medical-records など他 feature から再利用するため公開する。
export { transformExamResult, transformExamination } from "./api/transforms";
export type { ExamResult, ExaminationRecord } from "./api/transforms";
