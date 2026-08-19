import { useCallback, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { FlaskConical } from "lucide-react";
import { toast } from "sonner";

import {
  DataTable,
  DESIGN_TABLE_HEADER_CELL,
  DESIGN_TABLE_HEADER_ROW,
} from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TableCell } from "@/components/ui/table";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import { useGetAllExaminationTypes } from "../api/exam-types-master";
import {
  useEnsureLabDeviceItemMasters,
  useGetLabDeviceItemMasters,
  useUpdateLabDeviceItemMaster,
} from "../api/lab-device-item-masters";
import { LabDeviceItemMasterSidePanel } from "../components/LabDeviceItemMasterSidePanel";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import {
  collectDirtyLabDeviceUpdates,
  filterLabDeviceRows,
  itemsForLabDevice,
  toLabDeviceRows,
  type LabDeviceItemDraft,
  type LabDeviceRow,
} from "./lab-device-item-master-settings-model";

const COLUMNS = [
  { header: "機器", className: "flex-1" },
  { header: "項目数", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "未設定", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function LabDeviceItemMasterSettings() {
  const navigate = useNavigate();
  const { canEdit } = usePermission(ResourceLabImport);
  const { data: items = [] } = useGetLabDeviceItemMasters();
  const { data: examTypes = [] } = useGetAllExaminationTypes();
  const updateMutation = useUpdateLabDeviceItemMaster();
  const ensureMutation = useEnsureLabDeviceItemMasters();
  const dirty = useSidePeekDirty();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [selectedSource, setSelectedSource] = useState<string | null>(null);

  const rows = useMemo(() => toLabDeviceRows(items), [items]);
  const filteredRows = useMemo(
    () => filterLabDeviceRows(rows, searchTerm, activeFilters),
    [rows, searchTerm, activeFilters],
  );
  const selectedRow = rows.find((row) => row.sourceType === selectedSource) ?? null;
  const selectedItems = useMemo(
    () => (selectedSource === null ? [] : itemsForLabDevice(items, selectedSource)),
    [items, selectedSource],
  );

  const handleDirtyChange = useCallback((nextDirty: boolean) => {
    if (nextDirty) {
      dirty.markDirty();
      return;
    }
    dirty.markClean();
  }, [dirty]);

  const handleClose = useCallback(() => {
    if (!dirty.confirmDiscard()) {
      return;
    }
    setSelectedSource(null);
  }, [dirty]);

  const handleEdit = useCallback((row: LabDeviceRow) => {
    if (!dirty.confirmDiscard()) {
      return;
    }
    setSelectedSource(row.sourceType);
  }, [dirty]);

  const handleSave = useCallback(async (drafts: LabDeviceItemDraft[]) => {
    if (selectedSource === null) {
      return false;
    }
    const { error, updates } = collectDirtyLabDeviceUpdates(
      itemsForLabDevice(items, selectedSource),
      drafts,
    );
    if (error !== null) {
      toast.error(error);
      return false;
    }
    try {
      for (const update of updates) {
        await updateMutation.mutateAsync(update);
      }
      if (updates.length > 0) {
        toast.success("保存しました");
      }
      dirty.markClean();
      return true;
    } catch {
      return false;
    }
  }, [dirty, items, selectedSource, updateMutation]);

  return (
    <div className="flex h-full">
      <div className="flex-1 min-w-0">
        <PageLayout
          title="検査機器マスタ"
          icon={<FlaskConical className={`${ICON.page} ${C.text}`} />}
          resource={ResourceLabImport}
          onBack={() => navigate(paths.settings.getHref())}
          maxWidth={LAYOUT.pageContentMaxWidth.full}
          headerAction={
            canEdit ? (
              <PrimaryButton
                onClick={() => {
                  ensureMutation.mutate(undefined, {
                    onSuccess: (result) => {
                      toast.success(
                        result.insertedCount > 0
                          ? `既定項目を ${result.insertedCount} 件用意しました`
                          : "既定項目は揃っています",
                      );
                    },
                  });
                }}
                disabled={ensureMutation.isPending}
              >
                既定項目を用意
              </PrimaryButton>
            ) : null
          }
        >
          <div className="flex flex-col gap-4">
            <PropertyFilter
              properties={[MASTER_STATUS_FILTER]}
              activeFilters={activeFilters}
              onFilterChange={setActiveFilters}
              searchTerm={searchTerm}
              onSearchChange={setSearchTerm}
              searchPlaceholder="機器名で検索..."
              count={filteredRows.length}
            />
            <DataTable
              headerRowClassName={DESIGN_TABLE_HEADER_ROW}
              headerCellClassName={DESIGN_TABLE_HEADER_CELL}
              columns={COLUMNS}
              data={filteredRows}
              emptyMessage="該当する機器がありません"
              renderRow={(row) => (
                <DataTableRow key={row.id}>
                  <TableCell className={`font-medium ${C.text}`}>
                    <DataTableRowButton
                      aria-label={`詳細: 検査機器 ${row.name}`}
                      onClick={() => handleEdit(row)}
                    >
                      {row.name}
                    </DataTableRowButton>
                  </TableCell>
                  <TableCell className={`text-center ${C.text}`}>{row.itemCount}</TableCell>
                  <TableCell className={`text-center ${C.text}`}>{row.unmappedCount}</TableCell>
                  <TableCell className="text-center">
                    <StatusPill isActive={row.isActive} />
                  </TableCell>
                  <TableCell className="text-right">
                    <RowActionButton
                      onClick={() => handleEdit(row)}
                      aria-label={
                        canEdit
                          ? `検査機器「${row.name}」を編集`
                          : `検査機器「${row.name}」の詳細`
                      }
                    />
                  </TableCell>
                </DataTableRow>
              )}
            />
          </div>
        </PageLayout>
      </div>
      {selectedRow !== null ? (
        <LabDeviceItemMasterSidePanel
          key={selectedRow.sourceType}
          device={selectedRow}
          items={selectedItems}
          examTypes={examTypes}
          onClose={handleClose}
          onSave={handleSave}
          readOnly={!canEdit}
          isPending={updateMutation.isPending}
          onDirtyChange={handleDirtyChange}
        />
      ) : null}
    </div>
  );
}
