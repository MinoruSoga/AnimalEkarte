import { useEffect, useMemo, useState } from "react";
import type { ColumnData } from "@/types";
import { computeCheckedInWaitStats, computeReceptionTotalCount, type TelemetryWaitStats } from "../lib/reception-telemetry";

const RECOMPUTE_INTERVAL_MS = 60_000;

/**
 * Phase 2（BE: checked_in_at）の表示ゲート。
 *
 * `false` の間は「受付済 0 件」（実データはあるがまだ誰も待っていない）と
 * 「Phase 2 がまだ稼働していない」（migration 未適用 / codegen 未 regen）を
 * データ形状だけで区別できない。そのため実データの有無から推測せず、
 * この明示フラグで判定する。
 *
 * migration `005_add_appointment_checked_in_at.sql` の適用 と
 * `make codegen` による `Reservation.checked_in_at` の regen が完了したら、
 * 専用のフォローアップコミットで `true` に切り替える（本フラグ自体は
 * 恒久的なフィーチャーフラグ基盤ではなく、この移行を跨ぐための一時的なスイッチ）。
 */
export const RECEPTION_TELEMETRY_PHASE2_ENABLED = false;

export interface ReceptionTelemetryResult extends TelemetryWaitStats {
  totalCount: number;
}

/**
 * 受付ヘッダー テレメトリ・ストリップの集計フック（change-ui.md）。
 *
 * - `columns` は必ずフィルタ非適用の生データを渡すこと。`filteredColumns` を渡すと
 *   フィルタ操作のたびに「本日受付」件数が変動してしまう（仕様違反）。
 * - 待ち時間は checked_in_at からの経過分を 60 秒間隔でローカル再計算する。
 *   専用ポーリング/API 呼び出しは追加しない（既存の一覧取得に相乗り）。
 */
export function useReceptionTelemetry(columns: ColumnData[]): ReceptionTelemetryResult {
  const [tick, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), RECOMPUTE_INTERVAL_MS);
    return () => clearInterval(id);
  }, []);

  const totalCount = useMemo(() => computeReceptionTotalCount(columns), [columns]);

  const waitStats = useMemo(
    () => computeCheckedInWaitStats(columns, new Date()),
    // tick は値を使わないが、60秒毎の再計算トリガーとして意図的に依存配列へ含める
    [columns, tick],
  );

  return { totalCount, ...waitStats };
}
