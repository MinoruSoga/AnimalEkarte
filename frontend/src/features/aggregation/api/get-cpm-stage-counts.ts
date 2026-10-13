import { AGGREGATION_CPM_STAGES, type AggregationCPMStage } from "@/lib/cpm-stage";
import type { AggregationResponse } from "./get-aggregations";

export type CPMStageCounts = Record<AggregationCPMStage, number>;

export interface CPMStageCountsResult {
  counts: CPMStageCounts;
  total: number;
  isLoading: boolean;
  isError: boolean;
}

// EMR-334: CPM 区分人数は集計一覧レスポンスの cpm_stage_counts に同梱される
// (cpm_stage 絞り込み・ページネーション前の母集団)。区分ごとの追加 GET は廃止し、
// ここでは取得済みレスポンスからの純粋な派生のみ行う。ネットワークには触れない。
// BE は全 7 区分のキーを必ず返すが、欠損キーは 0 にフォールバックする。
export function toCPMStageCountsResult(
  data: AggregationResponse | undefined,
  isLoading: boolean,
  isError: boolean,
): CPMStageCountsResult {
  const counts = AGGREGATION_CPM_STAGES.reduce((acc, stage) => {
    acc[stage] = data?.cpm_stage_counts?.[stage] ?? 0;
    return acc;
  }, {} as CPMStageCounts);
  const total = AGGREGATION_CPM_STAGES.reduce((sum, stage) => sum + counts[stage], 0);

  return { counts, total, isLoading, isError };
}
