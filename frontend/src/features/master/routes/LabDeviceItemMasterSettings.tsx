import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
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
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceLabImport } from "@/types/generated/models";

import { useGetAllExaminationTypes } from "../api/exam-types-master";
import {
  useCreateLabDevice,
  useGetLabDevices,
  useUpdateLabDevice,
} from "../api/lab-devices";
import {
  useEnsureLabDeviceItemMasters,
  useGetLabDeviceItemMasters,
} from "../api/lab-device-item-masters";
import { LabDeviceItemMasterSidePanel } from "../components/LabDeviceItemMasterSidePanel";
import { MASTER_TABLE_COL } from "../constants/styles";
import {
  availableLabDeviceSourceTypes,
  buildLabDeviceCreateRequest,
  buildLabDeviceUpdateRequest,
  itemsForLabDevice,
  parseLabDeviceSourceQuery,
  toLabDeviceRows,
  validateLabDeviceDraft,
  type LabDeviceFormData,
  type LabDeviceRow,
} from "./lab-device-item-master-settings-model";

const COLUMNS = [
  { header: "機器", className: "flex-1" },
  { header: "検査", className: "flex-1" },
  { header: "項目数", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "未設定", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function LabDeviceItemMasterSettings() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission(ResourceLabImport);
  const { data: devices = [] } = useGetLabDevices();
  const { data: items = [] } = useGetLabDeviceItemMasters();
  const { data: examTypes = [] } = useGetAllExaminationTypes();
  const ensureMutation = useEnsureLabDeviceItemMasters();
  const createMutation = useCreateLabDevice();
  const updateMutation = useUpdateLabDevice();
  const dirty = useSidePeekDirty();
  const sourceFromQuery = parseLabDeviceSourceQuery(searchParams.get("source"));
  const fromBoard = searchParams.get("from") === "board";
  const [selectedId, setSelectedId] = useState<string | "new" | null>(null);

  const rows = useMemo(
    () => toLabDeviceRows(devices, items, examTypes),
    [devices, examTypes, items],
  );
  const unusedSourceTypes = useMemo(
    () => availableLabDeviceSourceTypes(devices),
    [devices],
  );
  const selectedRow = selectedId === null || selectedId === "new"
    ? null
    : rows.find((row) => row.id === selectedId) ?? null;
  const isCreating = selectedId === "new";
  const selectedItems = useMemo(
    () => (selectedRow === null ? [] : itemsForLabDevice(items, selectedRow.sourceType)),
    [items, selectedRow],
  );
  const showPanel = isCreating || selectedRow !== null;
  const readOnly = isCreating ? canCreate !== true : canEdit !== true;

  useEffect(() => {
    if (sourceFromQuery === null) {
      return;
    }
    const match = rows.find((row) => row.sourceType === sourceFromQuery);
    if (match !== undefined) {
      setSelectedId(match.id);
    }
  }, [rows, sourceFromQuery]);

  const handleClose = useCallback(() => {
    if (!dirty.confirmDiscard()) {
      return;
    }
    setSelectedId(null);
    if (searchParams.has("source")) {
      const next = new URLSearchParams(searchParams);
      next.delete("source");
      setSearchParams(next, { replace: true });
    }
    dirty.markClean();
  }, [dirty, searchParams, setSearchParams]);

  const handleEdit = useCallback((row: LabDeviceRow) => {
    if (!dirty.confirmDiscard()) {
      return;
    }
    setSelectedId(row.id);
  }, [dirty]);

  const handleNew = useCallback(() => {
    if (!dirty.confirmDiscard()) {
      return;
    }
    if (unusedSourceTypes.length === 0) {
      toast.error("対応プロトコルはすべて登録済みです");
      return;
    }
    setSelectedId("new");
  }, [dirty, unusedSourceTypes.length]);

  const handleDirtyChange = useCallback((nextDirty: boolean) => {
    if (nextDirty) {
      dirty.markDirty();
      return;
    }
    dirty.markClean();
  }, [dirty]);

  const handleSave = useCallback(async (form: LabDeviceFormData) => {
    const error = validateLabDeviceDraft({
      name: form.name,
      sourceType: form.sourceType,
      examTypeId: form.examTypeId,
      requireSourceType: isCreating,
    });
    if (error !== null) {
      toast.error(error);
      return;
    }
    try {
      if (isCreating) {
        if (canCreate !== true) {
          return;
        }
        await createMutation.mutateAsync(buildLabDeviceCreateRequest(form));
        toast.success("登録しました");
      } else if (selectedRow !== null) {
        if (canEdit !== true) {
          return;
        }
        await updateMutation.mutateAsync({
          id: selectedRow.id,
          req: buildLabDeviceUpdateRequest(form),
        });
        toast.success("更新しました");
      }
    } catch {
      return;
    }
    dirty.markClean();
    setSelectedId(null);
  }, [canCreate, canEdit, createMutation, dirty, isCreating, selectedRow, updateMutation]);

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
            <div className="flex items-center gap-2">
              {canEdit ? (
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
              ) : null}
              {canCreate ? (
                <PrimaryButton onClick={handleNew}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null}
            </div>
          }
        >
          <div className="flex flex-col gap-4">
            {fromBoard ? (
              <Link to={paths.labDevice.getHref()} className={`text-sm underline ${C.text}`}>
                検査受信へ戻る
              </Link>
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
                      onClick={() => handleEdit(row)}
                    >
                      {row.name}
                    </DataTableRowButton>
                  </TableCell>
                  <TableCell className={C.text}>{row.examLabel}</TableCell>
                  <TableCell className={`text-center ${C.text}`}>{row.itemCount}</TableCell>
                  <TableCell className={`text-center ${C.text}`}>{row.unmappedCount}</TableCell>
                  <TableCell className="text-right">
                    <RowActionButton
                      onClick={() => handleEdit(row)}
                      aria-label={`検査機器「${row.name}」の詳細`}
                    />
                  </TableCell>
                </DataTableRow>
              )}
            />
          </div>
        </PageLayout>
      </div>
      {showPanel ? (
        <LabDeviceItemMasterSidePanel
          key={selectedId ?? "closed"}
          device={selectedRow}
          items={selectedItems}
          examTypes={examTypes}
          unusedSourceTypes={unusedSourceTypes}
          readOnly={readOnly}
          isPending={createMutation.isPending || updateMutation.isPending}
          onClose={handleClose}
          onSave={(form) => {
            void handleSave(form);
          }}
          onDirtyChange={handleDirtyChange}
        />
      ) : null}
    </div>
  );
}
