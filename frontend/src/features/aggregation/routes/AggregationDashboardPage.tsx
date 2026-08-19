import { useState, useCallback, useMemo, useEffect } from "react";
import { useSearchParams } from "react-router";
import { Download } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Pagination } from "@/components/shared/Pagination";
import { Button } from "@/components/ui/button";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { C, ICON, STYLE, LAYOUT } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
import type { AggregationCPMStage } from "@/lib/cpm-stage";
import { useGetOwnerAggregations, type AggregationParams } from "../api/get-aggregations";
import { useGetCPMStageCounts } from "../api/get-cpm-stage-counts";
import { AggregationFilterPanel } from "../components/AggregationFilterPanel";
import { AggregationOwnerTable } from "../components/AggregationOwnerTable";
import { CPMStageSummary } from "../components/CPMStageSummary";
import { buildCsvContent } from "../components/aggregation-csv";
import type { AggregationTab } from "../components/aggregation-filter-panel-model";

const DEFAULT_AGGREGATION_TAB: AggregationTab = "revenue";

const AGGREGATION_TABS: readonly AggregationTab[] = ["revenue", "visit", "last_visit"] as const;

const CURRENT_YEAR = Number(todayJSTISO().slice(0, 4));

const TAB_DEFAULT_PARAMS: Record<AggregationTab, AggregationParams> = {
  revenue: {
    page: 1,
    per_page: 50,
    year: CURRENT_YEAR,
    amount_basis: "gross_total_amount",
    sort: "annual_amount",
    order: "desc",
  },
  // BUG-012: 来院/最終来院は売上0除外の対象外。UIに include_zero が無いため明示 true。
  visit: {
    page: 1,
    per_page: 50,
    period_preset: "last_12_months",
    sort: "period_visit_count",
    order: "desc",
    include_zero: true,
  },
  last_visit: {
    page: 1,
    per_page: 50,
    last_visit_bucket: "over_3m",
    sort: "last_visit_date",
    order: "asc",
    include_zero: true,
  },
};

function validateTab(value: unknown): AggregationTab | null {
  return AGGREGATION_TABS.find((t) => t === value) ?? null;
}

// エラーメッセージは「データの読み込みに失敗しました」をベースとし、
// 利用者に原因 (HTTP ステータス等) が分かる形で補助情報を併記する。
// axios エラーは Error インスタンスで `error.message` が `Request failed with status code 404` 等になるため、
// それをそのまま表示すると業務利用者には伝わりにくい。
function formatAggregationError(error: unknown): string {
  const baseMessage = "データの読み込みに失敗しました";
  if (error !== null && typeof error === "object") {
    const response = (error as { response?: { status?: number; statusText?: string; data?: { error?: string } } }).response;
    if (response?.status) {
      const detail = response.statusText
        ? `HTTP ${response.status} ${response.statusText}`
        : `HTTP ${response.status}`;
      const apiError = response.data?.error;
      return apiError
        ? `${baseMessage} (${detail}: ${apiError})`
        : `${baseMessage} (${detail})`;
    }
    if (error instanceof Error && error.message && !error.message.startsWith("Request failed")) {
      return `${baseMessage}: ${error.message}`;
    }
  }
  return baseMessage;
}

function downloadCsv(content: string, filename: string): void {
  const bom = "﻿";
  const blob = new Blob([bom + content], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

export function AggregationDashboardPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const rawTab = searchParams.get("tab");
  const activeTab: AggregationTab = validateTab(rawTab) ?? DEFAULT_AGGREGATION_TAB;
  const [params, setParams] = useState<AggregationParams>(TAB_DEFAULT_PARAMS[activeTab]);
  const [selectedOwnerIds, setSelectedOwnerIds] = useState<Set<string>>(new Set());

  // タブ変更 (ユーザー操作・ブラウザ戻る/進む共通) に伴う params / 選択状態のリセット。
  // useEffect 同期だと旧タブの params で 1 フレーム commit + 再レンダーが発生するため、
  // レンダー中の比較で同期する (rerender-derived-state-no-effect パターン)。
  const [prevTab, setPrevTab] = useState(activeTab);
  if (prevTab !== activeTab) {
    setPrevTab(activeTab);
    setParams(TAB_DEFAULT_PARAMS[activeTab]);
    setSelectedOwnerIds(new Set());
  }

  // URL 正規化: URL に tab がない、または不正値 → DEFAULT_AGGREGATION_TAB に揃える
  useEffect(() => {
    if (rawTab !== activeTab) {
      setSearchParams({ tab: activeTab }, { replace: true });
    }
  }, [activeTab, rawTab, setSearchParams]);

  const { data, isLoading, isError, error } = useGetOwnerAggregations(params);
  const owners = useMemo(() => data?.owners ?? [], [data?.owners]);

  // ISSUE-180: CPM セグメント別の人数（total 由来）。現在のフィルタ母集団を反映する。
  const cpmCounts = useGetCPMStageCounts(params);

  const handleTabChange = useCallback(
    (tab: string) => {
      const validTab = validateTab(tab) ?? DEFAULT_AGGREGATION_TAB;
      // setSearchParams のみで完結。レンダー中の prevTab 比較が params / 選択状態のリセットを担う。
      setSearchParams({ tab: validTab }, { replace: true });
    },
    [setSearchParams]
  );

  const handleParamsChange = useCallback((partial: Partial<AggregationParams>) => {
    // フィルタ・ソート変更時は仕様 §10.2 に従い必ず1ページ目に戻す。
    // Pagination からの呼び出しは partial.page を明示するためそちらを尊重する。
    setParams((prev) => ({
      ...prev,
      ...partial,
      page: partial.page ?? 1,
    }));
    setSelectedOwnerIds(new Set());
  }, []);

  const handleCpmStageSelect = useCallback(
    (stage: AggregationCPMStage | undefined) => {
      handleParamsChange({ cpm_stage: stage });
    },
    [handleParamsChange]
  );

  const handleSelectAll = useCallback(
    (checked: boolean) => {
      if (checked) {
        setSelectedOwnerIds(new Set(owners.map((o) => o.owner_id)));
      } else {
        setSelectedOwnerIds(new Set());
      }
    },
    [owners]
  );

  const handleSelectOwner = useCallback((ownerId: string, checked: boolean) => {
    setSelectedOwnerIds((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(ownerId);
      } else {
        next.delete(ownerId);
      }
      return next;
    });
  }, []);

  const selectedCount = selectedOwnerIds.size;

  const selectedOwners = useMemo(
    () => owners.filter((o) => selectedOwnerIds.has(o.owner_id)),
    [owners, selectedOwnerIds]
  );

  const handleExportCsv = useCallback(() => {
    // 誤操作防止: 選択 0 件のときは早期 return（disabled 属性と二重で防御）
    if (selectedCount === 0) return;
    const csv = buildCsvContent(selectedOwners, activeTab);
    const date = todayJSTISO();
    downloadCsv(csv, `aggregation-${activeTab}-${date}.csv`);
  }, [selectedCount, selectedOwners, activeTab]);

  const errorMessage = isError ? formatAggregationError(error) : undefined;

  const tabItems = [
    { value: "revenue" as const, label: "売上ランキング" },
    { value: "visit" as const, label: "来院回数" },
    { value: "last_visit" as const, label: "最終来院" },
  ];

  return (
    <PageLayout
      title="顧客集計ダッシュボード"
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      headerAction={
        <Button
          variant="outline"
          className={STYLE.btnOutline}
          onClick={handleExportCsv}
          disabled={isLoading || selectedCount === 0}
          title={selectedCount === 0 ? "出力対象を選択してください" : undefined}
          aria-label={selectedCount === 0 ? "CSV出力 (出力対象を選択してください)" : `${selectedCount}件をCSV出力`}
        >
          <Download className={`mr-1.5 ${ICON.action}`} />
          {selectedCount > 0 ? `${selectedCount}件をCSV出力` : "CSV出力"}
        </Button>
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* タブUI */}
        <UnifiedTabs
          items={tabItems}
          value={activeTab}
          onValueChange={handleTabChange}
          className="w-full"
        />

        {/* フィルタパネル */}
        <AggregationFilterPanel params={params} onParamsChange={handleParamsChange} activeTab={activeTab} />

        {/* CPM セグメント別の人数サマリー（ISSUE-180）。クリックで一覧を絞り込む。 */}
        <CPMStageSummary
          counts={cpmCounts.counts}
          total={cpmCounts.total}
          isLoading={cpmCounts.isLoading}
          isError={cpmCounts.isError}
          selected={params.cpm_stage}
          onSelect={handleCpmStageSelect}
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
          onSelectAll={handleSelectAll}
          onSelectOwner={handleSelectOwner}
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
            onPageChange={(page) => handleParamsChange({ page })}
            onPrev={() => handleParamsChange({ page: Math.max(1, data.page - 1) })}
            onNext={() => handleParamsChange({ page: data.page + 1 })}
          />
        ) : null}
      </div>
    </PageLayout>
  );
}
