/**
 * Reception API types
 * Backend source: {@link import("@/types/generated/models").ReservationAppointment}
 * ReservationStatus は src/types の union 型を使用（models.ts は string 型のため型安全性維持）
 */
import type { ReservationStatus } from "@/types";
/** 当日の受付カンバンカード用の変換後型 */
export interface ReceptionAppointment {
  id: string;
  time: string; // "HH:mm" 形式
  ownerName: string;
  petType: string;
  petName: string;
  visitType: "初診" | "再診";
  serviceType: string;
  nextAppointment?: "次回予約無" | "次回予約済" | "精算未確認" | "精算確認済";
  isDesignated: boolean;
  doctor?: string;
  petId: string;
  ownerId: string;
  status: ReservationStatus;
  notes?: string;
  source?: "manual" | "line";
}

/** 当日の受付カンバンカラム */
export interface ReceptionColumn {
  id: ReservationStatus;
  title: string;
  appointments: ReceptionAppointment[];
}

/** ステータス更新リクエスト */
export interface UpdateAppointmentStatusRequest {
  status: ReservationStatus;
}
