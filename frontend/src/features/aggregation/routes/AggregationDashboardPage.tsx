import { useState, useCallback, useMemo, useEffect } from "react";
import { useSearchParams } from "react-router";
import { todayJSTISO } from "@/lib/jst-date";
import type { AggregationCPMStage } from "@/lib/cpm-stage";
import { useGetOwnerAggregations, type AggregationParams } from "../api/get-aggregations";
import { useGetCPMStageCounts } from "../api/get-cpm-stage-counts";
import { buildCsvContent } from "../components/aggregation-csv";
import type { AggregationTab } from "../components/aggregation-filter-panel-model";
import {
  DEFAULT_AGGREGATION_TAB,
  TAB_DEFAULT_PARAMS,
  downloadCsv,
  formatAggregationError,
  validateTab,
} from "./aggregation-dashboard-model";
import { AggregationDashboardView } from "./AggregationDashboardPanels";

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

  return (
    <AggregationDashboardView
      activeTab={activeTab}
      onTabChange={handleTabChange}
      params={params}
      onParamsChange={handleParamsChange}
      onCpmStageSelect={handleCpmStageSelect}
      cpmCounts={cpmCounts}
      data={data}
      owners={owners}
      selectedOwnerIds={selectedOwnerIds}
      selectedCount={selectedCount}
      onSelectAll={handleSelectAll}
      onSelectOwner={handleSelectOwner}
      isLoading={isLoading}
      isError={isError}
      errorMessage={errorMessage}
      onExportCsv={handleExportCsv}
    />
  );
}
