// React/Framework
import { useState, useCallback, memo, useMemo, type ChangeEvent } from "react";

// External
import { Pencil, Plus, Check, X } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker/NotionDatePicker";

// Relative
import { usePermission } from "@/features/auth";
import { useGetCheckups } from "@/features/medical-records/api/checkups";
import { useCreateCheckup } from "@/features/medical-records/api/checkups";
import { useUpdateCheckup } from "@/features/medical-records/api/checkups";
import { useDeleteCheckup } from "@/features/medical-records/api/checkups";
import { useGetAllCheckupTypes } from "@/features/master";
import type { Checkup, CreateCheckupInput, UpdateCheckupInput } from "@/features/medical-records/api/checkups";
import type { CheckupTypeItem } from "@/features/master";

// ── 静的定数 ────────────────────────────────────────────────────────────

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

// ── 追加フォームの初期状態 ─────────────────────────────────────────────

interface AddFormState {
  checkup_type_id: string;
  date: string;
  next_date: string;
  result: string;
}

const EMPTY_ADD_FORM: AddFormState = {
  checkup_type_id: "",
  date: "",
  next_date: "",
  result: "",
};

// BUG-090: 実施日のデフォルト値として本日の日付を返す
function todayISODate(): string {
  return new Date().toISOString().slice(0, 10);
}

function makeDefaultAddForm(): AddFormState {
  return { ...EMPTY_ADD_FORM, date: todayISODate() };
}

// ── 編集行コンポーネント ───────────────────────────────────────────────

interface EditRowProps {
  checkup: Checkup;
  onSave: (checkupId: string, input: UpdateCheckupInput) => void;
  onCancel: () => void;
  isPending: boolean;
  checkupTypes: CheckupTypeItem[];
}

const EditRow = memo(function EditRow({ checkup, onSave, onCancel, isPending, checkupTypes }: EditRowProps) {
  const [form, setForm] = useState<UpdateCheckupInput>({
    checkup_type_id: Number(checkup.checkup_type_id),
    date: checkup.date,
    next_date: checkup.next_date ?? "",
    result: checkup.result,
  });

  const handleChange = useCallback(
    (field: keyof UpdateCheckupInput, value: string | number | null) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleSave = useCallback(() => {
    onSave(checkup.id, form);
  }, [checkup.id, form, onSave]);

  return (
    <tr className={`border-b ${C.borderLight} ${C.bgNotice40}`}>
      <td className="px-3 py-2">
        <NotionDatePicker
          value={form.date ?? ""}
          onChange={(v) => handleChange("date", v)}
          placeholder="日付"
          className="h-8 w-full"
        />
      </td>
      <td className="px-3 py-2">
        <select
          value={form.checkup_type_id ?? ""}
          onChange={(e: ChangeEvent<HTMLSelectElement>) =>
            handleChange("checkup_type_id", e.target.value ? Number(e.target.value) : null)
          }
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent} w-full`}
        >
          <option value="">選択してください</option>
          {checkupTypes.map((type: CheckupTypeItem) => (
            <option key={type.id} value={type.id}>
              {type.name}
            </option>
          ))}
        </select>
      </td>
      <td className="px-3 py-2">
        <NotionDatePicker
          value={form.next_date ?? ""}
          onChange={(v) => handleChange("next_date", v || null)}
          placeholder="次回日"
          className="h-8 w-full"
        />
      </td>
      <td className="px-3 py-2">
        <span className={`text-sm ${C.text60}`}>-</span>
      </td>
      <td className="px-3 py-2">
        <input
          type="text"
          value={form.result ?? ""}
          onChange={(e) => handleChange("result", e.target.value)}
          placeholder="結果を入力..."
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent} w-full`}
        />
      </td>
      <td className="px-2 py-2">
        <div className="flex items-center justify-end gap-1">
          <button
            onClick={handleSave}
            disabled={isPending}
            className={`${STYLE.iconBtn32} ${C.textStatusGreen} ${C.hoverBgStatusGreen}`}
            title="保存"
          >
            <Check className={`${ICON.xs}`} />
          </button>
          <button
            onClick={onCancel}
            disabled={isPending}
            className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverBgLight}`}
            title="キャンセル"
          >
            <X className={`${ICON.xs}`} />
          </button>
        </div>
      </td>
    </tr>
  );
});

// ── Props ──────────────────────────────────────────────────────────────

interface CheckupsTabProps {
  medicalRecordId: string;
}

// ── Component ──────────────────────────────────────────────────────────

export const CheckupsTab = memo(function CheckupsTab({ medicalRecordId }: CheckupsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const { data: checkups, isLoading } = useGetCheckups(medicalRecordId);
  const { data: checkupTypes = [] } = useGetAllCheckupTypes();
  const createMutation = useCreateCheckup(medicalRecordId);
  const { mutate: createCheckupFn } = createMutation;
  const updateMutation = useUpdateCheckup(medicalRecordId);
  const { mutate: updateCheckupFn } = updateMutation;
  const deleteMutation = useDeleteCheckup(medicalRecordId);
  const { mutate: deleteCheckupFn } = deleteMutation;

  const [editingId, setEditingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  // BUG-090: lazy init で本日日付をデフォルト値にセット
  const [addForm, setAddForm] = useState<AddFormState>(() => makeDefaultAddForm());

  // js-cache-function-results: checkupTypes 由来の option リストをキャッシュ
  const checkupTypeOptions = useMemo(
    () =>
      checkupTypes.map((type: CheckupTypeItem) => (
        <option key={type.id} value={type.id}>
          {type.name}
        </option>
      )),
    [checkupTypes],
  );

  // ── handlers ──

  const handleAddFormChange = useCallback(
    (field: keyof AddFormState, value: string) => {
      setAddForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleAddSubmit = useCallback(() => {
    if (!canCreate) return;
    if (!addForm.date || !addForm.checkup_type_id) {
      toast.error("日付と健診種別IDは必須です");
      return;
    }
    const input: CreateCheckupInput = {
      checkup_type_id: Number(addForm.checkup_type_id),
      date: addForm.date,
      next_date: addForm.next_date || null,
      result: addForm.result,
    };
    createCheckupFn(input, {
      onSuccess: () => {
        setAddForm(makeDefaultAddForm());
        setIsAdding(false);
        toast.success("健診記録を追加しました");
      },
    });
  }, [canCreate, addForm, createCheckupFn]);

  const handleAddCancel = useCallback(() => {
    setAddForm(EMPTY_ADD_FORM);
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (checkupId: string, input: UpdateCheckupInput) => {
      if (!canEdit) return;
      updateCheckupFn(
        { checkupId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("健診記録を更新しました");
          },
        }
      );
    },
    [canEdit, updateCheckupFn]
  );

  const handleEditCancel = useCallback(() => {
    setEditingId(null);
  }, []);

  const handleDelete = useCallback(
    (checkupId: string) => {
      if (!canDelete) return;
      deleteCheckupFn(checkupId, {
        onSuccess: () => {
          toast.success("健診記録を削除しました");
        },
      });
    },
    [canDelete, deleteCheckupFn]
  );

  // ── render ──

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
      <div className={`${STYLE.tableContainer} overflow-x-auto`}>
        <table className="w-full">
          {TABLE_HEADER}
          <tbody>
            {checkupList.length === 0 ? (
              <tr>
                <td colSpan={6} className={STYLE.tableEmptySm}>
                  健診記録がありません。下の「記録を追加」ボタンから追加してください。
                </td>
              </tr>
            ) : (
              checkupList.map((checkup) =>
                editingId === checkup.id ? (
                  <EditRow
                    key={checkup.id}
                    checkup={checkup}
                    onSave={handleEditSave}
                    onCancel={handleEditCancel}
                    isPending={updateMutation.isPending}
                    checkupTypes={checkupTypes}
                  />
                ) : (
                  <tr
                    key={checkup.id}
                    className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors h-12`}
                  >
                    <td className={`px-3 text-sm ${C.text}`}>{checkup.date}</td>
                    <td className={`px-3 text-sm ${C.text}`}>
                      {checkup.checkup_type?.name ?? checkup.checkup_type_id}
                    </td>
                    <td className={`px-3 text-sm ${C.text60}`}>
                      {checkup.next_date ? checkup.next_date : "-"}
                    </td>
                    <td className={`px-3 text-sm ${C.text60}`}>
                      {checkup.doctor?.name ?? (checkup.doctor_id ? checkup.doctor_id : "-")}
                    </td>
                    <td className={`px-3 text-sm ${C.text}`}>{checkup.result}</td>
                    <td className="px-2">
                      <div className="flex items-center justify-end gap-1">
                        {canEdit ? (
                          <button
                            onClick={() => setEditingId(checkup.id)}
                            className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverText} ${C.hoverBgLight}`}
                            title="編集"
                          >
                            <Pencil className={`${ICON.xs}`} />
                          </button>
                        ) : null}
                        {canDelete ? (
                          <DeleteIconButton
                            onClick={() => handleDelete(checkup.id)}
                            disabled={deleteMutation.isPending}
                          />
                        ) : null}
                      </div>
                    </td>
                  </tr>
                )
              )
            )}
          </tbody>
        </table>

        {/* インライン追加フォーム */}
        {isAdding ? (
          <div className={`flex flex-wrap items-center gap-2 px-3 py-2 border-t ${C.borderLight} ${C.bgPage30}`}>
            <NotionDatePicker
              value={addForm.date}
              onChange={(v) => handleAddFormChange("date", v)}
              placeholder="日付"
              className="h-8 w-32"
            />
            <select
              value={addForm.checkup_type_id}
              onChange={(e: ChangeEvent<HTMLSelectElement>) => handleAddFormChange("checkup_type_id", e.target.value)}
              className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent} w-32`}
            >
              <option value="">選択</option>
              {checkupTypeOptions}
            </select>
            <NotionDatePicker
              value={addForm.next_date}
              onChange={(v) => handleAddFormChange("next_date", v)}
              placeholder="次回日"
              className="h-8 w-32"
            />
            <input
              autoFocus
              type="text"
              placeholder="結果を入力..."
              value={addForm.result}
              onChange={(e) => handleAddFormChange("result", e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleAddSubmit();
                if (e.key === "Escape") handleAddCancel();
              }}
              className={`flex-1 min-w-[160px] h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent}`}
            />
            <Button
              size="sm"
              className={`${STYLE.btnPrimary} h-8 text-xs px-3`}
              onClick={handleAddSubmit}
              disabled={createMutation.isPending || !addForm.date || !addForm.checkup_type_id}
            >
              追加
            </Button>
            <Button
              size="sm"
              variant="outline"
              className={`h-8 text-xs px-3 ${C.borderMedium}`}
              onClick={handleAddCancel}
            >
              キャンセル
            </Button>
          </div>
        ) : (
          canCreate ? (
            <button className={STYLE.inlineAddBtn} onClick={() => setIsAdding(true)}>
              <Plus className={`${ICON.xs}`} />
              <span>記録を追加</span>
            </button>
          ) : null
        )}
      </div>

      {/* フッター: 件数表示 */}
      {checkupList.length > 0 ? (
        <div className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3`}>
          <span className={`text-sm ${C.text60}`}>
            健診記録 {checkupList.length} 件
          </span>
        </div>
      ) : null}
    </div>
  );
});
