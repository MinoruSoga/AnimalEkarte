import type {
  LegacyLineCustomerResponse,
  LegacyLineReservationSettingResponse,
} from "./legacy-wire";
import type {
  BusinessHours,
  BreakHour,
  BusinessHoursByWeekday,
} from "../components/LineReservationSettingsFormSections";

// ── Response types（EMR-527: handler の JSON を写した移行 DTO） ──
export type LineCustomer = LegacyLineCustomerResponse;
// Backward-compat aliases
export type ReservationSetting = LegacyLineReservationSettingResponse;

// ── API Request types ──
/**
 * 保存 model・応答 DTO から Omit で導出せず、binding struct が受け取る列のうち
 * 設定フォーム / ページ編集が送るキーだけを明示列挙する（EMR-527）。
 * line_access_token は送らない（L-step 設定 API 側の責務）。
 * JSONB 5フィールドは json.RawMessage 相当の object/array として送る。
 * @see backend/internal/reservation/line_reservation_setting_request.go
 *   upsertLineReservationSettingRequest
 */
export interface UpdateLineReservationSettingRequest {
  status: string;
  header_text: string;
  reservation_notice: string;
  cancel_notice: string;
  privacy_policy: string;
  closed_weekdays: number[];
  closed_dates: string[];
  national_holiday_closed: boolean;
  business_hours: BusinessHours;
  business_hours_by_weekday?: BusinessHoursByWeekday;
  break_hours: BreakHour[];
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
}
