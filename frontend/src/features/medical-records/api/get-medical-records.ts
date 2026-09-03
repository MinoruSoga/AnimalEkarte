/**
 * FE-RC-015 followup2: list / history hook の実体は @/hooks/use-medical-records。
 * 本ファイルは順方向 re-export のみ（実装を持たない）。
 */
export {
  useGetMedicalRecords,
  getMedicalRecords,
  useGetPetMedicalHistory,
  type MedicalRecordFilters,
  type MedicalRecordSortKey,
  type MedicalRecordsResult,
  type MedicalRecordInterviewHistoryItem,
} from "@/hooks/use-medical-records";
