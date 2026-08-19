import { memo, useCallback, useState } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";

import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { useGetStaffs } from "@/hooks/use-staffs";
import { useGetAllCheckupTypes } from "@/hooks/use-treatment-master";
import {
  useCreateCheckup,
  useDeleteCheckup,
  useGetCheckups,
  useUpdateCheckup,
  type CreateCheckupInput,
  type UpdateCheckupInput,
} from "../../api/checkups";
import {
  CheckupsTable,
  LstepStatusBadge,
  type LstepStatus,
} from "./CheckupsTabTable";
import {
  makeDefaultCheckupAddForm,
  type AddCheckupFormState,
} from "./checkups-tab-table-model";

interface CheckupsTabProps {
  medicalRecordId: string;
  lstepStatus?: LstepStatus;
  isFinalized?: boolean;
}

export const CheckupsTab = memo(function CheckupsTab({
  medicalRecordId,
  lstepStatus,
  isFinalized = false,
}: CheckupsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const { data: checkups, isLoading } = useGetCheckups(medicalRecordId);
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: staffs = [] } = useGetStaffs();
  const createMutation = useCreateCheckup(medicalRecordId);
  const updateMutation = useUpdateCheckup(medicalRecordId);
  const deleteMutation = useDeleteCheckup(medicalRecordId);

  const [searchParams] = useSearchParams();
  const [editingId, setEditingId] = useState<string | null>(
    () => searchParams.get("checkupId"),
  );
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddCheckupFormState>(() => makeDefaultCheckupAddForm());
  const [addFormErrors, setAddFormErrors] = useState<Record<string, string>>({});

  const handleAddFormChange = useCallback(
    (field: keyof AddCheckupFormState, value: string) => {
      setAddForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleAddSubmit = useCallback(() => {
    if (!canCreate) return;
    const errors: Record<string, string> = {};
    if (!addForm.date) errors.date = "日付は必須です";
    if (!addForm.checkup_type_id) errors.checkup_type_id = "健診種別は必須です";
    if (Object.keys(errors).length > 0) {
      setAddFormErrors(errors);
      return;
    }
    setAddFormErrors({});

    const input: CreateCheckupInput = {
      checkup_type_id: Number(addForm.checkup_type_id),
      date: addForm.date,
      next_date: addForm.next_date || null,
      doctor_id: addForm.doctor_id ? Number(addForm.doctor_id) : null,
      result: addForm.result,
    };
    createMutation.mutate(input, {
      onSuccess: () => {
        setAddForm(makeDefaultCheckupAddForm());
        setIsAdding(false);
        toast.success("健診記録を追加しました");
      },
    });
  }, [addForm, canCreate, createMutation]);

  const handleAddCancel = useCallback(() => {
    setAddForm(makeDefaultCheckupAddForm());
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (checkupId: string, input: UpdateCheckupInput) => {
      if (!canEdit) return;
      updateMutation.mutate(
        { checkupId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("健診記録を更新しました");
          },
        },
      );
    },
    [canEdit, updateMutation],
  );

  const handleDelete = useCallback(
    (checkupId: string) => {
      if (!canDelete) return;
      deleteMutation.mutate(checkupId, {
        onSuccess: () => {
          toast.success("健診記録を削除しました");
        },
      });
    },
    [canDelete, deleteMutation],
  );

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        読み込み中...
      </div>
    );
  }

  const checkupList = checkups ?? [];

  return (
    <div className="flex flex-col gap-3 pb-24">
      {lstepStatus !== undefined ? (
        <div className="flex items-center gap-2">
          <LstepStatusBadge status={lstepStatus} />
        </div>
      ) : null}

      <CheckupsTable
        checkups={checkupList}
        editingId={editingId}
        isFinalized={isFinalized}
        isAdding={isAdding}
        addForm={addForm}
        addFormErrors={addFormErrors}
        checkupTypes={checkupTypes}
        staffs={staffs}
        canCreate={canCreate}
        canEdit={canEdit}
        canDelete={canDelete}
        createPending={createMutation.isPending}
        updatePending={updateMutation.isPending}
        deletePending={deleteMutation.isPending}
        onStartAdd={() => setIsAdding(true)}
        onAddFormChange={handleAddFormChange}
        onAddSubmit={handleAddSubmit}
        onAddCancel={handleAddCancel}
        onStartEdit={setEditingId}
        onEditSave={handleEditSave}
        onEditCancel={() => setEditingId(null)}
        onDelete={handleDelete}
      />

      {checkupList.length > 0 ? (
        <div className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-3`}>
          <span className={`text-sm ${C.text60}`}>
            健診記録 {checkupList.length} 件
          </span>
        </div>
      ) : null}
    </div>
  );
});
