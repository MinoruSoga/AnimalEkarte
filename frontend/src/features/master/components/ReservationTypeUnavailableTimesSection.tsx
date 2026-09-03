import { useActionState, useCallback, useMemo, useState } from "react";
import { Trash2, Plus, Clock } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";
import {
  useGetUnavailableTimes,
  useCreateUnavailableTime,
  useDeleteUnavailableTime,
} from "../api/reservation-type-unavailable-times";
import {
  UnavailableTypeWeekly,
  UnavailableTypeSpecific,
} from "@/types/generated/models";
import { DAY_OF_WEEK_SELECT_ITEMS } from "./DayOfWeekSelectItems";
import type { CreateUnavailableTimeRequest } from "../api/reservation-type-unavailable-times";

// ─────────────────────────────────────────────────
// 静的定数（rendering-hoist-jsx）
// ─────────────────────────────────────────────────

/** 30分刻みの時刻選択肢 "HH:MM" */
const TIME_OPTIONS: string[] = [];
for (let h = 0; h < 24; h++) {
  for (const m of [0, 30]) {
    TIME_OPTIONS.push(`${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`);
  }
}

const TIME_SELECT_ITEMS = TIME_OPTIONS.map((t) => (
  <SelectItem key={t} value={t}>{t}</SelectItem>
));

// ─────────────────────────────────────────────────
// Default form state
// ─────────────────────────────────────────────────

interface FormState {
  unavailableType: string;
  dayOfWeek: string;
  specificDate: string;
  startTime: string;
  endTime: string;
}

const DEFAULT_FORM: FormState = {
  unavailableType: UnavailableTypeWeekly,
  dayOfWeek: "1",
  specificDate: "",
  startTime: "09:00",
  endTime: "18:00",
};

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

interface Props {
  clinicId: string;
  reservationTypeId: string;
}

export function ReservationTypeUnavailableTimesSection({ clinicId, reservationTypeId }: Props) {
  const { data: items = [], isLoading } = useGetUnavailableTimes(clinicId, reservationTypeId);
  const createMutation = useCreateUnavailableTime(clinicId, reservationTypeId);
  const deleteMutation = useDeleteUnavailableTime(clinicId, reservationTypeId);
  const { mutate } = deleteMutation;

  const [form, setForm] = useState<FormState>(DEFAULT_FORM);

  const handleFieldChange = useCallback(<K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  }, []);

  const [, formAction] = useActionState(async () => {
    try {
      const req: CreateUnavailableTimeRequest = {
        unavailable_type: form.unavailableType,
        start_time: form.startTime,
        end_time: form.endTime,
        ...(form.unavailableType === UnavailableTypeWeekly
          ? { day_of_week: Number(form.dayOfWeek) }
          : { specific_date: form.specificDate }),
      };
      await createMutation.mutateAsync(req);
      setForm(DEFAULT_FORM);
    } catch {
      // エラー通知は useCreateUnavailableTime の onError に一本化（二重トースト防止）
    }
  }, null);

  const handleDelete = useCallback((id: number) => {
    mutate(id);
  }, [mutate]);

  const itemList = useMemo(() => items.map((item) => {
    const label = item.unavailableType === UnavailableTypeWeekly
      ? `毎週${DAY_OF_WEEK_LABELS[item.dayOfWeek ?? 0]}曜日`
      : item.specificDate ?? "";

    return (
      <div
        key={item.id}
        className={`flex items-center justify-between gap-2 py-1.5 px-2 rounded-xxs ${C.hoverBgLight} transition-colors group`}
      >
        <Clock className={`${ICON.smXs} ${C.text40} shrink-0`} />
        <span className={`flex-1 text-sm ${C.text}`}>{label}</span>
        <span className={`text-sm ${C.text50} tabular-nums`}>{item.startTime}〜{item.endTime}</span>
        <button
          type="button"
          onClick={() => handleDelete(item.id)}
          className={`opacity-0 group-hover:opacity-100 ${ICON.smXs} ${C.text40} ${C.hoverTextDanger} transition-colors`}
          aria-label="削除"
        >
          <Trash2 className={ICON.smXs} />
        </button>
      </div>
    );
  }), [items, handleDelete]);

  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <Clock className={ICON.smXs} style={{ color: C.text50 }} />
        <p className={`text-xs font-medium ${C.text50}`}>予約不可時間</p>
      </div>

      {/* 既存リスト */}
      {isLoading ? (
        <p className={`text-sm ${C.text40} py-2`}>読み込み中...</p>
      ) : items.length > 0 ? (
        <div className="mb-3 space-y-0.5">{itemList}</div>
      ) : null}

      {/* 追加フォーム */}
      <form action={formAction} className="space-y-2">
        {/* 種別 */}
        <div className="flex items-center gap-2">
          <Select
            value={form.unavailableType}
            onValueChange={(v) => handleFieldChange("unavailableType", v)}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={UnavailableTypeWeekly}>毎週</SelectItem>
              <SelectItem value={UnavailableTypeSpecific}>特定日</SelectItem>
            </SelectContent>
          </Select>

          {/* 曜日 or 日付 */}
          {form.unavailableType === UnavailableTypeWeekly ? (
            <Select
              value={form.dayOfWeek}
              onValueChange={(v) => handleFieldChange("dayOfWeek", v)}
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{DAY_OF_WEEK_SELECT_ITEMS}</SelectContent>
            </Select>
          ) : (
            <input
              type="date"
              aria-label="特定日"
              value={form.specificDate}
              onChange={(e) => handleFieldChange("specificDate", e.target.value)}
              className={`rounded-xxs border ${C.borderMedium} px-2 py-1 text-sm ${C.text} ${C.bgWhite}`}
            />
          )}
        </div>

        {/* 時間帯 */}
        <div className="flex items-center gap-2">
          <Select
            value={form.startTime}
            onValueChange={(v) => handleFieldChange("startTime", v)}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
          </Select>
          <span className={`text-sm ${C.text50}`}>〜</span>
          <Select
            value={form.endTime}
            onValueChange={(v) => handleFieldChange("endTime", v)}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
          </Select>
          <SubmitButton loadingText="追加中..." className="h-8 text-sm px-3">
            <Plus className={ICON.smXs} />
            追加
          </SubmitButton>
        </div>
      </form>
    </div>
  );
}
