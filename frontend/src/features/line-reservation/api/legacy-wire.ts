/**
 * LINE予約設定・LINE顧客 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）、omitempty なしのポインタは null 可で表す。
 * JSONB 列（json.RawMessage）は unknown のまま受け、consumer 側の型ガードで解釈する。
 * 新しい LINE 予約契約の consumer を実装するときに、この adapter ごと置き換える（EMR-527）。
 */

/**
 * line_channel_secret / line_access_token はセキュリティ上レスポンスに含まれない。
 * @see backend/internal/reservation/line_reservation_setting_response.go
 *   lineReservationSettingResponse
 */
export interface LegacyLineReservationSettingResponse {
  id: number;
  clinic_id: number;
  status: string;
  header_text: string;
  reservation_notice: string;
  cancel_notice: string;
  privacy_policy: string;
  closed_weekdays: unknown;
  closed_dates: unknown;
  national_holiday_closed: boolean;
  business_hours: unknown;
  business_hours_by_weekday?: unknown;
  break_hours: unknown;
  daily_limit?: number;
  monthly_limit?: number;
  booking_window_max_days: number;
  booking_window_min_days: number;
  calendar_months: number;
  phone_number: string;
  notification_email: string;
  request_example: string;
  time_slot_mode: string;
  time_slot_interval_minutes: number;
  no_staff_mode: string;
  show_no_staff_option: boolean;
  additional_fields: unknown;
  line_channel_id: string;
  liff_id: string;
  created_at: string;
  updated_at: string;
}

/** @see backend/internal/lstep/line_customer_response.go lineCustomerResponse */
export interface LegacyLineCustomerResponse {
  id: number;
  clinic_id: number;
  line_user_id: string;
  display_name: string;
  real_name: string;
  additional_fields: unknown;
  owner_id?: number;
  owner_name?: string;
  created_at: string;
  updated_at: string;
}
