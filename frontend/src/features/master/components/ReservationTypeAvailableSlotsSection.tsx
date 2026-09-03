import { useActionState, useCallback, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { CalendarDays, Clock, Plus, Trash2 } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";
import {
  useGetAvailableSlots,
  useCreateAvailableSlot,
  useDeleteAvailableSlot,
} from "../api/reservation-type-available-slots";
import {
  AvailableSlotTypeWeekly,
  AvailableSlotTypeSpecific,
} from "@/types/generated/models";
import { TIME_SELECT_ITEMS } from "./AvailableSlotOptions";
import { DAY_OF_WEEK_SELECT_ITEMS } from "./DayOfWeekSelectItems";
import type { CreateAvailableSlotRequest } from "../api/reservation-type-available-slots";

interface FormState {
  availableType: string;
  dayOfWeek: string;
  specificDate: string;
  startTime: string;
}

const DEFAULT_FORM: FormState = {
  availableType: AvailableSlotTypeWeekly,
  dayOfWeek: "1",
  specificDate: "",
  startTime: "09:45",
};

interface Props {
  clinicId: string;
  reservationTypeId: string;
}

export function ReservationTypeAvailableSlotsSection({ clinicId, reservationTypeId }: Props) {
  const { data: items = [], isLoading } = useGetAvailableSlots(clinicId, reservationTypeId);
  const createMutation = useCreateAvailableSlot(clinicId, reservationTypeId);
  const deleteMutation = useDeleteAvailableSlot(clinicId, reservationTypeId);
  const { mutate } = deleteMutation;

  const [form, setForm] = useState<FormState>(DEFAULT_FORM);
  const navigate = useNavigate();

  const handleFieldChange = useCallback(<K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  }, []);

  const [, formAction] = useActionState(async () => {
    try {
      const req: CreateAvailableSlotRequest = {
        available_type: form.availableType,
        start_time: form.startTime,
        is_active: true,
        ...(form.availableType === AvailableSlotTypeWeekly
          ? { day_of_week: Number(form.dayOfWeek) }
          : { specific_date: form.specificDate }),
      };
      await createMutation.mutateAsync(req);
      setForm(DEFAULT_FORM);
    } catch {
      // エラー通知は useCreateAvailableSlot の onError に一本化（二重トースト防止）
    }
  }, null);

  const handleDelete = useCallback((id: number) => {
    mutate(id);
  }, [mutate]);

  const itemList = useMemo(() => items.map((item) => {
    const label = item.availableType === AvailableSlotTypeWeekly
      ? `毎週${DAY_OF_WEEK_LABELS[item.dayOfWeek ?? 0]}曜日`
      : item.specificDate ?? "";

    return (
      <div
        key={item.id}
        className={`flex items-center justify-between gap-2 py-1.5 px-2 rounded-xxs ${C.hoverBgLight} transition-colors group`}
      >
        <Clock className={`${ICON.smXs} ${C.text40} shrink-0`} />
        <span className={`flex-1 text-sm ${C.text}`}>{label}</span>
        <span className={`text-sm ${C.text50} tabular-nums`}>{item.startTime}</span>
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
        <Clock className={`${ICON.smXs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>予約可能枠</p>
        <button
          type="button"
          onClick={() => navigate(`${paths.lineReservation.slots.getHref()}?typeId=${reservationTypeId}`)}
          className={`ml-auto flex items-center gap-1 text-xs ${C.text50} ${C.hoverTextBrand} transition-colors`}
        >
          <CalendarDays className={ICON.smXs} />
          カレンダーで編集
        </button>
      </div>

      {isLoading ? (
        <p className={`text-sm ${C.text40} py-2`}>読み込み中...</p>
      ) : items.length > 0 ? (
        <div className="mb-3 space-y-0.5">{itemList}</div>
      ) : (
        <p className={`text-xs ${C.text40} mb-3`}>未設定の場合は営業時間内の空き枠を使用します</p>
      )}

      <form action={formAction} className="space-y-2">
        <div className="flex items-center gap-2">
          <Select
            value={form.availableType}
            onValueChange={(value) => handleFieldChange("availableType", value)}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={AvailableSlotTypeWeekly}>毎週</SelectItem>
              <SelectItem value={AvailableSlotTypeSpecific}>特定日</SelectItem>
            </SelectContent>
          </Select>

          {form.availableType === AvailableSlotTypeWeekly ? (
            <Select
              value={form.dayOfWeek}
              onValueChange={(value) => handleFieldChange("dayOfWeek", value)}
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
              onChange={(event) => handleFieldChange("specificDate", event.target.value)}
              className={`rounded-xxs border ${C.borderMedium} px-2 py-1 text-sm ${C.text} ${C.bgWhite}`}
            />
          )}
        </div>

        <div className="flex items-center gap-2">
          <Select
            value={form.startTime}
            onValueChange={(value) => handleFieldChange("startTime", value)}
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
