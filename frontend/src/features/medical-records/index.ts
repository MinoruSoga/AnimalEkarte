export { MedicalRecordForm } from "./routes/MedicalRecordForm";
export { MedicalRecords } from "./routes/MedicalRecords";
export { MedicalRecordPetSelection } from "./routes/MedicalRecordPetSelection";

export { MedicalRecordAddenda } from "./components/MedicalRecordAddenda";
export { useMedicalRecordAddenda, useCreateMedicalRecordAddendum } from "./hooks/use-medical-record-addenda";
export type { MedicalRecordAddendum } from "@/types/generated/models";
export type { MedicalRecord } from "./api/transforms";
