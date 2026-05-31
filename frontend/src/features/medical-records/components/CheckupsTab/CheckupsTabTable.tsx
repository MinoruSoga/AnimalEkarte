import { Plus } from "lucide-react";

import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { StaffItem } from "@/hooks/use-staffs";
import type { CheckupTypeItem } from "@/hooks/use-treatment-master";
import type { Checkup, UpdateCheckupInput } from "../../api/checkups";
import type { AddCheckupFormState } from "./CheckupsTabTableModel";
import { CheckupAddRow, CheckupDisplayRow, CheckupEditRow } from "./CheckupsTabRows";

export { LstepStatusBadge, type LstepStatus } from "./CheckupsTabBadges";
export { makeDefaultCheckupAddForm, type AddCheckupFormState } from "./CheckupsTabTableModel";

const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage30} h-10`}>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-32`}>日付</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-40`}>健診種別</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-32`}>次回予定日</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-32`}>担当医</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70}`}>結果</th>
      <th className={`px-2 text-right text-xs font-medium ${C.text70} w-24`}>操作</th>
    </tr>
  </thead>
);

interface CheckupsTableProps {
  checkups: Checkup[];
  editingId: string | null;
  isFinalized: boolean;
  isAdding: boolean;
  addForm: AddCheckupFormState;
  addFormErrors: Record<string, string>;
  checkupTypes: CheckupTypeItem[];
  staffs: StaffItem[];
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  createPending: boolean;
  updatePending: boolean;
  deletePending: boolean;
  onStartAdd: () => void;
  onAddFormChange: (field: keyof AddCheckupFormState, value: string) => void;
  onAddSubmit: () => void;
  onAddCancel: () => void;
  onStartEdit: (checkupId: string) => void;
  onEditSave: (checkupId: string, input: UpdateCheckupInput) => void;
  onEditCancel: () => void;
  onDelete: (checkupId: string) => void;
}

export function CheckupsTable({
  checkups,
  editingId,
  isFinalized,
  isAdding,
  addForm,
  addFormErrors,
  checkupTypes,
  staffs,
  canCreate,
  canEdit,
  canDelete,
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
  onDelete,
}: CheckupsTableProps) {
  return (
    <div className={`${STYLE.tableContainer} overflow-x-auto`}>
      <table className="w-full">
        {TABLE_HEADER}
        <tbody>
          {checkups.length === 0 ? (
            <tr>
              <td colSpan={6} className={STYLE.tableEmptySm}>
                健診記録がありません。下の「記録を追加」ボタンから追加してください。
              </td>
            </tr>
          ) : (
            checkups.map((checkup) =>
              editingId === checkup.id ? (
                <CheckupEditRow
                  key={checkup.id}
                  checkup={checkup}
                  onSave={onEditSave}
                  onCancel={onEditCancel}
                  isPending={updatePending}
                  checkupTypes={checkupTypes}
                  staffs={staffs}
                />
              ) : (
                <CheckupDisplayRow
                  key={checkup.id}
                  checkup={checkup}
                  canEdit={canEdit}
                  canDelete={canDelete}
                  isFinalized={isFinalized}
                  deletePending={deletePending}
                  onStartEdit={onStartEdit}
                  onDelete={onDelete}
                />
              )
            )
          )}
        </tbody>
      </table>

      {isFinalized ? (
        <div className={`px-4 py-3 text-sm ${C.text60} border-t ${C.borderLight}`}>
          確定済みカルテのため健診情報は編集できません
        </div>
      ) : null}

      {!isFinalized && isAdding ? (
        <CheckupAddRow
          addForm={addForm}
          errors={addFormErrors}
          checkupTypes={checkupTypes}
          staffs={staffs}
          isPending={createPending}
          onChange={onAddFormChange}
          onSubmit={onAddSubmit}
          onCancel={onAddCancel}
        />
      ) : canCreate && !isFinalized ? (
        <button type="button" className={STYLE.inlineAddBtn} onClick={onStartAdd}>
          <Plus className={ICON.xs} />
          <span>記録を追加</span>
        </button>
      ) : null}
    </div>
  );
}
