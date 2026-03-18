/**
 * 予約作成リクエスト
 * @see {@link import("@/types/generated/models").ReservationAppointment}
 */
export interface CreateReservationRequest {
  pet_id: string;
  owner_id: string;
  doctor_id?: string;
  start_time: string;
  end_time: string;
  visit_type: string;
  service_type: string;
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
  service_type?: string;
  doctor_id?: string;
  is_designated?: boolean;
  status?: string;
  notes?: string;
}
