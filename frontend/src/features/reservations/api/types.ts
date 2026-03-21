/**
 * 予約作成リクエスト
 * @see {@link import("@/types/generated/models").ReservationAppointment}
 */
export interface CreateReservationRequest {
  pet_id: number;
  owner_id: number;
  doctor_id?: number;
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type_id: number;
  is_designated?: boolean;
  notes?: string;
}

/**
 * 予約更新リクエスト
 * @see {@link import("@/types/generated/models").ReservationAppointment}
 */
export interface UpdateReservationRequest {
  start_time?: string;
  end_time?: string;
  visit_type?: string;
  service_type_id?: number;
  doctor_id?: string;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}
