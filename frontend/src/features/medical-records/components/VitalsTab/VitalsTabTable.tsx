import { Plus } from "lucide-react";

import { TableCell, TableHead } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { UpdateVitalInput, Vital } from "../../types";

import { VitalsAddRow, VitalsDisplayRow, VitalsEditRow } from "./VitalsTabRows";
import type { VitalsAddFormState } from "../../lib/vitals-tab-table-model";

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel）。
const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage} h-11`}>
      <TableHead className="w-40">記録日時</TableHead>
      <TableHead className="w-24 text-right">体温 (℃)</TableHead>
      <TableHead className="w-24 text-right">心拍数 (bpm)</TableHead>
      <TableHead className="w-24 text-right">呼吸数 (/min)</TableHead>
      <TableHead className="w-32 text-right">体重</TableHead>
      <TableHead>メモ</TableHead>
      <TableHead className="w-24 text-right">操作</TableHead>
    </tr>
  </thead>
);

interface VitalsTableProps {
  vitals: Vital[];
  editingId: string | null;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  isAdding: boolean;
  addForm: VitalsAddFormState;
  addFormErrors: Record<string, string>;
  addFormError?: string | null;
  createPending: boolean;
  updatePending: boolean;
  deletePending: boolean;
  onStartAdd: () => void;
  onAddFormChange: (patch: Partial<VitalsAddFormState>) => void;
  addFormAction: (payload: FormData) => void;
  onAddCancel: () => void;
  onStartEdit: (vitalId: string) => void;
  onEditSave: (vitalId: string, input: UpdateVitalInput) => void;
  onEditCancel: () => void;
  onDeleteClick: (vitalId: string) => void;
}

export function VitalsTable({
  vitals,
  editingId,
  canCreate,
  canEdit,
  canDelete,
  isAdding,
  addForm,
  addFormErrors,
  addFormError,
  createPending,
  updatePending,
  deletePending,
  onStartAdd,
  onAddFormChange,
  addFormAction,
  onAddCancel,
  onStartEdit,
  onEditSave,
  onEditCancel,
  onDeleteClick,
}: VitalsTableProps) {
  return (
    <div className={`${STYLE.tableContainer} overflow-x-auto`}>
      <table className="w-full">
        {TABLE_HEADER}
        <tbody>
          {vitals.length === 0 ? (
            <tr>
              <TableCell data-empty-state colSpan={7} className={STYLE.tableEmptySm}>
                バイタル記録がありません。下の「記録を追加」ボタンから追加してください。
              </TableCell>
            </tr>
          ) : (
            vitals.map((vital) =>
              editingId === vital.id ? (
                <VitalsEditRow
                  key={vital.id}
                  vital={vital}
                  onSave={onEditSave}
                  onCancel={onEditCancel}
                  isPending={updatePending}
                />
              ) : (
                <VitalsDisplayRow
                  key={vital.id}
                  vital={vital}
                  canEdit={canEdit}
                  canDelete={canDelete}
                  deletePending={deletePending}
                  onStartEdit={onStartEdit}
                  onDeleteClick={onDeleteClick}
                />
              ),
            )
          )}
        </tbody>
      </table>

      {isAdding ? (
        <VitalsAddRow
          addForm={addForm}
          errors={addFormErrors}
          actionError={addFormError}
          isPending={createPending}
          onChange={onAddFormChange}
          formAction={addFormAction}
          onCancel={onAddCancel}
        />
      ) : canCreate ? (
        <button type="button" className={STYLE.inlineAddBtn} onClick={onStartAdd}>
          <Plus className={ICON.xs} />
          <span>記録を追加</span>
        </button>
      ) : null}
    </div>
  );
}
