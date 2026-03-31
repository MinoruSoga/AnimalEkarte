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
  ResourceMaster,
  ResourceHospitalSettings,
  type Resource,
} from "@/types/generated/models";

export type { Resource };

export interface ResourceDefinition {
  key: Resource;
  label: string;
}

export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: ResourceDashboard, label: "ダッシュボード" },
  { key: ResourceOwners, label: "オーナー管理" },
  { key: ResourceReservations, label: "予約" },
  { key: ResourceMedicalRecords, label: "カルテ" },
  { key: ResourceHospitalization, label: "入院・ホテル" },
  { key: ResourceTrimming, label: "トリミング" },
  { key: ResourceExaminations, label: "検査" },
  { key: ResourceAccounting, label: "会計" },
  { key: ResourceVaccinations, label: "ワクチン" },
  { key: ResourceCheckups, label: "健診" },
  { key: ResourceInventory, label: "在庫" },
  { key: ResourceEstimates, label: "見積" },
  { key: ResourceShifts, label: "シフト" },
  { key: ResourceMaster, label: "マスタ管理" },
  { key: ResourceHospitalSettings, label: "病院設定" },
];
