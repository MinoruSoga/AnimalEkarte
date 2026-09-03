import { memo, useActionState, useCallback, useLayoutEffect, useMemo, useRef, useState, lazy, Suspense } from "react";
import { BarChart2, Table2 } from "lucide-react";
import { toast } from "sonner";

import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { ErrorFallback } from "@/components/shared/DataStates";
import { usePermission } from "@/hooks/use-permission";
import { C, ICON } from "@/lib/design-tokens";
import { jstDateTimeLocalToISOString } from "@/lib/jst-date";
import { INITIAL_ACTION_STATE, type ActionState } from "@/types/form";
import {
  useCreateVital,
  useDeleteVital,
  useGetVitals,
  useUpdateVital,
} from "../../api/vitals";
import type { CreateVitalInput, UpdateVitalInput, Vital } from "../../types";
// recharts(~100KB+ gzip)を引き込む唯一の依存。グラフは showGraph トグル
// ON 時のみ描画されるため遅延ロードし、医療記録ルートの初期バンドルから外す。
const VitalsGraph = lazy(() =>
  import("./VitalsGraph").then((m) => ({ default: m.VitalsGraph })),
);
import { VitalsTable } from "./VitalsTabTable";
import {
  EMPTY_VITALS_ADD_FORM,
  parseVitalsNumber,
  type VitalsAddFormState,
} from "./vitals-tab-table-model";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const DECEASED_VITALS_MESSAGE = "死亡したペットのバイタルは保存できません";

interface VitalsTabProps {
  medicalRecordId: string;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  isPetDeceased?: boolean;
}

export const VitalsTab = memo(function VitalsTab({
  medicalRecordId,
  recordClinicId,
  isPetDeceased = false,
}: VitalsTabProps) {
  const { canCreate, canEdit, canDelete } = usePermission("medical-records");
  const { data: vitals, isLoading, isError } = useGetVitals(medicalRecordId, recordClinicId);
  const { mutateAsync: createVital, isPending: createPending } = useCreateVital(medicalRecordId, recordClinicId);
  const { mutate: updateVital, isPending: updatePending } = useUpdateVital(medicalRecordId, recordClinicId);
  const { mutate: deleteVital, isPending: deletePending } = useDeleteVital(medicalRecordId, recordClinicId);

  const canCreateRef = useRef(canCreate);
  const canEditRef = useRef(canEdit);
  const canDeleteRef = useRef(canDelete);
  const isPetDeceasedRef = useRef(isPetDeceased);
  useLayoutEffect(() => {
    canCreateRef.current = canCreate;
  }, [canCreate]);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  useLayoutEffect(() => {
    isPetDeceasedRef.current = isPetDeceased;
  }, [isPetDeceased]);

  const [editingId, setEditingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [addForm, setAddForm] = useState<VitalsAddFormState>(EMPTY_VITALS_ADD_FORM);
  const [addFormErrors, setAddFormErrors] = useState<Record<string, string>>({});
  const [showGraph, setShowGraph] = useState(false);

  const sortedVitals = useMemo<Vital[]>(
    () =>
      vitals
        ? [...vitals].sort(
            (a, b) =>
              new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
          )
        : [],
    [vitals]
  );

  const handleAddFormChange = useCallback((patch: Partial<VitalsAddFormState>) => {
    setAddForm((prev) => ({ ...prev, ...patch }));
  }, []);
  const addFormRef = useRef(addForm);
  useLayoutEffect(() => {
    addFormRef.current = addForm;
  }, [addForm]);

  const [addFormState, addFormAction] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (isPetDeceasedRef.current === true) {
        return {
          success: false,
          error: DECEASED_VITALS_MESSAGE,
          timestamp: Date.now(),
        };
      }
      if (canCreateRef.current !== true) {
        return {
          success: false,
          error: PERMISSION_DENIED_MESSAGE,
          timestamp: Date.now(),
        };
      }

      const currentAddForm = addFormRef.current;
      const errors: Record<string, string> = {};
      if (!currentAddForm.recorded_at) {
        errors.recorded_at = "記録日時は必須です";
      } else {
        const recordedDate = new Date(currentAddForm.recorded_at);
        if (recordedDate > new Date()) {
          errors.recorded_at = "未来の日時は入力できません";
        }
      }

      const temperature = parseVitalsNumber(currentAddForm.temperature);
      if (temperature !== null && (temperature < 30 || temperature > 45)) {
        errors.temperature = "体温は30〜45℃の範囲で入力してください";
      }

      const heartRate = parseVitalsNumber(currentAddForm.heart_rate);
      const respiratoryRate = parseVitalsNumber(currentAddForm.respiration_rate);
      const bodyWeight = parseVitalsNumber(currentAddForm.weight);
      if (
        temperature === null &&
        heartRate === null &&
        respiratoryRate === null &&
        bodyWeight === null
      ) {
        errors.temperature =
          errors.temperature ?? "体温・心拍数・呼吸数・体重のいずれかを入力してください";
      }

      if (Object.keys(errors).length > 0) {
        setAddFormErrors(errors);
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }
      setAddFormErrors({});

      const input: CreateVitalInput = {
        recorded_at: jstDateTimeLocalToISOString(currentAddForm.recorded_at),
        temperature,
        heart_rate: heartRate,
        respiration_rate: respiratoryRate,
        weight: bodyWeight,
        weight_unit: currentAddForm.weight_unit,
        note: currentAddForm.note.trim() || null,
      };

      try {
        await createVital(input);
        setAddForm(EMPTY_VITALS_ADD_FORM);
        setIsAdding(false);
        toast.success("バイタルを追加しました");
        return { success: true, timestamp: Date.now() };
      } catch {
        // FE-RC-005 系: useCreateVital の onError が既に handleApiError でトースト
        // 表示済み。ここで再度呼ぶと二重トーストになるため何もしない。
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const handleAddCancel = useCallback(() => {
    setAddForm(EMPTY_VITALS_ADD_FORM);
    setAddFormErrors({});
    setIsAdding(false);
  }, []);

  const handleEditSave = useCallback(
    (vitalId: string, input: UpdateVitalInput) => {
      if (isPetDeceasedRef.current === true) {
        toast.error(DECEASED_VITALS_MESSAGE);
        return;
      }
      if (canEditRef.current !== true) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return;
      }
      updateVital(
        { vitalId, input },
        {
          onSuccess: () => {
            setEditingId(null);
            toast.success("バイタルを更新しました");
          },
          // FE-RC-005 系: useUpdateVital の onError が既に handleApiError 済み。
        }
      );
    },
    [updateVital]
  );

  const handleDeleteConfirm = useCallback(() => {
    if (isPetDeceasedRef.current === true) {
      toast.error(DECEASED_VITALS_MESSAGE);
      return;
    }
    if (canDeleteRef.current !== true || !deletingId) {
      if (canDeleteRef.current !== true) {
        toast.error(PERMISSION_DENIED_MESSAGE);
      }
      return;
    }
    deleteVital(deletingId, {
      onSuccess: () => {
        setDeletingId(null);
        toast.success("バイタルを削除しました");
      },
      // FE-RC-005 系: useDeleteVital の onError が既に handleApiError 済み。
    });
  }, [deletingId, deleteVital]);

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
      {sortedVitals.length > 0 ? (
        <VitalsViewToggle showGraph={showGraph} onChange={setShowGraph} />
      ) : null}

      {showGraph && sortedVitals.length > 0 ? (
        <Suspense fallback={null}>
          <VitalsGraph vitals={sortedVitals} />
        </Suspense>
      ) : null}

      <VitalsTable
        vitals={sortedVitals}
        editingId={editingId}
        canCreate={canCreate && !isPetDeceased}
        canEdit={canEdit && !isPetDeceased}
        canDelete={canDelete && !isPetDeceased}
        isAdding={isAdding}
        addForm={addForm}
        addFormErrors={addFormErrors}
        addFormError={addFormState.error}
        createPending={createPending}
        updatePending={updatePending}
        deletePending={deletePending}
        onStartAdd={() => setIsAdding(true)}
        onAddFormChange={handleAddFormChange}
        addFormAction={addFormAction}
        onAddCancel={handleAddCancel}
        onStartEdit={setEditingId}
        onEditSave={handleEditSave}
        onEditCancel={() => setEditingId(null)}
        onDeleteClick={setDeletingId}
      />

      {sortedVitals.length > 0 ? (
        <div className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-3`}>
          <span className={`text-sm ${C.text60}`}>
            バイタル記録 {sortedVitals.length} 件
          </span>
        </div>
      ) : null}

      <ConfirmDialog
        open={deletingId !== null}
        onClose={() => setDeletingId(null)}
        onConfirm={handleDeleteConfirm}
        title="バイタルを削除しますか？"
        description="このバイタル記録を削除します。この操作は元に戻せません。"
        confirmLabel={deletePending ? "削除中..." : "削除する"}
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={deletePending}
      />
    </div>
  );
});

interface VitalsViewToggleProps {
  showGraph: boolean;
  onChange: (showGraph: boolean) => void;
}

function VitalsViewToggle({ showGraph, onChange }: VitalsViewToggleProps) {
  return (
    <div className="flex items-center justify-end">
      <div className={`flex items-center border ${C.borderLight} rounded-xs overflow-hidden`}>
        <button
          type="button"
          onClick={() => onChange(false)}
          className={[
            "flex items-center gap-1.5 px-3 h-8 text-xs font-medium transition-colors",
            !showGraph
              ? `${C.bgWhite} ${C.text} border-r ${C.borderLight}`
              : `${C.text60} ${C.hoverBgLight} border-r ${C.borderLight}`,
          ].join(" ")}
          title="テーブル表示"
        >
          <Table2 className={ICON.xs} />
          テーブル
        </button>
        <button
          type="button"
          onClick={() => onChange(true)}
          className={[
            "flex items-center gap-1.5 px-3 h-8 text-xs font-medium transition-colors",
            showGraph ? `${C.bgWhite} ${C.text}` : `${C.text60} ${C.hoverBgLight}`,
          ].join(" ")}
          title="グラフ表示"
        >
          <BarChart2 className={ICON.xs} />
          グラフ
        </button>
      </div>
    </div>
  );
}
