import { ReactNode } from "react";
import type { Resource, VisitType, ReservationStatus } from "./generated/models";

/**
 * 共有UI型定義 (Single Source of Truth)
 *
 * ⚠️ 基本的なエンティティ型は @/types/generated/models.ts を参照してください。
 * ここには UI 表示用や機能横断的な共通型のみを定義します。
 */

// --- Re-exports for convenience ---
export type { Owner } from "./owner";
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
// FE6-4: 生成型 @/types/generated/models.ts の TreatmentPlan（backend DTO・snake_case・
// 数値ID）と名前が衝突していた stale twin。UI-facing（camelCase・string ID）である実体を
// 明確にするため HospitalizationTreatmentPlan へリネーム（hospitalization feature 専用の
// 消費者のみ。master/* の TreatmentPlan* コンポーネント名は無関係のため触っていない）。
export interface HospitalizationTreatmentPlan {
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

// FE6-3: VisitType/ReservationStatus は tygo enum_style: "union"（FE6-1/FE6-2）で
// 生成定数が真の literal union になったため、生成型からの re-export へ移行した
// （import は本ファイル冒頭にまとめている）。drift テストは union-drift.test.ts から削除済み。
export type { VisitType, ReservationStatus };
export type CalendarView = "month" | "week";
export const CALENDAR_VIEW_VALUES = ["month", "week"] as const;

// RESERVATION_STATUS_VALUES は Select 等での選択肢イテレーション用のランタイム値配列。
// tygo は union 型のみを生成し値配列は生成しないため、明示的な ReservationStatus[] 注釈で
// 生成型の値域からの逸脱を type-check が検知できるようにしたうえで手書きのまま維持する。
export const RESERVATION_STATUS_VALUES: readonly ReservationStatus[] = [
  "confirmed",
  "pending",
  "checked_in",
  "in_consultation",
  "accounting",
  "completed",
  "cancelled",
  "no_show",
] as const;

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
