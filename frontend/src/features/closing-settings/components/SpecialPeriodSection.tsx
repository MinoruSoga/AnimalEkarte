import { memo, useActionState, useState, useCallback } from "react";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { EmptyState } from "@/components/shared/DataStates";
import { handleApiError } from "@/lib/handle-api-error";
import type { ClosingSpecialPeriod } from "@/types/generated/models";
import { useCreateSpecialPeriod, useDeleteSpecialPeriod } from "../api/special-periods";

interface SpecialPeriodSectionProps {
  periods: ClosingSpecialPeriod[];
}

export const SpecialPeriodSection = memo(function SpecialPeriodSection({
  periods,
}: SpecialPeriodSectionProps) {
  const [showForm, setShowForm] = useState(false);
  const createMutation = useCreateSpecialPeriod();
  const deleteMutation = useDeleteSpecialPeriod();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await createMutation.mutateAsync({
        start_date: formData.get("start_date") as string,
        end_date: formData.get("end_date") as string,
        am_pm_boundary: formData.get("am_pm_boundary") as string,
        pm_end: formData.get("pm_end") as string,
        note: (formData.get("note") as string) || undefined,
      });
      toast.success("特別期間を追加しました");
      setShowForm(false);
    } catch (error) {
      handleApiError(error, "特別期間の追加");
    }
    return null;
  }, null);

  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await deleteMutation.mutateAsync(id);
        toast.success("特別期間を削除しました");
      } catch (error) {
        handleApiError(error, "特別期間の削除");
      }
    },
    [deleteMutation],
  );

  const handleShowForm = useCallback(() => setShowForm(true), []);
  const handleHideForm = useCallback(() => setShowForm(false), []);

  return (
    <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className={`text-base font-semibold ${C.text}`}>特別期間</h2>
        <button
          type="button"
          onClick={handleShowForm}
          className={`flex min-h-11 min-w-11 items-center gap-1.5 text-base ${C.textBrand} ${C.hoverBgBrand} hover:text-white rounded-xs px-3 transition-colors`}
        >
          <Plus className="size-4" />
          追加
        </button>
      </div>

      {showForm ? (
        <form action={formAction} className={`mb-4 p-4 rounded-lg border ${C.borderMedium} space-y-3`}>
          <p className={`text-base font-medium ${C.text}`}>新しい特別期間</p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label htmlFor="start_date" className={STYLE.formLabel}>
                開始日
              </label>
              <input
                id="start_date"
                name="start_date"
                type="date"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                required
              />
            </div>
            <div>
              <label htmlFor="end_date" className={STYLE.formLabel}>
                終了日
              </label>
              <input
                id="end_date"
                name="end_date"
                type="date"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                required
              />
            </div>
            <div>
              <label htmlFor="am_pm_boundary" className={STYLE.formLabel}>
                午前・午後 区切り時間
              </label>
              <input
                id="am_pm_boundary"
                name="am_pm_boundary"
                type="time"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                required
              />
            </div>
            <div>
              <label htmlFor="pm_end" className={STYLE.formLabel}>
                午後 終了時間
              </label>
              <input
                id="pm_end"
                name="pm_end"
                type="time"
                className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
                required
              />
            </div>
          </div>
          <div>
            <label htmlFor="note" className={STYLE.formLabel}>
              メモ
            </label>
            <input
              id="note"
              name="note"
              type="text"
              className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
              placeholder="例: 年末年始"
            />
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

      {periods.length > 0 ? (
        <div className="space-y-2">
          {periods.map((period) => (
            <div
              key={period.id}
              className={`flex items-center justify-between p-3 rounded-lg border ${C.borderLight} ${C.bgPage}`}
            >
              <div className="flex flex-col gap-0.5">
                <span className={`text-base font-medium ${C.text}`}>
                  {period.start_date} 〜 {period.end_date}
                </span>
                <span className={`text-base ${C.text60}`}>
                  区切り: {period.am_pm_boundary} / 終了: {period.pm_end}
                  {period.note ? ` — ${period.note}` : ""}
                </span>
              </div>
              <button
                type="button"
                onClick={() => handleDelete(period.id)}
                aria-label={`${period.start_date}から${period.end_date}の特別期間を削除`}
                className={`flex size-11 min-h-11 min-w-11 shrink-0 items-center justify-center rounded-xxs ${C.text50} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
              >
                <Trash2 className="size-4" />
              </button>
            </div>
          ))}
        </div>
      ) : (
        <EmptyState message="特別期間は登録されていません" />
      )}
    </section>
  );
});
