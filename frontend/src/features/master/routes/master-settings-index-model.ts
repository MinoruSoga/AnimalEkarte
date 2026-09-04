import type { LucideIcon } from "lucide-react";
import {
  Building2,
  Calendar,
  ClipboardList,
  Clock,
  CreditCard,
  FolderTree,
  Scissors,
  Stethoscope,
  Tag,
  FlaskConical,
  TestTube,
} from "lucide-react";
import { CATEGORY_CONFIG } from "../constants/category-config";
import type { MasterSettingsCategory } from "../constants/category-config";
import type { Resource } from "@/types/generated/models";
import {
  ResourceAccounting,
  ResourceClosingSettings,
  ResourceHospitalSettings,
  ResourceMasterMedical,
  ResourceMasterTrimming,
  ResourcePaymentMethod,
  ResourceShifts,
  ResourceLabImport,
} from "@/types/generated/models";

export type GroupKey =
  | "clinic"
  | "treatmentItems"
  | "diagnosisGroup"
  | "trimmingGroup"
  | "trimmingCourseTypes"
  | "inquiry_template"
  | "shift_template"
  | "paymentMethods"
  | "campaigns"
  | "closingTime"
  | "examinationItems"
  | "labDeviceItemMasters";

export type MasterCardKey = MasterSettingsCategory | GroupKey;

export interface SectionDef {
  title: string;
  keys: MasterCardKey[];
}

export interface GroupCardConfig {
  label: string;
  description: string;
  IconComponent: LucideIcon;
  path: string;
  resource: Resource;
  countCategories: MasterSettingsCategory[];
}

export const GROUP_CARD_CONFIG: Record<GroupKey, GroupCardConfig> = {
  clinic: {
    label: "医院マスタ",
    description: "院名、住所、電話番号などの医院基本情報を管理します",
    IconComponent: Building2,
    path: "/settings/clinic",
    resource: ResourceHospitalSettings,
    countCategories: [],
  },
  treatmentItems: {
    label: "診療項目マスタ",
    description: "診察・検査・処置・予防接種・定期健診の項目と単価(税込)を管理します",
    IconComponent: Stethoscope,
    path: "/settings/treatment-items",
    resource: ResourceMasterMedical,
    countCategories: ["consultation", "examination", "procedure", "vaccine", "checkup"],
  },
  diagnosisGroup: {
    label: "診断マスタ",
    description: "診断カテゴリと診断名を管理します",
    IconComponent: FolderTree,
    path: "/settings/diagnosis",
    resource: ResourceMasterMedical,
    countCategories: ["diagnosis_type", "diagnosis_name"],
  },
  trimmingGroup: {
    label: "トリミングマスタ",
    description: "トリミングコースとオプションを管理します",
    IconComponent: Scissors,
    path: "/settings/trimming",
    resource: ResourceMasterTrimming,
    countCategories: ["trimming_course", "trimming_option"],
  },
  trimmingCourseTypes: {
    label: "コース種別マスタ",
    description: "シャンプー・シャンプー＆カットなどコース種別を管理します",
    IconComponent: Scissors,
    path: "/settings/trimming-course-type",
    resource: ResourceMasterTrimming,
    countCategories: [],
  },
  inquiry_template: {
    label: "問診テンプレート",
    description: "問診票のテンプレートを管理します",
    IconComponent: ClipboardList,
    path: "/settings/inquiry-templates",
    resource: ResourceMasterMedical,
    countCategories: [],
  },
  shift_template: {
    label: "シフトテンプレートマスタ",
    description: "シフト登録で使用するテンプレートを管理します",
    IconComponent: Calendar,
    path: "/settings/shift-templates",
    resource: ResourceShifts,
    countCategories: [],
  },
  paymentMethods: {
    label: "支払方法マスタ",
    description: "現金・クレジットカードなど支払方法の種別を管理します",
    IconComponent: CreditCard,
    path: "/settings/payment-methods",
    resource: ResourcePaymentMethod,
    countCategories: [],
  },
  // #81: 割引キャンペーンは会計割引マスタのため ResourceAccounting（settings-routes campaigns と同権）
  campaigns: {
    label: "割引キャンペーンマスタ",
    description: "会計割引に使うキャンペーン期間と対象商品を管理します",
    IconComponent: Tag,
    path: "/settings/campaigns",
    resource: ResourceAccounting,
    countCategories: [],
  },
  closingTime: {
    label: "締め時間設定",
    description: "レジ締めのAM/PM境界・終了時刻・特別期間・休診日を設定します",
    IconComponent: Clock,
    path: "/settings/closing-time",
    // Align with settings-routes.tsx RequirePermission(ResourceClosingSettings)
    resource: ResourceClosingSettings,
    countCategories: [],
  },
  examinationItems: {
    label: "検査マスタ",
    description: "検査種別と測定項目・基準値を管理します。機器の対応表は検査機器マスタです",
    IconComponent: TestTube,
    path: "/settings/treatment-items?tab=examination",
    resource: ResourceMasterMedical,
    countCategories: [],
  },
  labDeviceItemMasters: {
    label: "検査機器マスタ",
    description: "NX600 / AU10V / 尿を検査へ対応付けます。日常の送信画面には出しません",
    IconComponent: FlaskConical,
    path: "/settings/lab-device-item-masters",
    resource: ResourceLabImport,
    countCategories: [],
  },
};

export const MASTER_SECTIONS: SectionDef[] = [
  { title: "基本設定", keys: ["clinic", "animal_species"] },
  {
    title: "カルテ",
    keys: [
      "treatmentItems",
      "diagnosisGroup",
      "inquiry_template",
      "chief_complaint",
      "medicine",
      "examinationItems",
      "labDeviceItemMasters",
    ],
  },
  { title: "予約管理マスタ", keys: ["reservationType"] },
  { title: "入院・ケージ管理", keys: ["hospitalization", "cage"] },
  { title: "トリミング関連", keys: ["trimmingGroup", "trimmingCourseTypes"] },
  {
    title: "会計・商品",
    keys: ["merchandise_item", "insurance", "paymentMethods", "campaigns", "closingTime"],
  },
  { title: "スタッフ・権限", keys: ["staff", "occupations", "permission_group"] },
  { title: "シフト管理", keys: ["shift_template"] },
];

export function isGroupCardKey(key: MasterCardKey): key is GroupKey {
  return key in GROUP_CARD_CONFIG;
}

export function getResourceForCardKey(key: MasterCardKey): Resource {
  if (isGroupCardKey(key)) {
    return GROUP_CARD_CONFIG[key].resource;
  }
  return CATEGORY_CONFIG[key].resource;
}
