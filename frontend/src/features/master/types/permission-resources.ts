import {
  ResourceDashboard,
  ResourceOwners,
  ResourceReservations,
  ResourceMedicalRecords,
  ResourceHospitalization,
  ResourceTrimming,
  ResourceExaminations,
  ResourceAccounting,
  ResourceVaccinations,
  ResourceCheckups,
  ResourceInventory,
  ResourceEstimates,
  ResourceShifts,
  ResourceHospitalSettings,
  ResourceMasterAnimalSpecies,
  ResourceMasterMedical,
  ResourceMasterServiceType,
  ResourceMasterHospitalization,
  ResourceMasterTrimming,
  ResourceMasterPermission,
  ResourceMasterStaff,
  ResourceMasterInsurance,
  ResourceMasterMerchandise,
  type Resource,
} from "@/types/generated/models";

export type { Resource };

export interface ResourceDefinition {
  key: Resource;
  label: string;
}

export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: ResourceDashboard, label: "当日の受付" },
  { key: ResourceOwners, label: "飼主・ペット" },
  { key: ResourceReservations, label: "予約管理" },
  { key: ResourceMedicalRecords, label: "カルテ" },
  { key: ResourceHospitalization, label: "入院・ホテル" },
  { key: ResourceTrimming, label: "トリミング" },
  { key: ResourceExaminations, label: "検査管理" },
  { key: ResourceAccounting, label: "会計管理" },
  { key: ResourceVaccinations, label: "予防接種" },
  { key: ResourceCheckups, label: "定期健診" },
  { key: ResourceInventory, label: "在庫管理" },
  { key: ResourceEstimates, label: "見積書" },
  { key: ResourceShifts, label: "シフト管理" },
  { key: ResourceHospitalSettings, label: "医院" },
  { key: ResourceMasterAnimalSpecies, label: "動物種類" },
  { key: ResourceMasterMedical, label: "カルテ関連" },
  { key: ResourceMasterServiceType, label: "診療サービス" },
  { key: ResourceMasterHospitalization, label: "入院・ケージ" },
  { key: ResourceMasterTrimming, label: "トリミングマスタ" },
  { key: ResourceMasterPermission, label: "権限グループ" },
  { key: ResourceMasterStaff, label: "スタッフ管理" },
  { key: ResourceMasterInsurance, label: "保険マスタ" },
  { key: ResourceMasterMerchandise, label: "物販・フード" },
];
