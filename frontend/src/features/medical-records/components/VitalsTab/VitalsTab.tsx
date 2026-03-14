// React/Framework
import { useState, useCallback, memo } from "react";

// External
import { Pencil, Trash2, Plus, Check, X } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { C, STYLE } from "@/lib/design-tokens";

// Relative
import { useVitals } from "../../api/vitals";
import { useCreateVital } from "../../api/vitals";
import { useUpdateVital } from "../../api/vitals";
import { useDeleteVital } from "../../api/vitals";
import type { Vital, CreateVitalInput, UpdateVitalInput } from "../../types";

// ── 静的定数 ─────────────────────────────────────────────────────────

const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage30} h-10`}>
      <th className={`px-3 text-left text-xs font-medium ${C.text70} w-40`}>記録日時</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>体温 (℃)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>心拍数 (bpm)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>呼吸数 (/min)</th>
      <th className={`px-3 text-right text-xs font-medium ${C.text70} w-24`}>体重 (kg)</th>
      <th className={`px-3 text-left text-xs font-medium ${C.text70}`}>メモ</th>
      <th className={`px-2 text-right text-xs font-medium ${C.text70} w-24`}>操作</th>
    </tr>
  </thead>
);

// ── ユーティリティ ────────────────────────────────────────────────────

/** ISO8601 → "YYYY-MM-DD HH:mm" 表示形式 */
function formatRecordedAt(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** 数値フィールドの表示: null/undefined → "-" */
function displayNum(value: number | null | undefined): string {
  return value != null ? String(value) : "-";
}

// ── 追加フォーム初期状態 ──────────────────────────────────────────────

interface AddFormState {
  recorded_at: string;
  temperature: string;
  heart_rate: string;
  respiratory_rate: string;
  body_weight: string;
  note: string;
}

const EMPTY_ADD_FORM: AddFormState = {
  recorded_at: "",
  temperature: "",
  heart_rate: "",
  respiratory_rate: "",
  body_weight: "",
  note: "",
};

/** string → number | null 変換（空文字は null） */
function parseNumField(s: string): number | null {
  if (s.trim() === "") return null;
  const n = Number(s);
  return isNaN(n) ? null : n;
}

// ── 編集行コンポーネント ──────────────────────────────────────────────

interface EditRowProps {
  vital: Vital;
  onSave: (vitalId: string, input: UpdateVitalInput) => void;
  onCancel: () => void;
  isPending: boolean;
}

const EditRow = memo(function EditRow({ vital, onSave, onCancel, isPending }: EditRowProps) {
  const [form, setForm] = useState({
    recorded_at: vital.recorded_at
      ? new Date(vital.recorded_at).toISOString().slice(0, 16)
      : "",
    temperature: vital.temperature != null ? String(vital.temperature) : "",
    heart_rate: vital.heart_rate != null ? String(vital.heart_rate) : "",
    respiratory_rate: vital.respiratory_rate != null ? String(vital.respiratory_rate) : "",
    body_weight: vital.body_weight != null ? String(vital.body_weight) : "",
    note: vital.note ?? "",
  });

  const handleChange = useCallback(
    (field: keyof typeof form, value: string) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleSave = useCallback(() => {
    if (!form.recorded_at) {
      toast.error("記録日時は必須です");
      return;
    }
    const input: UpdateVitalInput = {
      recorded_at: new Date(form.recorded_at).toISOString(),
      temperature: parseNumField(form.temperature),
      heart_rate: parseNumField(form.heart_rate),
      respiratory_rate: parseNumField(form.respiratory_rate),
      body_weight: parseNumField(form.body_weight),
      note: form.note.trim() || null,
    };
    onSave(vital.id, input);
  }, [vital.id, form, onSave]);

  const inputClass = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`;

  return (
    <tr className={`border-b ${C.borderLight} ${C.bgNotice40}`}>
      <td className="px-3 py-2">
        <input
          type="datetime-local"
          value={form.recorded_at}
          onChange={(e) => handleChange("recorded_at", e.target.value)}
          className={inputClass}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          step="0.1"
          value={form.temperature}
          onChange={(e) => handleChange("temperature", e.target.value)}
          placeholder="-"
          className={inputClass}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          value={form.heart_rate}
          onChange={(e) => handleChange("heart_rate", e.target.value)}
          placeholder="-"
          className={inputClass}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          value={form.respiratory_rate}
          onChange={(e) => handleChange("respiratory_rate", e.target.value)}
          placeholder="-"
          className={inputClass}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          step="0.01"
          value={form.body_weight}
          onChange={(e) => handleChange("body_weight", e.target.value)}
          placeholder="-"
          className={inputClass}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="text"
          value={form.note}
          onChange={(e) => handleChange("note", e.target.value)}
          placeholder="メモ"
          className={inputClass}
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

// ── 削除確認ダイアログ ────────────────────────────────────────────────

interface DeleteConfirmProps {
  onConfirm: () => void;
  onCancel: () => void;
  isPending: boolean;
}

function DeleteConfirmDialog({ onConfirm, onCancel, isPending }: DeleteConfirmProps) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={onCancel}
    >
      <div
        className="bg-white rounded-lg shadow-lg p-6 w-[380px]"
        onClick={(e) => e.stopPropagation()}
      >
        <p className={`text-base font-medium ${C.text} mb-2`}>バイタルを削除しますか？</p>
        <p className={`text-sm ${C.text60} mb-6`}>
          このバイタル記録を削除します。この操作は元に戻せません。
        </p>
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel} disabled={isPending}>
            キャンセル
          </Button>
          <Button
            className={STYLE.btnDanger}
            onClick={onConfirm}
            disabled={isPending}
          >
            削除
          </Button>
        </div>
      </div>
    </div>
  );
}

// ── Props ─────────────────────────────────────────────────────────────

interface VitalsTabProps {
  medicalRecordId: string;
}

// ── Component ─────────────────────────────────────────────────────────

export function VitalsTab({ medicalRecordId }: VitalsTabProps) {
  const { data: vitals, isLoading } = useVitals(medicalRecordId);
  const createMutation = useCreateVital(medicalRecordId);
  const updateMutation = useUpdateVital(medicalRecordId);
  const deleteMutation = useDeleteVital(medicalRecordId);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddFormState>(EMPTY_ADD_FORM);

  // recorded_at 昇順ソート済みリスト
  const sortedVitals: Vital[] = vitals
    ? [...vitals].sort(
        (a, b) =>
          new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
      )
    : [];

  // ── handlers ──

  const handleAddFormChange = useCallback(
    (field: keyof AddFormState, value: string) => {
      setAddForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleAddSubmit = useCallback(() => {
    if (!addForm.recorded_at) {
      toast.error("記録日時は必須です");
      return;
    }
    const input: CreateVitalInput = {
      recorded_at: new Date(addForm.recorded_at).toISOString(),
      temperature: parseNumField(addForm.temperature),
      heart_rate: parseNumField(addForm.heart_rate),
      respiratory_rate: parseNumField(addForm.respiratory_rate),
      body_weight: parseNumField(addForm.body_weight),
      note: addForm.note.trim() || null,
    };
    createMutation.mutate(input, {
      onSuccess: () => {
        setAddForm(EMPTY_ADD_FORM);
        setIsAdding(false);
        toast.success("バイタルを追加しました");
      },
    });
  }, [addForm, createMutation]);

  const handleAddCancel = useCallback(() => {
    setAddForm(EMPTY_ADD_FORM);
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (vitalId: string, input: UpdateVitalInput) => {
      updateMutation.mutate(
        { vitalId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("バイタルを更新しました");
          },
        }
      );
    },
    [updateMutation]
  );

  const handleEditCancel = useCallback(() => {
    setEditingId(null);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!deletingId) return;
    deleteMutation.mutate(deletingId, {
      onSuccess: () => {
        setDeletingId(null);
        toast.success("バイタルを削除しました");
      },
    });
  }, [deletingId, deleteMutation]);

  const handleDeleteCancel = useCallback(() => {
    setDeletingId(null);
  }, []);

  // ── render ──

  if (isLoading) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        読み込み中...
      </div>
    );
  }

  const inputClass = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2]`;

  return (
    <div className="flex flex-col gap-3 pb-24">
      <div className={`${STYLE.tableContainer} overflow-x-auto`}>
        <table className="w-full">
          {TABLE_HEADER}
          <tbody>
            {sortedVitals.length === 0 ? (
              <tr>
                <td colSpan={7} className={`text-center py-12 text-sm ${C.text40}`}>
                  バイタル記録がありません。下の「記録を追加」ボタンから追加してください。
                </td>
              </tr>
            ) : (
              sortedVitals.map((vital) =>
                editingId === vital.id ? (
                  <EditRow
                    key={vital.id}
                    vital={vital}
                    onSave={handleEditSave}
                    onCancel={handleEditCancel}
                    isPending={updateMutation.isPending}
                  />
                ) : (
                  <tr
                    key={vital.id}
                    className={`border-b ${C.borderLight} ${C.hoverBgPageHalf} transition-colors h-12`}
                  >
                    <td className={`px-3 text-sm ${C.text}`}>
                      {formatRecordedAt(vital.recorded_at)}
                    </td>
                    <td className={`px-3 text-sm text-right ${C.text}`}>
                      {displayNum(vital.temperature)}
                    </td>
                    <td className={`px-3 text-sm text-right ${C.text}`}>
                      {displayNum(vital.heart_rate)}
                    </td>
                    <td className={`px-3 text-sm text-right ${C.text}`}>
                      {displayNum(vital.respiratory_rate)}
                    </td>
                    <td className={`px-3 text-sm text-right ${C.text}`}>
                      {displayNum(vital.body_weight)}
                    </td>
                    <td className={`px-3 text-sm ${C.text60}`}>
                      {vital.note ? vital.note : "-"}
                    </td>
                    <td className="px-2">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={() => setEditingId(vital.id)}
                          className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverText} ${C.hoverBgLight} transition-colors`}
                          title="編集"
                        >
                          <Pencil className="size-3.5" />
                        </button>
                        <button
                          onClick={() => setDeletingId(vital.id)}
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
              autoFocus
              type="datetime-local"
              value={addForm.recorded_at}
              onChange={(e) => handleAddFormChange("recorded_at", e.target.value)}
              className={`${inputClass} w-40`}
            />
            <input
              type="number"
              step="0.1"
              value={addForm.temperature}
              onChange={(e) => handleAddFormChange("temperature", e.target.value)}
              placeholder="体温"
              className={`${inputClass} w-20`}
            />
            <input
              type="number"
              value={addForm.heart_rate}
              onChange={(e) => handleAddFormChange("heart_rate", e.target.value)}
              placeholder="心拍数"
              className={`${inputClass} w-20`}
            />
            <input
              type="number"
              value={addForm.respiratory_rate}
              onChange={(e) => handleAddFormChange("respiratory_rate", e.target.value)}
              placeholder="呼吸数"
              className={`${inputClass} w-20`}
            />
            <input
              type="number"
              step="0.01"
              value={addForm.body_weight}
              onChange={(e) => handleAddFormChange("body_weight", e.target.value)}
              placeholder="体重"
              className={`${inputClass} w-20`}
            />
            <input
              type="text"
              value={addForm.note}
              onChange={(e) => handleAddFormChange("note", e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleAddSubmit();
                if (e.key === "Escape") handleAddCancel();
              }}
              placeholder="メモ"
              className={`${inputClass} flex-1 min-w-[120px]`}
            />
            <Button
              size="sm"
              className={`${STYLE.btnPrimary} h-8 text-xs px-3`}
              onClick={handleAddSubmit}
              disabled={createMutation.isPending || !addForm.recorded_at}
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
      {sortedVitals.length > 0 ? (
        <div className={`bg-white border ${C.borderLight} rounded-[4px] px-4 py-3`}>
          <span className={`text-sm ${C.text60}`}>
            バイタル記録 {sortedVitals.length} 件
          </span>
        </div>
      ) : null}

      {/* 削除確認ダイアログ */}
      {deletingId ? (
        <DeleteConfirmDialog
          onConfirm={handleDeleteConfirm}
          onCancel={handleDeleteCancel}
          isPending={deleteMutation.isPending}
        />
      ) : null}
    </div>
  );
}
