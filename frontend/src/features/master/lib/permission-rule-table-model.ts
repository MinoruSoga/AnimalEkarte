export interface PermissionRule {
  resource: string;
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

export type PermissionRuleField = keyof Omit<PermissionRule, "resource">;

export const RESOURCE_LABELS: Record<string, string> = {
  reception: "当日の受付",
  owners: "飼主・ペット",
  reservations: "予約管理",
  "medical-records": "カルテ",
  hospitalization: "入院・ホテル",
  trimming: "トリミング",
  examinations: "検査管理",
  "examination-unconfirm": "検査確定解除",
  "lab-import": "検査取込",
  accounting: "会計管理",
  "accounting-cancel": "会計キャンセル",
  "accounting-post-close-edit": "締め後会計編集",
  vaccinations: "予防接種",
  checkups: "定期健診",
  inventory: "在庫管理",
  estimates: "見積書",
  shifts: "シフト管理",
  "hospital-settings": "医院",
  "master-animal-species": "動物種類",
  "master-medical": "カルテ関連",
  "master-reservation-type": "予約区分",
  "master-hospitalization": "入院・ケージ",
  "master-trimming": "トリミングマスタ",
  "master-permission": "権限グループ",
  "master-staff": "スタッフ管理",
  "master-insurance": "保険マスタ",
  "master-merchandise": "物販・フード",
  discount: "割引",
  "cash-register-close": "レジ締め",
  "accounting-reports": "月次売上レポート",
  "closing-settings": "締め設定",
  "master-payment-method": "支払方法マスタ",
  "lstep-csv-import": "Lステップ CSV 取込",
  "lstep-analytics": "Lステップ分析",
  "manual-edit": "取扱説明書 編集",
  "identity-links": "同一飼主・ペット連携",
};

export const ALL_PERMISSION_RESOURCES = Object.keys(RESOURCE_LABELS);

export const PERMISSION_ACTION_COLUMNS: { field: PermissionRuleField; label: string }[] = [
  { field: "canView", label: "表示" },
  { field: "canCreate", label: "作成" },
  { field: "canEdit", label: "編集" },
  { field: "canDelete", label: "削除" },
];

export function createEmptyPermissionRule(resource: string): PermissionRule {
  return {
    resource,
    canView: false,
    canCreate: false,
    canEdit: false,
    canDelete: false,
  };
}

export function buildPermissionRuleMap(rules: PermissionRule[]) {
  return new Map(rules.map((rule) => [rule.resource, rule]));
}
