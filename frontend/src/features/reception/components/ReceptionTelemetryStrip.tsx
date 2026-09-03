import { memo } from "react";
import { C } from "@/lib/design-tokens";
import type { TelemetryWaitStats } from "../lib/reception-telemetry";

interface ReceptionTelemetryStripProps {
  /** 本日受付件数（フィルタ非適用の全体値）。 */
  totalCount: number;
  /**
   * Phase 2（BE: checked_in_at）の待ち時間統計。
   * undefined = Phase 1（データ未到達、平均/最長セグメント自体を非表示）。
   * 値ありで averageMinutes/longest が null = 受付済 0 件（「—」表示）。
   */
  waitStats?: TelemetryWaitStats;
}

/** 受付ヘッダー テレメトリ表示（change-ui.md）。表示専用（props → JSX）。 */
export const ReceptionTelemetryStrip = memo(function ReceptionTelemetryStrip({
  totalCount,
  waitStats,
}: ReceptionTelemetryStripProps) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-2 text-base border-b ${C.borderLight} ${C.bgWhite}`}
    >
      <span>
        <span className={C.text60}>本日受付</span>{" "}
        <span className={`font-bold tabular-nums ${C.text}`}>{totalCount}</span>
        <span className={C.text60}>件</span>
      </span>

      {waitStats ? (
        <>
          <span className={C.text30} aria-hidden="true">
            ・
          </span>
          <span>
            <span className={C.text60}>平均待ち</span>{" "}
            <span className={`font-bold tabular-nums ${C.text}`}>
              {waitStats.averageMinutes === null ? "—" : `${waitStats.averageMinutes}分`}
            </span>
          </span>

          <span className={C.text30} aria-hidden="true">
            ・
          </span>
          <span>
            <span className={C.text60}>最長待ち</span>{" "}
            {waitStats.longest === null ? (
              <span className={`font-bold tabular-nums ${C.text}`}>—</span>
            ) : (
              <span className={`font-bold tabular-nums ${C.textDiscount}`}>
                {waitStats.longest.minutes}分 — {waitStats.longest.petName}
              </span>
            )}
          </span>
        </>
      ) : null}
    </div>
  );
});
