import { memo, useCallback, useState } from "react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { LoadingFallback } from "@/components/shared/DataStates";
import { C } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";
import { useGetStaffs } from "@/hooks/use-staffs";
import { useGetAllCheckupTypes } from "@/hooks/use-treatment-master";
import { replaceCheckupFieldResults, useGetCheckupTypeFields } from "@/hooks/use-checkup-fields";
import {
  buildCheckupResultsPayload,
  type CheckupFieldValue,
} from "@/components/shared/DynamicCheckupFields/DynamicCheckupFields";
import {
  useCreateCheckup,
  useDeleteCheckup,
  useGetCheckups,
  useUpdateCheckup,
  type Checkup,
  type CreateCheckupInput,
  type UpdateCheckupInput,
} from "../../api/checkups";
import { CheckupsTable, LstepStatusBadge, type LstepStatus } from "./CheckupsTabTable";
import {
  makeDefaultCheckupAddForm,
  type AddCheckupFormState,
} from "../../lib/checkups-tab-table-model";

const DECEASED_CHECKUPS_MESSAGE = "死亡したペットの健診記録は保存できません";

interface CheckupsTabProps {
  medicalRecordId: string;
  lstepStatus?: LstepStatus;
  isFinalized?: boolean;
  /** 死亡ペットのカルテは閲覧専用（VitalsTab と同じ二重ガード方針） */
  isPetDeceased?: boolean;
}

export const CheckupsTab = memo(function CheckupsTab({
  medicalRecordId,
  lstepStatus,
  isFinalized = false,
  isPetDeceased = false,
}: CheckupsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const { data: checkups, isLoading } = useGetCheckups(medicalRecordId);
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const { data: staffs = [] } = useGetStaffs();
  const createMutation = useCreateCheckup(medicalRecordId);
  const updateMutation = useUpdateCheckup(medicalRecordId);
  const deleteMutation = useDeleteCheckup(medicalRecordId);
  const { mutateAsync: createCheckupAsync } = createMutation;
  const { mutate: updateCheckup } = updateMutation;
  const { mutate: deleteCheckup } = deleteMutation;

  const [searchParams] = useSearchParams();
  const [editingId, setEditingId] = useState<string | null>(() => searchParams.get("checkupId"));
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddCheckupFormState>(() => makeDefaultCheckupAddForm());
  const [addFormErrors, setAddFormErrors] = useState<Record<string, string>>({});
  const [fieldValues, setFieldValues] = useState<Record<number, CheckupFieldValue>>({});
  const { data: checkupFields = [] } = useGetCheckupTypeFields(addForm.checkup_type_id);

  const handleAddFormChange = useCallback((field: keyof AddCheckupFormState, value: string) => {
    setAddForm((prev) => ({ ...prev, [field]: value }));
    if (field === "checkup_type_id") {
      setFieldValues({});
    }
  }, []);

  const handleFieldValueChange = useCallback((fieldId: number, value: CheckupFieldValue) => {
    setFieldValues((prev) => ({ ...prev, [fieldId]: value }));
  }, []);

  const handleAddSubmit = useCallback(async () => {
    if (isPetDeceased) {
      toast.error(DECEASED_CHECKUPS_MESSAGE);
      return;
    }
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

    let created: Checkup;
    try {
      created = await createCheckupAsync(input);
    } catch {
      return;
    }

    const resultsPayload = buildCheckupResultsPayload(checkupFields, fieldValues);
    if (resultsPayload.length > 0) {
      try {
        await replaceCheckupFieldResults(medicalRecordId, created.id, resultsPayload);
      } catch (error) {
        handleApiError(error, "健診項目の保存");
        return;
      }
    }

    setAddForm(makeDefaultCheckupAddForm());
    setFieldValues({});
    setIsAdding(false);
    toast.success("健診記録を追加しました");
  }, [
    addForm,
    canCreate,
    checkupFields,
    createCheckupAsync,
    fieldValues,
    isPetDeceased,
    medicalRecordId,
  ]);

  const handleAddCancel = useCallback(() => {
    setAddForm(makeDefaultCheckupAddForm());
    setFieldValues({});
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (checkupId: string, input: UpdateCheckupInput) => {
      if (isPetDeceased) {
        toast.error(DECEASED_CHECKUPS_MESSAGE);
        return;
      }
      if (!canEdit) return;
      updateCheckup(
        { checkupId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("健診記録を更新しました");
          },
        },
      );
    },
    [canEdit, isPetDeceased, updateCheckup],
  );

  const handleDeleteConfirm = useCallback(() => {
    if (isPetDeceased) {
      toast.error(DECEASED_CHECKUPS_MESSAGE);
      return;
    }
    if (!canDelete || !deletingId) return;
    deleteCheckup(deletingId, {
      onSuccess: () => {
        setDeletingId(null);
        toast.success("健診記録を削除しました");
      },
    });
  }, [canDelete, deletingId, deleteCheckup, isPetDeceased]);

  if (isLoading) {
    return <LoadingFallback />;
  }

  const checkupList = checkups ?? [];

  return (
    <div className="flex flex-col gap-3 pb-24 flex-1 min-h-0 overflow-y-auto relative">
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
        canCreate={Boolean(canCreate && !isPetDeceased)}
        canEdit={Boolean(canEdit && !isPetDeceased)}
        canDelete={Boolean(canDelete && !isPetDeceased)}
        createPending={createMutation.isPending}
        updatePending={updateMutation.isPending}
        deletePending={deleteMutation.isPending}
        checkupFields={checkupFields}
        fieldValues={fieldValues}
        onStartAdd={() => setIsAdding(true)}
        onAddFormChange={handleAddFormChange}
        onFieldValueChange={handleFieldValueChange}
        onAddSubmit={handleAddSubmit}
        onAddCancel={handleAddCancel}
        onStartEdit={setEditingId}
        onEditSave={handleEditSave}
        onEditCancel={() => setEditingId(null)}
        onDelete={setDeletingId}
      />

      {checkupList.length > 0 ? (
        <div className={`${C.bgWhite} border ${C.borderLight} rounded-lg px-4 py-3`}>
          <span className={`text-sm ${C.text60}`}>健診記録 {checkupList.length} 件</span>
        </div>
      ) : null}

      <ConfirmDialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="健診記録を削除しますか？"
        description="この健診記録を削除します。この操作は元に戻せません。"
        confirmLabel={deleteMutation.isPending ? "削除中..." : "削除する"}
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={deleteMutation.isPending}
      />
    </div>
  );
});
