import { memo } from "react";
import { C } from "@/lib/design-tokens";
import type { TelemetryWaitStats } from "../lib/reception-telemetry";

interface ReceptionTelemetryStripProps {
  /** 表示日の受付件数（フィルタ非適用の全体値）。 */
  totalCount: number;
  /**
   * Phase 2（BE: checked_in_at）の待ち時間統計。
   * undefined = Phase 1（データ未到達、平均/最長セグメント自体を非表示）。
   * 値ありで averageMinutes/longest が null = 受付済 0 件（「—」表示）。
   * EMR-243: 待ち時間は「今からの経過」で算出するため、非本日表示では
   * 呼び出し側の値に関わらず必ず非表示にする（isToday で強制）。
   */
  waitStats?: TelemetryWaitStats;
  /**
   * EMR-243: 表示日が当日なら true。非本日では件数ラベルを「対象日の受付」に変え
   * 待ち時間を隠す。省略時は当日扱い（従来表示と同じ）。
   */
  isToday?: boolean;
}

/** 受付ヘッダー テレメトリ表示（change-ui.md）。表示専用（props → JSX）。 */
export const ReceptionTelemetryStrip = memo(function ReceptionTelemetryStrip({
  totalCount,
  waitStats,
  isToday = true,
}: ReceptionTelemetryStripProps) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-2 text-base border-b ${C.borderLight} ${C.bgWhite}`}
    >
      <span>
        <span className={C.text60}>{isToday ? "本日受付" : "対象日の受付"}</span>{" "}
        <span className={`font-semibold tabular-nums ${C.text}`}>{totalCount}</span>
        <span className={C.text60}>件</span>
      </span>

      {isToday && waitStats ? (
        <>
          <span className={C.text30} aria-hidden="true">
            ・
          </span>
          <span>
            <span className={C.text60}>平均待ち</span>{" "}
            <span className={`font-semibold tabular-nums ${C.text}`}>
              {waitStats.averageMinutes === null ? "—" : `${waitStats.averageMinutes}分`}
            </span>
          </span>

          <span className={C.text30} aria-hidden="true">
            ・
          </span>
          <span>
            <span className={C.text60}>最長待ち</span>{" "}
            {waitStats.longest === null ? (
              <span className={`font-semibold tabular-nums ${C.text}`}>—</span>
            ) : (
              <span className={`font-semibold tabular-nums ${C.textDiscount}`}>
                {waitStats.longest.minutes}分 — {waitStats.longest.petName}
              </span>
            )}
          </span>
        </>
      ) : null}
    </div>
  );
});
