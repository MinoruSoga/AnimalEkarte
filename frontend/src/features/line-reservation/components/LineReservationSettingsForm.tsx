// React/Framework
import { useActionState, useState, useCallback } from "react";

// External
import { toast } from "sonner";

// Shared modules
import { handleApiError } from "@/lib/handle-api-error";
import { C } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { updateLineReservationSetting } from "../api/update-line-reservation-setting";
import type { ReservationSetting } from "../api/types";
import {
  BreakHoursEditor,
  ClosedDatesEditor,
  FieldRow,
  NO_STAFF_MODE_ITEMS,
  Section,
  TIME_SLOT_MODE_ITEMS,
  WeekdayHoursEditor,
  WEEKDAYS,
  type BreakHour,
  type BusinessHours,
  type BusinessHoursByWeekday,
} from "./LineReservationSettingsFormSections";
import { asJsonb, toDisplayTime, toStorageTime } from "./LineReservationSettingsFormModel";

// ── SettingsForm ──────────────────────────────────────────────────────────────

interface SettingsFormProps {
  setting: ReservationSetting;
  clinicId: string;
}

export function LineReservationSettingsForm({ setting, clinicId }: SettingsFormProps) {
  // スカラーフィールド
  const [status, setStatus] = useState(setting.status);
  const [nationalHolidayClosed, setNationalHolidayClosed] = useState(
    setting.national_holiday_closed,
  );
  const [timeSlotMode, setTimeSlotMode] = useState(setting.time_slot_mode);
  const [noStaffMode, setNoStaffMode] = useState(setting.no_staff_mode);

  // JSONB フィールド（lazy init — 初回レンダー時のみパース）
  const [closedWeekdays, setClosedWeekdays] = useState<string[]>(
    () => asJsonb<string[]>(setting.closed_weekdays, []),
  );
  const [businessHours, setBusinessHours] = useState<BusinessHours>(
    () => asJsonb<BusinessHours>(setting.business_hours, { start: "0900", end: "1900" }),
  );
  const [breakHours, setBreakHours] = useState<BreakHour[]>(
    () => asJsonb<BreakHour[]>(setting.break_hours, []),
  );
  const [closedDates, setClosedDates] = useState<string[]>(
    () => asJsonb<string[]>(setting.closed_dates, []),
  );
  const [enableWeekdayHours, setEnableWeekdayHours] = useState(() => {
    const parsed = asJsonb<BusinessHoursByWeekday>(setting.business_hours_by_weekday, {});
    return Object.keys(parsed).length > 0;
  });
  const [weekdayHours, setWeekdayHours] = useState<BusinessHoursByWeekday>(
    () => asJsonb<BusinessHoursByWeekday>(setting.business_hours_by_weekday, {}),
  );

  // ハンドラ（useCallback で安定化）
  const handleStatusToggle = useCallback(() => {
    setStatus((prev) => (prev === "running" ? "stopped" : "running"));
  }, []);

  const handleWeekdayToggle = useCallback((day: string, checked: boolean) => {
    setClosedWeekdays((prev) => (checked ? [...prev, day] : prev.filter((d) => d !== day)));
  }, []);

  const handleBusinessHoursChange = useCallback((field: "start" | "end", t: string) => {
    setBusinessHours((prev) => ({ ...prev, [field]: toStorageTime(t) }));
  }, []);

  const [formState, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    const payload = {
      status,
      national_holiday_closed: nationalHolidayClosed,
      time_slot_mode: timeSlotMode,
      no_staff_mode: noStaffMode,
      header_text: setting.header_text,
      reservation_notice: setting.reservation_notice,
      cancel_notice: setting.cancel_notice,
      privacy_policy: setting.privacy_policy,
      // JSONB: 実行時はオブジェクト/配列。tygo の型定義は `string` だが axios が正しくシリアライズする
      closed_weekdays: closedWeekdays as unknown as string,
      closed_dates: closedDates as unknown as string,
      business_hours: businessHours as unknown as string,
      business_hours_by_weekday: (enableWeekdayHours ? weekdayHours : {}) as unknown as string,
      break_hours: breakHours as unknown as string,
      daily_limit: setting.daily_limit,
      monthly_limit: setting.monthly_limit,
      booking_window_max_days: Number(formData.get("booking_window_max_days")),
      booking_window_min_days: Number(formData.get("booking_window_min_days")),
      calendar_months: Number(formData.get("calendar_months")),
      phone_number: formData.get("phone_number") as string,
      notification_email: formData.get("notification_email") as string,
      request_example: setting.request_example,
      time_slot_interval_minutes: Number(formData.get("time_slot_interval_minutes")),
      show_no_staff_option: setting.show_no_staff_option,
      additional_fields: setting.additional_fields,
      line_channel_id: formData.get("line_channel_id") as string,
      line_channel_secret: formData.get("line_channel_secret") as string,
      liff_id: formData.get("liff_id") as string,
      line_access_token: formData.get("line_access_token") as string,
    };
    try {
      await updateLineReservationSetting(clinicId, payload);
      toast.success("設定を保存しました");
    } catch (err) {
      handleApiError(err, "設定保存");
    }
    return null;
  }, null);

  return (
    <form action={formAction} className="space-y-8 max-w-[720px]">
      {/* 稼働状態 */}
      <Section title="稼働状態">
        <FieldRow label="LINE予約受付">
          <div className="flex items-center gap-3">
            <Switch
              id="status-toggle"
              checked={status === "running"}
              onCheckedChange={handleStatusToggle}
            />
            <span className={`text-sm ${status === "running" ? C.textBrand : C.textMuted}`}>
              {status === "running" ? "受付中" : "停止中"}
            </span>
          </div>
        </FieldRow>
      </Section>

      {/* 営業時間・定休日 */}
      <Section title="営業時間・定休日">
        <FieldRow label="定休曜日">
          <div className="flex flex-wrap gap-x-4 gap-y-2">
            {WEEKDAYS.map(({ value: day, label }) => (
              <div key={day} className="flex items-center gap-1.5">
                <Checkbox
                  id={`closed-weekday-${day}`}
                  checked={closedWeekdays.includes(day)}
                  onCheckedChange={(checked) => handleWeekdayToggle(day, checked === true)}
                />
                <Label htmlFor={`closed-weekday-${day}`} className={`text-sm ${C.text}`}>
                  {label}
                </Label>
              </div>
            ))}
          </div>
        </FieldRow>

        <FieldRow label="祝日休診">
          <Switch
            id="national-holiday"
            checked={nationalHolidayClosed}
            onCheckedChange={setNationalHolidayClosed}
          />
        </FieldRow>

        <FieldRow label="特定定休日">
          <ClosedDatesEditor value={closedDates} onChange={setClosedDates} />
        </FieldRow>

        <FieldRow label="通常営業時間">
          <div className="flex items-center gap-2">
            <Input
              type="time"
              value={toDisplayTime(businessHours.start)}
              onChange={(e) => handleBusinessHoursChange("start", e.target.value)}
              className="max-w-[120px]"
            />
            <span className={`text-sm ${C.textMuted}`}>〜</span>
            <Input
              type="time"
              value={toDisplayTime(businessHours.end)}
              onChange={(e) => handleBusinessHoursChange("end", e.target.value)}
              className="max-w-[120px]"
            />
          </div>
        </FieldRow>

        <FieldRow label="曜日別営業時間">
          <div className="space-y-3">
            <div className="flex items-center gap-3">
              <Switch
                id="weekday-hours-toggle"
                checked={enableWeekdayHours}
                onCheckedChange={setEnableWeekdayHours}
              />
              <span className={`text-sm ${C.textMuted}`}>曜日ごとに設定する</span>
            </div>
            {enableWeekdayHours ? (
              <WeekdayHoursEditor
                defaultHours={businessHours}
                value={weekdayHours}
                onChange={setWeekdayHours}
              />
            ) : null}
          </div>
        </FieldRow>

        <FieldRow label="休憩時間">
          <BreakHoursEditor value={breakHours} onChange={setBreakHours} />
        </FieldRow>
      </Section>

      {/* 予約枠設定 */}
      <Section title="予約枠設定">
        <FieldRow label="最短予約受付（日前）">
          <Input
            name="booking_window_min_days"
            type="number"
            min={0}
            defaultValue={setting.booking_window_min_days}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="最長予約受付（日前）">
          <Input
            name="booking_window_max_days"
            type="number"
            min={1}
            defaultValue={setting.booking_window_max_days}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="表示カレンダー月数">
          <Input
            name="calendar_months"
            type="number"
            min={1}
            max={6}
            defaultValue={setting.calendar_months}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="タイムスロットモード">
          <Select value={timeSlotMode} onValueChange={setTimeSlotMode}>
            <SelectTrigger className="max-w-[240px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{TIME_SLOT_MODE_ITEMS}</SelectContent>
          </Select>
        </FieldRow>
        <FieldRow label="スロット間隔（分）">
          <Input
            name="time_slot_interval_minutes"
            type="number"
            min={5}
            step={5}
            defaultValue={setting.time_slot_interval_minutes}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="スタッフ不在モード">
          <Select value={noStaffMode} onValueChange={setNoStaffMode}>
            <SelectTrigger className="max-w-[240px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{NO_STAFF_MODE_ITEMS}</SelectContent>
          </Select>
        </FieldRow>
      </Section>

      {/* 連絡先・通知 */}
      <Section title="連絡先・通知">
        <FieldRow label="電話番号">
          <Input
            name="phone_number"
            defaultValue={setting.phone_number}
            className="max-w-[240px]"
            placeholder="例: 03-1234-5678"
          />
        </FieldRow>
        <FieldRow label="通知メール">
          <Input
            name="notification_email"
            type="email"
            defaultValue={setting.notification_email}
            className="max-w-[320px]"
            placeholder="例: info@clinic.example.com"
          />
        </FieldRow>
      </Section>

      {/* LINE連携 */}
      <Section title="LINE連携設定">
        <FieldRow label="チャネルID">
          <Input
            name="line_channel_id"
            defaultValue={setting.line_channel_id}
            className="max-w-[320px]"
            placeholder="LINE チャネルID"
          />
        </FieldRow>
        <FieldRow label="LIFF ID">
          <Input
            name="liff_id"
            defaultValue={setting.liff_id}
            className="max-w-[320px]"
            placeholder="LIFF ID"
          />
        </FieldRow>
      </Section>

      <div className="pt-2">
        <SubmitButton colorVariant="brand">設定を保存</SubmitButton>
      </div>
      {formState !== undefined ? null : null}
    </form>
  );
}
