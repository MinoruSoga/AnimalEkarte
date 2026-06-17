export { MedicalRecordForm } from "./routes/MedicalRecordForm";
export { MedicalRecords } from "./routes/MedicalRecords";
export { MedicalRecordPetSelection } from "./routes/MedicalRecordPetSelection";

export { MedicalRecordAddenda } from "./components/MedicalRecordAddenda";
export { useMedicalRecordAddenda, useCreateMedicalRecordAddendum } from "./hooks/use-medical-record-addenda";
export type { MedicalRecordAddendum } from "@/types/generated/models";
export type { MedicalRecord } from "./api/transforms";

// #158 飼主レポートで再利用する pet 単位の予防接種履歴 hook
export { useGetPetVaccinations } from "./api/get-pet-vaccinations";
