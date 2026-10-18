// React/Framework
import { useActionState, useLayoutEffect, useRef, useState, useCallback } from "react";
import { Link } from "react-router";

// External
import { toast } from "sonner";

// Shared modules
import { handleApiError } from "@/lib/handle-api-error";
import { getFormString } from "@/lib/form-data";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";

import { updateLineReservationSetting } from "../api/update-line-reservation-setting";
import type { ReservationSetting } from "../api/types";
import {
  BreakHoursEditor,
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
import {
  asJsonb,
  parseClosedWeekdays,
  sameClosedWeekdays,
  isStringArray,
  isBusinessHours,
  isBreakHourArray,
  isBusinessHoursByWeekday,
  toDisplayTime,
  toStorageTime,
} from "../lib/line-reservation-settings-form-model";

// ── SettingsForm ──────────────────────────────────────────────────────────────

interface SettingsFormProps {
  setting: ReservationSetting;
  clinicId: string;
}

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const INVALID_CLOSED_WEEKDAYS_MESSAGE =
  "定休日の設定を読み込めません。内容を確認して修正してください。";

export function LineReservationSettingsForm({ setting, clinicId }: SettingsFormProps) {
  const { canEdit } = usePermission("reservations");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  // スカラーフィールド
  const [status, setStatus] = useState(setting.status);
  const [nationalHolidayClosed, setNationalHolidayClosed] = useState(
    setting.national_holiday_closed,
  );
  const [timeSlotMode, setTimeSlotMode] = useState(setting.time_slot_mode);
  const [noStaffMode, setNoStaffMode] = useState(setting.no_staff_mode);
  // BUG-028: defaultValue のままだと form action 完了後に初期 props へリセットされ、
  // 保存済みの新値（0 含む）が UI に残らない。controlled で保持する。
  const [bookingWindowMinDays, setBookingWindowMinDays] = useState(setting.booking_window_min_days);
  const [bookingWindowMaxDays, setBookingWindowMaxDays] = useState<number | string>(
    setting.booking_window_max_days,
  );
  const [calendarMonths, setCalendarMonths] = useState<number | string>(setting.calendar_months);
  const [timeSlotIntervalMinutes, setTimeSlotIntervalMinutes] = useState<number | string>(
    setting.time_slot_interval_minutes,
  );

  // JSONB フィールド（lazy init — 初回レンダー時のみパース）
  const [closedWeekdays, setClosedWeekdays] = useState<number[] | null>(() =>
    parseClosedWeekdays(setting.closed_weekdays),
  );
  // 直近にサーバーから受け取った定休曜日。ユーザー編集とサーバー更新を区別する基準値。
  const [closedWeekdaysBaseline, setClosedWeekdaysBaseline] = useState(closedWeekdays);
  // 編集中にサーバー側も更新された場合の最新値。undefined は競合なし（null は不正値との競合）。
  const [closedWeekdaysConflict, setClosedWeekdaysConflict] = useState<number[] | null | undefined>(
    undefined,
  );
  const hasInvalidClosedWeekdays = closedWeekdays === null;
  const hasClosedWeekdaysConflict = closedWeekdaysConflict !== undefined;

  // 再取得で setting.closed_weekdays が変わったときローカル状態へ反映する。
  // props変化時にsetStateで調整するReact推奨パターン。未編集なら自動同期し、
  // 編集中にサーバー側も変わった場合は編集を保持して保存を止める。
  const [prevRawClosedWeekdays, setPrevRawClosedWeekdays] = useState(setting.closed_weekdays);
  if (prevRawClosedWeekdays !== setting.closed_weekdays) {
    setPrevRawClosedWeekdays(setting.closed_weekdays);
    const incoming = parseClosedWeekdays(setting.closed_weekdays);
    if (!sameClosedWeekdays(incoming, closedWeekdaysBaseline)) {
      if (sameClosedWeekdays(closedWeekdays, closedWeekdaysBaseline)) {
        setClosedWeekdays(incoming);
      } else {
        setClosedWeekdaysConflict(incoming);
      }
      setClosedWeekdaysBaseline(incoming);
    }
  }

  const handleLoadLatestClosedWeekdays = useCallback(() => {
    if (closedWeekdaysConflict === undefined) return;
    setClosedWeekdays(closedWeekdaysConflict);
    setClosedWeekdaysBaseline(closedWeekdaysConflict);
    setClosedWeekdaysConflict(undefined);
  }, [closedWeekdaysConflict]);
  const [businessHours, setBusinessHours] = useState<BusinessHours>(() =>
    asJsonb<BusinessHours>(setting.business_hours, { start: "0900", end: "1900" }, isBusinessHours),
  );
  const [breakHours, setBreakHours] = useState<BreakHour[]>(() =>
    asJsonb<BreakHour[]>(setting.break_hours, [], isBreakHourArray),
  );
  // 個別定休日の write owner は ClinicHolidayModal（clinic_holidays）。
  // このフォームは既存 closed_dates を PUT で round-trip するだけにし、編集 UI は置かない。
  const closedDates = asJsonb<string[]>(setting.closed_dates, [], isStringArray);
  const [enableWeekdayHours, setEnableWeekdayHours] = useState(() => {
    const parsed = asJsonb<BusinessHoursByWeekday>(
      setting.business_hours_by_weekday,
      {},
      isBusinessHoursByWeekday,
    );
    return Object.keys(parsed).length > 0;
  });
  const [weekdayHours, setWeekdayHours] = useState<BusinessHoursByWeekday>(() =>
    asJsonb<BusinessHoursByWeekday>(
      setting.business_hours_by_weekday,
      {},
      isBusinessHoursByWeekday,
    ),
  );

  // ハンドラ（useCallback で安定化）
  const handleStatusToggle = useCallback(() => {
    setStatus((prev) => (prev === "running" ? "stopped" : "running"));
  }, []);

  const handleWeekdayToggle = useCallback((day: string, checked: boolean) => {
    const numericDay = Number(day);
    setClosedWeekdays((prev) => {
      if (prev === null) return null;
      return checked ? [...prev, numericDay] : prev.filter((weekday) => weekday !== numericDay);
    });
  }, []);

  const handleBusinessHoursChange = useCallback((field: "start" | "end", t: string) => {
    setBusinessHours((prev) => ({ ...prev, [field]: toStorageTime(t) }));
  }, []);

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    if (canEditRef.current !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return null;
    }
    if (closedWeekdays === null || closedWeekdaysConflict !== undefined) return null;
    const payload = {
      status,
      national_holiday_closed: nationalHolidayClosed,
      time_slot_mode: timeSlotMode,
      no_staff_mode: noStaffMode,
      header_text: setting.header_text,
      reservation_notice: setting.reservation_notice,
      cancel_notice: setting.cancel_notice,
      privacy_policy: setting.privacy_policy,
      // JSONB: 実行時はオブジェクト/配列。tygo の型定義は `string` だが
      // UpdateLineReservationSettingRequest 側で実行時型に上書きしているためキャスト不要
      closed_weekdays: closedWeekdays,
      closed_dates: closedDates,
      business_hours: businessHours,
      business_hours_by_weekday: enableWeekdayHours ? weekdayHours : {},
      break_hours: breakHours,
      daily_limit: setting.daily_limit,
      monthly_limit: setting.monthly_limit,
      booking_window_max_days: Number(formData.get("booking_window_max_days")),
      booking_window_min_days: bookingWindowMinDays,
      calendar_months: Number(formData.get("calendar_months")),
      phone_number: getFormString(formData, "phone_number"),
      notification_email: getFormString(formData, "notification_email"),
      request_example: setting.request_example,
      time_slot_interval_minutes: Number(formData.get("time_slot_interval_minutes")),
      show_no_staff_option: setting.show_no_staff_option,
      additional_fields: setting.additional_fields,
      line_channel_id: getFormString(formData, "line_channel_id"),
      liff_id: getFormString(formData, "liff_id"),
    };
    try {
      const updated = await updateLineReservationSetting(clinicId, payload);
      // 0 は falsy だが正当値。レスポンス優先、無ければ送信値で UI を同期
      setBookingWindowMinDays(
        typeof updated?.booking_window_min_days === "number"
          ? updated.booking_window_min_days
          : bookingWindowMinDays,
      );
      // 保存値を新しい baseline にし、直後の再取得を競合として扱わない
      const savedWeekdays =
        updated?.closed_weekdays === undefined
          ? closedWeekdays
          : parseClosedWeekdays(updated.closed_weekdays);
      setClosedWeekdays(savedWeekdays);
      setClosedWeekdaysBaseline(savedWeekdays);
      setClosedWeekdaysConflict(undefined);
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
              aria-label="LINE予約受付"
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
                  touchTarget
                  disabled={hasInvalidClosedWeekdays}
                  checked={closedWeekdays?.includes(Number(day)) ?? false}
                  onCheckedChange={(checked) => handleWeekdayToggle(day, checked === true)}
                />
                <Label htmlFor={`closed-weekday-${day}`} className={`${C.text}`}>
                  {label}
                </Label>
              </div>
            ))}
            {hasInvalidClosedWeekdays ? (
              <p role="alert" className={`text-sm ${C.danger}`}>
                {INVALID_CLOSED_WEEKDAYS_MESSAGE}
              </p>
            ) : null}
            {hasClosedWeekdaysConflict ? (
              <div className="flex w-full flex-col items-start gap-2">
                <p role="alert" className={`text-sm ${C.danger}`}>
                  定休曜日がサーバー上で更新されています。未保存の変更を上書きしないよう保存を停止しています。
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleLoadLatestClosedWeekdays}
                >
                  最新の定休曜日を読み込む
                </Button>
              </div>
            ) : null}
          </div>
        </FieldRow>

        <FieldRow label="祝日休診">
          <Switch
            id="national-holiday"
            aria-label="祝日休診"
            checked={nationalHolidayClosed}
            onCheckedChange={setNationalHolidayClosed}
          />
        </FieldRow>

        <FieldRow label="特定定休日">
          <p className={`text-sm ${C.textMuted}`}>
            個別の定休日は
            <Link
              to={paths.shifts.getHref()}
              className={`inline-flex min-h-11 items-center ${C.textBrand} underline underline-offset-2`}
            >
              シフト管理
            </Link>
            のカレンダーから設定します。
          </p>
        </FieldRow>

        <FieldRow label="通常営業時間">
          <div className="flex items-center gap-2">
            <Input
              type="time"
              aria-label="通常営業時間 開始"
              value={toDisplayTime(businessHours.start)}
              onChange={(e) => handleBusinessHoursChange("start", e.target.value)}
              className="max-w-[120px]"
            />
            <span className={`text-sm ${C.textMuted}`}>〜</span>
            <Input
              type="time"
              aria-label="通常営業時間 終了"
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
                aria-label="曜日別営業時間"
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
          <NumberInput
            name="booking_window_min_days"
            aria-label="最短予約受付（日数）"
            min={0}
            value={bookingWindowMinDays}
            onChange={(value) => {
              const n = Number(value);
              setBookingWindowMinDays(Number.isNaN(n) ? 0 : n);
            }}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="最長予約受付（日前）">
          <NumberInput
            name="booking_window_max_days"
            aria-label="最長予約受付（日数）"
            min={1}
            value={bookingWindowMaxDays}
            onChange={setBookingWindowMaxDays}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="表示カレンダー月数">
          <NumberInput
            name="calendar_months"
            aria-label="表示カレンダー月数"
            min={1}
            max={6}
            value={calendarMonths}
            onChange={setCalendarMonths}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="タイムスロットモード">
          <Select value={timeSlotMode} onValueChange={setTimeSlotMode}>
            <SelectTrigger className="max-w-[240px]" aria-label="タイムスロットモード">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{TIME_SLOT_MODE_ITEMS}</SelectContent>
          </Select>
        </FieldRow>
        <FieldRow label="スロット間隔（分）">
          <NumberInput
            name="time_slot_interval_minutes"
            aria-label="スロット間隔（分）"
            min={5}
            step={5}
            value={timeSlotIntervalMinutes}
            onChange={setTimeSlotIntervalMinutes}
            className="max-w-[120px]"
          />
        </FieldRow>
        <FieldRow label="スタッフ不在モード">
          <Select value={noStaffMode} onValueChange={setNoStaffMode}>
            <SelectTrigger className="max-w-[240px]" aria-label="スタッフ不在モード">
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
            aria-label="電話番号"
            defaultValue={setting.phone_number}
            className="max-w-[240px]"
            placeholder="例: 03-1234-5678"
          />
        </FieldRow>
        <FieldRow label="通知メール">
          <Input
            name="notification_email"
            aria-label="通知メール"
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
            aria-label="チャネルID"
            defaultValue={setting.line_channel_id}
            className="max-w-[320px]"
            placeholder="LINE チャネルID"
          />
        </FieldRow>
        <FieldRow label="LIFF ID">
          <Input
            name="liff_id"
            aria-label="LIFF ID"
            defaultValue={setting.liff_id}
            className="max-w-[320px]"
            placeholder="LIFF ID"
          />
        </FieldRow>
      </Section>

      <div className="pt-2">
        <SubmitButton disabled={hasInvalidClosedWeekdays || hasClosedWeekdaysConflict}>
          保存
        </SubmitButton>
      </div>
    </form>
  );
}
