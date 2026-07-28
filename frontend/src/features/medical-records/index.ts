export { MedicalRecordForm } from "./routes/MedicalRecordForm";
export { MedicalRecords } from "./routes/MedicalRecords";
export { MedicalRecordPetSelection } from "./routes/MedicalRecordPetSelection";

export type { MedicalRecord } from "./api/transforms";
export type { BackendMedicalRecord, CreateMedicalRecordRequest } from "./api/types";
export {
  useGetMedicalRecords,
  type MedicalRecordFilters,
  type MedicalRecordsResult,
} from "./api/get-medical-records";
