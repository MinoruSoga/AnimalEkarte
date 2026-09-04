import { useActionState, useLayoutEffect, useRef, useState, useCallback } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { C } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useGetLineReservationSetting } from "../api/get-line-reservation-setting";
import { updateLineReservationSetting } from "../api/update-line-reservation-setting";
import type { ReservationSetting } from "../api/types";
import {
  asJsonb,
  isStringArray,
  isBusinessHours,
  isBreakHourArray,
  isBusinessHoursByWeekday,
} from "../lib/line-reservation-settings-form-model";
import type {
  BusinessHours,
  BreakHour,
  BusinessHoursByWeekday,
} from "../components/LineReservationSettingsFormSections";

// ── Page Editor Form ──

interface PageEditorFormProps {
  setting: ReservationSetting;
  clinicId: string;
}

const FIELDS: Array<{
  key: keyof ReservationSetting;
  label: string;
  placeholder: string;
  rows: number;
}> = [
  {
    key: "header_text",
    label: "ヘッダーテキスト",
    placeholder: "LINE予約ページのヘッダーに表示するテキスト",
    rows: 3,
  },
  {
    key: "reservation_notice",
    label: "予約時の注意事項",
    placeholder: "予約時に顧客に表示する注意事項",
    rows: 5,
  },
  {
    key: "cancel_notice",
    label: "キャンセル時の注意事項",
    placeholder: "キャンセル時に表示する注意事項",
    rows: 4,
  },
  {
    key: "privacy_policy",
    label: "プライバシーポリシー",
    placeholder: "個人情報の取り扱いに関する説明",
    rows: 6,
  },
  {
    key: "request_example",
    label: "リクエスト例",
    placeholder: "予約リクエストの記入例",
    rows: 3,
  },
];

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

function PageEditorForm({ setting, clinicId }: PageEditorFormProps) {
  const { canEdit } = usePermission("reservations");
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  const [fieldValues, setFieldValues] = useState<Partial<ReservationSetting>>({});

  const handleChange = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const key = e.target.name as keyof ReservationSetting;
    const value = e.target.value;
    setFieldValues((prev) => ({ ...prev, [key]: value }));
  }, []);

  const [formState, formAction] = useActionState(async (_prev: null) => {
    if (canEditRef.current !== true) {
      toast.error(PERMISSION_DENIED_MESSAGE);
      return null;
    }
    const merged = { ...setting, ...fieldValues };
    try {
      await updateLineReservationSetting(clinicId, {
        status: merged.status,
        header_text: merged.header_text,
        reservation_notice: merged.reservation_notice,
        cancel_notice: merged.cancel_notice,
        privacy_policy: merged.privacy_policy,
        closed_weekdays: asJsonb<string[]>(merged.closed_weekdays, [], isStringArray),
        closed_dates: asJsonb<string[]>(merged.closed_dates, [], isStringArray),
        national_holiday_closed: merged.national_holiday_closed,
        business_hours: asJsonb<BusinessHours>(
          merged.business_hours,
          { start: "0900", end: "1900" },
          isBusinessHours,
        ),
        // optional フィールド: 未設定(null/undefined)はそのまま送出しない(PUT時にキー省略、挙動保存)
        business_hours_by_weekday:
          merged.business_hours_by_weekday == null
            ? undefined
            : asJsonb<BusinessHoursByWeekday>(
                merged.business_hours_by_weekday,
                {},
                isBusinessHoursByWeekday,
              ),
        break_hours: asJsonb<BreakHour[]>(merged.break_hours, [], isBreakHourArray),
        daily_limit: merged.daily_limit,
        monthly_limit: merged.monthly_limit,
        booking_window_max_days: merged.booking_window_max_days,
        booking_window_min_days: merged.booking_window_min_days,
        calendar_months: merged.calendar_months,
        phone_number: merged.phone_number,
        notification_email: merged.notification_email,
        request_example: merged.request_example,
        time_slot_mode: merged.time_slot_mode,
        time_slot_interval_minutes: merged.time_slot_interval_minutes,
        no_staff_mode: merged.no_staff_mode,
        show_no_staff_option: merged.show_no_staff_option,
        additional_fields: merged.additional_fields,
        line_channel_id: merged.line_channel_id,
        liff_id: merged.liff_id,
      });
      toast.success("ページ内容を保存しました");
    } catch (err) {
      handleApiError(err, "ページ編集保存");
    }
    return null;
  }, null);

  return (
    <form action={formAction} className="space-y-6 max-w-[720px]">
      {FIELDS.map((field) => (
        <div key={field.key} className="space-y-1.5">
          <Label htmlFor={`field-${field.key}`} className={`text-sm font-medium ${C.text}`}>
            {field.label}
          </Label>
          <Textarea
            id={`field-${field.key}`}
            name={field.key}
            rows={field.rows}
            placeholder={field.placeholder}
            defaultValue={String(setting[field.key] ?? "")}
            onChange={handleChange}
            className={`resize-y text-sm ${C.text}`}
          />
        </div>
      ))}

      <div className="pt-2">
        <SubmitButton>変更を保存</SubmitButton>
      </div>
      {formState !== undefined ? null : null}
    </form>
  );
}

// ── Main Page ──

export function LineReservationPageEditor() {
  const { currentClinicId } = useAuth();
  const { data: setting, isLoading } = useGetLineReservationSetting(currentClinicId);

  return (
    <PageLayout title="ページ編集">
      {isLoading ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
      ) : setting ? (
        <PageEditorForm setting={setting} clinicId={currentClinicId!} />
      ) : (
        <EmptyState message="設定データが見つかりません。" />
      )}
    </PageLayout>
  );
}
