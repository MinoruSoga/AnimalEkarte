export {
  getMedicalRecords,
  useGetMedicalRecords,
} from "./get-medical-records";
export {
  getMedicalRecord,
  useGetMedicalRecord,
  getMedicalRecordsByPetId,
  useGetMedicalRecordsByPetId,
  getMedicalRecordsByOwnerId,
  useGetMedicalRecordsByOwnerId,
} from "./get-medical-record";
export {
  createMedicalRecord,
  useCreateMedicalRecord,
} from "./create-medical-record";
export {
  updateMedicalRecord,
  useUpdateMedicalRecord,
} from "./update-medical-record";
export {
  deleteMedicalRecord,
  useDeleteMedicalRecord,
} from "./delete-medical-record";
export type {
  BackendMedicalRecord,
  CreateMedicalRecordRequest,
  UpdateMedicalRecordRequest,
} from "./types";
