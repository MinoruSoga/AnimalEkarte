import { memo, useActionState, useState, useCallback } from "react";
import { Calendar, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { MasterSidePanel, PropertyRow } from "@/components/shared/SidePeek";
import { getFormString } from "@/lib/form-data";
import type { ClosingSpecialPeriod } from "@/types/generated/models";
import { useCreateSpecialPeriod, useDeleteSpecialPeriod } from "../api/special-periods";

interface SpecialPeriodSectionProps {
  periods: ClosingSpecialPeriod[];
}

export const SpecialPeriodSection = memo(function SpecialPeriodSection({
  periods,
}: SpecialPeriodSectionProps) {
  const [showForm, setShowForm] = useState(false);
  const [note, setNote] = useState("");
  const createMutation = useCreateSpecialPeriod();
  const deleteMutation = useDeleteSpecialPeriod();

  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      await createMutation.mutateAsync({
        start_date: getFormString(formData, "start_date"),
        end_date: getFormString(formData, "end_date"),
        am_pm_boundary: getFormString(formData, "am_pm_boundary"),
        pm_end: getFormString(formData, "pm_end"),
        note: getFormString(formData, "note") || undefined,
      });
      toast.success("特別期間を追加しました");
      setShowForm(false);
      setNote("");
    } catch {
      // FE-RC-005: useCreateSpecialPeriod.onError が既に handleApiError で通知済み。
    }
    return null;
  }, null);

  const handleDelete = useCallback(
    async (id: number) => {
      try {
        await deleteMutation.mutateAsync(id);
        toast.success("特別期間を削除しました");
      } catch {
        // FE-RC-005: useDeleteSpecialPeriod.onError が既に handleApiError で通知済み。
      }
    },
    [deleteMutation],
  );

  const handleShowForm = useCallback(() => setShowForm(true), []);
  const handleHideForm = useCallback(() => {
    setShowForm(false);
    setNote("");
  }, []);

  return (
    <section className={`${C.bgWhite} rounded-lg border ${C.borderLight} p-6`}>
      <div className="flex items-center justify-between mb-4">
        <h2 className={`text-base font-semibold ${C.text}`}>特別期間</h2>
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
        <MasterSidePanel
          isNew
          title={note}
          onTitleChange={setNote}
          titlePlaceholder="メモ（例: 年末年始）"
          onClose={handleHideForm}
          action={formAction}
          icon={<Calendar className={LAYOUT.pageIcon.innerIcon} />}
        >
          <input type="hidden" name="note" value={note} />
          <PropertyRow label="開始日">
            <input
              id="start_date"
              name="start_date"
              type="date"
              aria-label="開始日"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow label="終了日">
            <input
              id="end_date"
              name="end_date"
              type="date"
              aria-label="終了日"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow label="午前・午後 区切り時間">
            <input
              id="am_pm_boundary"
              name="am_pm_boundary"
              type="time"
              aria-label="午前・午後 区切り時間"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
          <PropertyRow label="午後 終了時間">
            <input
              id="pm_end"
              name="pm_end"
              type="time"
              aria-label="午後 終了時間"
              className={`${STYLE.formInput} w-full rounded-xs border px-3`}
              required
            />
          </PropertyRow>
        </MasterSidePanel>
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
