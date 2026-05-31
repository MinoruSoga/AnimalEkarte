import { Plus } from "lucide-react";

import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { UpdateVitalInput, Vital } from "../../types";

import { VitalsAddRow, VitalsDisplayRow, VitalsEditRow } from "./VitalsTabRows";
import type { VitalsAddFormState } from "./VitalsTabTableModel";

export {
  EMPTY_VITALS_ADD_FORM,
  parseVitalsNumber,
  type VitalsAddFormState,
} from "./VitalsTabTableModel";

const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage30} h-10`}>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-40`}>記録日時</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>体温 (℃)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>心拍数 (bpm)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>呼吸数 (/min)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-32`}>体重</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70}`}>メモ</th>
      <th className={`px-2 text-right text-xs font-medium ${C.text70} w-24`}>操作</th>
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
  createPending: boolean;
  updatePending: boolean;
  deletePending: boolean;
  onStartAdd: () => void;
  onAddFormChange: (field: keyof VitalsAddFormState, value: string) => void;
  onAddSubmit: () => void;
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
  createPending,
  updatePending,
  deletePending,
  onStartAdd,
  onAddFormChange,
  onAddSubmit,
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
              <td colSpan={7} className={STYLE.tableEmptySm}>
                バイタル記録がありません。下の「記録を追加」ボタンから追加してください。
              </td>
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
              )
            )
          )}
        </tbody>
      </table>

      {isAdding ? (
        <VitalsAddRow
          addForm={addForm}
          errors={addFormErrors}
          isPending={createPending}
          onChange={onAddFormChange}
          onSubmit={onAddSubmit}
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
