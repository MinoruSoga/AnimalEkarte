import { ReactNode } from "react";
import type { Resource } from "./generated/models";

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
// ExamResult / ExaminationRecord は examination feature 固有の view model のため、
// examinations の feature index から直接 import すること（@/types で二重管理しない）。
export type { VaccinationRecord } from "@/features/vaccinations";
export type { TrimmingUI } from "@/features/trimming";
export type { InventoryItem } from "@/features/inventory";
// Feature domain types — sourced from transform ReturnType (FA9)
export type { Reservation } from "@/features/reservations";
export type { Hospitalization } from "@/features/hospitalization";
export type { MedicalRecord } from "@/features/medical-records";
import type { ReceptionAppointment } from "@/features/reception";
export type { ReceptionAppointment };


export interface MenuItem {
    icon?: ReactNode;
    label: string;
    path?: string;
    resource?: Resource; // 型安全な Resource 文字列を使用
    subItems?: MenuItem[];
}

// --- Reception（当日の受付）/ Calendar ---
// ReceptionAppointment は @/features/reception から re-export 済み

export interface ColumnData {
  title: string;
  appointments: ReceptionAppointment[];
}

// --- Common UI Components ---
export interface TreatmentPlan {
  id: string;
  treatmentContent: string;
  memo: string;
  is_insurance: boolean;
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
  "no_show",
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
  no_show: "未来院",
};


/** ナビゲーション遷移元情報（react-router location.state で使用） */
export interface NavigationState {
  from?: string | null;
}

// --- Feature UI Types ---
// Reservation は @/features/reservations から re-export 済み
// Hospitalization は @/features/hospitalization から re-export 済み

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

// MedicalRecord は @/features/medical-records から re-export 済み

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
  /** トリミングコース種別マスタ ID（#73・trimmingCourse 固有） */
  courseTypeId?: string;
}
