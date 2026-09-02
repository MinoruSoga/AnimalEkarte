import { memo, useActionState, useState, useCallback } from "react";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { EmptyState } from "@/components/shared/DataStates";
import { getFormString } from "@/lib/form-data";
import type { ClosingHoliday } from "../api/holidays";
import { useCreateHoliday, useDeleteHoliday } from "../api/holidays";

interface HolidaySectionProps {
  holidays: ClosingHoliday[];
}

export const HolidaySection = memo(function HolidaySection({ holidays }: HolidaySectionProps) {
  const [showForm, setShowForm] = useState(false);
  const createMutation = useCreateHoliday();
  const deleteMutation = useDeleteHoliday();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await createMutation.mutateAsync({
        date: getFormString(formData, "date"),
        reason: getFormString(formData, "reason") || undefined,
      });
      toast.success("休診日を追加しました");
      setShowForm(false);
    } catch {
      // FE-RC-005: useCreateHoliday.onError が既に handleApiError で通知済み。
    }
    return null;
  }, null);

  const handleDelete = useCallback(
    async (date: string) => {
      try {
        await deleteMutation.mutateAsync(date);
        toast.success("休診日を削除しました");
      } catch {
        // FE-RC-005: useDeleteHoliday.onError が既に handleApiError で通知済み。
      }
    },
    [deleteMutation],
  );

  const handleShowForm = useCallback(() => setShowForm(true), []);
  const handleHideForm = useCallback(() => setShowForm(false), []);

  return (
    <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className={`text-base font-semibold ${C.text}`}>個別休診日</h2>
        <button
          type="button"
          onClick={handleShowForm}
          className={`flex min-h-11 min-w-11 items-center gap-1.5 text-base ${C.textBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} rounded-xs px-3 transition-colors`}
        >
          <Plus className="size-4" />
          新規登録
        </button>
      </div>

      {showForm ? (
        <form action={formAction} className={`mb-4 p-4 rounded-lg border ${C.borderMedium} space-y-3`}>
          <p className={`text-base font-medium ${C.text}`}>新しい休診日</p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="holiday_date" className={STYLE.formLabel}>
                日付
              </label>
              <input
                id="holiday_date"
                name="date"
                type="date"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                required
              />
            </div>
            <div>
              <label htmlFor="holiday_reason" className={STYLE.formLabel}>
                理由・メモ
              </label>
              <input
                id="holiday_reason"
                name="reason"
                type="text"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                placeholder="例: 院内研修"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={handleHideForm}
              className={`min-h-11 px-4 text-base ${C.text60} ${C.hoverBgLight} rounded-xs transition-colors`}
            >
              キャンセル
            </button>
            <SubmitButton>追加</SubmitButton>
          </div>
        </form>
      ) : null}

      {holidays.length > 0 ? (
        <div className="space-y-2">
          {holidays.map((holiday) => (
            <div
              key={holiday.id}
              className={`flex items-center justify-between p-3 rounded-lg border ${C.borderLight} ${C.bgPage}`}
            >
              <div className="flex flex-col gap-0.5">
                <span className={`text-base font-medium ${C.text}`}>{holiday.date}</span>
                {holiday.reason ? (
                  <span className={`text-base ${C.text60}`}>{holiday.reason}</span>
                ) : null}
              </div>
              <button
                type="button"
                onClick={() => handleDelete(holiday.date)}
                aria-label={`${holiday.date}の休診日を削除`}
                className={`flex size-11 min-h-11 min-w-11 shrink-0 items-center justify-center rounded-xxs ${C.text50} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
              >
                <Trash2 className="size-4" />
              </button>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState message="個別休診日は登録されていません" />
      )}
    </section>
  );
});
