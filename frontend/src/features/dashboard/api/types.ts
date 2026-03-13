import type { ReservationStatus } from "@/types";
/** Dashboard カンバンカード用の変換後型 */
export interface DashboardAppointment {
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
}

/** Dashboard カンバンカラム */
export interface DashboardColumn {
  id: ReservationStatus;
  title: string;
  appointments: DashboardAppointment[];
}

/** ステータス更新リクエスト */
export interface UpdateAppointmentStatusRequest {
  status: ReservationStatus;
}
