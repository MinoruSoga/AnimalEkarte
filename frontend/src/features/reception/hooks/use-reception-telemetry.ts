import { useEffect, useMemo, useState } from "react";
import type { ColumnData } from "@/types";
import {
  computeCheckedInWaitStats,
  computeReceptionTotalCount,
  type TelemetryWaitStats,
} from "../lib/reception-telemetry";

const RECOMPUTE_INTERVAL_MS = 60_000;

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

  const waitStats = useMemo(() => {
    // tick は 60 秒毎の再計算トリガー。値自体は使わず Date を取り直す。
    void tick;
    return computeCheckedInWaitStats(columns, new Date());
  }, [columns, tick]);

  return { totalCount, ...waitStats };
}
