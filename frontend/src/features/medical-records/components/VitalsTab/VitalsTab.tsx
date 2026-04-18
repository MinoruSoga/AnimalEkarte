// React/Framework
import { useState, useCallback, memo } from "react";

// External
import { Pencil, Plus, Check, X, BarChart2, Table2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { handleApiError } from "@/lib/handle-api-error";
import { ErrorFallback } from "@/components/shared/DataStates";

// rendering-hoist-jsx: design token は定数なので module-level に巻き上げ
const EDIT_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent} w-full`;
const ADD_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 bg-white ${C.text} outline-none ${C.focusBorderAccent}`;

// Relative
import { VitalsGraph } from "./VitalsGraph";
import { usePermission } from "@/hooks/use-permission";
import { useGetVitals } from "../../api/vitals";
import { useCreateVital } from "../../api/vitals";
import { useUpdateVital } from "../../api/vitals";
import { useDeleteVital } from "../../api/vitals";
import type { Vital, CreateVitalInput, UpdateVitalInput, BodyWeightUnit } from "../../types";

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
  respiration_rate: string;
  weight: string;
  weight_unit: BodyWeightUnit;
  note: string;
}

const EMPTY_ADD_FORM: AddFormState = {
  recorded_at: "",
  temperature: "",
  heart_rate: "",
  respiration_rate: "",
  weight: "",
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

// rerender-lazy-state-init: computed object の初期化をモジュールレベル関数に抽出
function buildEditRowForm(vital: Vital) {
  return {
    recorded_at: vital.recorded_at
      ? new Date(vital.recorded_at).toISOString().slice(0, 16)
      : "",
    temperature: vital.temperature != null ? String(vital.temperature) : "",
    heart_rate: vital.heart_rate != null ? String(vital.heart_rate) : "",
    respiration_rate: vital.respiration_rate != null ? String(vital.respiration_rate) : "",
    weight: vital.weight != null ? String(vital.weight) : "",
    weight_unit: vital.weight_unit ?? "Kg",
    note: vital.note ?? "",
  };
}

const EditRow = memo(function EditRow({ vital, onSave, onCancel, isPending }: EditRowProps) {
  // lazy initializer — 初回マウント時のみ buildEditRowForm() が実行される
  const [form, setForm] = useState(() => buildEditRowForm(vital));
  const [editFormErrors, setEditFormErrors] = useState<Record<string, string>>({});

  const handleChange = useCallback(
    (field: string, value: string | BodyWeightUnit) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    []
  );

  const handleSave = useCallback(() => {
    const errors: Record<string, string> = {};
    if (!form.recorded_at) {
      errors.recorded_at = "記録日時は必須です";
    } else {
      const recordedDate = new Date(form.recorded_at);
      if (recordedDate > new Date()) {
        errors.recorded_at = "未来の日時は入力できません";
      }
    }
    const temp = parseNumField(form.temperature);
    if (temp !== null && (temp < 30 || temp > 45)) {
      errors.temperature = "体温は30〜45℃の範囲で入力してください";
    }
    if (Object.keys(errors).length > 0) {
      setEditFormErrors(errors);
      return;
    }
    setEditFormErrors({});

    const recordedDate = new Date(form.recorded_at);
    const input: UpdateVitalInput = {
      recorded_at: recordedDate.toISOString(),
      temperature: temp,
      heart_rate: parseNumField(form.heart_rate),
      respiration_rate: parseNumField(form.respiration_rate),
      weight: parseNumField(form.weight),
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
        <FormFieldError message={editFormErrors.recorded_at} />
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
        <FormFieldError message={editFormErrors.temperature} />
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
          value={form.respiration_rate}
          onChange={(e) => handleChange("respiration_rate", e.target.value)}
          placeholder="-"
          className={EDIT_INPUT_CLASS}
        />
      </td>
      <td className="px-3 py-2">
        <div className="flex items-center gap-1">
          <input
            type="number"
            step="0.01"
            value={form.weight}
            onChange={(e) => handleChange("weight", e.target.value)}
            placeholder="-"
            className={`${EDIT_INPUT_CLASS} text-right`}
          />
          <button
            type="button"
            onClick={() => handleChange("weight_unit", form.weight_unit === "Kg" ? "g" : "Kg")}
            className={`text-[10px] px-1 h-6 rounded border ${C.borderMedium} ${C.bgPage} ${C.hoverBgPage} min-w-[24px]`}
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

export const VitalsTab = memo(function VitalsTab({ medicalRecordId }: VitalsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const { data: vitals, isLoading, isError } = useGetVitals(medicalRecordId);
  const createMutation = useCreateVital(medicalRecordId);
  const { mutate: createVitalFn } = createMutation;
  const updateMutation = useUpdateVital(medicalRecordId);
  const { mutate: updateVitalFn } = updateMutation;
  const deleteMutation = useDeleteVital(medicalRecordId);
  const { mutate: deleteVitalFn } = deleteMutation;

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<AddFormState>(EMPTY_ADD_FORM);
  const [addFormErrors, setAddFormErrors] = useState<Record<string, string>>({});
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
    if (!canCreate) return;
    const errors: Record<string, string> = {};
    if (!addForm.recorded_at) {
      errors.recorded_at = "記録日時は必須です";
    } else {
      const recordedDate = new Date(addForm.recorded_at);
      if (recordedDate > new Date()) {
        errors.recorded_at = "未来の日時は入力できません";
      }
    }
    const temp = parseNumField(addForm.temperature);
    if (temp !== null && (temp < 30 || temp > 45)) {
      errors.temperature = "体温は30〜45℃の範囲で入力してください";
    }
    // BUG-044: 数値フィールドがすべて未入力の場合は保存しない（サーバー500回避）
    const heartRate = parseNumField(addForm.heart_rate);
    const respiratoryRate = parseNumField(addForm.respiration_rate);
    const bodyWeight = parseNumField(addForm.weight);
    if (temp === null && heartRate === null && respiratoryRate === null && bodyWeight === null) {
      errors.temperature = errors.temperature ?? "体温・心拍数・呼吸数・体重のいずれかを入力してください";
    }
    if (Object.keys(errors).length > 0) {
      setAddFormErrors(errors);
      return;
    }
    setAddFormErrors({});

    const recordedDate = new Date(addForm.recorded_at);
    const input: CreateVitalInput = {
      recorded_at: recordedDate.toISOString(),
      temperature: temp,
      heart_rate: heartRate,
      respiration_rate: respiratoryRate,
      weight: bodyWeight,
      weight_unit: addForm.weight_unit,
      note: addForm.note.trim() || null,
    };
    createVitalFn(input, {
      onSuccess: () => {
        setAddForm(EMPTY_ADD_FORM);
        setIsAdding(false);
        toast.success("バイタルを追加しました");
      },
      onError: (error) => handleApiError(error, "バイタル追加"),
    });
  }, [canCreate, addForm, createVitalFn]);

  const handleAddCancel = useCallback(() => {
    setAddForm(EMPTY_ADD_FORM);
    setAddFormErrors({});
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (vitalId: string, input: UpdateVitalInput) => {
      if (!canEdit) return;
      updateVitalFn(
        { vitalId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("バイタルを更新しました");
          },
          onError: (error) => handleApiError(error, "バイタル更新"),
        }
      );
    },
    [canEdit, updateVitalFn]
  );

  const handleEditCancel = useCallback(() => {
    setEditingId(null);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!canDelete || !deletingId) return;
    deleteVitalFn(deletingId, {
      onSuccess: () => {
        setDeletingId(null);
        toast.success("バイタルを削除しました");
      },
      onError: (error) => handleApiError(error, "バイタル削除"),
    });
  }, [canDelete, deletingId, deleteVitalFn]);

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

  if (isError) return <ErrorFallback />;

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
                <td colSpan={7} className={STYLE.tableEmptySm}>
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
                      {displayNum(vital.respiration_rate)}
                    </td>
                    <td className={`px-3 text-sm text-right ${C.text}`}>
                      {displayNum(vital.weight)}
                      <span className={`ml-0.5 text-[10px] ${C.text40}`}>{vital.weight_unit}</span>
                    </td>
                    <td className={`px-3 text-sm ${C.text60}`}>
                      {vital.note ? vital.note : "-"}
                    </td>
                    <td className="px-2">
                      <div className="flex items-center justify-end gap-1">
                        {canEdit ? (
                          <button
                            onClick={() => setEditingId(vital.id)}
                            className={`${STYLE.iconBtn32} ${C.text60} ${C.hoverText} ${C.hoverBgLight}`}
                            title="編集"
                          >
                            <Pencil className={`${ICON.xs}`} />
                          </button>
                        ) : null}
                        {canDelete ? (
                          <DeleteIconButton
                            onClick={() => setDeletingId(vital.id)}
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
          <div className={`flex flex-wrap items-start gap-2 px-3 py-2 border-t ${C.borderLight} ${C.bgPage30}`}>
            <div className="flex flex-col">
              <input
                autoFocus
                type="datetime-local"
                value={addForm.recorded_at}
                onChange={(e) => handleAddFormChange("recorded_at", e.target.value)}
                className={`${ADD_INPUT_CLASS} w-40`}
              />
              <FormFieldError message={addFormErrors.recorded_at} />
            </div>
            <div className="flex flex-col">
              <input
                type="number"
                step="0.1"
                value={addForm.temperature}
                onChange={(e) => handleAddFormChange("temperature", e.target.value)}
                placeholder="体温"
                className={`${ADD_INPUT_CLASS} w-20`}
              />
              <FormFieldError message={addFormErrors.temperature} />
            </div>
            <input
              type="number"
              value={addForm.heart_rate}
              onChange={(e) => handleAddFormChange("heart_rate", e.target.value)}
              placeholder="心拍数"
              className={`${ADD_INPUT_CLASS} w-20`}
            />
            <input
              type="number"
              value={addForm.respiration_rate}
              onChange={(e) => handleAddFormChange("respiration_rate", e.target.value)}
              placeholder="呼吸数"
              className={`${ADD_INPUT_CLASS} w-20`}
            />
            <div className="flex items-center gap-1">
              <input
                type="number"
                step="0.01"
                value={addForm.weight}
                onChange={(e) => handleAddFormChange("weight", e.target.value)}
                placeholder="体重"
                className={`${ADD_INPUT_CLASS} w-20 text-right`}
              />
              <button
                type="button"
                onClick={() => handleAddFormChange("weight_unit", addForm.weight_unit === "Kg" ? "g" : "Kg")}
                className={`text-[10px] px-1 h-6 rounded border ${C.borderMedium} ${C.bgPage} ${C.hoverBgPage} min-w-[24px]`}
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
          canCreate ? (
            <button className={STYLE.inlineAddBtn} onClick={() => setIsAdding(true)}>
              <Plus className={`${ICON.xs}`} />
              <span>記録を追加</span>
            </button>
          ) : null
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
});
