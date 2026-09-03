/**
 * FE-RC-015: Shared import surface for medical-records list query.
 * Cross-feature consumers (owner-report) must import from here — never from
 * `@/features/medical-records`. Implementation stays in the feature API module
 * (transformMedicalRecord remains the domain source of truth).
 */
export {
  useGetMedicalRecords,
  useGetPetMedicalHistory,
  type MedicalRecordFilters,
  type MedicalRecordsResult,
  type MedicalRecordSortKey,
} from "../features/medical-records/api/get-medical-records";
