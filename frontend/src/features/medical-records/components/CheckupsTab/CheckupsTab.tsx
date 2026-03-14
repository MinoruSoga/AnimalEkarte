// React/Framework
import { useState, useCallback, memo } from "react";

// External
import { Pencil, Trash2, Plus, Check, X } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { C, STYLE } from "@/lib/design-tokens";

// Relative
import { useCheckups } from "../../api/checkups";
import { useCreateCheckup } from "../../api/checkups";
import { useUpdateCheckup } from "../../api/checkups";
import { useDeleteCheckup } from "../../api/checkups";
import type { Checkup, CreateCheckupInput, UpdateCheckupInput } from "../../api/checkups";

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

// ── 編集行コンポーネント ───────────────────────────────────────────────

interface EditRowProps {
  checkup: Checkup;
  onSave: (checkupId: string, input: UpdateCheckupInput) => void;
  onCancel: () => void;
  isPending: boolean;
}

const EditRow = memo(function EditRow({ checkup, onSave, onCancel, isPending }: EditRowProps) {
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
        <input
          type="date"
          value={form.date ?? ""}
          onChange={(e) => handleChange("date", e.target.value)}
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          value={form.checkup_type_id ?? ""}
          onChange={(e) =>
            handleChange("checkup_type_id", e.target.value ? Number(e.target.value) : null)
          }
          placeholder="健診種別ID"
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="date"
          value={form.next_date ?? ""}
          onChange={(e) => handleChange("next_date", e.target.value || null)}
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`}
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
          className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`}
        />
      </td>
      <td className="px-2 py-2">
        <div className="flex items-center justify-end gap-1">
          <button
            onClick={handleSave}
            disabled={isPending}
            className={`size-8 flex items-center justify-center rounded-[3px] ${C.textStatusGreen} ${C.hoverBgStatusGreen} transition-colors`}
            title="保存"
          >
            <Check className="size-3.5" />
          </button>
          <button
            onClick={onCancel}
            disabled={isPending}
            className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverBgLight} transition-colors`}
            title="キャンセル"
          >
            <X className="size-3.5" />
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

export function CheckupsTab({ medicalRecordId }: CheckupsTabProps) {
  const { data: checkups, isLoading } = useCheckups(medicalRecordId);
  const createMutation = useCreateCheckup(medicalRecordId);
  const updateMutation = useUpdateCheckup(medicalRecordId);
  const deleteMutation = useDeleteCheckup(medicalRecordId);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddFormState>(EMPTY_ADD_FORM);

  // ── handlers ──

  const handleAddFormChange = useCallback(
    (field: keyof AddFormState, value: string) => {
      setAddForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleAddSubmit = useCallback(() => {
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
    createMutation.mutate(input, {
      onSuccess: () => {
        setAddForm(EMPTY_ADD_FORM);
        setIsAdding(false);
        toast.success("健診記録を追加しました");
      },
    });
  }, [addForm, createMutation]);

  const handleAddCancel = useCallback(() => {
    setAddForm(EMPTY_ADD_FORM);
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (checkupId: string, input: UpdateCheckupInput) => {
      updateMutation.mutate(
        { checkupId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("健診記録を更新しました");
          },
        }
      );
    },
    [updateMutation]
  );

  const handleEditCancel = useCallback(() => {
    setEditingId(null);
  }, []);

  const handleDelete = useCallback(
    (checkupId: string) => {
      deleteMutation.mutate(checkupId, {
        onSuccess: () => {
          toast.success("健診記録を削除しました");
        },
      });
    },
    [deleteMutation]
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
                <td colSpan={6} className={`text-center py-12 text-sm ${C.text40}`}>
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
                        <button
                          onClick={() => setEditingId(checkup.id)}
                          className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverText} ${C.hoverBgLight} transition-colors`}
                          title="編集"
                        >
                          <Pencil className="size-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(checkup.id)}
                          disabled={deleteMutation.isPending}
                          className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverTextDanger} ${C.hoverBgDanger5} transition-colors`}
                          title="削除"
                        >
                          <Trash2 className="size-3.5" />
                        </button>
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
            <input
              type="date"
              value={addForm.date}
              onChange={(e) => handleAddFormChange("date", e.target.value)}
              className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-32`}
            />
            <input
              type="number"
              value={addForm.checkup_type_id}
              onChange={(e) => handleAddFormChange("checkup_type_id", e.target.value)}
              placeholder="健診種別ID"
              className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-32`}
            />
            <input
              type="date"
              value={addForm.next_date}
              onChange={(e) => handleAddFormChange("next_date", e.target.value)}
              className={`h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-32`}
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
              className={`flex-1 min-w-[160px] h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2]`}
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
          <button className={STYLE.inlineAddBtn} onClick={() => setIsAdding(true)}>
            <Plus className="size-3.5" />
            <span>記録を追加</span>
          </button>
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
}
