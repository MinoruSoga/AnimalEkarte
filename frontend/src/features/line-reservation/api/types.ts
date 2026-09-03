import type {
  LineReservationSetting,
  LineCustomer as LineCustomerModel,
} from "@/types/generated/models";
import type {
  BusinessHours,
  BreakHour,
  BusinessHoursByWeekday,
} from "../components/LineReservationSettingsFormSections";

// ── Re-export backend types ──
export type { LineCustomerModel as LineCustomer };
// Backward-compat aliases
export type ReservationSetting = LineReservationSetting;

// ── API Request types ──
// JSONB 5フィールドは tygo が `string /* []byte */` と宣言するが、実行時は
// json.RawMessage 相当の object/array なので実行時型で上書きする（generated models は編集しない）。
type LineReservationSettingJsonbFields =
  | "closed_weekdays"
  | "closed_dates"
  | "business_hours"
  | "business_hours_by_weekday"
  | "break_hours";

export type UpdateLineReservationSettingRequest = Omit<
  LineReservationSetting,
  "id" | "clinic_id" | "created_at" | "updated_at" | LineReservationSettingJsonbFields
> & {
  closed_weekdays: string[];
  closed_dates: string[];
  business_hours: BusinessHours;
  business_hours_by_weekday?: BusinessHoursByWeekday;
  break_hours: BreakHour[];
};
