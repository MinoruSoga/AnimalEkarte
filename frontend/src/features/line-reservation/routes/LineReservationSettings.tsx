import { useActionState, useState, useCallback } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C } from "@/lib/design-tokens";
import { useAuth } from "@/features/auth";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useGetReservationSetting } from "../api/get-reservation-setting";
import { updateReservationSetting } from "../api/update-reservation-setting";
import type { ReservationSetting } from "../api/types";

// ── Property Row ──

interface SectionProps {
  title: string;
  children: React.ReactNode;
}

function Section({ title, children }: SectionProps) {
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
    <div className="grid grid-cols-[200px_1fr] gap-4 items-center">
      <Label className={`text-sm ${C.text65}`}>{label}</Label>
      <div>{children}</div>
    </div>
  );
}

// ── Settings Form ──

interface SettingsFormProps {
  setting: ReservationSetting;
  clinicId: string;
}

function SettingsForm({ setting, clinicId }: SettingsFormProps) {
  const [status, setStatus] = useState(setting.status);
  const [nationalHolidayClosed, setNationalHolidayClosed] = useState(
    setting.national_holiday_closed
  );
  const [timeSlotMode, setTimeSlotMode] = useState(setting.time_slot_mode);
  const [noStaffMode, setNoStaffMode] = useState(setting.no_staff_mode);

  const handleStatusToggle = useCallback(() => {
    setStatus((prev) => (prev === "running" ? "stopped" : "running"));
  }, []);

  const [formState, formAction] = useActionState(
    async (_prev: null, formData: FormData) => {
      const payload = {
        status,
        national_holiday_closed: nationalHolidayClosed,
        time_slot_mode: timeSlotMode,
        no_staff_mode: noStaffMode,
        header_text: setting.header_text,
        reservation_notice: setting.reservation_notice,
        cancel_notice: setting.cancel_notice,
        privacy_policy: setting.privacy_policy,
        closed_weekdays: setting.closed_weekdays,
        closed_dates: setting.closed_dates,
        business_hours: setting.business_hours,
        break_hours: setting.break_hours,
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
        await updateReservationSetting(clinicId, payload);
        toast.success("設定を保存しました");
      } catch (err) {
        handleApiError(err, "設定保存");
      }
      return null;
    },
    null
  );

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
        <FieldRow label="祝日休診">
          <Switch
            id="national-holiday"
            checked={nationalHolidayClosed}
            onCheckedChange={setNationalHolidayClosed}
          />
        </FieldRow>
        <FieldRow label="タイムスロットモード">
          <Select value={timeSlotMode} onValueChange={setTimeSlotMode}>
            <SelectTrigger className="max-w-[240px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="minimize_gaps">空き時間を最小化</SelectItem>
              <SelectItem value="allow_gaps">空き時間を許容</SelectItem>
            </SelectContent>
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
            <SelectContent>
              <SelectItem value="first_available">最初の空き</SelectItem>
              <SelectItem value="top_priority">優先度最上位</SelectItem>
            </SelectContent>
          </Select>
        </FieldRow>
      </Section>

      {/* 連絡先 */}
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
        <FieldRow label="チャネルシークレット">
          <Input
            name="line_channel_secret"
            defaultValue={setting.line_channel_secret}
            className="max-w-[320px]"
            placeholder="LINE チャネルシークレット"
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
        <FieldRow label="アクセストークン">
          <Input
            name="line_access_token"
            defaultValue={setting.line_access_token}
            className="max-w-[320px]"
            placeholder="LINE Channel Access Token"
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

// ── Main Page ──

export function LineReservationSettings() {
  const { currentClinicId } = useAuth();
  const { data: setting, isLoading } = useGetReservationSetting(currentClinicId);

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
