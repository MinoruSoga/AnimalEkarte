import { ReactNode } from "react";
import type { 
  Staff, 
  Insurance, 
  TrimmingCourse as BackendTrimmingCourse, 
  TrimmingOption as BackendTrimmingOption,
  ExaminationType as BackendExaminationType,
  InventoryItem as BackendInventoryItem,
  Resource
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

// --- Dashboard / Calendar ---
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

// --- Specialized Master Items (UI Aliases) ---
export type StaffMember = Staff;
export type InsuranceCompany = Insurance;
export type TrimmingCourse = BackendTrimmingCourse;
export type TrimmingOption = BackendTrimmingOption;
export type ExaminationType = BackendExaminationType;
export type InventoryItem = BackendInventoryItem;

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

export const STAFF_ROLE_LABELS: Record<string, string> = {
  veterinarian: '獣医師',
  nurse: '看護師',
  trimmer: 'トリマー',
  reception: '受付',
  manager: '管理職',
};
