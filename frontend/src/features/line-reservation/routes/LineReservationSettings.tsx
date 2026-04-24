// React/Framework
import { useActionState, useState, useCallback } from "react";

// External
import { toast } from "sonner";

// Shared modules
import { handleApiError } from "@/lib/handle-api-error";
import { C } from "@/lib/design-tokens";
import { useAuth } from "@/hooks/use-auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Feature-internal
import { useGetLineReservationSetting } from "../api/get-line-reservation-setting";
import { updateLineReservationSetting } from "../api/update-line-reservation-setting";
import type { ReservationSetting } from "../api/types";

// ── JSONB field types ─────────────────────────────────────────────────────────
// Go の []byte JSONB フィールドは tygo では `string` に生成されるが、
// axios が JSON を自動パースするため実行時はオブジェクト/配列として届く。

type BusinessHours = { start: string; end: string };
type BreakHour = { start: string; end: string };
type BusinessHoursByWeekday = Record<string, BusinessHours>;

/**
 * JSONB フィールドを期待する型にキャストする。
 * axios が JSON をパース済みなので実行時は文字列ではなくオブジェクト/配列が入っている。
 * fallback は null/undefined のときのみ使用する。
 */
function asJsonb<T>(value: unknown, fallback: T): T {
  if (value == null) return fallback;
  if (typeof value === "string") {
    try {
      return JSON.parse(value) as T;
    } catch {
      return fallback;
    }
  }
  return value as T;
}

/** "0900" → "09:00"（<input type="time"> の value 形式） */
function toDisplayTime(t: string): string {
  if (t.length === 4) return `${t.slice(0, 2)}:${t.slice(2)}`;
  return t;
}

/** "09:00" → "0900"（DB 保存形式） */
function toStorageTime(t: string): string {
  return t.replace(":", "");
}

// ── 静的定数 JSX（rendering-hoist-jsx） ──────────────────────────────────────

const TIME_SLOT_MODE_ITEMS = (
  <>
    <SelectItem value="minimize_gaps">空き時間を最小化</SelectItem>
    <SelectItem value="allow_gaps">空き時間を許容</SelectItem>
  </>
);

const NO_STAFF_MODE_ITEMS = (
  <>
    <SelectItem value="first_available">最初の空き</SelectItem>
    <SelectItem value="top_priority">優先度最上位</SelectItem>
  </>
);

const WEEKDAYS = [
  { value: "1", label: "月" },
  { value: "2", label: "火" },
  { value: "3", label: "水" },
  { value: "4", label: "木" },
  { value: "5", label: "金" },
  { value: "6", label: "土" },
  { value: "0", label: "日" },
] as const;

// ── 共通レイアウトコンポーネント ──────────────────────────────────────────────

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-3">
      <h2 className={`text-sm font-semibold ${C.text} border-b ${C.borderLight} pb-2`}>
        {title}
      </h2>
      <div className="space-y-4">{children}</div>
    </div>
  );
}

function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[200px_1fr] gap-4 items-start pt-1">
      <Label className={`text-sm ${C.text65} pt-1`}>{label}</Label>
      <div>{children}</div>
    </div>
  );
}

// ── BreakHoursEditor ──────────────────────────────────────────────────────────

interface BreakHoursEditorProps {
  value: BreakHour[];
  onChange: (v: BreakHour[]) => void;
}

function BreakHoursEditor({ value, onChange }: BreakHoursEditorProps) {
  const handleAdd = useCallback(() => {
    onChange([...value, { start: "1200", end: "1300" }]);
  }, [value, onChange]);

  const handleRemove = useCallback(
    (idx: number) => {
      onChange(value.filter((_, i) => i !== idx));
    },
    [value, onChange],
  );

  const handleChange = useCallback(
    (idx: number, field: "start" | "end", t: string) => {
      onChange(value.map((b, i) => (i === idx ? { ...b, [field]: toStorageTime(t) } : b)));
    },
    [value, onChange],
  );

  return (
    <div className="space-y-2">
      {value.map((b, idx) => (
        <div key={idx} className="flex items-center gap-2">
          <Input
            type="time"
            value={toDisplayTime(b.start)}
            onChange={(e) => handleChange(idx, "start", e.target.value)}
            className="max-w-[120px]"
          />
          <span className={`text-sm ${C.textMuted}`}>〜</span>
          <Input
            type="time"
            value={toDisplayTime(b.end)}
            onChange={(e) => handleChange(idx, "end", e.target.value)}
            className="max-w-[120px]"
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => handleRemove(idx)}
            className={`text-xs ${C.textMuted}`}
          >
            削除
          </Button>
        </div>
      ))}
      <Button type="button" variant="outline" size="sm" onClick={handleAdd}>
        + 追加
      </Button>
    </div>
  );
}

// ── ClosedDatesEditor ─────────────────────────────────────────────────────────

interface ClosedDatesEditorProps {
  value: string[];
  onChange: (v: string[]) => void;
}

function ClosedDatesEditor({ value, onChange }: ClosedDatesEditorProps) {
  const handleAdd = useCallback(() => {
    const today = new Date().toISOString().slice(0, 10);
    onChange([...value, today]);
  }, [value, onChange]);

  const handleRemove = useCallback(
    (idx: number) => {
      onChange(value.filter((_, i) => i !== idx));
    },
    [value, onChange],
  );

  const handleChange = useCallback(
    (idx: number, v: string) => {
      onChange(value.map((d, i) => (i === idx ? v : d)));
    },
    [value, onChange],
  );

  return (
    <div className="space-y-2">
      {value.length === 0 ? (
        <p className={`text-sm ${C.textMuted}`}>特定定休日は設定されていません</p>
      ) : (
        value.map((d, idx) => (
          <div key={idx} className="flex items-center gap-2">
            <Input
              type="date"
              value={d}
              onChange={(e) => handleChange(idx, e.target.value)}
              className="max-w-[160px]"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => handleRemove(idx)}
              className={`text-xs ${C.textMuted}`}
            >
              削除
            </Button>
          </div>
        ))
      )}
      <Button type="button" variant="outline" size="sm" onClick={handleAdd}>
        + 追加
      </Button>
    </div>
  );
}

// ── WeekdayHoursEditor ────────────────────────────────────────────────────────

interface WeekdayHoursEditorProps {
  defaultHours: BusinessHours;
  value: BusinessHoursByWeekday;
  onChange: (v: BusinessHoursByWeekday) => void;
}

function WeekdayHoursEditor({ defaultHours, value, onChange }: WeekdayHoursEditorProps) {
  const handleChange = useCallback(
    (day: string, field: "start" | "end", t: string) => {
      onChange({
        ...value,
        [day]: { ...(value[day] ?? defaultHours), [field]: toStorageTime(t) },
      });
    },
    [value, defaultHours, onChange],
  );

  return (
    <div className="space-y-2 pl-1">
      {WEEKDAYS.map(({ value: day, label }) => {
        const hours = value[day] ?? defaultHours;
        return (
          <div key={day} className="flex items-center gap-3">
            <span className={`text-sm w-6 text-center font-medium ${C.text65}`}>{label}</span>
            <Input
              type="time"
              value={toDisplayTime(hours.start)}
              onChange={(e) => handleChange(day, "start", e.target.value)}
              className="max-w-[120px]"
            />
            <span className={`text-sm ${C.textMuted}`}>〜</span>
            <Input
              type="time"
              value={toDisplayTime(hours.end)}
              onChange={(e) => handleChange(day, "end", e.target.value)}
              className="max-w-[120px]"
            />
          </div>
        );
      })}
    </div>
  );
}

// ── SettingsForm ──────────────────────────────────────────────────────────────

interface SettingsFormProps {
  setting: ReservationSetting;
  clinicId: string;
}

function SettingsForm({ setting, clinicId }: SettingsFormProps) {
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
        <SubmitButton>設定を保存</SubmitButton>
      </div>
      {formState !== undefined ? null : null}
    </form>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function LineReservationSettings() {
  const { currentClinicId } = useAuth();
  const { data: setting, isLoading } = useGetLineReservationSetting(currentClinicId);

  return (
    <PageLayout title="基本設定">
      {isLoading ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
      ) : setting ? (
        <SettingsForm setting={setting} clinicId={currentClinicId!} />
      ) : (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>
          設定データが見つかりません。
        </div>
      )}
    </PageLayout>
  );
}
