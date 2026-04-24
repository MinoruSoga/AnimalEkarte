export { MasterSettingsIndex } from "./routes/MasterSettingsIndex";
export { AnimalSpeciesSettings } from "./routes/AnimalSpeciesSettings";
export { CageSettings } from "./routes/CageSettings";
export { ChiefComplaintSettings } from "./routes/ChiefComplaintSettings";
export { DiagnosisSettings } from "./routes/DiagnosisSettings";
export { HospitalizationSettings } from "./routes/HospitalizationSettings";
export { InsuranceSettings } from "./routes/InsuranceSettings";
export { InterviewTemplateSettings } from "./routes/InterviewTemplateSettings";
export { OccupationSettings } from "./routes/OccupationSettings";
export { PaymentMethodSettings } from "./routes/PaymentMethodSettings";
export { PermissionGroupSettings } from "./routes/PermissionGroupSettings";
export { MedicineSettings } from "./routes/MedicineSettings";
export { MerchandiseItemSettings } from "./routes/MerchandiseItemSettings";
export { ReservationTypeSettings } from "./routes/ReservationTypeSettings";
export { StaffSettings } from "./routes/StaffSettings";
export { TreatmentPlanMaster } from "./routes/TreatmentPlanMaster";
export { TrimmingSettings } from "./routes/TrimmingSettings";

export { useMasterItems } from "./hooks/use-master-items";
export { useReservationTypeColorMap } from "./hooks/use-reservation-type-color-map";
export { useGetCompany } from "./api/company";
export { useGetAllConsultations } from "./api/consultations";
export { useGetAllProcedures } from "./api/procedures";
export { useGetAllVaccinesMaster } from "./api/vaccines-master";
export { useGetAllCheckupTypes } from "./api/checkup-types";
export type { CheckupTypeItem } from "./api/checkup-types";
export { useGetStaffs } from "./api/staffs";
export {
  useGetReservationTypeGroups,
  useCreateReservationTypeGroup,
  useUpdateReservationTypeGroup,
  useDeleteReservationTypeGroup,
} from "./api/reservation-type-groups";
export type { ReservationTypeGroup } from "./api/reservation-type-groups";
