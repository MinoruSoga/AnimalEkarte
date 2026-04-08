import { ReactNode } from "react";
import type {
  Staff,
  Insurance,
  TrimmingCourse as BackendTrimmingCourse,
  TrimmingOption as BackendTrimmingOption,
  ExaminationType as BackendExaminationType,
  InventoryItem as BackendInventoryItem,
  Resource,
  Hospitalization as BackendHospitalization,
  CarePlanItem as BackendCarePlanItem,
  DailyRecord as BackendDailyRecord,
  MedicalRecord as BackendMedicalRecord,
} from "./generated/models";

/**
 * 共有UI型定義 (Single Source of Truth)
 * 
 * ⚠️ 基本的なエンティティ型は @/types/generated/models.ts を参照してください。
 * ここには UI 表示用や機能横断的な共通型のみを定義します。
 */

// --- Re-exports for convenience ---
export type { Owner, CreateOwnerRequest, UpdateOwnerRequest } from './owner';
export type { Pet } from "@/lib/transforms/pet";
export type { Medicine } from "@/lib/transforms/medicine";

// --- Clinic / App Shell ---
export interface ClinicInfo {
  name: string;
  postalCode: string;
  address: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
}

export interface MenuItem {
    icon?: ReactNode;
    label: string;
    path?: string;
    resource?: Resource; // 型安全な Resource 文字列を使用
    subItems?: MenuItem[];
}

// --- Reception（当日の受付）/ Calendar ---
export interface Appointment {
  id: string;
  time: string;
  ownerName: string;
  petType: string;
  petName: string;
  visitType: "初診" | "再診";
  serviceType: string;
  nextAppointment?: "次回予約無" | "次回予約済" | "精算未確認" | "精算確認済";
  isDesignated?: boolean;
  doctor?: string;
  petId?: string;
  ownerId?: string;
}

export interface ColumnData {
  title: string;
  appointments: Appointment[];
}

// --- Common UI Components ---
export interface TreatmentPlan {
  id: string;
  treatmentContent: string;
  memo: string;
  insurance: boolean;
  unitPrice: number;
  quantity: number;
  discount: number;
  discountAmount: number;
  subtotal: number;
}

// --- Status & Enums ---
export const SORT_ORDER_VALUES = ["desc", "asc"] as const;
export type SortOrder = (typeof SORT_ORDER_VALUES)[number];

export type VisitType = "first" | "revisit";
export type CalendarView = "month" | "week";
export const CALENDAR_VIEW_VALUES = ["month", "week"] as const;

export const RESERVATION_STATUS_VALUES = [
  "confirmed",
  "pending",
  "checked_in",
  "in_consultation",
  "accounting",
  "completed",
  "cancelled",
] as const;

export type ReservationStatus = (typeof RESERVATION_STATUS_VALUES)[number];

export const RESERVATION_STATUS_LABELS: Record<ReservationStatus, string> = {
  confirmed: "予約確定",
  pending: "仮予約",
  checked_in: "受付済",
  in_consultation: "診療中",
  accounting: "会計待ち",
  completed: "完了",
  cancelled: "キャンセル",
};

export const RESERVATION_TYPE_VALUES = ["診療", "検診", "手術", "トリミング", "ワクチン", "入院"] as const;

/** 予約種別 */
export type ReservationType = (typeof RESERVATION_TYPE_VALUES)[number];

/** ナビゲーション遷移元情報（react-router location.state で使用） */
export interface NavigationState {
  from?: string | null;
}

// --- Feature UI Types ---

/**
 * フロントエンド予約アポイントメント型（UI 表示用 - id:string, start/end:Date）
 * transforms.ts の変換結果として使用
 */
export interface ReservationAppointment {
  id: string;
  start: Date;
  end: Date;
  ownerName: string;
  petName: string;
  visitType: "first" | "revisit";
  type: string;
  serviceTypeId?: string;
  doctor: string;
  doctorId?: string;
  isDesignated: boolean;
  status: ReservationStatus;
  notes?: string;
  petId?: string;
}

/**
 * フロントエンド入院記録型（UI 表示用 - id:string, 日本語ステータス）
 * transforms.ts の変換結果として使用
 */
export interface Hospitalization {
  id: string;
  hospitalizationNo: string;
  ownerName: string;
  petName: string;
  species: string;
  hospitalizationType: "入院" | "ホテル";
  startDate: string;
  endDate: string;
  status: "入院中" | "退院済" | "予約" | "一時帰宅";
  cageId?: string;
  /** 退院後の会計連携で使用 */
  petId?: string;
  /** 臨床安全ガード: ペットが死亡済みの場合 true */
  petIsDeceased?: boolean;
}

/**
 * フロントエンドケアプラン項目型（UI 表示用 - camelCase フィールド）
 * CarePlanDialog, CarePlanItemRow 等で使用
 */
export type CarePlanItemType = "food" | "medicine" | "treatment" | "instruction" | "item";
export type CarePlanItemStatus = "active" | "completed" | "discontinued";
export type CarePlanTiming = "morning" | "noon" | "night";

export interface CarePlanItem {
  id: string;
  hospitalizationId: string;
  type: CarePlanItemType;
  name: string;
  description: string;
  timing: CarePlanTiming[];
  status: CarePlanItemStatus;
  notes: string;
  medicineId?: string | null;
  procedureId?: string | null;
  hospitalizationPlanId?: string | null;
  unitPrice?: number;
  masterId?: string | null;
  category?: string;
  sortOrder?: number;
  createdAt?: string;
  updatedAt?: string;
}

/** ケアログの種別 */
export type CareLogType = "food" | "medicine" | "treatment" | "other" | "excretion";

/**
 * フロントエンドデイリーレコード型（UI 表示用）
 * DailyRecord["vitals"], DailyRecord["careLogs"] アクセスのために必要
 */
export interface DailyRecord {
  id: string;
  hospitalizationId: string;
  date: string;
  vitals: Array<{
    id: string;
    time: string;
    temperature?: number;
    heartRate?: number;
    respirationRate?: number;
    weight?: number;
    notes?: string;
    staff?: string;
  }>;
  careLogs: Array<{
    id: string;
    time: string;
    type: CareLogType;
    status?: string;
    value?: string;
    notes?: string;
    staff?: string;
  }>;
  staffNotes: Array<{
    id: string;
    time: string;
    content: string;
    staff?: string;
  }>;
}

/**
 * フロントエンド電子カルテ型（UI 表示用）
 * transforms.ts の変換結果として使用
 */
export interface MedicalRecord {
  id: string;
  recordNo: string;
  date: string;
  ownerId?: string;
  ownerName: string;
  petId?: string;
  petName: string;
  species: string;
  chiefComplaint: string;
  doctor: string;
  status: "作成中" | "確定済";
  visitType?: string;
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
  surgeryNotes?: string;
  diagnosis?: string;
  treatment?: string;
  prescription?: string;
  notes?: string;
  accountingId?: string;
  visitCount?: number;
  version: number;
}

/**
 * フロントエンド検査項目型（UI 表示用）
 * transforms.ts の変換結果として使用
 */
export interface ExaminationItem {
  id: string;
  name: string;
  result: string;
  unit: string;
  referenceRange: string;
}

/**
 * フロントエンド検査記録型（UI 表示用）
 * transforms.ts の変換結果として使用
 */
export interface ExaminationRecord {
  id: string;
  date: string;
  ownerName: string;
  petName: string;
  petId?: string;
  testType: string;
  testTypeId: string;
  doctor: string;
  doctorId: string;
  status: "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定";
  resultSummary?: string;
  machine?: string;
  items?: ExaminationItem[];
}

/**
 * フロントエンドワクチン接種記録型（UI 表示用）
 * transforms.ts の変換結果として使用
 */
export interface VaccinationRecord {
  id: string;
  petId?: string;
  ownerName: string;
  petName: string;
  vaccineId: string;
  vaccineName: string;
  doctor: string;
  date: string;
  nextDate: string;
}

/**
 * フロントエンドトリミング記録型（UI 表示用）
 * transforms.ts の変換結果として使用
 */
export interface TrimmingUI {
  id: string;
  date: string;
  petId?: string;
  ownerId?: string;
  petNumber: string;
  petName: string;
  ownerName: string;
  species: string;
  weight: string;
  styleRequest: string;
  staff: string;
  status: "完了" | "予約" | "進行中";
  // Form fields
  staffId: string;
  courseId: string;
  optionIds: string[];
  bw: string;
  bwUnit: "Kg" | "g";
  bt: string;
  usedShampoo: string;
  usedRibbon: string;
  remarks: string;
}

// Re-export backend types for use in hospitalization types (type alias to distinguish)
export type { BackendHospitalization, BackendCarePlanItem, BackendDailyRecord, BackendMedicalRecord };

// --- Specialized Master Items (UI Aliases) ---
export type StaffMember = Staff;
export type InsuranceCompany = Insurance;
export type TrimmingCourse = BackendTrimmingCourse;
export type TrimmingOption = BackendTrimmingOption;
export type ExaminationType = BackendExaminationType;

/**
 * フロントエンド在庫品目型（UI 表示用 - id:string, camelCase フィールド名）
 * use-inventory.ts の transformInventoryItem が返す型
 */
export interface InventoryItem {
  id: string;
  name: string;
  category: BackendInventoryItem["category"];
  quantity: number;
  unit: string;
  minStockLevel: number;
  location?: string;
  expiryDate?: string;
  supplier?: string;
  lastRestocked?: string;
  status: BackendInventoryItem["status"];
}

export interface MasterItem {
  id: string | number;
  code?: string;
  name: string;
  category?: string;
  price: number;
  status: "active" | "inactive";
  description?: string;
  inventoryId?: string | number;
  defaultQuantity?: number;
  timeCondition?: string;
  duration?: number | null;
}

export type MasterCategory = "vaccine" | "serviceType" | "consultation" | "procedure" | "hospitalization" | "diagnosis_category" | "diagnosis_name" | "checkup";

