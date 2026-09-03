import type { LucideIcon } from "lucide-react";
import TestTube from "lucide-react/dist/esm/icons/test-tube";
import Syringe from "lucide-react/dist/esm/icons/syringe";
import Pill from "lucide-react/dist/esm/icons/pill";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import Activity from "lucide-react/dist/esm/icons/activity";
import Bed from "lucide-react/dist/esm/icons/bed";
import Building2 from "lucide-react/dist/esm/icons/building-2";
import Users from "lucide-react/dist/esm/icons/users";
import ShieldCheck from "lucide-react/dist/esm/icons/shield-check";
import Lock from "lucide-react/dist/esm/icons/lock";
import Scissors from "lucide-react/dist/esm/icons/scissors";
import Sparkles from "lucide-react/dist/esm/icons/sparkles";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";
import FileText from "lucide-react/dist/esm/icons/file-text";
import ClipboardCheck from "lucide-react/dist/esm/icons/clipboard-check";
import PawPrint from "lucide-react/dist/esm/icons/paw-print";
import ShoppingBag from "lucide-react/dist/esm/icons/shopping-bag";
import MessageSquareText from "lucide-react/dist/esm/icons/message-square-text";

import type { Resource } from "@/types/generated/models";
import {
  ResourceMasterAnimalSpecies,
  ResourceMasterMedical,
  ResourceMasterReservationType,
  ResourceMasterHospitalization,
  ResourceMasterTrimming,
  ResourceMasterPermission,
  ResourceMasterStaff,
  ResourceMasterInsurance,
  ResourceMasterMerchandise,
  ResourceCheckups,
} from "@/types/generated/models";

// All master categories for the settings UI
// Note: some of these have their own backend tables, but all appear in settings UI
export type MasterSettingsCategory =
  | "examination"
  | "vaccine"
  | "medicine"
  | "consultation"
  | "reservationType"
  | "procedure"
  | "hospitalization"
  | "cage"
  | "trimming_course"
  | "trimming_option"
  | "staff"
  | "insurance"
  | "diagnosis_type"
  | "diagnosis_name"
  | "checkup"
  | "occupations"
  | "animal_species"
  | "inquiry_template"
  | "merchandise_item"
  | "permission_group"
  | "chief_complaint";

export interface CategoryConfig {
  label: string;
  description: string;
  settingsPath: string;
  IconComponent: LucideIcon;
  /** 権限チェック用リソース定数（BUG-123） */
  resource: Resource;
  labels: { code: string; name: string; category: string };
  showPrice: boolean;
  showCode: boolean;
  showCategory: boolean;
  showParentItem: boolean;
  namePlaceholder: string;
  codePlaceholder: string;
}

export const CATEGORY_CONFIG: Record<MasterSettingsCategory, CategoryConfig> = {
  examination: {
    label: "検査マスタ",
    description: "血液検査、レントゲン、超音波検査などを管理します",
    settingsPath: "/settings/treatment-items?tab=examination",
    IconComponent: TestTube,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "血液検査(CBC)",
    codePlaceholder: "",
  },
  vaccine: {
    label: "予防接種マスタ",
    description: "予防接種種別、接種料金、接種間隔を管理します",
    settingsPath: "/settings/treatment-items?tab=vaccine",
    IconComponent: Syringe,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "狂犬病ワクチン",
    codePlaceholder: "",
  },
  medicine: {
    label: "薬剤マスタ",
    description: "処方薬の名称、単価(税込)、薬効分類を管理します",
    settingsPath: "/settings/medicine",
    IconComponent: Pill,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "薬品名", category: "薬効分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "アモキシシリン",
    codePlaceholder: "",
  },
  consultation: {
    label: "診察マスタ",
    description: "初診料、再診料、時間外診察料などを管理します",
    settingsPath: "/settings/treatment-items?tab=consultation",
    IconComponent: Stethoscope,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "再診料",
    codePlaceholder: "",
  },
  reservationType: {
    label: "予約区分マスタ",
    description: "予約の区分（診療、トリミング入院等）を管理します",
    settingsPath: "/settings/reservation-type",
    IconComponent: Activity,
    resource: ResourceMasterReservationType,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "診療",
    codePlaceholder: "",
  },
  procedure: {
    label: "処置マスタ",
    description: "爪切り、耳掃除、肛門腺絞りなどを管理します",
    settingsPath: "/settings/treatment-items?tab=procedure",
    IconComponent: Activity,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "爪切り",
    codePlaceholder: "",
  },
  hospitalization: {
    label: "入院マスタ",
    description: "入院料金の体格別単価(税込)を管理します",
    settingsPath: "/settings/hospitalization",
    IconComponent: Bed,
    resource: ResourceMasterHospitalization,
    labels: { code: "コード", name: "名称", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "入院料(小型)",
    codePlaceholder: "",
  },
  cage: {
    label: "ケージマスタ",
    description: "ICU、犬舎、猫舎などのケージ情報を管理します",
    settingsPath: "/settings/cage",
    IconComponent: Building2,
    resource: ResourceMasterHospitalization,
    labels: { code: "コード", name: "ケージ名", category: "エリア" },
    showPrice: false,
    showCode: true,
    showCategory: true,
    showParentItem: false,
    namePlaceholder: "ICU-1",
    codePlaceholder: "ICU-01",
  },
  trimming_course: {
    label: "トリミングコースマスタ",
    description: "シャンプーコース、カットコースなどを管理します",
    settingsPath: "/settings/trimming?tab=course",
    IconComponent: Scissors,
    resource: ResourceMasterTrimming,
    labels: { code: "コード", name: "コース名", category: "対象サイズ" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "シャンプーコース(小型)",
    codePlaceholder: "",
  },
  trimming_option: {
    label: "トリミングオプションマスタ",
    description: "薬用シャンプー、炭酸泉、泥パックなどを管理します",
    settingsPath: "/settings/trimming?tab=option",
    IconComponent: Sparkles,
    resource: ResourceMasterTrimming,
    labels: { code: "コード", name: "オプション名", category: "種別" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "薬用シャンプー",
    codePlaceholder: "",
  },
  staff: {
    label: "スタッフマスタ",
    description: "獣医師、スタッフの情報を管理します",
    settingsPath: "/settings/staff",
    IconComponent: Users,
    resource: ResourceMasterStaff,
    labels: { code: "社員番号", name: "氏名", category: "職種" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "山田 太郎",
    codePlaceholder: "ST-001",
  },
  insurance: {
    label: "保険マスタ",
    description: "アニコム、アイペットなどの保険会社を管理します",
    settingsPath: "/settings/insurance",
    IconComponent: ShieldCheck,
    resource: ResourceMasterInsurance,
    labels: { code: "コード", name: "保険会社名", category: "種別" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "アニコム",
    codePlaceholder: "",
  },
  diagnosis_type: {
    label: "診断カテゴリマスタ",
    description: "消化器疾患、呼吸器疾患などの診断カテゴリを管理します",
    settingsPath: "/settings/diagnosis?tab=diagnosis_type",
    IconComponent: FolderTree,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "カテゴリ名", category: "分類" },
    showPrice: false,
    showCode: false,
    showCategory: true,
    showParentItem: false,
    namePlaceholder: "消化器疾患",
    codePlaceholder: "",
  },
  diagnosis_name: {
    label: "診断名マスタ",
    description: "カテゴリに紐づく病名（食道炎、膀胱炎等）を管理します",
    settingsPath: "/settings/diagnosis?tab=diagnosis_name",
    IconComponent: FileText,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "診断名", category: "分類" },
    showPrice: false,
    showCode: false,
    showCategory: true,
    showParentItem: false,
    namePlaceholder: "食道炎",
    codePlaceholder: "",
  },
  checkup: {
    label: "定期健診マスタ",
    description: "年次健康診断、シニア健診などの健診種別を管理します",
    settingsPath: "/settings/treatment-items",
    IconComponent: ClipboardCheck,
    resource: ResourceCheckups,
    labels: { code: "コード", name: "健診種別名", category: "分類" },
    showPrice: true,
    showCode: false,
    showCategory: false,
    showParentItem: true,
    namePlaceholder: "年次健康診断",
    codePlaceholder: "",
  },
  occupations: {
    label: "職種マスタ",
    description: "獣医師、動物看護師などの職種を管理します",
    settingsPath: "/settings/occupations",
    IconComponent: Users,
    resource: ResourceMasterStaff,
    labels: { code: "コード", name: "職種名", category: "分類" },
    showPrice: false,
    showCode: true,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "獣医師",
    codePlaceholder: "veterinarian",
  },
  animal_species: {
    label: "動物種類マスタ",
    description: "犬、猫、鳥などの動物種類を管理します",
    settingsPath: "/settings/animal-species",
    IconComponent: PawPrint,
    resource: ResourceMasterAnimalSpecies,
    labels: { code: "", name: "動物種類名", category: "" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "犬",
    codePlaceholder: "",
  },
  inquiry_template: {
    label: "問診テンプレートマスタ",
    description: "問診の質問項目テンプレートを管理します",
    settingsPath: "/settings/inquiry-templates",
    IconComponent: FileText,
    resource: ResourceMasterMedical,
    labels: { code: "コード", name: "テンプレート名", category: "分類" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "初診問診",
    codePlaceholder: "",
  },
  merchandise_item: {
    label: "商品マスタ",
    description: "フード、サプリメント、グッズ等の販売品目を管理します",
    settingsPath: "/settings/merchandise-items",
    IconComponent: ShoppingBag,
    resource: ResourceMasterMerchandise,
    labels: { code: "コード", name: "品目名", category: "カテゴリ" },
    showPrice: true,
    showCode: false,
    showCategory: true,
    showParentItem: false,
    namePlaceholder: "ロイヤルカナン",
    codePlaceholder: "",
  },
  permission_group: {
    label: "権限グループマスタ",
    description: "スタッフに付与する権限グループを管理します",
    settingsPath: "/settings/permission-groups",
    IconComponent: Lock,
    resource: ResourceMasterPermission,
    labels: { code: "コード", name: "グループ名", category: "分類" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "獣医師権限",
    codePlaceholder: "",
  },
  chief_complaint: {
    label: "主訴カテゴリ",
    description: "問診で使用する主訴カテゴリを管理します",
    settingsPath: "/settings/interview/chief-complaint",
    IconComponent: MessageSquareText,
    resource: ResourceMasterMedical,
    labels: { code: "", name: "カテゴリ名", category: "" },
    showPrice: false,
    showCode: false,
    showCategory: false,
    showParentItem: false,
    namePlaceholder: "消化器症状",
    codePlaceholder: "",
  },
};
