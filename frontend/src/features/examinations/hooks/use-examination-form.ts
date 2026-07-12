import { useState, useEffect, useTransition, useCallback, useActionState, useRef } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import { useNavigate, useSearchParams } from "react-router";
import type { ExaminationRecord } from "../api/transforms";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetExamination } from "../api/get-examination";
import { useCreateExamination } from "../api/create-examination";
import { useUpdateExamination } from "../api/update-examination";
import { useDeleteExamination } from "../api/delete-examination";
import { useGetExaminationItems } from "../api/get-examination-items";
import { useUpdateExaminationItems } from "../api/update-examination-items";
import { useGetExamTypeFields, type ExamTypeFieldRow } from "../api/get-exam-type-fields";
import type {
  CreateExaminationRequest,
  ReplaceExamItemsRequest,
  UpdateExaminationRequest,
  UpsertExamItemRequest,
} from "../api/types";
import type { ExamItemRow } from "../components/ExamItemsTable";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import { EXAM_STATUS_EN_TO_JA } from "@/lib/transforms/examination";

/** EXAM_STATUS_EN_TO_JA（正本）の逆写像を導出する（FE5-10）。両写像は完全対称であることを確認済み。 */
const EXAM_STATUS_JA_TO_EN = Object.fromEntries(
  Object.entries(EXAM_STATUS_EN_TO_JA).map(([en, ja]) => [ja, en]),
) as Record<string, "pending" | "in_progress" | "result_entered" | "completed" | "confirmed">;

// テンプレ（exam_type_fields）から ExamItemRow の初期行を組み立てる。
// status/isAbnormal は backend が保存後に導出するため未設定で開始する。
function buildRowsFromTemplate(fields: ExamTypeFieldRow[]): ExamItemRow[] {
  return fields.map((f, idx) => ({
    key: `tmpl-${f.id}`,
    examTypeFieldId: f.id,
    name: f.name,
    inspectionValue: "",
    unit: f.unit,
    normalValue: f.normalValue,
    referenceValue: f.normalValue,
    refMin: f.refMin,
    refMax: f.refMax,
    sortOrder: f.sortOrder !== 0 ? f.sortOrder : idx,
  }));
}

// formItems → PUT リクエスト形式へ変換。空の項目（name 空 & 値空）は送信しない。
function rowsToRequest(items: ExamItemRow[]): UpsertExamItemRequest[] {
  return items
    .filter((it) => it.name.trim() !== "")
    .map((it, idx) => ({
      exam_type_field_id: it.examTypeFieldId ?? null,
      name: it.name,
      inspection_value: it.inspectionValue,
      normal_value: it.normalValue,
      unit: it.unit,
      reference_value: it.referenceValue,
      ref_min: it.refMin ?? null,
      ref_max: it.refMax ?? null,
      sort_order: it.sortOrder !== 0 ? it.sortOrder : idx,
    }));
}

// v2: added handleDelete, isDeleting
export function useExaminationForm(id?: string, medicalRecordIdParam?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const doctorId = searchParams.get("doctorId");
  const medicalRecordId = medicalRecordIdParam ?? searchParams.get("medicalRecordId") ?? "";
  const isEdit = !!id;

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks
  const { data: existingExam } = useGetExamination(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateExamination();
  const updateMutation = useUpdateExamination();
  const deleteMutation = useDeleteExamination();
  const { data: existingItems } = useGetExaminationItems(id ?? "");
  const updateItemsMutation = useUpdateExaminationItems();

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // Local overrides applied on top of server data (only tracks user edits in edit mode)
  // useActionState の前に宣言: callback 内で formDataWithPet を参照するため
  const [localOverrides, setLocalOverrides] = useState<Partial<ExaminationRecord>>({});

  // Merge: server data as base + user edits on top
  const formData: Partial<ExaminationRecord> =
    isEdit && existingExam
      ? { ...existingExam, ...localOverrides }
      : {
          status: "依頼中" as const,
          ownerName: "",
          petName: "",
          ...(doctorId && { doctorId }),
          ...localOverrides
        };

  const setFormData = (next: Partial<ExaminationRecord>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  };

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const formDataWithPet =
    selectedPets.length > 0
      ? {
          ...formData,
          ownerName: selectedPets[0].ownerName,
          petName: selectedPets[0].name,
        }
      : formData;

  // useActionState の stale closure 対策: 最新の formDataWithPet を ref で保持
  // (use-medical-record-form.ts の activeTabRef と同じパターン)
  const formDataWithPetRef = useRef(formDataWithPet);
  useEffect(() => {
    formDataWithPetRef.current = formDataWithPet;
  });

  // ─────────────────────────────────────────────────
  // 検査項目テーブルの state
  // ─────────────────────────────────────────────────
  const [formItems, setFormItems] = useState<ExamItemRow[]>([]);
  const formItemsRef = useRef(formItems);
  useEffect(() => {
    formItemsRef.current = formItems;
  });

  // 検査種別テンプレ（exam_type_fields）取得 — testTypeId 変更検知に使う
  const currentTestTypeId = formData.testTypeId ?? "";
  const { data: examTypeFields } = useGetExamTypeFields(currentTestTypeId);

  // 編集モード: 既存 items を一度だけ formItems に流し込む
  // 同期目的のため useEffect 内 setState は許容。
  const itemsInitializedRef = useRef(false);
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!isEdit) return;
    if (itemsInitializedRef.current) return;
    if (!existingItems) return;
    if (existingItems.length === 0 && !examTypeFields) return; // テンプレ未到着なら待つ
    if (existingItems.length > 0) {
      const rows: ExamItemRow[] = existingItems.map((it) => ({
        key: `srv-${it.id}`,
        examTypeFieldId: it.examTypeFieldId,
        name: it.name,
        inspectionValue: it.inspectionValue,
        unit: it.unit,
        normalValue: it.normalValue,
        referenceValue: it.referenceValue,
        refMin: it.refMin,
        refMax: it.refMax,
        sortOrder: it.sortOrder,
        status: it.status,
        isAbnormal: it.isAbnormal,
      }));
      setFormItems(rows);
    } else if (examTypeFields && examTypeFields.length > 0) {
      // 既存 items が空ならテンプレで初期化
      setFormItems(buildRowsFromTemplate(examTypeFields));
    }
    itemsInitializedRef.current = true;
  }, [isEdit, existingItems, examTypeFields]);

  // 検査種別変更検知 — ユーザーが testTypeId を変えたらテンプレで再構築
  // 初回レンダー（編集モードで existingExam が後から到着するケース）は既存 items の流入を尊重するためスキップ
  const previousTestTypeIdRef = useRef<string | undefined>(undefined);
  useEffect(() => {
    const next = currentTestTypeId;
    if (!next) return;
    if (previousTestTypeIdRef.current === undefined) {
      previousTestTypeIdRef.current = next;
      return;
    }
    if (previousTestTypeIdRef.current === next) return;
    previousTestTypeIdRef.current = next;
    // 種別が変わった → テンプレで再構築（テンプレ未到着なら空行）
    if (examTypeFields) {
      setFormItems(buildRowsFromTemplate(examTypeFields));
    } else {
      setFormItems([]);
    }
  }, [currentTestTypeId, examTypeFields]);

  // 新規モード: testTypeId 選択 & テンプレ到着で初期化（一度だけ）
  const newModeInitializedRef = useRef(false);
  useEffect(() => {
    if (isEdit) return;
    if (newModeInitializedRef.current) return;
    if (!currentTestTypeId) return;
    if (!examTypeFields) return;
    setFormItems(buildRowsFromTemplate(examTypeFields));
    newModeInitializedRef.current = true;
  }, [isEdit, currentTestTypeId, examTypeFields]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const setInspectionValue = useCallback((key: string, value: string) => {
    setFormItems((prev) =>
      prev.map((row) => (row.key === key ? { ...row, inspectionValue: value } : row)),
    );
  }, []);

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const current = formDataWithPetRef.current;
      // フロントエンド・バリデーション
      const errors: Record<string, string> = {};
      if (!current.testTypeId) errors.testTypeId = "検査種別を選択してください";
      if (!current.doctorId) errors.doctorId = "担当医を選択してください";

      if (Object.keys(errors).length > 0) {
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }

      try {
        // 親 exam (PATCH/POST) と検査項目 (PUT items) を順次保存する。
        // 確定済みの場合は items の更新も backend で 400 になるため、ここでは送らない。
        const itemsReq: ReplaceExamItemsRequest = { items: rowsToRequest(formItemsRef.current) };
        const isConfirmed = current.status === "確定";

        if (isEdit && id) {
          const req: UpdateExaminationRequest = {
            status: current.status ? EXAM_STATUS_JA_TO_EN[current.status] : undefined,
            result_summary: current.resultSummary,
            machine: current.machine,
            date: current.date
              ? current.date.includes("T")
                ? current.date
                : jstDateStartISOString(current.date)
              : undefined,
          };
          await updateMutation.mutateAsync({ id, req });
          if (!isConfirmed) {
            await updateItemsMutation.mutateAsync({ id, req: itemsReq });
          }
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          const req: CreateExaminationRequest = {
            medical_record_id: medicalRecordId ? Number(medicalRecordId) : null,
            pet_id: Number(pet.id) || null,
            exam_type_id: Number(current.testTypeId) || 0,
            doctor_id: current.doctorId ? Number(current.doctorId) : null,
            date: current.date ?? jstDateStartISOString(todayJSTISO()),
            result_summary: current.resultSummary,
            machine: current.machine,
          };
          const created = await createMutation.mutateAsync(req);
          // 新規作成後に items を保存。items が空なら呼ばない（不要な PUT を回避）。
          if (!isConfirmed && itemsReq.items.length > 0 && created?.id) {
            await updateItemsMutation.mutateAsync({ id: String(created.id), req: itemsReq });
          }
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  // New mode: populate pet selection from petId query param
  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        // No petId provided and not loading — redirect to pet selection
        navigate(paths.examinations.selectPet.getHref());
      }
      // If petId is provided but petFromQuery is not yet resolved, wait
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    startDeleteTransition(() => {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          toast.success("検査記録を削除しました");
          onSuccess?.();
        },
      });
    });
  }, [isEdit, id, deleteMutation, startDeleteTransition]);

  const isSaving = isPending || updateItemsMutation.isPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    formAction,
    formState,
    handleDelete,
    isEdit,
    isSaving,
    isDeleting,
    // 検査項目テーブル
    formItems,
    setInspectionValue,
  };
}
