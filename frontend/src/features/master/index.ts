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
export { CampaignSettings } from "./routes/CampaignSettings";
export { PermissionGroupSettings } from "./routes/PermissionGroupSettings";
export { MedicineSettings } from "./routes/MedicineSettings";
export { MerchandiseItemSettings } from "./routes/MerchandiseItemSettings";
export { ReservationTypeSettings } from "./routes/ReservationTypeSettings";
export { LineReservationSlotsSettings } from "./routes/LineReservationSlotsSettings";
export { StaffSettings } from "./routes/StaffSettings";
export { TreatmentPlanMaster } from "./routes/TreatmentPlanMaster";
export { LabDeviceItemMasterSettings } from "./routes/LabDeviceItemMasterSettings";
export { TrimmingSettings } from "./routes/TrimmingSettings";
export { TrimmingCourseTypeSettings } from "./routes/TrimmingCourseTypeSettings";

export { useGetCompany, useUpdateCompany } from "./api/company";
export { useGetAllMedicines } from "./api/medicines";
export { useGetAllProcedures } from "./api/procedures";
export { useGetAllHospitalizationPlans } from "./api/hospitalization-plans";
export { useGetStaffs } from "./api/staffs";
export {
  useGetLabDeviceItemMasters,
  useEnsureLabDeviceItemMasters,
  useUpdateLabDeviceItemMaster,
} from "./api/lab-device-item-masters";
export {
  useGetLabDevices,
  useCreateLabDevice,
  useUpdateLabDevice,
} from "./api/lab-devices";
