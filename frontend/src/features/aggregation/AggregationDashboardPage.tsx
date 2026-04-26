import { useState, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { Download, RefreshCw } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import {
  Tabs,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { useGetOwnerAggregations, type LtvOwnersParams, type LtvOwner } from "./api/get-aggregations";
import { AggregationFilterPanel } from "./AggregationFilterPanel";
import { AggregationOwnerTable } from "./AggregationOwnerTable";

export type LtvTab = "revenue" | "visit" | "last_visit";

const CURRENT_YEAR = new Date().getFullYear();

const TAB_DEFAULT_PARAMS: Record<LtvTab, LtvOwnersParams> = {
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

function validateTab(value: unknown): LtvTab | null {
  if (value === "revenue" || value === "visit" || value === "last_visit") {
    return value;
  }
  return null;
}

interface CsvColumnDef {
  header: string;
  getValue: (o: LtvOwner) => string;
}

const CSV_COMMON_COLUMNS: CsvColumnDef[] = [
  { header: "owner_id", getValue: (o) => o.owner_id },
  { header: "owner_name", getValue: (o) => `"${o.owner_name.replace(/"/g, '""')}"` },
];

const CSV_COLUMNS: Record<LtvTab, CsvColumnDef[]> = {
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

function buildCsvContent(owners: LtvOwner[], tab: LtvTab): string {
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
  const activeTab: LtvTab = validateTab(searchParams.get("tab")) ?? "revenue";
  const [params, setParams] = useState<LtvOwnersParams>(TAB_DEFAULT_PARAMS[activeTab]);
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

  const handleParamsChange = useCallback((partial: Partial<LtvOwnersParams>) => {
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
    const targets = selectedCount > 0 ? selectedOwners : owners;
    const csv = buildCsvContent(targets, activeTab);
    const date = new Date().toISOString().slice(0, 10);
    downloadCsv(csv, `ltv-${activeTab}-${date}.csv`);
  }, [selectedCount, selectedOwners, owners, activeTab]);

  const errorMessage = isError
    ? (error instanceof Error ? error.message : "データの読み込みに失敗しました")
    : undefined;

  return (
    <PageLayout
      title="顧客集計ダッシュボード"
      maxWidth="max-w-full"
      headerAction={
        <Button
          variant="outline"
          className={STYLE.btnOutline}
          onClick={handleExportCsv}
          disabled={isLoading}
        >
          <Download className={`mr-1.5 ${ICON.sm}`} />
          {selectedCount > 0 ? `${selectedCount}件をCSV出力` : "CSV出力"}
        </Button>
      }
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* タブUI */}
        <Tabs value={activeTab} onValueChange={handleTabChange} className="w-full">
          <TabsList className={`grid w-full grid-cols-3 ${C.bgLight}`}>
            <TabsTrigger
              value="revenue"
              className={`data-[state=active]:${C.dataActiveBorderB} data-[state=active]:${C.dataActiveText}`}
            >
              売上ランキング
            </TabsTrigger>
            <TabsTrigger
              value="visit"
              className={`data-[state=active]:${C.dataActiveBorderB} data-[state=active]:${C.dataActiveText}`}
            >
              来院回数
            </TabsTrigger>
            <TabsTrigger
              value="last_visit"
              className={`data-[state=active]:${C.dataActiveBorderB} data-[state=active]:${C.dataActiveText}`}
            >
              最終来院
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {/* フィルタパネル */}
        <AggregationFilterPanel params={params} onParamsChange={handleParamsChange} activeTab={activeTab} />

        {/* ツールバー */}
        <div className={`flex items-center gap-3 py-2 px-1 border-b ${C.borderLight}`}>
          <span className={`text-sm ${C.text65}`}>
            {data ? `全${data.total}件` : ""}
          </span>
          {selectedCount > 0 ? (
            <span className={`text-sm font-medium ${C.textBrand}`}>
              {selectedCount}件選択中
            </span>
          ) : null}
          <div className="flex items-center gap-2 ml-auto">
            <Button
              variant="outline"
              className={STYLE.btnOutline}
              onClick={() => handleParamsChange({ page: 1 })}
              disabled={isLoading}
            >
              <RefreshCw className={`mr-1.5 ${ICON.sm} ${isLoading ? "animate-spin" : ""}`} />
              更新
            </Button>
          </div>
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

        {/* ページネーション */}
        {data ? (
          <div className={`flex items-center justify-between text-sm ${C.text50}`}>
            <span>
              {Math.max(1, (data.page - 1) * data.per_page + 1)}
              〜
              {Math.min(data.total, data.page * data.per_page)}
              件（全{data.total}件）
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleParamsChange({ page: Math.max(1, data.page - 1) })}
                disabled={data.page <= 1 || isLoading}
              >
                前へ
              </Button>
              <span className={C.text65}>
                {data.page} / {Math.ceil(data.total / data.per_page)}
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleParamsChange({ page: data.page + 1 })}
                disabled={data.page >= Math.ceil(data.total / data.per_page) || isLoading}
              >
                次へ
              </Button>
            </div>
          </div>
        ) : null}
      </div>
    </PageLayout>
  );
}
