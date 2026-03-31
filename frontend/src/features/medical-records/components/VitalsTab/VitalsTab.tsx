// React/Framework
import { useState, useCallback, memo } from "react";

// External
import { Pencil, Plus, Check, X, BarChart2, Table2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { C, STYLE, ICON } from "@/lib/design-tokens";

// rendering-hoist-jsx: design token は定数なので module-level に巻き上げ
const EDIT_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2] w-full`;
const ADD_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none focus:border-[#2383E2]`;

// Relative
import { VitalsGraph } from "./VitalsGraph";
import { useGetVitals } from "@/features/medical-records/api/vitals";
import { useCreateVital } from "@/features/medical-records/api/vitals";
import { useUpdateVital } from "@/features/medical-records/api/vitals";
import { useDeleteVital } from "@/features/medical-records/api/vitals";
import type { Vital, CreateVitalInput, UpdateVitalInput, BodyWeightUnit } from "@/features/medical-records/types";

// ── 静的定数 ─────────────────────────────────────────────────────────

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
  weight_unit: BodyWeightUnit;
  note: string;
}

const EMPTY_ADD_FORM: AddFormState = {
  recorded_at: "",
  temperature: "",
  heart_rate: "",
  respiratory_rate: "",
  body_weight: "",
  weight_unit: "Kg",
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
    weight_unit: vital.weight_unit ?? "Kg",
    note: vital.note ?? "",
  });

  const handleChange = useCallback(
    (field: string, value: string | BodyWeightUnit) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleSave = useCallback(() => {
    if (!form.recorded_at) {
      toast.error("記録日時は必須です");
      return;
    }
    const recordedDate = new Date(form.recorded_at);
    if (recordedDate > new Date()) {
      toast.error("未来の日時は入力できません");
      return;
    }

    const temp = parseNumField(form.temperature);
    if (temp !== null && (temp < 30 || temp > 45)) {
      toast.error("体温は30〜45℃の範囲で入力してください");
      return;
    }

    const input: UpdateVitalInput = {
      recorded_at: recordedDate.toISOString(),
      temperature: temp,
      heart_rate: parseNumField(form.heart_rate),
      respiratory_rate: parseNumField(form.respiratory_rate),
      body_weight: parseNumField(form.body_weight),
      weight_unit: form.weight_unit as BodyWeightUnit,
      note: form.note.trim() || null,
    };
    onSave(vital.id, input);
  }, [vital.id, form, onSave]);

  return (
    <tr className={`border-b ${C.borderLight} ${C.bgNotice40}`}>
      <td className="px-3 py-2">
        <input
          type="datetime-local"
          value={form.recorded_at}
          onChange={(e) => handleChange("recorded_at", e.target.value)}
          className={EDIT_INPUT_CLASS}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          step="0.1"
          value={form.temperature}
          onChange={(e) => handleChange("temperature", e.target.value)}
          placeholder="-"
          className={EDIT_INPUT_CLASS}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          value={form.heart_rate}
          onChange={(e) => handleChange("heart_rate", e.target.value)}
          placeholder="-"
          className={EDIT_INPUT_CLASS}
        />
      </td>
      <td className="px-3 py-2">
        <input
          type="number"
          value={form.respiratory_rate}
          onChange={(e) => handleChange("respiratory_rate", e.target.value)}
          placeholder="-"
          className={EDIT_INPUT_CLASS}
        />
      </td>
      <td className="px-3 py-2">
        <div className="flex items-center gap-1">
          <input
            type="number"
            step="0.01"
            value={form.body_weight}
            onChange={(e) => handleChange("body_weight", e.target.value)}
            placeholder="-"
            className={`${EDIT_INPUT_CLASS} text-right`}
          />
          <button
            type="button"
            onClick={() => handleChange("weight_unit", form.weight_unit === "Kg" ? "g" : "Kg")}
            className={`text-[10px] px-1 h-6 rounded border ${C.borderMedium} bg-gray-50 hover:bg-gray-100 min-w-[24px]`}
          >
            {form.weight_unit}
          </button>
        </div>
      </td>
      <td className="px-3 py-2">
        <input
          type="text"
          value={form.note}
          onChange={(e) => handleChange("note", e.target.value)}
          placeholder="メモ"
          className={EDIT_INPUT_CLASS}
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
            <Check className={`${ICON.xs}`} />
          </button>
          <button
            onClick={onCancel}
            disabled={isPending}
            className={`size-8 flex items-center justify-center rounded-[3px] ${C.text60} ${C.hoverBgLight} transition-colors`}
            title="キャンセル"
          >
            <X className={`${ICON.xs}`} />
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
  const { data: vitals, isLoading } = useGetVitals(medicalRecordId);
  const createMutation = useCreateVital(medicalRecordId);
  const updateMutation = useUpdateVital(medicalRecordId);
  const deleteMutation = useDeleteVital(medicalRecordId);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddFormState>(EMPTY_ADD_FORM);
  const [showGraph, setShowGraph] = useState(false);

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
    const recordedDate = new Date(addForm.recorded_at);
    if (recordedDate > new Date()) {
      toast.error("未来の日時は入力できません");
      return;
    }

    const temp = parseNumField(addForm.temperature);
    if (temp !== null && (temp < 30 || temp > 45)) {
      toast.error("体温は30〜45℃の範囲で入力してください");
      return;
    }

    const input: CreateVitalInput = {
      recorded_at: recordedDate.toISOString(),
      temperature: temp,
      heart_rate: parseNumField(addForm.heart_rate),
      respiratory_rate: parseNumField(addForm.respiratory_rate),
      body_weight: parseNumField(addForm.body_weight),
      weight_unit: addForm.weight_unit,
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

  return (
    <div className="flex flex-col gap-3 pb-24">
      {/* ツールバー: 表示切り替え */}
      {sortedVitals.length > 0 ? (
        <div className="flex items-center justify-end">
          <div className={`flex items-center border ${C.borderLight} rounded-[4px] overflow-hidden`}>
            <button
              type="button"
              onClick={() => setShowGraph(false)}
              className={[
                "flex items-center gap-1.5 px-3 h-8 text-xs font-medium transition-colors",
                !showGraph
                  ? `bg-white ${C.text} border-r ${C.borderLight}`
                  : `${C.text60} ${C.hoverBgLight} border-r ${C.borderLight}`,
              ].join(" ")}
              title="テーブル表示"
            >
              <Table2 className={ICON.xs} />
              テーブル
            </button>
            <button
              type="button"
              onClick={() => setShowGraph(true)}
              className={[
                "flex items-center gap-1.5 px-3 h-8 text-xs font-medium transition-colors",
                showGraph ? `bg-white ${C.text}` : `${C.text60} ${C.hoverBgLight}`,
              ].join(" ")}
              title="グラフ表示"
            >
              <BarChart2 className={ICON.xs} />
              グラフ
            </button>
          </div>
        </div>
      ) : null}

      {/* グラフ表示 */}
      {showGraph && sortedVitals.length > 0 ? (
        <VitalsGraph vitals={sortedVitals} />
      ) : null}

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
                      <span className="ml-0.5 text-[10px] text-gray-400">{vital.weight_unit}</span>
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
                          <Pencil className={`${ICON.xs}`} />
                        </button>
                        <DeleteIconButton
                          onClick={() => setDeletingId(vital.id)}
                          disabled={deleteMutation.isPending}
                        />
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
              className={`${ADD_INPUT_CLASS} w-40`}
            />
            <input
              type="number"
              step="0.1"
              value={addForm.temperature}
              onChange={(e) => handleAddFormChange("temperature", e.target.value)}
              placeholder="体温"
              className={`${ADD_INPUT_CLASS} w-20`}
            />
            <input
              type="number"
              value={addForm.heart_rate}
              onChange={(e) => handleAddFormChange("heart_rate", e.target.value)}
              placeholder="心拍数"
              className={`${ADD_INPUT_CLASS} w-20`}
            />
            <input
              type="number"
              value={addForm.respiratory_rate}
              onChange={(e) => handleAddFormChange("respiratory_rate", e.target.value)}
              placeholder="呼吸数"
              className={`${ADD_INPUT_CLASS} w-20`}
            />
            <div className="flex items-center gap-1">
              <input
                type="number"
                step="0.01"
                value={addForm.body_weight}
                onChange={(e) => handleAddFormChange("body_weight", e.target.value)}
                placeholder="体重"
                className={`${ADD_INPUT_CLASS} w-20 text-right`}
              />
              <button
                type="button"
                onClick={() => handleAddFormChange("weight_unit", addForm.weight_unit === "Kg" ? "g" : "Kg")}
                className={`text-[10px] px-1 h-6 rounded border ${C.borderMedium} bg-gray-50 hover:bg-gray-100 min-w-[24px]`}
              >
                {addForm.weight_unit}
              </button>
            </div>
            <input
              type="text"
              value={addForm.note}
              onChange={(e) => handleAddFormChange("note", e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleAddSubmit();
                if (e.key === "Escape") handleAddCancel();
              }}
              placeholder="メモ"
              className={`${ADD_INPUT_CLASS} flex-1 min-w-[120px]`}
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
            <Plus className={`${ICON.xs}`} />
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
