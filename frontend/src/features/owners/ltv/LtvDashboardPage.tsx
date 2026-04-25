import { useState, useCallback, useMemo } from "react";
import { Download, RefreshCw, Tags } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { Button } from "@/components/ui/button";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { useGetLtvOwners, type LtvOwnersParams, type LtvOwner } from "../api/get-ltv-owners";
import { LtvFilterPanel } from "./LtvFilterPanel";
import { LtvOwnerTable } from "./LtvOwnerTable";
import { LtvBulkSyncDialog } from "./LtvBulkSyncDialog";

const DEFAULT_PARAMS: LtvOwnersParams = {
  page: 1,
  per_page: 50,
};

function buildCsvContent(owners: LtvOwner[]): string {
  const header = [
    "owner_id",
    "owner_name",
    "has_line",
    "cpm_stage",
    "total_fee",
    "annual_visit_count",
    "total_visit_count",
    "last_visit_date",
    "first_visit_date",
  ].join(",");

  const rows = owners.map((o) =>
    [
      o.owner_id,
      `"${o.owner_name.replace(/"/g, '""')}"`,
      o.has_line ? "true" : "false",
      o.cpm_stage ?? "",
      o.total_fee,
      o.annual_visit_count,
      o.total_visit_count,
      o.last_visit_date ?? "",
      o.first_visit_date ?? "",
    ].join(",")
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

export function LtvDashboardPage() {
  const [params, setParams] = useState<LtvOwnersParams>(DEFAULT_PARAMS);
  const [selectedOwnerIds, setSelectedOwnerIds] = useState<Set<string>>(new Set());
  const [syncDialogOpen, setSyncDialogOpen] = useState(false);

  const { data, isLoading } = useGetLtvOwners(params);
  const owners = data?.owners ?? [];

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
    const csv = buildCsvContent(targets);
    const date = new Date().toISOString().slice(0, 10);
    downloadCsv(csv, `ltv-owners-${date}.csv`);
  }, [selectedCount, selectedOwners, owners]);

  const handleSyncClick = useCallback(() => {
    setSyncDialogOpen(true);
  }, []);

  return (
    <PageLayout
      title="LTV/CPMダッシュボード"
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
        {/* フィルタパネル */}
        <LtvFilterPanel params={params} onParamsChange={handleParamsChange} />

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
            {selectedCount > 0 ? (
              <Button
                className={STYLE.btnPrimary}
                onClick={handleSyncClick}
                disabled={isLoading}
              >
                <Tags className={`mr-1.5 ${ICON.sm}`} />
                Lステップに連携
              </Button>
            ) : null}
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
        <LtvOwnerTable
          owners={owners}
          selectedOwnerIds={selectedOwnerIds}
          onSelectAll={handleSelectAll}
          onSelectOwner={handleSelectOwner}
          isLoading={isLoading}
        />

        {/* ページ情報 */}
        {data ? (
          <div className={`text-sm ${C.text50} text-right`}>
            {data.page}ページ / {Math.ceil(data.total / data.per_page)}ページ（全{data.total}件）
          </div>
        ) : null}
      </div>

      {/* Lステップ連携ダイアログ */}
      <LtvBulkSyncDialog
        open={syncDialogOpen}
        onOpenChange={setSyncDialogOpen}
        selectedOwnerIds={Array.from(selectedOwnerIds)}
        selectedCount={selectedCount}
      />
    </PageLayout>
  );
}
