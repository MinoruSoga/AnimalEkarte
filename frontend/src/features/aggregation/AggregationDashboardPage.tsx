import { useState, useCallback, useMemo, useEffect } from "react";
import { useSearchParams } from "react-router";
import { Download } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Pagination } from "@/components/shared/Pagination";
import { Button } from "@/components/ui/button";
import { UnifiedTabs } from "@/components/shared/UnifiedTabs";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { useGetOwnerAggregations, type AggregationParams, type AggregationOwner } from "./api/get-aggregations";
import { AggregationFilterPanel } from "./AggregationFilterPanel";
import { AggregationOwnerTable } from "./AggregationOwnerTable";

// 集計軸は仕様書 §6.2 で revenue / visit / last_visit の3つに固定。
// 既定タブは売上ランキング (revenue)。それ以外の値は URL に書き込めない。
export type AggregationTab = "revenue" | "visit" | "last_visit";

export const DEFAULT_AGGREGATION_TAB: AggregationTab = "revenue";

const AGGREGATION_TABS: readonly AggregationTab[] = ["revenue", "visit", "last_visit"] as const;

const CURRENT_YEAR = new Date().getFullYear();

const TAB_DEFAULT_PARAMS: Record<AggregationTab, AggregationParams> = {
  revenue: {
    page: 1,
    per_page: 50,
    year: CURRENT_YEAR,
    amount_basis: "gross_total_amount",
    sort: "annual_amount",
    order: "desc",
  },
  visit: {
    page: 1,
    per_page: 50,
    period_preset: "last_12_months",
    sort: "period_visit_count",
    order: "desc",
  },
  last_visit: {
    page: 1,
    per_page: 50,
    last_visit_bucket: "over_3m",
    sort: "last_visit_date",
    order: "asc",
  },
};

function validateTab(value: unknown): AggregationTab | null {
  return AGGREGATION_TABS.find((t) => t === value) ?? null;
}

interface CsvColumnDef {
  header: string;
  getValue: (o: AggregationOwner) => string;
}

const CSV_COMMON_COLUMNS: CsvColumnDef[] = [
  { header: "owner_id", getValue: (o) => o.owner_id },
  { header: "owner_name", getValue: (o) => `"${o.owner_name.replace(/"/g, '""')}"` },
];

const CSV_COLUMNS: Record<AggregationTab, CsvColumnDef[]> = {
  revenue: [
    ...CSV_COMMON_COLUMNS,
    { header: "annual_amount", getValue: (o) => String(o.annual_amount ?? "") },
    { header: "billing_count", getValue: (o) => String(o.billing_count ?? "") },
    { header: "period_visit_count", getValue: (o) => String(o.period_visit_count ?? "") },
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "first_visit_date", getValue: (o) => o.first_visit_date ?? "" },
  ],
  visit: [
    ...CSV_COMMON_COLUMNS,
    { header: "period_visit_count", getValue: (o) => String(o.period_visit_count ?? "") },
    { header: "total_visit_count", getValue: (o) => String(o.total_visit_count) },
    { header: "annual_visit_count", getValue: (o) => String(o.annual_visit_count) },
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "first_visit_date", getValue: (o) => o.first_visit_date ?? "" },
  ],
  // 仕様書 §4.3 表示項目に合わせる: 飼い主名 / 最終来院日 / 経過日数 / 分類 / 累計来院回数 / 年間来院回数 / 累計診療費
  // 画面 (TAB_SPECIFIC_COLUMNS.last_visit) と1対1で揃えること。drift 防止。
  last_visit: [
    ...CSV_COMMON_COLUMNS,
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "days_since_last_visit", getValue: (o) => String(o.days_since_last_visit ?? "") },
    { header: "last_visit_bucket", getValue: (o) => o.last_visit_bucket ?? "" },
    { header: "total_visit_count", getValue: (o) => String(o.total_visit_count) },
    { header: "annual_visit_count", getValue: (o) => String(o.annual_visit_count) },
    { header: "total_amount", getValue: (o) => String(o.total_amount ?? o.total_fee ?? "") },
  ],
};

function buildCsvContent(owners: AggregationOwner[], tab: AggregationTab): string {
  const columns = CSV_COLUMNS[tab];
  const header = columns.map((col) => col.header).join(",");

  const rows = owners.map((o) =>
    columns.map((col) => col.getValue(o)).join(",")
  );

  return [header, ...rows].join("\n");
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

  // URL → state 同期:
  //   1. URL に tab がない、または不正値 → URL を正規化し DEFAULT_AGGREGATION_TAB に揃える
  //   2. ブラウザ戻る/進む等で URL の tab が変わった場合、params も該当タブの初期条件にリセットする
  //      （ユーザー操作のタブ切り替えは handleTabChange が同時に setParams するので冪等）
  // 同期目的のため useEffect 内 setState は許容。
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (rawTab !== activeTab) {
      setSearchParams({ tab: activeTab }, { replace: true });
    }
    setParams(TAB_DEFAULT_PARAMS[activeTab]);
    setSelectedOwnerIds(new Set());
    // activeTab が変わったときだけ実行する。
  }, [activeTab, rawTab, setSearchParams]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const { data, isLoading, isError, error } = useGetOwnerAggregations(params);
  const owners = useMemo(() => data?.owners ?? [], [data?.owners]);

  const handleTabChange = useCallback(
    (tab: string) => {
      const validTab = validateTab(tab) ?? DEFAULT_AGGREGATION_TAB;
      // setSearchParams のみで完結。useEffect が params / 選択状態のリセットを担う。
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
    const date = new Date().toISOString().slice(0, 10);
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
      maxWidth="max-w-full"
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

        {/* 件数 + 選択件数 (NotionFilter のツールバーと同じ密度) */}
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
