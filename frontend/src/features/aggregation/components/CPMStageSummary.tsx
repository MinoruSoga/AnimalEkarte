import { C } from "@/lib/design-tokens";
import {
  AGGREGATION_CPM_STAGE_OPTIONS,
  AGGREGATION_CPM_STAGE_SHORT_LABELS,
  type AggregationCPMStage,
} from "@/lib/cpm-stage";

import type { CPMStageCounts } from "../api/get-cpm-stage-counts";

interface CPMStageSummaryProps {
  counts: CPMStageCounts;
  total: number;
  isLoading: boolean;
  isError: boolean;
  selected?: AggregationCPMStage;
  onSelect: (stage: AggregationCPMStage | undefined) => void;
}

// 人数表示: エラーは "—"、読み込み中は "…"（"読み込み中..." 文言はテーブルの空状態と
// 衝突するため使わない）、それ以外は実数。
function countLabel(value: number, isLoading: boolean, isError: boolean): string {
  if (isError) return "—";
  if (isLoading) return "…";
  return String(value);
}

function chipClass(active: boolean): string {
  const base =
    "inline-flex min-h-11 min-w-11 items-center gap-1.5 px-3 py-1 rounded-full text-sm border transition-colors";
  return active
    ? `${base} ${C.bgBrand10} ${C.textBrand} border-current`
    : `${base} ${C.bgWhite} ${C.text} ${C.borderLight} ${C.hoverBgStatusGray}`;
}

// ISSUE-180: CPM セグメント別の人数サマリー。各チップはクリックで一覧を絞り込む。
// 人数は集計エンドポイントの total（フィルタ後の全件数）由来で、現在のフィルタ母集団を反映する。
// ページ内の件数ではない（per_page=1 でも total を読むため母集団全体が出る）。
export function CPMStageSummary({
  counts,
  total,
  isLoading,
  isError,
  selected,
  onSelect,
}: CPMStageSummaryProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-2"
      role="group"
      aria-label="CPMセグメント別人数"
    >
      <button
        type="button"
        onClick={() => onSelect(undefined)}
        aria-pressed={!selected}
        className={chipClass(!selected)}
      >
        すべて
        <span className="font-mono">
          {countLabel(total, isLoading, isError)}
        </span>
      </button>

      {AGGREGATION_CPM_STAGE_OPTIONS.map((opt) => {
        const isActive = selected === opt.value;
        return (
          <button
            key={opt.value}
            type="button"
            onClick={() => onSelect(isActive ? undefined : opt.value)}
            aria-pressed={isActive}
            className={chipClass(isActive)}
            title={opt.label}
          >
            {AGGREGATION_CPM_STAGE_SHORT_LABELS[opt.value]}
            <span className="font-mono">
              {countLabel(counts[opt.value], isLoading, isError)}
            </span>
          </button>
        );
      })}
    </div>
  );
}
