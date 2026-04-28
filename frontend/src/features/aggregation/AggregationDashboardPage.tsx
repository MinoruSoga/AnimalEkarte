import { useState, useCallback, useMemo } from "react";
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

export type AggregationTab = "revenue" | "visit" | "last_visit";

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
  if (value === "revenue" || value === "visit" || value === "last_visit") {
    return value;
  }
  return null;
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
  last_visit: [
    ...CSV_COMMON_COLUMNS,
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "days_since_last_visit", getValue: (o) => String(o.days_since_last_visit ?? "") },
    { header: "last_visit_bucket", getValue: (o) => o.last_visit_bucket ?? "" },
    { header: "first_visit_date", getValue: (o) => o.first_visit_date ?? "" },
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
  const activeTab: AggregationTab = validateTab(searchParams.get("tab")) ?? "revenue";
  const [params, setParams] = useState<AggregationParams>(TAB_DEFAULT_PARAMS[activeTab]);
  const [selectedOwnerIds, setSelectedOwnerIds] = useState<Set<string>>(new Set());

  const { data, isLoading, isError, error } = useGetOwnerAggregations(params);
  const owners = data?.owners ?? [];

  const handleTabChange = useCallback(
    (tab: string) => {
      const validTab = validateTab(tab) ?? "revenue";
      setSearchParams({ tab: validTab }, { replace: true });
      setParams(TAB_DEFAULT_PARAMS[validTab]);
      setSelectedOwnerIds(new Set());
    },
    [setSearchParams]
  );

  const handleParamsChange = useCallback((partial: Partial<AggregationParams>) => {
    setParams((prev) => ({ ...prev, ...partial }));
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

  const errorMessage = isError
    ? (error instanceof Error ? error.message : "データの読み込みに失敗しました")
    : undefined;

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
