import { Plus } from "lucide-react";

import { TableCell, TableHead } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { StaffItem } from "@/hooks/use-staffs";
import type { CheckupTypeItem } from "@/hooks/use-treatment-master";
import type { CheckupTypeFieldRow } from "@/features/checkups/api/get-checkup-type-fields";
import type { CheckupFieldValue } from "@/features/checkups/components/DynamicCheckupFields";
import type { Checkup, UpdateCheckupInput } from "../../api/checkups";
import type { AddCheckupFormState } from "./checkups-tab-table-model";
import { CheckupAddRow, CheckupDisplayRow, CheckupEditRow } from "./CheckupsTabRows";

export { LstepStatusBadge, type LstepStatus } from "./CheckupsTabBadges";

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel）。
const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage} h-11`}>
      <TableHead className="min-w-[10rem] w-40">日付</TableHead>
      <TableHead className="w-40">健診種別</TableHead>
      <TableHead className="min-w-[10rem] w-40">次回の予定</TableHead>
      <TableHead className="w-32">担当医</TableHead>
      <TableHead>結果</TableHead>
      <TableHead className="w-24 text-right">操作</TableHead>
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
  checkupFields: CheckupTypeFieldRow[];
  fieldValues: Record<number, CheckupFieldValue>;
  onStartAdd: () => void;
  onAddFormChange: (field: keyof AddCheckupFormState, value: string) => void;
  onFieldValueChange: (fieldId: number, value: CheckupFieldValue) => void;
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
  checkupFields,
  fieldValues,
  onStartAdd,
  onAddFormChange,
  onFieldValueChange,
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
              <TableCell data-empty-state colSpan={6} className={STYLE.tableEmptySm}>
                健診記録がありません。下の「記録を追加」ボタンから追加してください。
              </TableCell>
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
          checkupFields={checkupFields}
          fieldValues={fieldValues}
          onChange={onAddFormChange}
          onFieldValueChange={onFieldValueChange}
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
