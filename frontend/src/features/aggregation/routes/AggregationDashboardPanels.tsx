import { Download } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Pagination } from "@/components/shared/Pagination";
import { Button } from "@/components/ui/button";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import type { AggregationCPMStage } from "@/lib/cpm-stage";
import type { AggregationOwner, AggregationParams, AggregationResponse } from "../api/get-aggregations";
import type { CPMStageCountsResult } from "../api/get-cpm-stage-counts";
import { AggregationFilterPanel } from "../components/AggregationFilterPanel";
import { AggregationOwnerTable } from "../components/AggregationOwnerTable";
import { CPMStageSummary } from "../components/CPMStageSummary";
import type { AggregationTab } from "../components/aggregation-filter-panel-model";
import { AGGREGATION_TAB_ITEMS } from "./aggregation-dashboard-model";

interface AggregationDashboardViewProps {
  activeTab: AggregationTab;
  onTabChange: (tab: string) => void;
  params: AggregationParams;
  onParamsChange: (partial: Partial<AggregationParams>) => void;
  onCpmStageSelect: (stage: AggregationCPMStage | undefined) => void;
  cpmCounts: CPMStageCountsResult;
  data: AggregationResponse | undefined;
  owners: AggregationOwner[];
  selectedOwnerIds: Set<string>;
  selectedCount: number;
  onSelectAll: (checked: boolean) => void;
  onSelectOwner: (ownerId: string, checked: boolean) => void;
  isLoading: boolean;
  isError: boolean;
  errorMessage: string | undefined;
  onExportCsv: () => void;
}

function AggregationDashboardHeaderAction({
  isLoading,
  selectedCount,
  onExportCsv,
}: {
  isLoading: boolean;
  selectedCount: number;
  onExportCsv: () => void;
}) {
  return (
    <Button
      variant="outline"
      className={STYLE.btnOutline}
      onClick={onExportCsv}
      disabled={isLoading || selectedCount === 0}
      title={selectedCount === 0 ? "出力対象を選択してください" : undefined}
      aria-label={selectedCount === 0 ? "CSV出力 (出力対象を選択してください)" : `${selectedCount}件をCSV出力`}
    >
      <Download className={`mr-1.5 ${ICON.action}`} />
      {selectedCount > 0 ? `${selectedCount}件をCSV出力` : "CSV出力"}
    </Button>
  );
}

export function AggregationDashboardView({
  activeTab,
  onTabChange,
  params,
  onParamsChange,
  onCpmStageSelect,
  cpmCounts,
  data,
  owners,
  selectedOwnerIds,
  selectedCount,
  onSelectAll,
  onSelectOwner,
  isLoading,
  isError,
  errorMessage,
  onExportCsv,
}: AggregationDashboardViewProps) {
  return (
    <PageLayout
      title="顧客集計ダッシュボード"
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        <AggregationDashboardHeaderAction
          isLoading={isLoading}
          selectedCount={selectedCount}
          onExportCsv={onExportCsv}
        />
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* タブUI */}
        <UnifiedTabs
          items={AGGREGATION_TAB_ITEMS}
          value={activeTab}
          onValueChange={onTabChange}
          className="w-full"
        />

        {/* フィルタパネル */}
        <AggregationFilterPanel params={params} onParamsChange={onParamsChange} activeTab={activeTab} />

        {/* CPM セグメント別の人数サマリー（ISSUE-180）。クリックで一覧を絞り込む。 */}
        <CPMStageSummary
          counts={cpmCounts.counts}
          total={cpmCounts.total}
          isLoading={cpmCounts.isLoading}
          isError={cpmCounts.isError}
          selected={params.cpm_stage}
          onSelect={onCpmStageSelect}
        />

        {/* 件数 + 選択件数 (PropertyFilter のツールバーと同じ密度) */}
        <div className="flex flex-wrap items-center gap-2">
          {data ? <span className={STYLE.searchCount}>{data.total} 件</span> : null}
          {selectedCount > 0 ? (
            <span className={`text-base font-medium ${C.textBrand}`}>
              {selectedCount}件選択中
            </span>
          ) : null}
        </div>

        {/* テーブル */}
        <AggregationOwnerTable
          owners={owners}
          selectedOwnerIds={selectedOwnerIds}
          onSelectAll={onSelectAll}
          onSelectOwner={onSelectOwner}
          isLoading={isLoading}
          activeTab={activeTab}
          isError={isError}
          errorMessage={errorMessage}
        />

        {/* ページネーション (他ページと同じ共通コンポーネント) */}
        {data && data.total > data.per_page ? (
          <Pagination
            currentPage={data.page}
            totalPages={Math.ceil(data.total / data.per_page)}
            totalCount={data.total}
            startIndex={Math.max(1, (data.page - 1) * data.per_page + 1)}
            endIndex={Math.min(data.total, data.page * data.per_page)}
            onPageChange={(page) => onParamsChange({ page })}
            onPrev={() => onParamsChange({ page: Math.max(1, data.page - 1) })}
            onNext={() => onParamsChange({ page: data.page + 1 })}
          />
        ) : null}
      </div>
    </PageLayout>
  );
}
