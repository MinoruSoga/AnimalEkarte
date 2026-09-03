import {
  useState,
  useEffect,
  useLayoutEffect,
  useCallback,
  useRef,
  useMemo,
} from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { paths } from "@/config/paths";
import { useGetExamination } from "../api/get-examination";
import { useGetExaminationItems } from "../api/get-examination-items";
import { useGetPet } from "@/hooks/use-pet";
import {
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import type { Pet } from "@/types";
import type { ExaminationRecord } from "../api/transforms";
import type { ExamResult } from "../api/transforms";
import type { CreateExaminationRequest, UpdateExaminationRequest } from "../api/types";
import type { ActionState } from "@/types/form";
import type { ExamItemRow } from "../components/ExamItemsTable";
import { useGetExamTypeFields } from "../api/get-exam-type-fields";
import {
  buildCreateExaminationRequest,
  buildRowsFromTemplate,
  buildUpdateExaminationRequest,
  decideExaminationPatientChange,
  mapExamResultsToFormRows,
  omitCorrectedExaminationFieldErrors,
  rowsToRequest,
  validateExaminationSave,
  type ExaminationMutationPermissions,
} from "./use-examination-form-model";
import { UNCONFIRM_REASON_MAX_LENGTH } from "../constants";

export function useExaminationFormOverrides(id: string | undefined) {
  const [localOverrideScope, setLocalOverrideScope] = useState<{
    examinationID: string | undefined;
    values: Partial<ExaminationRecord>;
  }>({ examinationID: id, values: {} });
  const localOverrides =
    localOverrideScope.examinationID === id ? localOverrideScope.values : {};

  const [manualFieldErrors, setManualFieldErrors] = useState<
    Record<string, string>
  >({});

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
      setManualFieldErrors((previous) =>
        omitCorrectedExaminationFieldErrors(previous, next),
      );
    },
    [id],
  );

  return {
    localOverrides,
    setFormData,
    manualFieldErrors,
    setManualFieldErrors,
  };
}

function useExaminationFormItemEffects(input: {
  id: string | undefined;
  isEdit: boolean;
  existingItems: ExamResult[] | undefined;
  existingItemsQuerySucceeded: boolean;
  currentTestTypeId: string;
  formItemsRef: { current: ExamItemRow[] };
  itemsReadyForIDRef: { current: string | undefined };
  setFormItems: (rows: ExamItemRow[]) => void;
  setFormItemsOwnerID: (id: string | undefined) => void;
}) {
  const {
    id,
    isEdit,
    existingItems,
    existingItemsQuerySucceeded,
    currentTestTypeId,
    formItemsRef,
    itemsReadyForIDRef,
    setFormItems,
    setFormItemsOwnerID,
  } = input;
  const { data: examTypeFields } = useGetExamTypeFields(currentTestTypeId);
  const itemsInitializedForIDRef = useRef<string | undefined>(undefined);
  const formItemsExamIDRef = useRef(id);
  const emptyItemsAwaitingTemplateForIDRef = useRef<string | undefined>(undefined);

  useLayoutEffect(() => {
    if (
      isEdit &&
      !existingItemsQuerySucceeded &&
      itemsReadyForIDRef.current === id
    ) {
      itemsReadyForIDRef.current = undefined;
    }
  }, [existingItemsQuerySucceeded, id, isEdit, itemsReadyForIDRef]);

  useEffect(() => {
    if (!isEdit || formItemsExamIDRef.current === id) return;
    formItemsExamIDRef.current = id;
    itemsInitializedForIDRef.current = undefined;
    itemsReadyForIDRef.current = undefined;
    emptyItemsAwaitingTemplateForIDRef.current = undefined;
    formItemsRef.current = [];
    setFormItemsOwnerID(id);
    setFormItems([]);
  }, [id, isEdit, formItemsRef, itemsReadyForIDRef, setFormItems, setFormItemsOwnerID]);

  useEffect(() => {
    if (!isEdit) return;
    if (!existingItemsQuerySucceeded) return;
    if (itemsInitializedForIDRef.current === id) {
      itemsReadyForIDRef.current = id;
      return;
    }
    if (!existingItems) return;
    if (existingItems.length > 0) {
      const rows = mapExamResultsToFormRows(existingItems);
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
  }, [id, isEdit, existingItems, existingItemsQuerySucceeded, examTypeFields, formItemsRef, itemsReadyForIDRef, setFormItems, setFormItemsOwnerID]);

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
  }, [examTypeFields, id, formItemsRef, setFormItems, setFormItemsOwnerID]);

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
    setFormItemsOwnerID(id);
    if (examTypeFields) {
      setFormItems(buildRowsFromTemplate(examTypeFields));
    } else {
      setFormItems([]);
    }
  }, [currentTestTypeId, examTypeFields, id, setFormItems, setFormItemsOwnerID]);

  const newModeInitializedRef = useRef(false);
  useEffect(() => {
    if (isEdit) return;
    if (newModeInitializedRef.current) return;
    if (!currentTestTypeId) return;
    if (!examTypeFields) return;
    setFormItemsOwnerID(id);
    setFormItems(buildRowsFromTemplate(examTypeFields));
    newModeInitializedRef.current = true;
  }, [isEdit, currentTestTypeId, examTypeFields, id, setFormItems, setFormItemsOwnerID]);
}

export function useExaminationFormItems(input: {
  id: string | undefined;
  isEdit: boolean;
  existingItems: ExamResult[] | undefined;
  existingItemsQuerySucceeded: boolean;
  currentTestTypeId: string;
}) {
  const { id, isEdit, existingItems, existingItemsQuerySucceeded, currentTestTypeId } =
    input;
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

  const itemsReadyForIDRef = useRef<string | undefined>(undefined);
  useExaminationFormItemEffects({
    id,
    isEdit,
    existingItems,
    existingItemsQuerySucceeded,
    currentTestTypeId,
    formItemsRef,
    itemsReadyForIDRef,
    setFormItems,
    setFormItemsOwnerID,
  });

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

  return {
    visibleFormItems,
    formItemsRef,
    itemsReadyForIDRef,
    setInspectionValue,
    addManualItem,
    removeItem,
    setItemName,
  };
}

export function useExaminationFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  mutationPet: Pet | undefined;
  isPetLoading: boolean;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const { isEdit, petId, mutationPet, isPetLoading, setSelectedPets } = input;
  const navigate = useNavigate();
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
}

interface ExaminationSaveDeps {
  id: string | undefined;
  isEdit: boolean;
  medicalRecordId: string;
  formDataWithPetRef: { current: Partial<ExaminationRecord> };
  formItemsRef: { current: readonly ExamItemRow[] };
  itemsReadyForIDRef: { current: string | undefined };
  activeExaminationIDRef: { current: string | undefined };
  isPersistedConfirmedRef: { current: boolean };
  isPersistedResultsLockedRef: { current: boolean };
  isPatientChangeLockedRef: { current: boolean };
  existingPetIdRef: { current: string | undefined };
  entityReadRef: { current: EntityReadResult<ExaminationRecord> };
  activePetRef: { current: Pet | undefined };
  isPetExplicitlyDeceased: () => boolean;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  updateMutation: {
    mutateAsync: (vars: { id: string; req: UpdateExaminationRequest }) => Promise<unknown>;
  };
  createMutation: {
    mutateAsync: (req: CreateExaminationRequest) => Promise<unknown>;
  };
}

export async function runExaminationSave(
  deps: ExaminationSaveDeps,
): Promise<ActionState> {
  const current = deps.formDataWithPetRef.current;
  const isPersistedConfirmed = deps.isPersistedConfirmedRef.current;
  const resultsLocked = deps.isPersistedResultsLockedRef.current;
  const errors = validateExaminationSave({
    current,
    isEdit: deps.isEdit,
    isCurrentEditTarget: deps.activeExaminationIDRef.current === deps.id,
    resultsLocked,
    areCurrentItemsReady: deps.itemsReadyForIDRef.current === deps.id,
    formItems: deps.formItemsRef.current,
  });

  if (Object.keys(errors).length > 0) {
    return { success: false, fieldErrors: errors, timestamp: Date.now() };
  }

  try {
    // サーバ保存済みの結果ロック中は items を省略（BE は confirmed / 初回 completed への item 更新を 409）。
    // ドラフトでステータス「確定」を選んだだけの遷移保存では items を送る（A-S02-01）。
    const items = rowsToRequest(deps.formItemsRef.current);

    if (deps.isPetExplicitlyDeceased()) {
      return { success: false, timestamp: Date.now() };
    }

    if (deps.isEdit && deps.id) {
      // BUG-016: edit route without a found entity must not create/update
      if (deps.entityReadRef.current.status !== "found") {
        return { success: false, timestamp: Date.now() };
      }
      const patientChange = decideExaminationPatientChange({
        currentPetId: current.petId,
        existingPetId: deps.existingPetIdRef.current,
        isPatientChangeLocked: deps.isPatientChangeLockedRef.current,
        changedPatient: deps.activePetRef.current,
      });
      if (patientChange.kind === "blocked") {
        toast.error(
          "患者変更の条件が変わりました。検査記録を再読み込みしてください",
        );
        return { success: false, timestamp: Date.now() };
      }
      const req = buildUpdateExaminationRequest({
        current,
        items,
        resultsLocked,
        patientChange,
      });
      // 確定済みは親フィールド更新も拒否。完了シールは status 遷移保存を許可。
      if (isPersistedConfirmed) {
        return { success: false, timestamp: Date.now() };
      }
      if (!deps.isMutationAllowed("canEdit")) {
        return { success: false, timestamp: Date.now() };
      }
      await deps.updateMutation.mutateAsync({ id: deps.id, req });
    } else {
      const pet = deps.activePetRef.current;
      if (!pet) return { success: false, timestamp: Date.now() };
      const req = buildCreateExaminationRequest({
        current,
        medicalRecordId: deps.medicalRecordId,
        petId: pet.id,
        items,
      });
      if (
        !deps.isMutationAllowed("canCreate") ||
        !deps.isMutationAllowed("canEdit")
      ) {
        return { success: false, timestamp: Date.now() };
      }
      await deps.createMutation.mutateAsync(req);
    }
    return { success: true, timestamp: Date.now() };
  } catch {
    // FE-RC-005: useCreateExamination/useUpdateExamination の onError が
    // handleApiError 済み（api/create-examination.ts, api/update-examination.ts）。
    // ここでは二重 toast を避け、失敗を呼び出し元へ伝えるだけにする。
    return { success: false, timestamp: Date.now() };
  }
}

export function createExaminationUnconfirmHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isPersistedConfirmed: () => boolean;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  unconfirm: (vars: { id: string; reason: string }) => Promise<unknown>;
}): (rawReason: string) => Promise<boolean> {
  return async (rawReason: string): Promise<boolean> => {
    const reason = rawReason.trim();
    if (!input.isEdit || !input.id) return false;
    if (!input.isPersistedConfirmed()) return false;
    if (!input.isMutationAllowed("canUnconfirm")) return false;
    if (!reason || reason.length > UNCONFIRM_REASON_MAX_LENGTH) return false;

    try {
      await input.unconfirm({ id: input.id, reason });
      toast.success("検査記録の確定を解除しました");
      return true;
    } catch {
      // onError は useUnconfirmExamination 側で handleApiError 済み。ここでは失敗を呼び出し元へ伝えるだけ。
      return false;
    }
  };
}

export function createExaminationDeleteHandler(input: {
  isEdit: boolean;
  id: string | undefined;
  isMutationAllowed: (action: keyof ExaminationMutationPermissions) => boolean;
  isResultsLocked: () => boolean;
  isPetExplicitlyDeceased: () => boolean;
  startDeleteTransition: (fn: () => void) => void;
  deleteExamination: (id: string, opts: { onSuccess: () => void }) => void;
}): (onSuccess?: () => void) => void {
  return (onSuccess?: () => void) => {
    if (!input.isEdit || !input.id) return;
    if (!input.isMutationAllowed("canDelete")) return;
    if (input.isResultsLocked()) return;
    if (input.isPetExplicitlyDeceased()) return;
    const examinationId = input.id;
    input.startDeleteTransition(() => {
      input.deleteExamination(examinationId, {
        onSuccess: () => {
          toast.success("検査記録を削除しました");
          onSuccess?.();
        },
      });
    });
  };
}

export function useExaminationFormLoad(id: string | undefined, isEdit: boolean, petId: string | null) {
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
  const mutationPetId = isEdit ? (existingExam?.petId ?? "") : (petId ?? "");
  const { data: mutationPet, isLoading: isPetLoading } = useGetPet(mutationPetId);
  const { data: existingItems, isSuccess: existingItemsQuerySucceeded } =
    useGetExaminationItems(id ?? "");
  return {
    entityRead,
    existingExam,
    mutationPet,
    isPetLoading,
    existingItems,
    existingItemsQuerySucceeded,
  };
}
