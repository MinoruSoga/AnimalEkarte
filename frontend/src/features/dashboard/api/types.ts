// Backend Reservation を Dashboard で使用するための型定義
import type { ReservationStatus } from "@/types";

/** Backend から返ってくる Reservation の raw 型（snake_case） */
export interface BackendDashboardReservation {
  id: string;
  pet_id: string;
  owner_id: string;
  doctor_id?: string;
  start_time: string;
  end_time: string;
  visit_type: string; // "first" | "revisit"
  service_type: string;
  is_designated: boolean;
  status: string; // "pending" | "confirmed" | "checked_in" | "in_consultation" | "accounting" | "completed" | "canceled"
  notes?: string;
  created_at: string;
  updated_at: string;

  // Preload済みのリレーション
  pet?: BackendPet;
  owner?: BackendOwner;
}

export interface BackendPet {
  id: string;
  name: string;
  species: string;
  breed?: string;
  gender?: string;
}

export interface BackendOwner {
  id: string;
  name: string;
  phone?: string;
}

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
