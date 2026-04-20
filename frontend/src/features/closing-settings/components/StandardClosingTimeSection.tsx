import { memo, useActionState } from "react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { handleApiError } from "@/lib/handle-api-error";
import type { ClinicSettings } from "@/types/generated/models";
import { updateClosingSettings } from "../api/update-closing-settings";

const WEEKDAY_LABELS: Record<number, string> = {
  0: "日",
  1: "月",
  2: "火",
  3: "水",
  4: "木",
  5: "金",
  6: "土",
};

interface StandardClosingTimeSectionProps {
  settings: ClinicSettings;
}

export const StandardClosingTimeSection = memo(function StandardClosingTimeSection({
  settings,
}: StandardClosingTimeSectionProps) {
  const [, formAction] = useActionState(async (_prev: null, formData: FormData) => {
    try {
      const closedWeekdays: number[] = [];
      for (let i = 0; i <= 6; i++) {
        if (formData.get(`closed_weekday_${i}`) === "on") {
          closedWeekdays.push(i);
        }
      }
      await updateClosingSettings({
        closing_am_pm_boundary: formData.get("closing_am_pm_boundary") as string,
        closing_weekday_end: formData.get("closing_weekday_end") as string,
        closing_sunday_end: formData.get("closing_sunday_end") as string,
        closed_weekdays: closedWeekdays,
      });
      toast.success("標準締め時間を更新しました");
    } catch (error) {
      handleApiError(error, "標準締め時間の更新");
    }
    return null;
  }, null);

  return (
    <section className={`bg-white rounded-lg border ${C.borderLight} p-6`}>
      <h2 className={`text-base font-semibold ${C.text} mb-4`}>標準締め時間</h2>
      <form action={formAction} className="space-y-4">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label htmlFor="closing_am_pm_boundary" className={STYLE.formLabel}>
              午前・午後 区切り時間
            </label>
            <input
              id="closing_am_pm_boundary"
              name="closing_am_pm_boundary"
              type="time"
              defaultValue={settings.closing_am_pm_boundary}
              className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
              required
            />
          </div>
          <div>
            <label htmlFor="closing_weekday_end" className={STYLE.formLabel}>
              平日 終了時間
            </label>
            <input
              id="closing_weekday_end"
              name="closing_weekday_end"
              type="time"
              defaultValue={settings.closing_weekday_end}
              className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
              required
            />
          </div>
          <div>
            <label htmlFor="closing_sunday_end" className={STYLE.formLabel}>
              日曜 終了時間
            </label>
            <input
              id="closing_sunday_end"
              name="closing_sunday_end"
              type="time"
              defaultValue={settings.closing_sunday_end}
              className={`${STYLE.formInput} mt-1 w-full rounded-[4px] border px-3`}
              required
            />
          </div>
        </div>
        <div>
          <p className={`${STYLE.formLabel} mb-2`}>休診曜日</p>
          <div className="flex flex-wrap gap-3">
            {[0, 1, 2, 3, 4, 5, 6].map((day) => (
              <label key={day} className="flex items-center gap-1.5 cursor-pointer">
                <input
                  type="checkbox"
                  name={`closed_weekday_${day}`}
                  defaultChecked={settings.closed_weekdays.includes(day)}
                  className="rounded"
                />
                <span className={`text-base ${C.text}`}>{WEEKDAY_LABELS[day]}</span>
              </label>
            ))}
          </div>
        </div>
        <div className="flex justify-end pt-2">
          <SubmitButton>保存</SubmitButton>
        </div>
      </form>
    </section>
  );
});
