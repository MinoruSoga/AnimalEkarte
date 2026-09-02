import { Link, useNavigate } from "react-router";
import { FlaskConical, Plus } from "lucide-react";
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
import { RowActionButton } from "@/components/shared/RowActionButton";
import { TableCell } from "@/components/ui/table";
import { paths } from "@/config/paths";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import { LabDeviceItemMasterSidePanel } from "../components/LabDeviceItemMasterSidePanel";
import { MASTER_TABLE_COL } from "../constants/styles";
import { labDeviceSourceLabel, type LabDeviceRow } from "./lab-device-item-master-settings-model";
import { useLabDeviceItemMasterSettings } from "./use-lab-device-item-master-settings";

const COLUMNS = [
  { header: "機器", className: "flex-1" },
  { header: "検査", className: "flex-1" },
  { header: "項目数", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "未設定", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function LabDeviceItemMasterSettings() {
  const navigate = useNavigate();
  const s = useLabDeviceItemMasterSettings();

  return (
    <>
    <div className="flex h-full">
      <div className="flex-1 min-w-0">
        <PageLayout
          title="検査機器マスタ"
          icon={<FlaskConical className={`${ICON.page} ${C.text}`} />}
          resource={ResourceLabImport}
          onBack={() => navigate(paths.settings.getHref())}
          maxWidth={LAYOUT.pageContentMaxWidth.full}
          headerAction={
            <LabDeviceItemMasterHeaderActions
              canEdit={s.canEdit}
              canCreate={s.canCreate}
              ensurePending={s.ensureMutation.isPending}
              onEnsure={() => {
                s.ensureMutation.mutate(undefined, {
                  onSuccess: (result) => {
                    toast.success(
                      result.insertedCount > 0
                        ? `既定項目を ${result.insertedCount} 件用意しました`
                        : "既定項目は揃っています",
                    );
                  },
                });
              }}
              onNew={s.handleNew}
            />
          }
        >
          <LabDeviceItemMasterTable
            fromBoard={s.fromBoard}
            sourceFromQuery={s.sourceFromQuery}
            devicesFetched={s.devicesFetched}
            rows={s.rows}
            onEdit={s.handleEdit}
          />
        </PageLayout>
      </div>
      {s.showPanel ? (
        <LabDeviceItemMasterSidePanel
          key={s.selectedId ?? "closed"}
          device={s.selectedRow}
          items={s.selectedItems}
          examTypes={s.examTypes}
          unusedSourceTypes={s.unusedSourceTypes}
          readOnly={s.readOnly}
          isPending={s.createMutation.isPending || s.saveConfigurationMutation.isPending}
          onClose={s.handleClose}
          onSave={(form, drafts) => {
            void s.handleSave(form, drafts);
          }}
          onDirtyChange={s.handleDirtyChange}
        />
      ) : null}
    </div>
    {s.dirty.discardDialog}
    </>
  );
}

interface LabDeviceItemMasterHeaderActionsProps {
  canEdit: boolean;
  canCreate: boolean;
  ensurePending: boolean;
  onEnsure: () => void;
  onNew: () => void;
}

function LabDeviceItemMasterHeaderActions({
  canEdit,
  canCreate,
  ensurePending,
  onEnsure,
  onNew,
}: LabDeviceItemMasterHeaderActionsProps) {
  return (
    <div className="flex items-center gap-2">
      {canEdit ? (
        <PrimaryButton onClick={onEnsure} disabled={ensurePending}>
          既定項目を用意
        </PrimaryButton>
      ) : null}
      {canCreate ? (
        <PrimaryButton onClick={onNew}>
          <Plus className={`mr-1.5 ${ICON.action}`} />
          新規登録
        </PrimaryButton>
      ) : null}
    </div>
  );
}

interface LabDeviceItemMasterTableProps {
  fromBoard: boolean;
  sourceFromQuery: string | null;
  devicesFetched: boolean;
  rows: LabDeviceRow[];
  onEdit: (row: LabDeviceRow) => void;
}

function LabDeviceItemMasterTable({
  fromBoard,
  sourceFromQuery,
  devicesFetched,
  rows,
  onEdit,
}: LabDeviceItemMasterTableProps) {
  return (
    <div className="flex flex-col gap-4">
      {fromBoard ? (
        <Link to={paths.labDevice.getHref()} className={`text-sm underline ${C.text}`}>
          検査受信へ戻る
        </Link>
      ) : null}
      {sourceFromQuery !== null
        && devicesFetched
        && !rows.some((row) => row.sourceType === sourceFromQuery) ? (
          <p className={`text-sm ${C.textWarning}`}>
            {labDeviceSourceLabel(sourceFromQuery)} はまだ登録されていません。「既定項目を用意」で投入してください
          </p>
        ) : null}
      <DataTable
        headerRowClassName={DESIGN_TABLE_HEADER_ROW}
        headerCellClassName={DESIGN_TABLE_HEADER_CELL}
        columns={COLUMNS}
        data={rows}
        emptyMessage="機器がありません。新規登録するか、既定項目を用意してください"
        renderRow={(row) => (
          <DataTableRow key={row.id}>
            <TableCell className={`font-medium ${C.text}`}>
              <DataTableRowButton
                aria-label={`詳細: 検査機器 ${row.name}`}
                onClick={() => onEdit(row)}
              >
                {row.name}
              </DataTableRowButton>
            </TableCell>
            <TableCell className={C.text}>{row.examLabel}</TableCell>
            <TableCell className={`text-center ${C.text}`}>{row.itemCount}</TableCell>
            <TableCell className={`text-center ${C.text}`}>{row.unmappedCount}</TableCell>
            <TableCell className="text-right">
              <RowActionButton
                onClick={() => onEdit(row)}
                aria-label={`検査機器「${row.name}」の詳細`}
              />
            </TableCell>
          </DataTableRow>
        )}
      />
    </div>
  );
}
