export interface ResourceDefinition {
  key: string;
  label: string;
}

export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: "dashboard", label: "ダッシュボード" },
  { key: "owners", label: "オーナー管理" },
  { key: "reservations", label: "予約" },
  { key: "medical-records", label: "カルテ" },
  { key: "hospitalization", label: "入院・ホテル" },
  { key: "trimming", label: "トリミング" },
  { key: "examinations", label: "検査" },
  { key: "accounting", label: "会計" },
  { key: "vaccinations", label: "ワクチン" },
  { key: "checkups", label: "健診" },
  { key: "inventory", label: "在庫" },
  { key: "estimates", label: "見積" },
  { key: "shifts", label: "シフト" },
  { key: "master", label: "マスタ管理" },
  { key: "hospital-settings", label: "病院設定" },
];
