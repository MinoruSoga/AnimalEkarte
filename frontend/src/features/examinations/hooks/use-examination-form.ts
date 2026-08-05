import {
  useState,
  useEffect,
  useLayoutEffect,
  useTransition,
  useCallback,
  useActionState,
  useRef,
  useMemo,
} from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import { useNavigate, useSearchParams } from "react-router";
import type { ExaminationRecord } from "../api/transforms";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import { useGetExamination } from "../api/get-examination";
import { useCreateExamination } from "../api/create-examination";
import { useUpdateExamination } from "../api/update-examination";
import { useDeleteExamination } from "../api/delete-examination";
import { useUnconfirmExamination } from "../api/unconfirm-examination";
import { useGetExaminationItems } from "../api/get-examination-items";
import {
  useGetExamTypeFields,
  type ExamTypeFieldRow,
} from "../api/get-exam-type-fields";
import type {
  CreateExaminationRequest,
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
) as Record<
  string,
  "pending" | "in_progress" | "result_entered" | "completed" | "confirmed"
>;

interface ExaminationMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canUnconfirm: boolean;
}

const DENIED_MUTATION_PERMISSIONS: Readonly<ExaminationMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
  canUnconfirm: false,
};

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
      sort_order: it.sortOrder !== 0 ? it.sortOrder : idx,
    }));
}

// v2: added handleDelete, isDeleting
export function useExaminationForm(
  id?: string,
  medicalRecordIdParam?: string,
  permissions: Readonly<ExaminationMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const doctorId = searchParams.get("doctorId");
  const medicalRecordId =
    medicalRecordIdParam ?? searchParams.get("medicalRecordId") ?? "";
  const isEdit = !!id;
  const activeExaminationIDRef = useRef(id);
  useLayoutEffect(() => {
    activeExaminationIDRef.current = id;
  }, [id]);

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks — BUG-016: classify read failures; never fold into blank edit model
  const {
    data: examinationData,
    isLoading: isExaminationLoading,
    isError: isExaminationError,
    error: examinationError,
    refetch: refetchExamination,
  } = useGetExamination(id ?? "");
  const entityRead: EntityReadResult<ExaminationRecord> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: examinationData,
    isLoading: isExaminationLoading,
    isError: isExaminationError,
    error: examinationError,
    refetch: refetchExamination,
  });
  const existingExam =
    entityRead.status === "found" ? entityRead.data : undefined;
  const entityReadRef = useRef(entityRead);
  useLayoutEffect(() => {
    entityReadRef.current = entityRead;
  }, [entityRead]);
  const mutationPetId = isEdit ? (existingExam?.petId ?? "") : (petId ?? "");
  const { data: mutationPet, isLoading: isPetLoading } =
    useGetPet(mutationPetId);
  const createMutation = useCreateExamination();
  const updateMutation = useUpdateExamination();
  const deleteMutation = useDeleteExamination();
  const unconfirmMutation = useUnconfirmExamination();
  const { data: existingItems, isSuccess: existingItemsQuerySucceeded } =
    useGetExaminationItems(id ?? "");
  const { canCreate, canEdit, canDelete, canUnconfirm } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete, canUnconfirm };
  }, [canCreate, canDelete, canEdit, canUnconfirm]);
  const hasExplicitlyDeceasedPet =
    mutationPet?.status === "死亡" || selectedPets[0]?.status === "死亡";
  const hasExplicitlyDeceasedPetRef = useRef(hasExplicitlyDeceasedPet);
  useLayoutEffect(() => {
    hasExplicitlyDeceasedPetRef.current = hasExplicitlyDeceasedPet;
  }, [hasExplicitlyDeceasedPet]);
  const isMutationAllowed = useCallback(
    (action: keyof ExaminationMutationPermissions) =>
      permissionsRef.current[action] === true,
    [],
  );
  const isPetExplicitlyDeceased = useCallback(
    () => hasExplicitlyDeceasedPetRef.current === true,
    [],
  );

  // サーバ保存済み確定フラグ（useActionState 内の stale closure を避ける）
  const isPersistedConfirmedRef = useRef(false);
  useLayoutEffect(() => {
    isPersistedConfirmedRef.current = isEdit && existingExam?.status === "確定";
  }, [isEdit, existingExam?.status]);

  const isPatientChangeLocked =
    !isEdit ||
    !canEdit ||
    !existingExam ||
    existingExam.status === "確定" ||
    existingExam.currentRevisionVersion !== undefined;
  const isPatientChangeLockedRef = useRef(isPatientChangeLocked);
  const existingPetIdRef = useRef(existingExam?.petId);
  useLayoutEffect(() => {
    isPatientChangeLockedRef.current = isPatientChangeLocked;
    existingPetIdRef.current = existingExam?.petId;
  }, [existingExam?.petId, isPatientChangeLocked]);

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // Local overrides applied on top of server data (only tracks user edits in edit mode)
  // useActionState の前に宣言: callback 内で formDataWithPet を参照するため
  const [localOverrideScope, setLocalOverrideScope] = useState<{
    examinationID: string | undefined;
    values: Partial<ExaminationRecord>;
  }>({ examinationID: id, values: {} });
  const localOverrides =
    localOverrideScope.examinationID === id ? localOverrideScope.values : {};

  // Direct hook consumers can change id without remounting. Scope overrides to
  // the active record immediately, then discard the previous record's values.
  /* eslint-disable react-hooks/set-state-in-effect -- defensive reset for non-route hook consumers */
  useEffect(() => {
    if (localOverrideScope.examinationID === id) return;
    setLocalOverrideScope((previous) =>
      previous.examinationID === id
        ? previous
        : { examinationID: id, values: {} },
    );
  }, [id, localOverrideScope.examinationID]);
  /* eslint-enable react-hooks/set-state-in-effect */

  // Merge: server data as base + user edits on top
  const formData: Partial<ExaminationRecord> =
    isEdit && existingExam
      ? { ...existingExam, ...localOverrides }
      : {
          status: "依頼中" as const,
          ownerName: "",
          petName: "",
          ...(doctorId && { doctorId }),
          ...localOverrides,
        };

  // BUG-017: display field errors with field-local clear (owner form pattern)
  const [manualFieldErrors, setManualFieldErrors] = useState<
    Record<string, string>
  >({});

  const setFormData = useCallback(
    (next: Partial<ExaminationRecord>) => {
      setLocalOverrideScope((previous) => ({
        examinationID: id,
        values:
          previous.examinationID === id
            ? { ...previous.values, ...next }
            : { ...next },
      }));
      // Clear only errors for fields the user just corrected.
      if ("testTypeId" in next || "doctorId" in next) {
        setManualFieldErrors((previous) => {
          let changed = false;
          const updated = { ...previous };
          if ("testTypeId" in next && updated.testTypeId) {
            delete updated.testTypeId;
            changed = true;
          }
          if ("doctorId" in next && updated.doctorId) {
            delete updated.doctorId;
            changed = true;
          }
          return changed ? updated : previous;
        });
      }
    },
    [id],
  );

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const activePet = selectedPets[0] ?? mutationPet;
  const formDataWithPet = activePet
    ? {
        ...formData,
        ownerName: activePet.ownerName,
        petName: activePet.name,
        petId: activePet.id,
      }
    : formData;

  // useActionState の stale closure 対策: 最新の formDataWithPet を ref で保持
  // (use-medical-record-form.ts の activeTabRef と同じパターン)
  const formDataWithPetRef = useRef(formDataWithPet);
  useLayoutEffect(() => {
    formDataWithPetRef.current = formDataWithPet;
  });
  const activePetRef = useRef(activePet);
  useLayoutEffect(() => {
    activePetRef.current = activePet;
  }, [activePet]);

  // ─────────────────────────────────────────────────
  // 検査項目テーブルの state
  // ─────────────────────────────────────────────────
  const [formItems, setFormItems] = useState<ExamItemRow[]>([]);
  const [formItemsOwnerID, setFormItemsOwnerID] = useState(id);
  const visibleFormItems = useMemo(
    () => (isEdit && formItemsOwnerID !== id ? [] : formItems),
    [formItems, formItemsOwnerID, id, isEdit],
  );
  const formItemsRef = useRef(visibleFormItems);
  useLayoutEffect(() => {
    formItemsRef.current = visibleFormItems;
  }, [visibleFormItems]);

  // 検査種別テンプレ（exam_type_fields）取得 — testTypeId 変更検知に使う
  const currentTestTypeId = formData.testTypeId ?? "";
  const { data: examTypeFields } = useGetExamTypeFields(currentTestTypeId);

  // 編集モード: 既存 items を一度だけ formItems に流し込む
  // 同期目的のため useEffect 内 setState は許容。
  const itemsInitializedForIDRef = useRef<string | undefined>(undefined);
  const itemsReadyForIDRef = useRef<string | undefined>(undefined);
  const formItemsExamIDRef = useRef(id);
  const emptyItemsAwaitingTemplateForIDRef = useRef<string | undefined>(
    undefined,
  );
  useLayoutEffect(() => {
    if (
      isEdit &&
      !existingItemsQuerySucceeded &&
      itemsReadyForIDRef.current === id
    ) {
      itemsReadyForIDRef.current = undefined;
    }
  }, [existingItemsQuerySucceeded, id, isEdit]);
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!isEdit || formItemsExamIDRef.current === id) return;
    formItemsExamIDRef.current = id;
    itemsInitializedForIDRef.current = undefined;
    itemsReadyForIDRef.current = undefined;
    emptyItemsAwaitingTemplateForIDRef.current = undefined;
    formItemsRef.current = [];
    setFormItemsOwnerID(id);
    setFormItems([]);
  }, [id, isEdit]);

  useEffect(() => {
    if (!isEdit) return;
    if (!existingItemsQuerySucceeded) return;
    if (itemsInitializedForIDRef.current === id) {
      itemsReadyForIDRef.current = id;
      return;
    }
    if (!existingItems) return;
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
        isAssessed: it.isAssessed,
        sortOrder: it.sortOrder,
        status: it.status,
        isAbnormal: it.isAbnormal,
      }));
      formItemsRef.current = rows;
      setFormItemsOwnerID(id);
      setFormItems(rows);
    } else {
      // 成功した空応答は権威ある初期状態。テンプレが遅れて届いた場合だけ、
      // ユーザーが行を追加していなければ後続 effect で補完する。
      emptyItemsAwaitingTemplateForIDRef.current =
        examTypeFields === undefined ? id : undefined;
      if (
        formItemsRef.current.length === 0 &&
        examTypeFields &&
        examTypeFields.length > 0
      ) {
        const rows = buildRowsFromTemplate(examTypeFields);
        formItemsRef.current = rows;
        setFormItemsOwnerID(id);
        setFormItems(rows);
      }
    }
    formItemsExamIDRef.current = id;
    itemsInitializedForIDRef.current = id;
    itemsReadyForIDRef.current = id;
  }, [id, isEdit, existingItems, existingItemsQuerySucceeded, examTypeFields]);

  useEffect(() => {
    if (emptyItemsAwaitingTemplateForIDRef.current !== id) return;
    if (examTypeFields === undefined) return;
    emptyItemsAwaitingTemplateForIDRef.current = undefined;
    if (formItemsRef.current.length === 0 && examTypeFields.length > 0) {
      const rows = buildRowsFromTemplate(examTypeFields);
      formItemsRef.current = rows;
      setFormItemsOwnerID(id);
      setFormItems(rows);
    }
  }, [examTypeFields, id]);

  // 検査種別変更検知 — ユーザーが testTypeId を変えたらテンプレで再構築
  // 初回レンダー（編集モードで existingExam が後から到着するケース）は既存 items の流入を尊重するためスキップ
  const previousTestTypeRef = useRef<{
    examinationID: string | undefined;
    testTypeID: string | undefined;
  }>({ examinationID: id, testTypeID: undefined });
  useEffect(() => {
    const next = currentTestTypeId;
    if (previousTestTypeRef.current.examinationID !== id) {
      previousTestTypeRef.current = {
        examinationID: id,
        testTypeID: next || undefined,
      };
      return;
    }
    if (!next) return;
    if (previousTestTypeRef.current.testTypeID === undefined) {
      previousTestTypeRef.current = { examinationID: id, testTypeID: next };
      return;
    }
    if (previousTestTypeRef.current.testTypeID === next) return;
    previousTestTypeRef.current = { examinationID: id, testTypeID: next };
    // 種別が変わった → テンプレで再構築（テンプレ未到着なら空行）
    setFormItemsOwnerID(id);
    if (examTypeFields) {
      setFormItems(buildRowsFromTemplate(examTypeFields));
    } else {
      setFormItems([]);
    }
  }, [currentTestTypeId, examTypeFields, id]);

  // 新規モード: testTypeId 選択 & テンプレ到着で初期化（一度だけ）
  const newModeInitializedRef = useRef(false);
  useEffect(() => {
    if (isEdit) return;
    if (newModeInitializedRef.current) return;
    if (!currentTestTypeId) return;
    if (!examTypeFields) return;
    setFormItemsOwnerID(id);
    setFormItems(buildRowsFromTemplate(examTypeFields));
    newModeInitializedRef.current = true;
  }, [isEdit, currentTestTypeId, examTypeFields, id]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const setInspectionValue = useCallback((key: string, value: string) => {
    setFormItems((prev) =>
      prev.map((row) =>
        row.key === key ? { ...row, inspectionValue: value } : row,
      ),
    );
  }, []);

  const manualItemSequenceRef = useRef(0);
  const addManualItem = useCallback(() => {
    manualItemSequenceRef.current += 1;
    const key = `manual-${manualItemSequenceRef.current}`;
    setFormItems((previous) => {
      const nextSortOrder =
        previous.reduce(
          (maximum, row) => Math.max(maximum, row.sortOrder),
          -1,
        ) + 1;
      return [
        ...previous,
        {
          key,
          name: "",
          inspectionValue: "",
          unit: "",
          normalValue: "",
          referenceValue: "",
          sortOrder: nextSortOrder,
        },
      ];
    });
  }, []);

  const removeItem = useCallback((key: string) => {
    setFormItems((previous) => previous.filter((row) => row.key !== key));
  }, []);

  const setItemName = useCallback((key: string, value: string) => {
    setFormItems((previous) =>
      previous.map((row) => (row.key === key ? { ...row, name: value } : row)),
    );
  }, []);

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (
      _prevState: ActionState,
      _formData: FormData,
    ): Promise<ActionState> => {
      const current = formDataWithPetRef.current;
      const isPersistedConfirmed = isPersistedConfirmedRef.current;
      // フロントエンド・バリデーション
      const errors: Record<string, string> = {};
      if (!current.testTypeId) errors.testTypeId = "検査種別を選択してください";
      if (!current.doctorId) errors.doctorId = "担当医を選択してください";
      const isCurrentEditTarget = activeExaminationIDRef.current === id;
      const areCurrentItemsReady = itemsReadyForIDRef.current === id;
      if (
        isEdit &&
        (!isCurrentEditTarget ||
          (!isPersistedConfirmed && !areCurrentItemsReady))
      ) {
        errors.examItems = "検査項目の読み込み完了後に保存してください";
      }
      if (
        formItemsRef.current.some(
          (item) =>
            item.name.trim() === "" && item.inspectionValue.trim() !== "",
        )
      ) {
        errors.examItems =
          "結果値を入力した手動項目には項目名が必要です";
      }

      if (Object.keys(errors).length > 0) {
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }

      try {
        // サーバ保存済みの確定のみ items を省略する（BE は confirmed への item 更新を 409）。
        // ドラフトでステータス「確定」を選んだだけの遷移保存では items を送る（A-S02-01）。
        const items = rowsToRequest(formItemsRef.current);

        if (isPetExplicitlyDeceased()) {
          return { success: false, timestamp: Date.now() };
        }

        if (isEdit && id) {
          // BUG-016: edit route without a found entity must not create/update
          if (entityReadRef.current.status !== "found") {
            return { success: false, timestamp: Date.now() };
          }
          const patientChanged =
            current.petId !== undefined &&
            current.petId !== existingPetIdRef.current;
          const changedPetID = Number(current.petId);
          const changedPatient = activePetRef.current;
          const canApplyPatientChange =
            patientChanged &&
            !isPatientChangeLockedRef.current &&
            changedPatient?.id === current.petId &&
            changedPatient.status === "生存" &&
            Number.isSafeInteger(changedPetID) &&
            changedPetID > 0;
          if (patientChanged && !canApplyPatientChange) {
            toast.error(
              "患者変更の条件が変わりました。検査記録を再読み込みしてください",
            );
            return { success: false, timestamp: Date.now() };
          }
          const req: UpdateExaminationRequest = {
            status: current.status
              ? EXAM_STATUS_JA_TO_EN[current.status]
              : undefined,
            result_summary: current.resultSummary,
            machine: current.machine,
            date: current.date
              ? current.date.includes("T")
                ? current.date
                : jstDateStartISOString(current.date)
              : undefined,
            ...(!isPersistedConfirmed ? { items } : {}),
            ...(canApplyPatientChange ? { pet_id: changedPetID } : {}),
          };
          if (!isMutationAllowed("canEdit")) {
            return { success: false, timestamp: Date.now() };
          }
          await updateMutation.mutateAsync({ id, req });
        } else {
          const pet = activePetRef.current;
          if (!pet) return { success: false, timestamp: Date.now() };
          const req: CreateExaminationRequest = {
            medical_record_id: medicalRecordId ? Number(medicalRecordId) : null,
            pet_id: Number(pet.id) || null,
            exam_type_id: Number(current.testTypeId) || 0,
            doctor_id: current.doctorId ? Number(current.doctorId) : null,
            date: current.date ?? jstDateStartISOString(todayJSTISO()),
            result_summary: current.resultSummary,
            machine: current.machine,
            // 新規は常に items を送る（作成時点で「確定」を選んでもロック対象はサーバ確定後のみ）
            items,
          };
          if (
            !isMutationAllowed("canCreate") ||
            !isMutationAllowed("canEdit")
          ) {
            return { success: false, timestamp: Date.now() };
          }
          await createMutation.mutateAsync(req);
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  // Queryまたは保存済みexamの患者を、create/edit双方の表示・mutation候補へ一度だけ同期する。
  const initializedPetIDRef = useRef<string | null>(null);
  useEffect(() => {
    if (mutationPet && initializedPetIDRef.current !== mutationPet.id) {
      initializedPetIDRef.current = mutationPet.id;
      setSelectedPets([mutationPet]);
      return;
    }
    if (!isEdit && !petId && !isPetLoading) {
      navigate(paths.examinations.selectPet.getHref());
    }
  }, [isEdit, petId, mutationPet, isPetLoading, setSelectedPets, navigate]);

  const handleUnconfirm = useCallback(
    async (rawReason: string): Promise<boolean> => {
      const reason = rawReason.trim();
      if (!isEdit || !id) return false;
      if (!isPersistedConfirmedRef.current) return false;
      if (!isMutationAllowed("canUnconfirm")) return false;
      if (!reason || reason.length > 500) return false;

      try {
        await unconfirmMutation.mutateAsync({ id, reason });
        toast.success("検査記録の確定を解除しました");
        return true;
      } catch {
        return false;
      }
    },
    [id, isEdit, isMutationAllowed, unconfirmMutation],
  );

  const handleDelete = useCallback(
    (onSuccess?: () => void) => {
      if (!isEdit || !id) return;
      if (!isMutationAllowed("canDelete")) return;
      if (isPetExplicitlyDeceased()) return;
      startDeleteTransition(() => {
        deleteMutation.mutate(id, {
          onSuccess: () => {
            toast.success("検査記録を削除しました");
            onSuccess?.();
          },
        });
      });
    },
    [
      isEdit,
      id,
      isMutationAllowed,
      isPetExplicitlyDeceased,
      deleteMutation,
      startDeleteTransition,
    ],
  );

  const isSaving = isPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;
  const isUnconfirming = unconfirmMutation.isPending;
  // UI ロックはサーバ保存済み確定のみ（ドラフトで「確定」を選んだだけではロックしない — A-S02-01）
  const isPersistedConfirmed = isEdit && existingExam?.status === "確定";

  // Sync ActionState field errors into display map so field-local clear can omit keys.
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- ActionState errors → form field display
    setManualFieldErrors(formState.fieldErrors || {});
  }, [formState.fieldErrors, formState.timestamp]);

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    formAction,
    formState,
    fieldErrors: manualFieldErrors,
    handleDelete,
    isEdit,
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
    isSaving,
    isDeleting,
    handleUnconfirm,
    isUnconfirming,
    isPersistedConfirmed,
    isPatientChangeLocked,
    // 検査項目テーブル
    formItems: visibleFormItems,
    setInspectionValue,
    addManualItem,
    removeItem,
    setItemName,
  };
}
