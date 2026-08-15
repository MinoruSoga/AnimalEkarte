import { memo, useActionState, useState } from "react";
import { Check } from "lucide-react";
import { toast } from "sonner";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { TableCell, TableHead } from "@/components/ui/table";
import { handleApiError } from "@/lib/handle-api-error";
import type { ClinicSettings } from "@/types/generated/models";
import { DAY_OF_WEEK_LABELS as WEEKDAY_LABELS } from "@/constants/day-of-week";
import { updateClosingSettings } from "../api/update-closing-settings";
import {
  computeClosingTimeRanges,
  formatRangeText,
  toTimeInputValue,
} from "../lib/closing-time-ranges";

// 時間帯プレビューの行定義（key は ClosingTimeRanges のキーと一致）。
const PERIOD_ROWS = [
  { key: "am", label: "AM", caption: "午前" },
  { key: "pm", label: "PM", caption: "午後" },
  { key: "emg", label: "EMG", caption: "緊急" },
] as const;

interface StandardClosingTimeSectionProps {
  settings: ClinicSettings;
}

export const StandardClosingTimeSection = memo(function StandardClosingTimeSection({
  settings,
}: StandardClosingTimeSectionProps) {
  // 時間帯プレビュー用に境界値を制御値として保持する。
  // 送信は form action が FormData(DOM 値) から読むため、このプレビュー state とは独立。
  const [amPmBoundary, setAmPmBoundary] = useState(() =>
    toTimeInputValue(settings.closing_am_pm_boundary),
  );
  const [weekdayEnd, setWeekdayEnd] = useState(() =>
    toTimeInputValue(settings.closing_weekday_end),
  );
  const [sundayEnd, setSundayEnd] = useState(() => toTimeInputValue(settings.closing_sunday_end));

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

  // 派生値は描画時に計算する（useEffect で同期しない）。AM は am_start と境界に依存し
  // 平日・日曜で共通。PM/EMG は各曜日の終了時刻に追従する（EMG は翌 am_start までの越日 #215）。
  const weekdayRanges = computeClosingTimeRanges(amPmBoundary, weekdayEnd, settings.closing_am_start);
  const sundayRanges = computeClosingTimeRanges(amPmBoundary, sundayEnd, settings.closing_am_start);

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
              value={amPmBoundary}
              onChange={(event) => setAmPmBoundary(event.target.value)}
              className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
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
              value={weekdayEnd}
              onChange={(event) => setWeekdayEnd(event.target.value)}
              className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
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
              value={sundayEnd}
              onChange={(event) => setSundayEnd(event.target.value)}
              className={`${STYLE.formInput} mt-1 w-full rounded-xs border px-3`}
              required
            />
          </div>
        </div>

        {/* 時間帯プレビュー: 境界編集に即時追従する（Issue #151）。 */}
        <div className={`overflow-x-auto rounded-xs border ${C.borderLight} ${C.bgSubtle} p-4`}>
          <table className="w-full border-collapse text-left">
            <caption className={`mb-3 text-left text-sm font-medium ${C.textMuted}`}>
              時間帯プレビュー
            </caption>
            <thead>
              <tr>
                <TableHead scope="col" className={C.textMuted}>
                  区分
                </TableHead>
                <TableHead scope="col" className={C.textMuted}>
                  平日
                </TableHead>
                <TableHead scope="col" className={C.textMuted}>
                  日曜
                </TableHead>
              </tr>
            </thead>
            <tbody>
              {PERIOD_ROWS.map((row) => (
                <tr key={row.key}>
                  <TableHead scope="row" className={C.text}>
                    <span className="font-semibold">{row.label}</span>
                    <span className={`ml-2 text-sm ${C.textMuted}`}>{row.caption}</span>
                  </TableHead>
                  <TableCell className={`tabular-nums whitespace-nowrap ${C.text}`}>
                    {formatRangeText(weekdayRanges[row.key])}
                  </TableCell>
                  <TableCell className={`tabular-nums whitespace-nowrap ${C.text}`}>
                    {formatRangeText(sundayRanges[row.key])}
                  </TableCell>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div>
          <p className={`${STYLE.formLabel} mb-2`}>休診曜日</p>
          <div className="flex flex-wrap gap-3">
            {[0, 1, 2, 3, 4, 5, 6].map((day) => (
              <label
                key={day}
                className="relative flex min-h-11 min-w-11 items-center gap-1.5 cursor-pointer"
              >
                <input
                  type="checkbox"
                  name={`closed_weekday_${day}`}
                  defaultChecked={(settings.closed_weekdays ?? []).includes(day)}
                  className="peer absolute inset-0 size-full cursor-pointer opacity-0"
                />
                <span
                  aria-hidden="true"
                  className={`flex size-4 shrink-0 items-center justify-center rounded-xs border ${C.borderMedium} bg-white text-white transition-colors peer-checked:border-primary peer-checked:bg-primary peer-focus-visible:ring-[3px] peer-focus-visible:ring-ring`}
                >
                  <Check className="size-3.5" />
                </span>
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
