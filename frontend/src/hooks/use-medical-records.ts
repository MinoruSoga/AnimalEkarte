/**
 * FE-RC-015: Shared import surface for medical-records list query.
 * Cross-feature consumers (owner-report) must import from here — never from
 * `@/features/medical-records`.
 */
export { useGetMedicalRecords } from "../features/medical-records/api/get-medical-records";
