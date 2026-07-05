import type { ColumnData } from "@/types";

export interface TelemetryLongestWait {
  minutes: number;
  petName: string;
}

export interface TelemetryWaitStats {
  averageMinutes: number | null;
  longest: TelemetryLongestWait | null;
}

/**
 * 「本日受付」= 全カラム（フィルタ非適用）の予約件数合計。
 * cancelled / no_show は transformReservationsToReceptionColumns で上流除外済みのため、
 * ここでは columns の appointments を単純合算するだけでよい。
 *
 * 呼び出し側は必ず `columns`（フィルタ非適用の生データ）を渡すこと。`filteredColumns` を渡すと
 * フィルタ操作のたびに「今日の全体量」の意味が壊れる（change-ui.md 仕様）。
 */
export function computeReceptionTotalCount(columns: ColumnData[]): number {
  return columns.reduce((sum, col) => sum + col.appointments.length, 0);
}

/**
 * checked_in（受付済）ステータスの待ち時間統計。
 * 待ち時間 = now - checkedInAt。checkedInAt 未設定（Phase 2 データ未到達）の患者は集計対象外。
 * 対象 0 件の場合は averageMinutes/longest ともに null（呼び出し側で「—」表示）。
 */
export function computeCheckedInWaitStats(columns: ColumnData[], now: Date): TelemetryWaitStats {
  const waits: TelemetryLongestWait[] = [];

  for (const col of columns) {
    for (const appt of col.appointments) {
      if (appt.status !== "checked_in" || !appt.checkedInAt) continue;
      const checkedInAt = new Date(appt.checkedInAt);
      const minutes = Math.max(0, Math.round((now.getTime() - checkedInAt.getTime()) / 60_000));
      waits.push({ minutes, petName: appt.petName });
    }
  }

  if (waits.length === 0) {
    return { averageMinutes: null, longest: null };
  }

  const averageMinutes = Math.round(waits.reduce((sum, w) => sum + w.minutes, 0) / waits.length);
  const longest = waits.reduce((max, w) => (w.minutes > max.minutes ? w : max));

  return { averageMinutes, longest };
}
